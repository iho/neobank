package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/saga"
	"github.com/iho/neobank/pkg/userclient"
	"github.com/iho/neobank/services/card/internal/adapter/processor"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/port"
	"github.com/jackc/pgx/v5"
)

type IssueCardInput struct {
	UserID         string
	WalletID       string
	CardholderName string
	DailyLimit     string
	OnlineOnly     bool
	IdempotencyKey string
}

type IssueCardUseCase struct {
	cards     port.CardRepository
	users     *userclient.Client
	processor processor.Processor
	fraud     *fraud.Checker
	fraudRepo port.FraudDecisionRepository
	outbox    outbox.TxPublisher
	audit     audit.Recorder
	sagaStore saga.InstanceStore
	tx        *pgutil.TxRunner
}

func NewIssueCardUseCase(
	cards port.CardRepository,
	users *userclient.Client,
	proc processor.Processor,
	fraudChecker *fraud.Checker,
	fraudRepo port.FraudDecisionRepository,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	sagaStore saga.InstanceStore,
	tx *pgutil.TxRunner,
) *IssueCardUseCase {
	return &IssueCardUseCase{
		cards:     cards,
		users:     users,
		processor: proc,
		fraud:     fraudChecker,
		fraudRepo: fraudRepo,
		outbox:    outboxPublisher,
		audit:     auditRecorder,
		sagaStore: sagaStore,
		tx:        tx,
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
		"daily_limit":     in.DailyLimit,
		"online_only":     fmt.Sprintf("%t", in.OnlineOnly),
		"idempotency_key": in.IdempotencyKey,
	})

	orchestrator := saga.New("card_issuance", uc.steps(), uc.sagaStore)
	if err := orchestrator.Run(ctx, in.IdempotencyKey, state); err != nil {
		if biz, ok := err.(*saga.BusinessError); ok {
			return IssueCardOutput{}, biz
		}
		return IssueCardOutput{}, err
	}

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
				if uc.fraudRepo != nil {
					if recErr := uc.fraudRepo.Record(ctx, "card", state.Get("card_id"), state.Get("user_id"),
						"card_issue", "0", "USD", result); recErr != nil {
						return fmt.Errorf("record fraud decision: %w", recErr)
					}
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
				return uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
					if err := uc.cards.WithTx(tx).Create(ctx, domain.Card{
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
						DailyLimit:     state.Get("daily_limit"),
						OnlineOnly:     state.Get("online_only") == "true",
					}); err != nil {
						return err
					}
					return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
						EntityType: "card",
						EntityID:   state.Get("card_id"),
						Action:     "issued",
						ToStatus:   string(domain.CardStatusActive),
						Metadata:   map[string]any{"last_four": state.Get("last_four")},
					})
				})
			},
			Compensate: func(ctx context.Context, state *saga.State) error {
				if state.Has("card_id") {
					return uc.cards.MarkCancelled(ctx, state.Get("card_id"))
				}
				return nil
			},
		},
		{
			Name: "publish_event",
			Execute: func(ctx context.Context, state *saga.State) error {
				return uc.outbox.Publish(ctx, events.CardIssued{
					CardID:      state.Get("card_id"),
					UserID:      state.Get("user_id"),
					WalletID:    state.Get("wallet_id"),
					LastFour:    state.Get("last_four"),
					ExpiryMonth: parseInt(state.Get("expiry_month")),
					ExpiryYear:  parseInt(state.Get("expiry_year")),
				})
			},
		},
	}
}

func parseInt(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}
