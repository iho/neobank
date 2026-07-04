package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
	"github.com/iho/neobank/services/simulators/cardproc/internal/port"
)

// EventChargebackOpened is delivered when an admin disputes a captured
// transaction; the card service issues the cardholder a provisional credit
// in response.
const EventChargebackOpened = "card.chargeback.opened"

// ChargebackWebhookPayload is the body delivered on the chargeback
// lifecycle webhooks (opened, won, lost). ChargebackID is this simulator's
// own ID, which the card service tracks as its dispute_id.
type ChargebackWebhookPayload struct {
	ChargebackID    string `json:"chargeback_id"`
	AuthorizationID string `json:"authorization_id"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	Reason          string `json:"reason,omitempty"`
}

// OpenChargebackUseCase is the admin entry point for disputing a captured
// transaction — the cardproc simulator's chargeback flow.
type OpenChargebackUseCase struct {
	txs         port.TransactionRepository
	chargebacks port.ChargebackRepository
	dispatcher  WebhookDispatcher
	eventsURL   string
}

func NewOpenChargebackUseCase(
	txs port.TransactionRepository,
	chargebacks port.ChargebackRepository,
	dispatcher WebhookDispatcher,
	eventsURL string,
) *OpenChargebackUseCase {
	return &OpenChargebackUseCase{txs: txs, chargebacks: chargebacks, dispatcher: dispatcher, eventsURL: eventsURL}
}

func (uc *OpenChargebackUseCase) Execute(ctx context.Context, transactionID, reason string) (domain.Chargeback, error) {
	tx, err := uc.txs.GetByID(ctx, transactionID)
	if err != nil {
		return domain.Chargeback{}, err
	}

	if tx == nil {
		return domain.Chargeback{}, fmt.Errorf("transaction %q not found", transactionID)
	}

	if tx.Status != domain.TransactionStatusCaptured {
		return domain.Chargeback{}, fmt.Errorf("only captured transactions can be charged back, got status %q", tx.Status)
	}

	cb, err := uc.chargebacks.Create(ctx, tx.ID, tx.AuthorizationID, tx.Amount, tx.Currency, reason)
	if err != nil {
		return domain.Chargeback{}, err
	}

	if _, err := uc.dispatcher.Enqueue(ctx, uc.eventsURL, EventChargebackOpened, ChargebackWebhookPayload{
		ChargebackID:    cb.ID,
		AuthorizationID: cb.AuthorizationID,
		Amount:          cb.Amount,
		Currency:        cb.Currency,
		Reason:          cb.Reason,
	}); err != nil {
		return domain.Chargeback{}, fmt.Errorf("schedule chargeback webhook: %w", err)
	}

	return cb, nil
}
