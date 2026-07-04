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
	"github.com/iho/neobank/services/payment/internal/adapter/railsclient"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/port"
	"github.com/jackc/pgx/v5"
)

type SendBankTransferInput struct {
	UserID           string
	Amount           string
	Currency         string
	CounterpartyIBAN string
	Reference        string
}

// SendBankTransferUseCase is the neobank-initiated entry point for sending
// money out over the rails simulator: it debits the wallet immediately
// (the ledger transfer is what makes the money actually leave), then hands
// off to the rail. Settlement/return arrives later via webhook
// (ProcessOutboundPaymentWebhookUseCase).
type SendBankTransferUseCase struct {
	orders            port.BankTransferOrderRepository
	getOrCreateAcct   *GetOrCreateBankAccountUseCase
	users             *userclient.Client
	ledger            *ledgerclient.Client
	rails             *railsclient.Client
	outbox            outbox.TxPublisher
	audit             audit.Recorder
	settlementAccount string
	tx                *pgutil.TxRunner
}

func NewSendBankTransferUseCase(
	orders port.BankTransferOrderRepository,
	getOrCreateAcct *GetOrCreateBankAccountUseCase,
	users *userclient.Client,
	ledger *ledgerclient.Client,
	rails *railsclient.Client,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	settlementAccount string,
	tx *pgutil.TxRunner,
) *SendBankTransferUseCase {
	return &SendBankTransferUseCase{
		orders:            orders,
		getOrCreateAcct:   getOrCreateAcct,
		users:             users,
		ledger:            ledger,
		rails:             rails,
		outbox:            outboxPublisher,
		audit:             auditRecorder,
		settlementAccount: settlementAccount,
		tx:                tx,
	}
}

func (uc *SendBankTransferUseCase) Execute(ctx context.Context, in SendBankTransferInput) (domain.BankTransferOrder, error) {
	if in.UserID == "" || in.Amount == "" || in.Currency == "" || in.CounterpartyIBAN == "" {
		return domain.BankTransferOrder{}, fmt.Errorf("user_id, amount, currency, and counterparty_iban are required")
	}

	if uc.settlementAccount == "" {
		return domain.BankTransferOrder{}, fmt.Errorf("rails settlement account is not configured")
	}

	if uc.ledger == nil || uc.rails == nil {
		return domain.BankTransferOrder{}, fmt.Errorf("ledger or rails simulator unavailable")
	}

	wallet, err := uc.users.GetWallet(ctx, in.UserID, in.Currency)
	if err != nil {
		return domain.BankTransferOrder{}, fmt.Errorf("source wallet: %w", err)
	}

	bankAccount, err := uc.getOrCreateAcct.Execute(ctx, GetOrCreateBankAccountInput{UserID: in.UserID, Currency: in.Currency})
	if err != nil {
		return domain.BankTransferOrder{}, fmt.Errorf("rails account: %w", err)
	}

	// The debit is idempotent on a client-chosen key would be ideal, but
	// this use case doesn't take one today; the rails payment ID (created
	// next) becomes the durable idempotency anchor for everything downstream.
	ledgerTransfer, err := uc.ledger.CreateTransfer(ctx, ledgerclient.CreateTransferInput{
		FromAccountID: wallet.LedgerAccountID,
		ToAccountID:   uc.settlementAccount,
		Amount:        in.Amount,
		Metadata: map[string]string{
			"type": "bank_transfer_order_debit",
		},
	})
	if err != nil {
		return domain.BankTransferOrder{}, fmt.Errorf("debit wallet: %w", err)
	}

	railsPayment, err := uc.rails.CreatePayment(ctx, bankAccount.RailsAccountID, in.Amount, in.Currency, in.CounterpartyIBAN, in.Reference)
	if err != nil {
		return domain.BankTransferOrder{}, fmt.Errorf("initiate rails payment: %w", err)
	}

	order := domain.BankTransferOrder{
		RailsPaymentID:   railsPayment.ID,
		UserID:           in.UserID,
		Amount:           in.Amount,
		Currency:         in.Currency,
		CounterpartyIBAN: in.CounterpartyIBAN,
		Reference:        in.Reference,
		LedgerTransferID: ledgerTransfer.Id,
		Status:           domain.BankTransferOrderStatusProcessing,
	}

	event := events.BankTransferSent{
		UserID:           in.UserID,
		LedgerTransferID: ledgerTransfer.Id,
		Amount:           in.Amount,
		Currency:         in.Currency,
		CounterpartyIBAN: in.CounterpartyIBAN,
		Reference:        in.Reference,
	}

	var created domain.BankTransferOrder
	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var createErr error
		created, createErr = uc.orders.WithTx(tx).Create(ctx, order)
		if createErr != nil {
			return createErr
		}

		event.OrderID = created.ID

		if err := uc.outbox.WithTx(tx).Publish(ctx, event); err != nil {
			return err
		}

		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "bank_transfer_order",
			EntityID:   created.ID,
			Action:     "sent",
			ToStatus:   domain.BankTransferOrderStatusProcessing,
			Metadata: map[string]any{
				"rails_payment_id":   railsPayment.ID,
				"user_id":            in.UserID,
				"amount":             in.Amount,
				"currency":           in.Currency,
				"counterparty_iban":  in.CounterpartyIBAN,
				"ledger_transfer_id": ledgerTransfer.Id,
			},
		})
	}); err != nil {
		return domain.BankTransferOrder{}, err
	}

	return created, nil
}
