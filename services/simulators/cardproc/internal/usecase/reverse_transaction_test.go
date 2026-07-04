package usecase

import (
	"context"
	"testing"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
)

func TestReverseTransactionSchedulesWebhook(t *testing.T) {
	ctx := context.Background()
	txs := newFakeTransactionRepository()
	tx, _ := txs.Create(ctx, "card-1", "10.00", "USD", "Coffee Shop", "5812")
	_, _ = txs.SetAuthResult(ctx, tx.ID, domain.TransactionStatusApproved, "auth-1", "")

	dispatcher := &fakeDispatcher{}
	uc := NewReverseTransactionUseCase(txs, dispatcher, "http://card/webhooks/cardproc/events")

	reversed, err := uc.Execute(ctx, tx.ID, "merchant_voided")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reversed.Status != domain.TransactionStatusReversed {
		t.Fatalf("expected reversed, got %q", reversed.Status)
	}

	if len(dispatcher.calls) != 1 || dispatcher.calls[0].eventType != EventAuthReversed {
		t.Fatalf("unexpected dispatcher calls: %+v", dispatcher.calls)
	}

	payload, ok := dispatcher.calls[0].payload.(CardEventPayload)
	if !ok || payload.AuthorizationID != "auth-1" || payload.Reason != "merchant_voided" {
		t.Fatalf("unexpected payload: %+v", dispatcher.calls[0].payload)
	}
}

func TestReverseTransactionRejectsNonApproved(t *testing.T) {
	ctx := context.Background()
	txs := newFakeTransactionRepository()
	tx, _ := txs.Create(ctx, "card-1", "10.00", "USD", "Coffee Shop", "5812")

	dispatcher := &fakeDispatcher{}
	uc := NewReverseTransactionUseCase(txs, dispatcher, "http://card/webhooks/cardproc/events")

	if _, err := uc.Execute(ctx, tx.ID, "reason"); err == nil {
		t.Fatal("expected error reversing a non-approved transaction")
	}
}
