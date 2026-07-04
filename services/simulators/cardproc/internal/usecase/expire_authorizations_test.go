package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
)

func TestExpireAuthorizationsSweepsOldApproved(t *testing.T) {
	ctx := context.Background()
	txs := newFakeTransactionRepository()

	old, _ := txs.Create(ctx, "card-1", "10.00", "USD", "Coffee Shop", "5812")
	old, _ = txs.SetAuthResult(ctx, old.ID, domain.TransactionStatusApproved, "auth-old", "")
	old.CreatedAt = time.Now().UTC().Add(-2 * time.Hour)
	txs.txs[old.ID] = old

	fresh, _ := txs.Create(ctx, "card-1", "20.00", "USD", "Coffee Shop", "5812")
	fresh, _ = txs.SetAuthResult(ctx, fresh.ID, domain.TransactionStatusApproved, "auth-fresh", "")
	fresh.CreatedAt = time.Now().UTC()
	txs.txs[fresh.ID] = fresh

	dispatcher := &fakeDispatcher{}
	uc := NewExpireAuthorizationsUseCase(txs, dispatcher, "http://card/webhooks/cardproc/events")

	n, err := uc.Sweep(ctx, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != 1 {
		t.Fatalf("expected 1 expired transaction, got %d", n)
	}

	updatedOld, _ := txs.GetByID(ctx, old.ID)
	if updatedOld.Status != domain.TransactionStatusExpired {
		t.Fatalf("expected old transaction expired, got %q", updatedOld.Status)
	}

	updatedFresh, _ := txs.GetByID(ctx, fresh.ID)
	if updatedFresh.Status != domain.TransactionStatusApproved {
		t.Fatalf("expected fresh transaction untouched, got %q", updatedFresh.Status)
	}

	if len(dispatcher.calls) != 1 || dispatcher.calls[0].eventType != EventAuthExpired {
		t.Fatalf("unexpected dispatcher calls: %+v", dispatcher.calls)
	}

	payload, ok := dispatcher.calls[0].payload.(CardEventPayload)
	if !ok || payload.AuthorizationID != "auth-old" || payload.Reason != "expired" {
		t.Fatalf("unexpected payload: %+v", dispatcher.calls[0].payload)
	}
}

func TestExpireAuthorizationsNoOpWhenNothingDue(t *testing.T) {
	ctx := context.Background()
	txs := newFakeTransactionRepository()
	dispatcher := &fakeDispatcher{}
	uc := NewExpireAuthorizationsUseCase(txs, dispatcher, "http://card/webhooks/cardproc/events")

	n, err := uc.Sweep(ctx, time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != 0 || len(dispatcher.calls) != 0 {
		t.Fatalf("expected no-op sweep, got n=%d calls=%+v", n, dispatcher.calls)
	}
}
