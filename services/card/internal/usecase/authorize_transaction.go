package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/money"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/saga"
	"github.com/iho/neobank/pkg/userclient"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/port"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthorizeTransactionInput struct {
	UserID         string
	CardID         string
	Amount         string
	Currency       string
	MerchantName   string
	IdempotencyKey string
}

type AuthorizeTransactionOutput struct {
	Authorization *domain.Authorization
	Replayed      bool
}

type AuthorizeTransactionUseCase struct {
	cards     port.CardRepository
	auths     port.AuthorizationRepository
	users     *userclient.Client
	ledger    *ledgerclient.Client
	fraud     *fraud.Checker
	fraudRepo port.FraudDecisionRepository
	outbox    outbox.TxPublisher
	audit     audit.Recorder
	sagaStore saga.InstanceStore
	tx        *pgutil.TxRunner
}

func NewAuthorizeTransactionUseCase(
	cards port.CardRepository,
	auths port.AuthorizationRepository,
	users *userclient.Client,
	ledger *ledgerclient.Client,
	fraudChecker *fraud.Checker,
	fraudRepo port.FraudDecisionRepository,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	sagaStore saga.InstanceStore,
	tx *pgutil.TxRunner,
) *AuthorizeTransactionUseCase {
	return &AuthorizeTransactionUseCase{
		cards:     cards,
		auths:     auths,
		users:     users,
		ledger:    ledger,
		fraud:     fraudChecker,
		fraudRepo: fraudRepo,
		outbox:    outboxPublisher,
		audit:     auditRecorder,
		sagaStore: sagaStore,
		tx:        tx,
	}
}

func (uc *AuthorizeTransactionUseCase) Execute(ctx context.Context, in AuthorizeTransactionInput) (AuthorizeTransactionOutput, error) {
	if in.CardID == "" || in.IdempotencyKey == "" || in.Amount == "" {
		return AuthorizeTransactionOutput{}, fmt.Errorf("card_id, amount, and idempotency_key are required")
	}
	if in.Currency == "" {
		in.Currency = "USD"
	}
	if _, err := money.Parse(in.Amount); err != nil {
		return AuthorizeTransactionOutput{}, err
	}

	existing, err := uc.auths.GetByCardAndIdempotencyKey(ctx, in.CardID, in.IdempotencyKey)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return AuthorizeTransactionOutput{}, err
	}
	if existing != nil {
		return AuthorizeTransactionOutput{Authorization: existing, Replayed: true}, nil
	}

	card, err := uc.cards.GetByID(ctx, in.CardID)
	if err != nil {
		return AuthorizeTransactionOutput{}, err
	}
	if in.UserID != "" && card.UserID != in.UserID {
		return AuthorizeTransactionOutput{}, fmt.Errorf("card not found")
	}
	if card.Status != domain.CardStatusActive {
		return AuthorizeTransactionOutput{}, saga.NewBusinessError("card_not_active", "card is not active")
	}

	wallet, err := uc.users.GetWallet(ctx, card.UserID, in.Currency)
	if err != nil {
		return AuthorizeTransactionOutput{}, fmt.Errorf("wallet: %w", err)
	}

	authID := uuid.NewString()
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.auths.WithTx(tx).Create(ctx, domain.Authorization{
			ID:             authID,
			CardID:         in.CardID,
			UserID:         card.UserID,
			IdempotencyKey: in.IdempotencyKey,
			MerchantName:   in.MerchantName,
			Amount:         in.Amount,
			Currency:       in.Currency,
			Status:         domain.AuthStatusAuthorized,
		}); err != nil {
			return err
		}
		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "authorization",
			EntityID:   authID,
			Action:     "created",
			ToStatus:   string(domain.AuthStatusAuthorized),
			Metadata:   map[string]any{"amount": in.Amount, "currency": in.Currency, "card_id": in.CardID},
		})
	}); err != nil {
		return AuthorizeTransactionOutput{}, err
	}

	state := saga.NewState(map[string]string{
		"authorization_id":  authID,
		"card_id":           in.CardID,
		"user_id":           card.UserID,
		"ledger_account_id": wallet.LedgerAccountID,
		"amount":            in.Amount,
		"currency":          in.Currency,
		"merchant_name":     in.MerchantName,
		"idempotency_key":   in.IdempotencyKey,
	})

	orchestrator := saga.New("card_authorization", uc.steps(), uc.sagaStore)
	if err := orchestrator.Run(ctx, in.IdempotencyKey, state); err != nil {
		reason := err.Error()
		if biz, ok := err.(*saga.BusinessError); ok {
			reason = biz.Code
		}
		if txErr := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
			if err := uc.auths.WithTx(tx).MarkFailed(ctx, authID, reason); err != nil {
				return err
			}
			return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
				EntityType: "authorization",
				EntityID:   authID,
				Action:     "mark_failed",
				FromStatus: string(domain.AuthStatusAuthorized),
				ToStatus:   "declined",
				Metadata:   map[string]any{"reason": reason},
			})
		}); txErr != nil {
			return AuthorizeTransactionOutput{}, txErr
		}
		auth, getErr := uc.auths.GetByID(ctx, authID)
		if getErr != nil {
			return AuthorizeTransactionOutput{}, getErr
		}
		return AuthorizeTransactionOutput{Authorization: auth}, nil
	}

	approvedEvent := events.CardAuthApproved{
		AuthorizationID: authID,
		CardID:          in.CardID,
		UserID:          card.UserID,
		Amount:          in.Amount,
		Currency:        in.Currency,
		MerchantName:    in.MerchantName,
	}
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.auths.WithTx(tx).MarkHold(ctx, authID, state.Get("ledger_hold_id")); err != nil {
			return err
		}
		if err := uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "authorization",
			EntityID:   authID,
			Action:     "mark_hold",
			ToStatus:   string(domain.AuthStatusAuthorized),
			Metadata:   map[string]any{"ledger_hold_id": state.Get("ledger_hold_id")},
		}); err != nil {
			return err
		}
		return uc.outbox.WithTx(tx).Publish(ctx, approvedEvent)
	}); err != nil {
		return AuthorizeTransactionOutput{}, err
	}

	auth, err := uc.auths.GetByID(ctx, authID)
	if err != nil {
		return AuthorizeTransactionOutput{}, err
	}
	return AuthorizeTransactionOutput{Authorization: auth}, nil
}

func (uc *AuthorizeTransactionUseCase) GetByID(ctx context.Context, id string) (*domain.Authorization, error) {
	return uc.auths.GetByID(ctx, id)
}

func (uc *AuthorizeTransactionUseCase) steps() []saga.Step {
	return []saga.Step{
		{
			Name: "fraud_check",
			Execute: func(ctx context.Context, state *saga.State) error {
				result, err := uc.fraud.Evaluate(
					state.Get("user_id"), "card_auth", state.Get("amount"), state.Get("currency"),
					fraud.EvaluateOpts{},
				)
				if err != nil {
					return err
				}
				if uc.fraudRepo != nil {
					if recErr := uc.fraudRepo.Record(ctx, "authorization", state.Get("authorization_id"), state.Get("user_id"),
						"card_auth", state.Get("amount"), state.Get("currency"), result); recErr != nil {
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
			Name: "ledger_hold",
			Execute: func(ctx context.Context, state *saga.State) error {
				if uc.ledger == nil {
					return fmt.Errorf("ledger unavailable")
				}
				hold, err := uc.ledger.HoldFunds(ctx, ledgerclient.HoldFundsInput{
					AccountID: state.Get("ledger_account_id"),
					Amount:    state.Get("amount"),
				})
				if err != nil {
					if st, ok := status.FromError(err); ok {
						if st.Code() == codes.FailedPrecondition || strings.Contains(strings.ToLower(st.Message()), "insufficient") {
							return saga.NewBusinessError("insufficient_funds", st.Message())
						}
					}
					return err
				}
				state.Set("ledger_hold_id", hold.Id)
				return nil
			},
			Compensate: func(ctx context.Context, state *saga.State) error {
				if !state.Has("ledger_hold_id") || uc.ledger == nil {
					return nil
				}
				return uc.ledger.VoidHold(ctx, state.Get("ledger_hold_id"))
			},
		},
	}
}
