package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
	"github.com/iho/neobank/services/simulators/cardproc/internal/port"
)

const (
	// EventChargebackWon is delivered when the cardholder's dispute is
	// upheld — the provisional credit is finalized.
	EventChargebackWon = "card.chargeback.won"
	// EventChargebackLost is delivered when the dispute is rejected — the
	// provisional credit is clawed back.
	EventChargebackLost = "card.chargeback.lost"
)

// ResolveChargebackUseCase is the admin entry point for closing out a
// dispute, one way or the other.
type ResolveChargebackUseCase struct {
	chargebacks port.ChargebackRepository
	dispatcher  WebhookDispatcher
	eventsURL   string
}

func NewResolveChargebackUseCase(chargebacks port.ChargebackRepository, dispatcher WebhookDispatcher, eventsURL string) *ResolveChargebackUseCase {
	return &ResolveChargebackUseCase{chargebacks: chargebacks, dispatcher: dispatcher, eventsURL: eventsURL}
}

func (uc *ResolveChargebackUseCase) Execute(ctx context.Context, chargebackID, outcome string) (domain.Chargeback, error) {
	if outcome != domain.ChargebackStatusWon && outcome != domain.ChargebackStatusLost {
		return domain.Chargeback{}, fmt.Errorf("outcome must be %q or %q, got %q", domain.ChargebackStatusWon, domain.ChargebackStatusLost, outcome)
	}

	cb, err := uc.chargebacks.GetByID(ctx, chargebackID)
	if err != nil {
		return domain.Chargeback{}, err
	}

	if cb == nil {
		return domain.Chargeback{}, fmt.Errorf("chargeback %q not found", chargebackID)
	}

	if cb.Status != domain.ChargebackStatusOpened {
		return domain.Chargeback{}, fmt.Errorf("chargeback cannot be resolved from status %q", cb.Status)
	}

	updated, err := uc.chargebacks.SetStatus(ctx, cb.ID, outcome)
	if err != nil {
		return domain.Chargeback{}, err
	}

	eventType := EventChargebackWon
	if outcome == domain.ChargebackStatusLost {
		eventType = EventChargebackLost
	}

	if _, err := uc.dispatcher.Enqueue(ctx, uc.eventsURL, eventType, ChargebackWebhookPayload{
		ChargebackID:    updated.ID,
		AuthorizationID: updated.AuthorizationID,
		Amount:          updated.Amount,
		Currency:        updated.Currency,
		Reason:          updated.Reason,
	}); err != nil {
		return domain.Chargeback{}, fmt.Errorf("schedule chargeback resolution webhook: %w", err)
	}

	return updated, nil
}
