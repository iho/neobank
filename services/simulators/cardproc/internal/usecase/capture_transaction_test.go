package usecase

import (
	"context"
	"testing"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
)

func TestCaptureTransactionSchedulesWebhook(t *testing.T) {
	ctx := context.Background()
	txs := newFakeTransactionRepository()
	tx, _ := txs.Create(ctx, "card-1", "10.00", "USD", "Coffee Shop", "5812")
	_, _ = txs.SetAuthResult(ctx, tx.ID, domain.TransactionStatusApproved, "auth-1", "")

	dispatcher := &fakeDispatcher{}
	uc := NewCaptureTransactionUseCase(txs, dispatcher, "http://card/webhooks/cardproc/events")

	captured, err := uc.Execute(ctx, tx.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured.Status != domain.TransactionStatusCaptured {
		t.Fatalf("expected captured, got %q", captured.Status)
	}

	if len(dispatcher.calls) != 1 || dispatcher.calls[0].eventType != EventCaptured {
		t.Fatalf("unexpected dispatcher calls: %+v", dispatcher.calls)
	}
}

func TestCaptureTransactionRejectsNonApproved(t *testing.T) {
	ctx := context.Background()
	txs := newFakeTransactionRepository()
	tx, _ := txs.Create(ctx, "card-1", "10.00", "USD", "Coffee Shop", "5812")

	dispatcher := &fakeDispatcher{}
	uc := NewCaptureTransactionUseCase(txs, dispatcher, "http://card/webhooks/cardproc/events")

	if _, err := uc.Execute(ctx, tx.ID); err == nil {
		t.Fatal("expected error capturing a non-approved transaction")
	}
}

func TestCaptureTransactionRejectsUnknown(t *testing.T) {
	txs := newFakeTransactionRepository()
	dispatcher := &fakeDispatcher{}
	uc := NewCaptureTransactionUseCase(txs, dispatcher, "http://card/webhooks/cardproc/events")

	if _, err := uc.Execute(context.Background(), "missing"); err == nil {
		t.Fatal("expected error for unknown transaction")
	}
}
