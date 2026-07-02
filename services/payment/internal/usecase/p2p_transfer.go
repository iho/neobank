package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/money"
	"github.com/iho/neobank/pkg/saga"
	"github.com/iho/neobank/pkg/userclient"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransferRepository interface {
	Create(ctx context.Context, t domain.Transfer) error
	GetBySenderAndIdempotencyKey(ctx context.Context, senderUserID, key string) (*domain.Transfer, error)
	GetByID(ctx context.Context, id string) (*domain.Transfer, error)
	MarkCompleted(ctx context.Context, id, ledgerTransferID string) error
	MarkFailed(ctx context.Context, id, reason string) error
	ListByUser(ctx context.Context, userID string, limit int) ([]domain.Transfer, error)
}

type OutboxPublisher interface {
	Publish(ctx context.Context, evt events.Event) error
}

type P2PTransferInput struct {
	SenderUserID   string
	RecipientPhone string
	Amount         string
	Currency       string
	Memo           string
	IdempotencyKey string
}

type P2PTransferUseCase struct {
	transfers TransferRepository
	users     *userclient.Client
	ledger    *ledgerclient.Client
	fraud     *fraud.Checker
	outbox    OutboxPublisher
	sagaStore saga.InstanceStore
}

func NewP2PTransferUseCase(
	transfers TransferRepository,
	users *userclient.Client,
	ledger *ledgerclient.Client,
	fraudChecker *fraud.Checker,
	outbox OutboxPublisher,
	sagaStore saga.InstanceStore,
) *P2PTransferUseCase {
	return &P2PTransferUseCase{
		transfers: transfers,
		users:     users,
		ledger:    ledger,
		fraud:     fraudChecker,
		outbox:    outbox,
		sagaStore: sagaStore,
	}
}

func (uc *P2PTransferUseCase) Execute(ctx context.Context, in P2PTransferInput) (*domain.Transfer, error) {
	if in.SenderUserID == "" || in.RecipientPhone == "" || in.IdempotencyKey == "" {
		return nil, fmt.Errorf("sender_user_id, recipient_phone, and idempotency_key are required")
	}
	if in.Currency == "" {
		in.Currency = "USD"
	}
	if _, err := money.Parse(in.Amount); err != nil {
		return nil, err
	}

	existing, err := uc.transfers.GetBySenderAndIdempotencyKey(ctx, in.SenderUserID, in.IdempotencyKey)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	senderWallet, err := uc.users.GetWallet(ctx, in.SenderUserID, in.Currency)
	if err != nil {
		return nil, fmt.Errorf("sender wallet: %w", err)
	}

	recipient, err := uc.users.GetByPhone(ctx, in.RecipientPhone)
	if err != nil {
		return nil, fmt.Errorf("recipient: %w", err)
	}
	if recipient.ID == in.SenderUserID {
		return nil, saga.NewBusinessError("self_transfer", "cannot transfer to yourself")
	}
	recipientWallet, err := uc.users.GetWallet(ctx, recipient.ID, in.Currency)
	if err != nil {
		return nil, fmt.Errorf("recipient wallet: %w", err)
	}

	transferID := uuid.NewString()
	transfer := domain.Transfer{
		ID:              transferID,
		IdempotencyKey:  in.IdempotencyKey,
		Type:            "p2p",
		Status:          domain.TransferStatusPending,
		SenderUserID:    in.SenderUserID,
		RecipientUserID: recipient.ID,
		Amount:          in.Amount,
		Currency:        in.Currency,
		Memo:            in.Memo,
	}
	if err := uc.transfers.Create(ctx, transfer); err != nil {
		return nil, err
	}

	state := saga.NewState(map[string]string{
		"transfer_id":                transferID,
		"sender_user_id":             in.SenderUserID,
		"sender_ledger_account_id":   senderWallet.LedgerAccountID,
		"recipient_user_id":          recipient.ID,
		"recipient_ledger_account_id": recipientWallet.LedgerAccountID,
		"amount":                     in.Amount,
		"currency":                   in.Currency,
		"idempotency_key":            in.IdempotencyKey,
	})

	orchestrator := saga.New("p2p_transfer", uc.steps(), uc.sagaStore)
	if err := orchestrator.Run(ctx, in.IdempotencyKey, state); err != nil {
		reason := err.Error()
		if biz, ok := err.(*saga.BusinessError); ok {
			reason = biz.Code
		}
		_ = uc.transfers.MarkFailed(ctx, transferID, reason)
		t, getErr := uc.transfers.GetByID(ctx, transferID)
		if getErr != nil {
			return nil, getErr
		}
		return t, nil
	}

	if err := uc.transfers.MarkCompleted(ctx, transferID, state.Get("ledger_transfer_id")); err != nil {
		return nil, err
	}
	_ = uc.outbox.Publish(ctx, events.TransferCompleted{
		TransferID:       transferID,
		LedgerTransferID: state.Get("ledger_transfer_id"),
		SenderUserID:     in.SenderUserID,
		RecipientUserID:  state.Get("recipient_user_id"),
		Amount:           in.Amount,
		Currency:         in.Currency,
	})

	return uc.transfers.GetByID(ctx, transferID)
}

func (uc *P2PTransferUseCase) GetByID(ctx context.Context, id string) (*domain.Transfer, error) {
	return uc.transfers.GetByID(ctx, id)
}

func (uc *P2PTransferUseCase) List(ctx context.Context, userID string, limit int) ([]domain.Transfer, error) {
	return uc.transfers.ListByUser(ctx, userID, limit)
}

func (uc *P2PTransferUseCase) steps() []saga.Step {
	return []saga.Step{
		{
			Name: "fraud_check",
			Execute: func(ctx context.Context, state *saga.State) error {
				result, err := uc.fraud.Evaluate(
					state.Get("sender_user_id"), "p2p", state.Get("amount"), state.Get("currency"),
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
			Name: "ledger_transfer",
			Execute: func(ctx context.Context, state *saga.State) error {
				if uc.ledger == nil {
					return fmt.Errorf("ledger unavailable")
				}
				ledgerTransfer, err := uc.ledger.CreateTransfer(ctx, ledgerclient.CreateTransferInput{
					FromAccountID:  state.Get("sender_ledger_account_id"),
					ToAccountID:    state.Get("recipient_ledger_account_id"),
					Amount:         state.Get("amount"),
					IdempotencyKey: state.Get("idempotency_key"),
					Metadata: map[string]string{
						"transfer_id": state.Get("transfer_id"),
						"type":        "p2p",
					},
				})
				if err != nil {
					if st, ok := status.FromError(err); ok {
						if st.Code() == codes.FailedPrecondition || strings.Contains(strings.ToLower(st.Message()), "insufficient") {
							return saga.NewBusinessError("insufficient_funds", st.Message())
						}
					}
					return err
				}
				state.Set("ledger_transfer_id", ledgerTransfer.Id)
				return nil
			},
			Compensate: func(ctx context.Context, state *saga.State) error {
				if !state.Has("ledger_transfer_id") || uc.ledger == nil {
					return nil
				}
				_, err := uc.ledger.ReverseTransfer(ctx, state.Get("ledger_transfer_id"), map[string]string{
					"reason": "saga_compensation",
				})
				return err
			},
		},
	}
}