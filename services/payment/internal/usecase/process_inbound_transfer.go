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
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/port"
	"github.com/jackc/pgx/v5"
)

// InboundTransferWebhookPayload is the body the rails simulator (or, later,
// a real payment rail) delivers on its "transfer received" webhook. It is
// this service's own copy of that external contract, not shared Go types.
type InboundTransferWebhookPayload struct {
	TransferID  string `json:"transfer_id"`
	AccountID   string `json:"account_id"`
	ExternalRef string `json:"external_ref"`
	IBAN        string `json:"iban"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	SenderName  string `json:"sender_name"`
	Reference   string `json:"reference"`
}

// ProcessInboundTransferUseCase credits a user's wallet for money that
// arrived on the rails simulator, idempotently on the rail's transfer ID so
// a redelivered or duplicated webhook never double-credits.
type ProcessInboundTransferUseCase struct {
	bankTransfers     port.BankTransferRepository
	users             *userclient.Client
	ledger            *ledgerclient.Client
	outbox            outbox.TxPublisher
	audit             audit.Recorder
	settlementAccount string
	tx                *pgutil.TxRunner
}

func NewProcessInboundTransferUseCase(
	bankTransfers port.BankTransferRepository,
	users *userclient.Client,
	ledger *ledgerclient.Client,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	settlementAccount string,
	tx *pgutil.TxRunner,
) *ProcessInboundTransferUseCase {
	return &ProcessInboundTransferUseCase{
		bankTransfers:     bankTransfers,
		users:             users,
		ledger:            ledger,
		outbox:            outboxPublisher,
		audit:             auditRecorder,
		settlementAccount: settlementAccount,
		tx:                tx,
	}
}

func (uc *ProcessInboundTransferUseCase) Execute(ctx context.Context, in InboundTransferWebhookPayload) (domain.BankTransfer, error) {
	if in.TransferID == "" || in.ExternalRef == "" || in.Amount == "" || in.Currency == "" {
		return domain.BankTransfer{}, fmt.Errorf("transfer_id, external_ref, amount, and currency are required")
	}

	if existing, err := uc.bankTransfers.GetByRailsTransferID(ctx, in.TransferID); err != nil {
		return domain.BankTransfer{}, err
	} else if existing != nil {
		return *existing, nil
	}

	if uc.settlementAccount == "" {
		return domain.BankTransfer{}, fmt.Errorf("rails settlement account is not configured")
	}

	if uc.ledger == nil {
		return domain.BankTransfer{}, fmt.Errorf("ledger unavailable")
	}

	// external_ref is the user ID the rails account was issued for; see
	// GetOrCreateBankAccountUseCase.
	userID := in.ExternalRef

	wallet, err := uc.users.GetWallet(ctx, userID, in.Currency)
	if err != nil {
		return domain.BankTransfer{}, fmt.Errorf("recipient wallet: %w", err)
	}

	ledgerTransfer, err := uc.ledger.CreateTransfer(ctx, ledgerclient.CreateTransferInput{
		FromAccountID:  uc.settlementAccount,
		ToAccountID:    wallet.LedgerAccountID,
		Amount:         in.Amount,
		IdempotencyKey: in.TransferID,
		Metadata: map[string]string{
			"rails_transfer_id": in.TransferID,
			"type":              "bank_transfer",
		},
	})
	if err != nil {
		return domain.BankTransfer{}, fmt.Errorf("ledger transfer: %w", err)
	}

	bankTransfer := domain.BankTransfer{
		RailsTransferID:  in.TransferID,
		UserID:           userID,
		Amount:           in.Amount,
		Currency:         in.Currency,
		SenderName:       in.SenderName,
		Reference:        in.Reference,
		LedgerTransferID: ledgerTransfer.Id,
		Status:           "completed",
	}

	event := events.BankTransferReceived{
		TransferID:       in.TransferID,
		UserID:           userID,
		LedgerTransferID: ledgerTransfer.Id,
		Amount:           in.Amount,
		Currency:         in.Currency,
		SenderName:       in.SenderName,
		Reference:        in.Reference,
	}

	var created domain.BankTransfer
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var createErr error
		created, createErr = uc.bankTransfers.WithTx(tx).Create(ctx, bankTransfer)
		if createErr != nil {
			return createErr
		}

		if err := uc.outbox.WithTx(tx).Publish(ctx, event); err != nil {
			return err
		}

		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "bank_transfer",
			EntityID:   created.ID,
			Action:     "completed",
			ToStatus:   "completed",
			Metadata: map[string]any{
				"rails_transfer_id":  in.TransferID,
				"user_id":            userID,
				"amount":             in.Amount,
				"currency":           in.Currency,
				"ledger_transfer_id": ledgerTransfer.Id,
			},
		})
	}); err != nil {
		return domain.BankTransfer{}, err
	}

	return created, nil
}
