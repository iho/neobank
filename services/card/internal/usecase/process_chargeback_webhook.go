package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/userclient"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/port"
	"github.com/jackc/pgx/v5"
)

// ChargebackWebhookPayload is the body the cardproc simulator delivers on
// its chargeback lifecycle webhooks (opened, won, lost).
type ChargebackWebhookPayload struct {
	ChargebackID    string `json:"chargeback_id"`
	AuthorizationID string `json:"authorization_id"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	Reason          string `json:"reason,omitempty"`
}

// ProcessChargebackWebhookUseCase advances a dispute from the cardproc
// simulator's async chargeback lifecycle: opening one issues an immediate
// provisional credit (genuinely new ledger movement, unlike auth expiry
// which just reuses ReverseAuthorizationUseCase); resolving it either
// finalizes that credit (won) or claws it back (lost).
type ProcessChargebackWebhookUseCase struct {
	auths      port.AuthorizationRepository
	disputes   port.DisputeRepository
	users      *userclient.Client
	ledger     *ledgerclient.Client
	outbox     outbox.TxPublisher
	audit      audit.Recorder
	settlement string
	tx         *pgutil.TxRunner
}

func NewProcessChargebackWebhookUseCase(
	auths port.AuthorizationRepository,
	disputes port.DisputeRepository,
	users *userclient.Client,
	ledger *ledgerclient.Client,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	settlementAccountID string,
	tx *pgutil.TxRunner,
) *ProcessChargebackWebhookUseCase {
	return &ProcessChargebackWebhookUseCase{
		auths:      auths,
		disputes:   disputes,
		users:      users,
		ledger:     ledger,
		outbox:     outboxPublisher,
		audit:      auditRecorder,
		settlement: settlementAccountID,
		tx:         tx,
	}
}

// Opened issues the cardholder a provisional credit and records the
// dispute; idempotent on the chargeback ID (a redelivered webhook is a
// no-op, not a second credit).
func (uc *ProcessChargebackWebhookUseCase) Opened(ctx context.Context, in ChargebackWebhookPayload) (domain.Dispute, error) {
	if in.ChargebackID == "" || in.AuthorizationID == "" {
		return domain.Dispute{}, fmt.Errorf("chargeback_id and authorization_id are required")
	}

	if existing, err := uc.disputes.GetByChargebackID(ctx, in.ChargebackID); err != nil {
		return domain.Dispute{}, err
	} else if existing != nil {
		return *existing, nil
	}

	auth, err := uc.auths.GetByID(ctx, in.AuthorizationID)
	if err != nil {
		return domain.Dispute{}, err
	}
	if auth.Status != domain.AuthStatusCaptured {
		return domain.Dispute{}, fmt.Errorf("authorization %q is not captured, cannot charge back", in.AuthorizationID)
	}

	if uc.settlement == "" {
		return domain.Dispute{}, fmt.Errorf("settlement account is not configured")
	}
	if uc.ledger == nil {
		return domain.Dispute{}, fmt.Errorf("ledger unavailable")
	}

	wallet, err := uc.users.GetWallet(ctx, auth.UserID, auth.Currency)
	if err != nil {
		return domain.Dispute{}, fmt.Errorf("wallet: %w", err)
	}

	creditTransfer, err := uc.ledger.CreateTransfer(ctx, ledgerclient.CreateTransferInput{
		FromAccountID:  uc.settlement,
		ToAccountID:    wallet.LedgerAccountID,
		Amount:         in.Amount,
		IdempotencyKey: in.ChargebackID + ":credit",
		Metadata: map[string]string{
			"chargeback_id":    in.ChargebackID,
			"authorization_id": in.AuthorizationID,
			"type":             "chargeback_provisional_credit",
		},
	})
	if err != nil {
		return domain.Dispute{}, fmt.Errorf("provisional credit transfer: %w", err)
	}

	dispute := domain.Dispute{
		ChargebackID:                in.ChargebackID,
		AuthorizationID:             auth.ID,
		CardID:                      auth.CardID,
		UserID:                      auth.UserID,
		Amount:                      in.Amount,
		Currency:                    auth.Currency,
		Reason:                      in.Reason,
		ProvisionalCreditTransferID: creditTransfer.Id,
	}

	event := events.CardChargebackOpened{
		DisputeID:                   in.ChargebackID,
		AuthorizationID:             auth.ID,
		CardID:                      auth.CardID,
		UserID:                      auth.UserID,
		Amount:                      in.Amount,
		Currency:                    auth.Currency,
		Reason:                      in.Reason,
		ProvisionalCreditTransferID: creditTransfer.Id,
	}

	var created *domain.Dispute
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var createErr error
		created, createErr = uc.disputes.WithTx(tx).Create(ctx, dispute)
		if createErr != nil {
			return createErr
		}

		if err := uc.outbox.WithTx(tx).Publish(ctx, event); err != nil {
			return err
		}

		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "dispute",
			EntityID:   in.ChargebackID,
			Action:     "opened",
			ToStatus:   domain.DisputeStatusOpen,
			Metadata: map[string]any{
				"authorization_id":               auth.ID,
				"provisional_credit_transfer_id": creditTransfer.Id,
			},
		})
	}); err != nil {
		return domain.Dispute{}, err
	}

	return *created, nil
}

// Resolved closes out a dispute: "won" leaves the provisional credit in
// place (no ledger action); "lost" claws it back.
func (uc *ProcessChargebackWebhookUseCase) Resolved(ctx context.Context, in ChargebackWebhookPayload, outcome string) (domain.Dispute, error) {
	if in.ChargebackID == "" {
		return domain.Dispute{}, fmt.Errorf("chargeback_id is required")
	}
	if outcome != domain.DisputeStatusWon && outcome != domain.DisputeStatusLost {
		return domain.Dispute{}, fmt.Errorf("outcome must be %q or %q, got %q", domain.DisputeStatusWon, domain.DisputeStatusLost, outcome)
	}

	dispute, err := uc.disputes.GetByChargebackID(ctx, in.ChargebackID)
	if err != nil {
		return domain.Dispute{}, err
	}
	if dispute == nil {
		return domain.Dispute{}, fmt.Errorf("dispute for chargeback %q not found", in.ChargebackID)
	}
	if dispute.Status != domain.DisputeStatusOpen {
		// Already resolved — a redelivered or duplicated webhook is a no-op.
		return *dispute, nil
	}

	reversalTransferID := ""
	if outcome == domain.DisputeStatusLost {
		if uc.settlement == "" {
			return domain.Dispute{}, fmt.Errorf("settlement account is not configured")
		}
		if uc.ledger == nil {
			return domain.Dispute{}, fmt.Errorf("ledger unavailable")
		}

		wallet, err := uc.users.GetWallet(ctx, dispute.UserID, dispute.Currency)
		if err != nil {
			return domain.Dispute{}, fmt.Errorf("wallet: %w", err)
		}

		reversal, err := uc.ledger.CreateTransfer(ctx, ledgerclient.CreateTransferInput{
			FromAccountID:  wallet.LedgerAccountID,
			ToAccountID:    uc.settlement,
			Amount:         dispute.Amount,
			IdempotencyKey: in.ChargebackID + ":reversal",
			Metadata: map[string]string{
				"chargeback_id": in.ChargebackID,
				"type":          "chargeback_credit_reversal",
			},
		})
		if err != nil {
			return domain.Dispute{}, fmt.Errorf("reversal transfer: %w", err)
		}
		reversalTransferID = reversal.Id
	}

	event := events.CardChargebackResolved{
		DisputeID:          dispute.ChargebackID,
		AuthorizationID:    dispute.AuthorizationID,
		CardID:             dispute.CardID,
		UserID:             dispute.UserID,
		Amount:             dispute.Amount,
		Currency:           dispute.Currency,
		Outcome:            outcome,
		ReversalTransferID: reversalTransferID,
	}

	var resolved *domain.Dispute
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var markErr error
		resolved, markErr = uc.disputes.WithTx(tx).MarkResolved(ctx, in.ChargebackID, outcome, reversalTransferID)
		if markErr != nil {
			return markErr
		}

		if err := uc.outbox.WithTx(tx).Publish(ctx, event); err != nil {
			return err
		}

		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "dispute",
			EntityID:   in.ChargebackID,
			Action:     outcome,
			FromStatus: domain.DisputeStatusOpen,
			ToStatus:   outcome,
			Metadata: map[string]any{
				"reversal_transfer_id": reversalTransferID,
			},
		})
	}); err != nil {
		return domain.Dispute{}, err
	}

	return *resolved, nil
}
