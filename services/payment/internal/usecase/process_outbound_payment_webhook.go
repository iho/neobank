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

// OutboundPaymentWebhookPayload is the body the rails simulator delivers on
// its settled/returned/failed webhooks for a neobank-initiated payment.
type OutboundPaymentWebhookPayload struct {
	PaymentID        string `json:"payment_id"`
	AccountID        string `json:"account_id"`
	ExternalRef      string `json:"external_ref"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	CounterpartyIBAN string `json:"counterparty_iban"`
	Reference        string `json:"reference"`
}

// ProcessOutboundPaymentWebhookUseCase advances a bank transfer order from
// the rail's async outcome. Settlement needs no ledger action (the debit
// already happened at send time); a return or failure reverses that debit
// back to the wallet.
type ProcessOutboundPaymentWebhookUseCase struct {
	orders            port.BankTransferOrderRepository
	users             *userclient.Client
	ledger            *ledgerclient.Client
	outbox            outbox.TxPublisher
	audit             audit.Recorder
	settlementAccount string
	tx                *pgutil.TxRunner
}

func NewProcessOutboundPaymentWebhookUseCase(
	orders port.BankTransferOrderRepository,
	users *userclient.Client,
	ledger *ledgerclient.Client,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
	settlementAccount string,
	tx *pgutil.TxRunner,
) *ProcessOutboundPaymentWebhookUseCase {
	return &ProcessOutboundPaymentWebhookUseCase{
		orders:            orders,
		users:             users,
		ledger:            ledger,
		outbox:            outboxPublisher,
		audit:             auditRecorder,
		settlementAccount: settlementAccount,
		tx:                tx,
	}
}

// Settled marks a processing order settled; no ledger action needed since
// the debit already moved the funds when the order was sent.
func (uc *ProcessOutboundPaymentWebhookUseCase) Settled(ctx context.Context, in OutboundPaymentWebhookPayload) (domain.BankTransferOrder, error) {
	if in.PaymentID == "" {
		return domain.BankTransferOrder{}, fmt.Errorf("payment_id is required")
	}

	order, err := uc.orders.GetByRailsPaymentID(ctx, in.PaymentID)
	if err != nil {
		return domain.BankTransferOrder{}, err
	}

	if order == nil {
		return domain.BankTransferOrder{}, fmt.Errorf("bank transfer order for payment %q not found", in.PaymentID)
	}

	if order.Status != domain.BankTransferOrderStatusProcessing {
		// Already settled/returned/failed — a redelivered webhook is a no-op.
		return *order, nil
	}

	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.orders.WithTx(tx).MarkSettled(ctx, in.PaymentID); err != nil {
			return err
		}

		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "bank_transfer_order",
			EntityID:   order.ID,
			Action:     "settled",
			FromStatus: domain.BankTransferOrderStatusProcessing,
			ToStatus:   domain.BankTransferOrderStatusSettled,
		})
	}); err != nil {
		return domain.BankTransferOrder{}, err
	}

	order.Status = domain.BankTransferOrderStatusSettled

	return *order, nil
}

// ReturnedOrFailed reverses the original debit back to the user's wallet —
// "the money bounced after it looked done" (returned) or "the rail never
// accepted it" (failed) look the same from here: funds go back either way.
func (uc *ProcessOutboundPaymentWebhookUseCase) ReturnedOrFailed(ctx context.Context, in OutboundPaymentWebhookPayload, status, reason string) (domain.BankTransferOrder, error) {
	if in.PaymentID == "" {
		return domain.BankTransferOrder{}, fmt.Errorf("payment_id is required")
	}

	order, err := uc.orders.GetByRailsPaymentID(ctx, in.PaymentID)
	if err != nil {
		return domain.BankTransferOrder{}, err
	}

	if order == nil {
		return domain.BankTransferOrder{}, fmt.Errorf("bank transfer order for payment %q not found", in.PaymentID)
	}

	if order.Status == domain.BankTransferOrderStatusReturned || order.Status == domain.BankTransferOrderStatusFailed {
		// Already reversed — a redelivered or duplicated webhook is a no-op.
		return *order, nil
	}

	if uc.settlementAccount == "" {
		return domain.BankTransferOrder{}, fmt.Errorf("rails settlement account is not configured")
	}

	if uc.ledger == nil {
		return domain.BankTransferOrder{}, fmt.Errorf("ledger unavailable")
	}

	wallet, err := uc.users.GetWallet(ctx, order.UserID, order.Currency)
	if err != nil {
		return domain.BankTransferOrder{}, fmt.Errorf("recipient wallet: %w", err)
	}

	returnTransfer, err := uc.ledger.CreateTransfer(ctx, ledgerclient.CreateTransferInput{
		FromAccountID:  uc.settlementAccount,
		ToAccountID:    wallet.LedgerAccountID,
		Amount:         order.Amount,
		IdempotencyKey: in.PaymentID + ":return",
		Metadata: map[string]string{
			"rails_payment_id": in.PaymentID,
			"type":             "bank_transfer_order_return",
		},
	})
	if err != nil {
		return domain.BankTransferOrder{}, fmt.Errorf("reverse ledger transfer: %w", err)
	}

	event := events.BankTransferReturned{
		OrderID:          order.ID,
		UserID:           order.UserID,
		Amount:           order.Amount,
		Currency:         order.Currency,
		ReturnTransferID: returnTransfer.Id,
		Reason:           reason,
	}

	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.orders.WithTx(tx).MarkReturned(ctx, in.PaymentID, status, returnTransfer.Id); err != nil {
			return err
		}

		if err := uc.outbox.WithTx(tx).Publish(ctx, event); err != nil {
			return err
		}

		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "bank_transfer_order",
			EntityID:   order.ID,
			Action:     status,
			FromStatus: order.Status,
			ToStatus:   status,
			Metadata: map[string]any{
				"reason":             reason,
				"return_transfer_id": returnTransfer.Id,
			},
		})
	}); err != nil {
		return domain.BankTransferOrder{}, err
	}

	order.Status = status
	order.ReturnTransferID = returnTransfer.Id

	return *order, nil
}
