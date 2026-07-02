package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/amlmonitor"
	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/money"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pagination"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/reqctx"
	"github.com/iho/neobank/pkg/saga"
	"github.com/iho/neobank/pkg/screening"
	"github.com/iho/neobank/pkg/userclient"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/port"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type P2PTransferInput struct {
	SenderUserID      string
	RecipientPhone    string
	RecipientEmail    string
	RecipientUserID   string
	Amount            string
	Currency          string
	Memo              string
	IdempotencyKey    string
}

type P2PTransferUseCase struct {
	transfers port.TransferRepository
	users     *userclient.Client
	ledger    *ledgerclient.Client
	fraud       *fraud.Checker
	fraudRepo   port.FraudDecisionRepository
	aml         *amlmonitor.Monitor
	amlRepo     port.AMLRepository
	screening   port.ScreeningRepository
	screener    screening.Screener
	outbox      outbox.TxPublisher
	audit     audit.Recorder
	sagaStore saga.InstanceStore
	tx        *pgutil.TxRunner
}

func NewP2PTransferUseCase(
	transfers port.TransferRepository,
	users *userclient.Client,
	ledger *ledgerclient.Client,
	fraudChecker *fraud.Checker,
	fraudRepo port.FraudDecisionRepository,
	amlMonitor *amlmonitor.Monitor,
	amlRepo port.AMLRepository,
	screeningRepo port.ScreeningRepository,
	screener screening.Screener,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	sagaStore saga.InstanceStore,
	tx *pgutil.TxRunner,
) *P2PTransferUseCase {
	return &P2PTransferUseCase{
		transfers: transfers,
		users:     users,
		ledger:    ledger,
		fraud:       fraudChecker,
		fraudRepo:   fraudRepo,
		aml:         amlMonitor,
		amlRepo:     amlRepo,
		screening:   screeningRepo,
		screener:    screener,
		outbox:      outboxPublisher,
		audit:     auditRecorder,
		sagaStore: sagaStore,
		tx:        tx,
	}
}

func (uc *P2PTransferUseCase) Execute(ctx context.Context, in P2PTransferInput) (*domain.Transfer, error) {
	if in.SenderUserID == "" || in.IdempotencyKey == "" {
		return nil, fmt.Errorf("sender_user_id and idempotency_key are required")
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

	recipient, err := uc.resolveRecipient(ctx, in)
	if err != nil {
		return nil, err
	}
	if recipient.ID == in.SenderUserID {
		return nil, saga.NewBusinessError("self_transfer", "cannot transfer to yourself")
	}
	recipientWallet, err := uc.users.GetWallet(ctx, recipient.ID, in.Currency)
	if err != nil {
		return nil, fmt.Errorf("recipient wallet: %w", err)
	}

	transferID := uuid.NewString()
	screenCtx := screening.Context{
		CheckType:     screening.CheckTransferCounterparty,
		EntityType:    "transfer",
		EntityID:      transferID,
		CorrelationID: reqctx.CorrelationID(ctx),
	}
	screenResult, err := uc.screener.ScreenCounterparty(screening.Counterparty{
		UserID: recipient.ID,
		Phone:  recipient.Phone,
	}, screenCtx)
	if err != nil {
		return nil, fmt.Errorf("counterparty screening: %w", err)
	}
	if uc.screening != nil {
		if recErr := uc.screening.Record(ctx, port.ScreeningCheck{
			CheckType:         screenCtx.CheckType,
			SubjectUserID:     recipient.ID,
			RelatedUserID:     in.SenderUserID,
			EntityType:        screenCtx.EntityType,
			EntityID:          screenCtx.EntityID,
			Decision:          screenResult.Decision,
			ReasonCode:        screenResult.ReasonCode,
			Provider:          screenResult.Provider,
			ProviderReference: screenResult.ProviderRef,
			RawResponse:       screenResult.RawResponse,
			CorrelationID:     screenCtx.CorrelationID,
		}); recErr != nil {
			return nil, fmt.Errorf("record screening check: %w", recErr)
		}
	}
	if screenResult.Decision == screening.DecisionBlock {
		return nil, saga.NewBusinessError("screening_blocked", screenResult.ReasonCode)
	}
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
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.transfers.WithTx(tx).Create(ctx, transfer); err != nil {
			return err
		}
		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "transfer",
			EntityID:   transferID,
			Action:     "created",
			ToStatus:   string(domain.TransferStatusPending),
			Metadata: map[string]any{
				"amount":            in.Amount,
				"currency":          in.Currency,
				"recipient_user_id": recipient.ID,
			},
		})
	}); err != nil {
		return nil, err
	}

	state := saga.NewState(map[string]string{
		"transfer_id":                 transferID,
		"sender_user_id":              in.SenderUserID,
		"sender_ledger_account_id":    senderWallet.LedgerAccountID,
		"recipient_user_id":           recipient.ID,
		"recipient_ledger_account_id": recipientWallet.LedgerAccountID,
		"amount":                      in.Amount,
		"currency":                    in.Currency,
		"idempotency_key":             in.IdempotencyKey,
	})

	orchestrator := saga.New("p2p_transfer", uc.steps(), uc.sagaStore)
	if err := orchestrator.Run(ctx, in.IdempotencyKey, state); err != nil {
		reason := err.Error()
		if biz, ok := err.(*saga.BusinessError); ok {
			reason = biz.Code
		}
		if txErr := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
			if err := uc.transfers.WithTx(tx).MarkFailed(ctx, transferID, reason); err != nil {
				return err
			}
			return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
				EntityType: "transfer",
				EntityID:   transferID,
				Action:     "mark_failed",
				FromStatus: string(domain.TransferStatusPending),
				ToStatus:   "failed",
				Metadata:   map[string]any{"reason": reason},
			})
		}); txErr != nil {
			return nil, txErr
		}
		t, getErr := uc.transfers.GetByID(ctx, transferID)
		if getErr != nil {
			return nil, getErr
		}
		return t, nil
	}

	uc.runAMLMonitoring(ctx, state)

	senderUser, _ := uc.users.GetByID(ctx, in.SenderUserID)
	completedEvent := events.TransferCompleted{
		TransferID:           transferID,
		LedgerTransferID:       state.Get("ledger_transfer_id"),
		SenderUserID:         in.SenderUserID,
		RecipientUserID:      state.Get("recipient_user_id"),
		Amount:               in.Amount,
		Currency:             in.Currency,
		Memo:                 in.Memo,
		SenderDisplayName:    displayName(senderUser),
		RecipientDisplayName: displayName(recipient),
	}
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.transfers.WithTx(tx).MarkCompleted(ctx, transferID, state.Get("ledger_transfer_id")); err != nil {
			return err
		}
		if err := uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "transfer",
			EntityID:   transferID,
			Action:     "mark_completed",
			FromStatus: string(domain.TransferStatusPending),
			ToStatus:   "completed",
			Metadata:   map[string]any{"ledger_transfer_id": state.Get("ledger_transfer_id")},
		}); err != nil {
			return err
		}
		return uc.outbox.WithTx(tx).Publish(ctx, completedEvent)
	}); err != nil {
		return nil, err
	}

	if upsertErr := uc.users.UpsertPayee(ctx, in.SenderUserID, recipient.ID, "", in.IdempotencyKey+"-payee"); upsertErr != nil {
		slog.WarnContext(ctx, "upsert saved payee failed", "error", upsertErr, "sender", in.SenderUserID, "recipient", recipient.ID)
	}

	return uc.transfers.GetByID(ctx, transferID)
}

func displayName(user userclient.User) string {
	if user.Email != "" {
		return user.Email
	}
	if user.Phone != "" {
		return user.Phone
	}
	return user.ID
}

func (uc *P2PTransferUseCase) GetByID(ctx context.Context, id string) (*domain.Transfer, error) {
	return uc.transfers.GetByID(ctx, id)
}

type TransferListResult struct {
	Transfers  []domain.Transfer
	NextCursor string
}

func (uc *P2PTransferUseCase) List(ctx context.Context, userID string, limit int, cursor string) (TransferListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	pageSize := limit + 1
	var cursorAt *time.Time
	var cursorID string
	if cursor != "" {
		decoded, err := pagination.Decode(cursor)
		if err != nil {
			return TransferListResult{}, err
		}
		at := decoded.CreatedAt
		cursorAt = &at
		cursorID = decoded.ID
	}
	rows, err := uc.transfers.ListByUser(ctx, userID, pageSize, cursorAt, cursorID)
	if err != nil {
		return TransferListResult{}, err
	}
	items, next := pagination.Trim(rows, limit, func(t domain.Transfer) pagination.Cursor {
		return pagination.Cursor{CreatedAt: t.CreatedAt, ID: t.ID}
	})
	return TransferListResult{Transfers: items, NextCursor: next}, nil
}

func (uc *P2PTransferUseCase) runAMLMonitoring(ctx context.Context, state *saga.State) {
	if uc.aml == nil {
		return
	}
	result, err := uc.aml.Evaluate(
		state.Get("sender_user_id"), "p2p", state.Get("amount"), state.Get("currency"),
		amlmonitor.EvaluateOpts{},
	)
	if err != nil {
		slog.WarnContext(ctx, "aml evaluation failed", "error", err, "transfer_id", state.Get("transfer_id"))
		return
	}
	if uc.amlRepo == nil {
		return
	}
	evalID, err := uc.amlRepo.RecordEvaluation(ctx, "transfer", state.Get("transfer_id"), state.Get("sender_user_id"),
		"p2p", state.Get("amount"), state.Get("currency"), result)
	if err != nil {
		slog.WarnContext(ctx, "record aml evaluation failed", "error", err, "transfer_id", state.Get("transfer_id"))
		return
	}
	if result.Disposition == amlmonitor.DispositionClear {
		return
	}
	caseType := amlCaseType(result)
	if caseType == "" {
		return
	}
	if err := uc.amlRepo.OpenCase(ctx, evalID, state.Get("sender_user_id"), "transfer", state.Get("transfer_id"),
		caseType, result.ReasonCode, reqctx.CorrelationID(ctx)); err != nil {
		slog.WarnContext(ctx, "open aml case failed", "error", err, "transfer_id", state.Get("transfer_id"))
	}
}

func amlCaseType(result amlmonitor.Result) string {
	switch result.Disposition {
	case amlmonitor.DispositionReport:
		switch result.ReasonCode {
		case "CTR_THRESHOLD":
			return "ctr"
		case "STRUCTURING":
			return "sar"
		default:
			return "sar"
		}
	case amlmonitor.DispositionReview:
		return "review"
	default:
		return ""
	}
}

func (uc *P2PTransferUseCase) resolveRecipient(ctx context.Context, in P2PTransferInput) (userclient.User, error) {
	provided := 0
	if in.RecipientPhone != "" {
		provided++
	}
	if in.RecipientEmail != "" {
		provided++
	}
	if in.RecipientUserID != "" {
		provided++
	}
	if provided != 1 {
		return userclient.User{}, fmt.Errorf("exactly one of recipient_phone, recipient_email, or recipient_user_id is required")
	}

	switch {
	case in.RecipientPhone != "":
		user, err := uc.users.GetByPhone(ctx, in.RecipientPhone)
		if err != nil {
			return userclient.User{}, fmt.Errorf("recipient: %w", err)
		}
		return user, nil
	case in.RecipientEmail != "":
		user, err := uc.users.GetByEmail(ctx, in.RecipientEmail)
		if err != nil {
			return userclient.User{}, fmt.Errorf("recipient: %w", err)
		}
		return user, nil
	default:
		user, err := uc.users.GetByID(ctx, in.RecipientUserID)
		if err != nil {
			return userclient.User{}, fmt.Errorf("recipient: %w", err)
		}
		return user, nil
	}
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
				if uc.fraudRepo != nil {
					if recErr := uc.fraudRepo.Record(ctx, "transfer", state.Get("transfer_id"), state.Get("sender_user_id"),
						"p2p", state.Get("amount"), state.Get("currency"), result); recErr != nil {
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
