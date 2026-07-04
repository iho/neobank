package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/services/simulators/cardproc/internal/adapter/cardclient"
	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
	"github.com/iho/neobank/services/simulators/cardproc/internal/port"
)

const (
	// EventCaptured is delivered once a previously-approved transaction settles.
	EventCaptured = "card.captured"
	// EventAuthReversed is delivered when a hold is released without capture.
	EventAuthReversed = "card.auth.reversed"
)

// CardEventPayload is the webhook body delivered on EventCaptured and
// EventAuthReversed, carrying the neobank's own authorization ID (learned
// from the synchronous auth response), not this simulator's transaction ID.
type CardEventPayload struct {
	AuthorizationID string `json:"authorization_id"`
	Reason          string `json:"reason,omitempty"`
}

type SimulateTransactionInput struct {
	CardRef      string
	Amount       string
	Currency     string
	MerchantName string
	MCC          string
	// Capture requests an immediate settlement (a "sale"); false leaves the
	// transaction authorized-only, to be captured or reversed later via the
	// admin API.
	Capture bool
}

// SimulateTransactionUseCase is the admin/test entry point that simulates a
// merchant charge: it calls the card service's real-time authorization
// webhook synchronously and, if approved and requested, schedules a capture
// webhook — see docs/vendor-simulators-plan.md Phase 2.
type SimulateTransactionUseCase struct {
	cards      port.CardRepository
	txs        port.TransactionRepository
	cardClient *cardclient.Client
	dispatcher WebhookDispatcher
	eventsURL  string
}

func NewSimulateTransactionUseCase(
	cards port.CardRepository,
	txs port.TransactionRepository,
	cardClient *cardclient.Client,
	dispatcher WebhookDispatcher,
	eventsURL string,
) *SimulateTransactionUseCase {
	return &SimulateTransactionUseCase{
		cards:      cards,
		txs:        txs,
		cardClient: cardClient,
		dispatcher: dispatcher,
		eventsURL:  eventsURL,
	}
}

func (uc *SimulateTransactionUseCase) Execute(ctx context.Context, in SimulateTransactionInput) (domain.Transaction, error) {
	if in.CardRef == "" || in.Amount == "" || in.Currency == "" {
		return domain.Transaction{}, fmt.Errorf("card_ref, amount, and currency are required")
	}

	card, err := uc.cards.GetByID(ctx, in.CardRef)
	if err != nil {
		return domain.Transaction{}, err
	}

	if card == nil {
		return domain.Transaction{}, fmt.Errorf("card %q not found", in.CardRef)
	}

	tx, err := uc.txs.Create(ctx, card.ID, in.Amount, in.Currency, in.MerchantName, in.MCC)
	if err != nil {
		return domain.Transaction{}, err
	}

	result, err := uc.cardClient.Authorize(ctx, cardclient.AuthorizeRequest{
		TransactionID:        tx.ID,
		CardRef:              card.ID,
		Amount:               in.Amount,
		Currency:             in.Currency,
		MerchantName:         in.MerchantName,
		MerchantCategoryCode: in.MCC,
	})
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("authorize: %w", err)
	}

	status := domain.TransactionStatusDeclined
	if result.Decision == "approved" {
		status = domain.TransactionStatusApproved
	}

	tx, err = uc.txs.SetAuthResult(ctx, tx.ID, status, result.AuthorizationID, result.ReasonCode)
	if err != nil {
		return domain.Transaction{}, err
	}

	if status == domain.TransactionStatusApproved && in.Capture {
		if err := uc.scheduleCapture(ctx, tx); err != nil {
			return domain.Transaction{}, err
		}

		tx.Status = domain.TransactionStatusCaptured
	}

	return tx, nil
}

func (uc *SimulateTransactionUseCase) scheduleCapture(ctx context.Context, tx domain.Transaction) error {
	if _, err := uc.dispatcher.Enqueue(ctx, uc.eventsURL, EventCaptured, CardEventPayload{
		AuthorizationID: tx.AuthorizationID,
	}); err != nil {
		return fmt.Errorf("schedule capture webhook: %w", err)
	}

	return uc.txs.MarkCaptured(ctx, tx.ID)
}
