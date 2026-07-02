package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/pkg/saga"
	"github.com/iho/neobank/pkg/userclient"
	"github.com/iho/neobank/services/card/internal/adapter/processor"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/jackc/pgx/v5"
)

type CardRepository interface {
	Create(ctx context.Context, c domain.Card) error
	GetByID(ctx context.Context, id string) (*domain.Card, error)
	GetByUserAndIdempotencyKey(ctx context.Context, userID, key string) (*domain.Card, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Card, error)
	UpdateStatus(ctx context.Context, id, userID string, status domain.CardStatus) error
	MarkCancelled(ctx context.Context, id string) error
}

type OutboxPublisher interface {
	Publish(ctx context.Context, evt events.Event) error
}

type IssueCardInput struct {
	UserID         string
	WalletID       string
	CardholderName string
	IdempotencyKey string
}

type IssueCardUseCase struct {
	cards     CardRepository
	users     *userclient.Client
	processor processor.Processor
	fraud     *fraud.Checker
	outbox    OutboxPublisher
	sagaStore saga.InstanceStore
}

func NewIssueCardUseCase(
	cards CardRepository,
	users *userclient.Client,
	proc processor.Processor,
	fraudChecker *fraud.Checker,
	outbox OutboxPublisher,
	sagaStore saga.InstanceStore,
) *IssueCardUseCase {
	return &IssueCardUseCase{
		cards:     cards,
		users:     users,
		processor: proc,
		fraud:     fraudChecker,
		outbox:    outbox,
		sagaStore: sagaStore,
	}
}

type IssueCardOutput struct {
	Card     *domain.Card
	Replayed bool
}

func (uc *IssueCardUseCase) Execute(ctx context.Context, in IssueCardInput) (IssueCardOutput, error) {
	if in.UserID == "" || in.IdempotencyKey == "" || in.CardholderName == "" {
		return IssueCardOutput{}, fmt.Errorf("user_id, cardholder_name, and idempotency_key are required")
	}

	existing, err := uc.cards.GetByUserAndIdempotencyKey(ctx, in.UserID, in.IdempotencyKey)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return IssueCardOutput{}, err
	}
	if existing != nil {
		return IssueCardOutput{Card: existing, Replayed: true}, nil
	}

	walletID := in.WalletID
	if walletID == "" {
		wallet, err := uc.users.GetWallet(ctx, in.UserID, "USD")
		if err != nil {
			return IssueCardOutput{}, fmt.Errorf("wallet: %w", err)
		}
		walletID = wallet.ID
	}

	cardID := uuid.NewString()
	state := saga.NewState(map[string]string{
		"card_id":         cardID,
		"user_id":         in.UserID,
		"wallet_id":       walletID,
		"cardholder_name": in.CardholderName,
		"idempotency_key": in.IdempotencyKey,
	})

	orchestrator := saga.New("card_issuance", uc.steps(), uc.sagaStore)
	if err := orchestrator.Run(ctx, in.IdempotencyKey, state); err != nil {
		if biz, ok := err.(*saga.BusinessError); ok {
			return IssueCardOutput{}, biz
		}
		return IssueCardOutput{}, err
	}

	_ = uc.outbox.Publish(ctx, events.CardIssued{
		CardID:      cardID,
		UserID:      in.UserID,
		WalletID:    walletID,
		LastFour:    state.Get("last_four"),
		ExpiryMonth: parseInt(state.Get("expiry_month")),
		ExpiryYear:  parseInt(state.Get("expiry_year")),
	})

	card, err := uc.cards.GetByID(ctx, cardID)
	if err != nil {
		return IssueCardOutput{}, err
	}
	return IssueCardOutput{Card: card}, nil
}

func (uc *IssueCardUseCase) GetByID(ctx context.Context, id string) (*domain.Card, error) {
	return uc.cards.GetByID(ctx, id)
}

func (uc *IssueCardUseCase) List(ctx context.Context, userID string) ([]domain.Card, error) {
	return uc.cards.ListByUser(ctx, userID)
}

func (uc *IssueCardUseCase) steps() []saga.Step {
	return []saga.Step{
		{
			Name: "fraud_check",
			Execute: func(ctx context.Context, state *saga.State) error {
				result, err := uc.fraud.Evaluate(
					state.Get("user_id"), "card_issue", "0", "USD",
					fraud.EvaluateOpts{},
				)
				if err != nil {
					return err
				}
				if result.Decision == fraud.DecisionDeny {
					return saga.NewBusinessError("fraud_denied", result.ReasonCode)
				}
				return nil
			},
		},
		{
			Name: "processor_create_card",
			Execute: func(ctx context.Context, state *saga.State) error {
				card, err := uc.processor.CreateVirtualCard(ctx, state.Get("user_id"), state.Get("cardholder_name"))
				if err != nil {
					return err
				}
				state.Set("processor_ref", card.Ref)
				state.Set("pan_token", card.PANToken)
				state.Set("last_four", card.LastFour)
				state.Set("expiry_month", fmt.Sprintf("%d", card.ExpiryMonth))
				state.Set("expiry_year", fmt.Sprintf("%d", card.ExpiryYear))
				return nil
			},
			Compensate: func(ctx context.Context, state *saga.State) error {
				if state.Has("processor_ref") {
					return uc.processor.CancelCard(ctx, state.Get("processor_ref"))
				}
				return nil
			},
		},
		{
			Name: "persist_card",
			Execute: func(ctx context.Context, state *saga.State) error {
				return uc.cards.Create(ctx, domain.Card{
					ID:             state.Get("card_id"),
					UserID:         state.Get("user_id"),
					WalletID:       state.Get("wallet_id"),
					ProcessorRef:   state.Get("processor_ref"),
					PANToken:       state.Get("pan_token"),
					LastFour:       state.Get("last_four"),
					ExpiryMonth:    parseInt(state.Get("expiry_month")),
					ExpiryYear:     parseInt(state.Get("expiry_year")),
					Status:         domain.CardStatusActive,
					IdempotencyKey: state.Get("idempotency_key"),
				})
			},
			Compensate: func(ctx context.Context, state *saga.State) error {
				if state.Has("card_id") {
					return uc.cards.MarkCancelled(ctx, state.Get("card_id"))
				}
				return nil
			},
		},
	}
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}