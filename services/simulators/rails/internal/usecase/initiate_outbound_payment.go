package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/iho/neobank/pkg/vendorsim"
	"github.com/iho/neobank/services/simulators/rails/internal/domain"
	"github.com/iho/neobank/services/simulators/rails/internal/port"
	"github.com/shopspring/decimal"
)

const (
	// EventPaymentSettled is delivered when a neobank-initiated outbound
	// payment clears the rail.
	EventPaymentSettled = "rails.payment.settled"
	// EventPaymentReturned is delivered when a payment that already looked
	// settled bounces back — the interesting saga this unlocks.
	EventPaymentReturned = "rails.payment.returned"
	// EventPaymentFailed is delivered when a payment never settles at all
	// (an upfront validation failure at the rail).
	EventPaymentFailed = "rails.payment.failed"
)

// settleDelay and returnDelay are deliberately different so a returned
// payment always arrives after its own settled webhook, modeling "the
// money bounced after it looked done."
const (
	settleDelay = 2 * time.Second
	returnDelay = 10 * time.Second
)

// PaymentWebhookPayload is the webhook body delivered on EventPaymentSettled,
// EventPaymentReturned, and EventPaymentFailed.
type PaymentWebhookPayload struct {
	PaymentID        string `json:"payment_id"`
	AccountID        string `json:"account_id"`
	ExternalRef      string `json:"external_ref"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	CounterpartyIBAN string `json:"counterparty_iban"`
	Reference        string `json:"reference"`
}

type InitiateOutboundPaymentInput struct {
	AccountID        string
	Amount           string
	Currency         string
	CounterpartyIBAN string
	Reference        string
}

// InitiateOutboundPaymentUseCase is the neobank-initiated entry point for
// sending money out over the rail: accepted synchronously, settled or
// returned asynchronously via webhook, driven by the shared magic-value
// conventions (reference containing RETURN bounces after settlement;
// amount ending in .99 fails validation with no settlement at all).
type InitiateOutboundPaymentUseCase struct {
	accounts   port.AccountRepository
	payments   port.OutboundPaymentRepository
	dispatcher WebhookDispatcher
	eventsURL  string
}

func NewInitiateOutboundPaymentUseCase(
	accounts port.AccountRepository,
	payments port.OutboundPaymentRepository,
	dispatcher WebhookDispatcher,
	eventsURL string,
) *InitiateOutboundPaymentUseCase {
	return &InitiateOutboundPaymentUseCase{
		accounts:   accounts,
		payments:   payments,
		dispatcher: dispatcher,
		eventsURL:  eventsURL,
	}
}

func (uc *InitiateOutboundPaymentUseCase) Execute(ctx context.Context, in InitiateOutboundPaymentInput) (domain.OutboundPayment, error) {
	if in.AccountID == "" || in.Amount == "" || in.Currency == "" || in.CounterpartyIBAN == "" {
		return domain.OutboundPayment{}, fmt.Errorf("account_id, amount, currency, and counterparty_iban are required")
	}

	account, err := uc.accounts.GetByID(ctx, in.AccountID)
	if err != nil {
		return domain.OutboundPayment{}, err
	}

	if account == nil {
		return domain.OutboundPayment{}, fmt.Errorf("account %q not found", in.AccountID)
	}

	payment, err := uc.payments.Create(ctx, account.ID, in.Amount, in.Currency, in.CounterpartyIBAN, in.Reference)
	if err != nil {
		return domain.OutboundPayment{}, err
	}

	payload := PaymentWebhookPayload{
		PaymentID:        payment.ID,
		AccountID:        account.ID,
		ExternalRef:      account.ExternalRef,
		Amount:           payment.Amount,
		Currency:         payment.Currency,
		CounterpartyIBAN: payment.CounterpartyIBAN,
		Reference:        payment.Reference,
	}

	if err := uc.scheduleOutcome(ctx, payload); err != nil {
		return domain.OutboundPayment{}, fmt.Errorf("schedule outcome webhook: %w", err)
	}

	return payment, nil
}

func (uc *InitiateOutboundPaymentUseCase) scheduleOutcome(ctx context.Context, payload PaymentWebhookPayload) error {
	amount, err := decimal.NewFromString(payload.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount %q: %w", payload.Amount, err)
	}

	amountMinor := amount.Mul(decimal.NewFromInt(100)).IntPart()

	switch {
	case vendorsim.AmountEndsInCents(amountMinor, 99):
		_, err := uc.dispatcher.EnqueueAfter(ctx, uc.eventsURL, EventPaymentFailed, payload, settleDelay)
		return err

	case vendorsim.ContainsToken(payload.Reference, "RETURN"):
		if _, err := uc.dispatcher.EnqueueAfter(ctx, uc.eventsURL, EventPaymentSettled, payload, settleDelay); err != nil {
			return err
		}

		_, err := uc.dispatcher.EnqueueAfter(ctx, uc.eventsURL, EventPaymentReturned, payload, returnDelay)
		return err

	default:
		_, err := uc.dispatcher.EnqueueAfter(ctx, uc.eventsURL, EventPaymentSettled, payload, settleDelay)
		return err
	}
}
