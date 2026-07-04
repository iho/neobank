package usecase

import (
	"context"
	"testing"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
)

func TestOpenChargebackOnCapturedTransaction(t *testing.T) {
	ctx := context.Background()
	txs := newFakeTransactionRepository()
	tx, _ := txs.Create(ctx, "card-1", "25.00", "USD", "Coffee Shop", "5812")
	tx, _ = txs.SetAuthResult(ctx, tx.ID, domain.TransactionStatusApproved, "auth-1", "")
	_ = txs.MarkCaptured(ctx, tx.ID)

	chargebacks := newFakeChargebackRepository()
	dispatcher := &fakeDispatcher{}
	uc := NewOpenChargebackUseCase(txs, chargebacks, dispatcher, "http://card/webhooks/cardproc/events")

	cb, err := uc.Execute(ctx, tx.ID, "fraud")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cb.Status != domain.ChargebackStatusOpened || cb.AuthorizationID != "auth-1" || cb.Amount != "25.00" {
		t.Fatalf("unexpected chargeback: %+v", cb)
	}

	if len(dispatcher.calls) != 1 || dispatcher.calls[0].eventType != EventChargebackOpened {
		t.Fatalf("unexpected dispatcher calls: %+v", dispatcher.calls)
	}

	payload, ok := dispatcher.calls[0].payload.(ChargebackWebhookPayload)
	if !ok || payload.ChargebackID != cb.ID || payload.Reason != "fraud" {
		t.Fatalf("unexpected payload: %+v", dispatcher.calls[0].payload)
	}
}

func TestOpenChargebackRejectsNonCaptured(t *testing.T) {
	ctx := context.Background()
	txs := newFakeTransactionRepository()
	tx, _ := txs.Create(ctx, "card-1", "25.00", "USD", "Coffee Shop", "5812")
	tx, _ = txs.SetAuthResult(ctx, tx.ID, domain.TransactionStatusApproved, "auth-1", "")

	chargebacks := newFakeChargebackRepository()
	dispatcher := &fakeDispatcher{}
	uc := NewOpenChargebackUseCase(txs, chargebacks, dispatcher, "http://card/webhooks/cardproc/events")

	if _, err := uc.Execute(ctx, tx.ID, "fraud"); err == nil {
		t.Fatal("expected error charging back a non-captured transaction")
	}
}
