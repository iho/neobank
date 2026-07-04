package usecase

import (
	"context"
	"testing"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
)

func TestResolveChargebackLost(t *testing.T) {
	ctx := context.Background()
	chargebacks := newFakeChargebackRepository()
	cb, _ := chargebacks.Create(ctx, "tx-1", "auth-1", "25.00", "USD", "fraud")

	dispatcher := &fakeDispatcher{}
	uc := NewResolveChargebackUseCase(chargebacks, dispatcher, "http://card/webhooks/cardproc/events")

	resolved, err := uc.Execute(ctx, cb.ID, domain.ChargebackStatusLost)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Status != domain.ChargebackStatusLost {
		t.Fatalf("expected lost, got %q", resolved.Status)
	}

	if len(dispatcher.calls) != 1 || dispatcher.calls[0].eventType != EventChargebackLost {
		t.Fatalf("unexpected dispatcher calls: %+v", dispatcher.calls)
	}
}

func TestResolveChargebackWon(t *testing.T) {
	ctx := context.Background()
	chargebacks := newFakeChargebackRepository()
	cb, _ := chargebacks.Create(ctx, "tx-1", "auth-1", "25.00", "USD", "fraud")

	dispatcher := &fakeDispatcher{}
	uc := NewResolveChargebackUseCase(chargebacks, dispatcher, "http://card/webhooks/cardproc/events")

	resolved, err := uc.Execute(ctx, cb.ID, domain.ChargebackStatusWon)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Status != domain.ChargebackStatusWon {
		t.Fatalf("expected won, got %q", resolved.Status)
	}

	if len(dispatcher.calls) != 1 || dispatcher.calls[0].eventType != EventChargebackWon {
		t.Fatalf("unexpected dispatcher calls: %+v", dispatcher.calls)
	}
}

func TestResolveChargebackRejectsAlreadyResolved(t *testing.T) {
	ctx := context.Background()
	chargebacks := newFakeChargebackRepository()
	cb, _ := chargebacks.Create(ctx, "tx-1", "auth-1", "25.00", "USD", "fraud")
	_, _ = chargebacks.SetStatus(ctx, cb.ID, domain.ChargebackStatusWon)

	dispatcher := &fakeDispatcher{}
	uc := NewResolveChargebackUseCase(chargebacks, dispatcher, "http://card/webhooks/cardproc/events")

	if _, err := uc.Execute(ctx, cb.ID, domain.ChargebackStatusLost); err == nil {
		t.Fatal("expected error resolving an already-resolved chargeback")
	}
}

func TestResolveChargebackRejectsInvalidOutcome(t *testing.T) {
	ctx := context.Background()
	chargebacks := newFakeChargebackRepository()
	cb, _ := chargebacks.Create(ctx, "tx-1", "auth-1", "25.00", "USD", "fraud")

	dispatcher := &fakeDispatcher{}
	uc := NewResolveChargebackUseCase(chargebacks, dispatcher, "http://card/webhooks/cardproc/events")

	if _, err := uc.Execute(ctx, cb.ID, "maybe"); err == nil {
		t.Fatal("expected error for invalid outcome")
	}
}
