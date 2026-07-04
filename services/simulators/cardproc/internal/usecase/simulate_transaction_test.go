package usecase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iho/neobank/services/simulators/cardproc/internal/adapter/cardclient"
	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
)

func newTestCardClient(t *testing.T, decision, authorizationID string) (*cardclient.Client, func()) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cardclient.AuthorizeResult{
			Decision:        decision,
			AuthorizationID: authorizationID,
			ReasonCode:      "test_reason",
		})
	}))

	client := cardclient.New(cardclient.Config{AuthorizeURL: srv.URL, Secret: []byte("test-secret")})

	return client, srv.Close
}

func TestSimulateTransactionApprovedWithoutCapture(t *testing.T) {
	cards := newFakeCardRepository()
	ctx := context.Background()
	card, _ := cards.Create(ctx, "user-1", "Jane Doe", "tok", "4242", 12, 2030)

	txs := newFakeTransactionRepository()
	dispatcher := &fakeDispatcher{}
	client, closeSrv := newTestCardClient(t, "approved", "auth-1")
	defer closeSrv()

	uc := NewSimulateTransactionUseCase(cards, txs, client, dispatcher, "http://card/webhooks/cardproc/events")

	tx, err := uc.Execute(ctx, SimulateTransactionInput{
		CardRef: card.ID, Amount: "10.00", Currency: "USD", MerchantName: "Coffee Shop", MCC: "5812",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tx.Status != domain.TransactionStatusApproved {
		t.Fatalf("expected approved, got %q", tx.Status)
	}

	if tx.AuthorizationID != "auth-1" {
		t.Fatalf("expected authorization_id auth-1, got %q", tx.AuthorizationID)
	}

	if len(dispatcher.calls) != 0 {
		t.Fatalf("expected no webhook scheduled without capture, got %d", len(dispatcher.calls))
	}
}

func TestSimulateTransactionApprovedWithCaptureSchedulesWebhook(t *testing.T) {
	cards := newFakeCardRepository()
	ctx := context.Background()
	card, _ := cards.Create(ctx, "user-1", "Jane Doe", "tok", "4242", 12, 2030)

	txs := newFakeTransactionRepository()
	dispatcher := &fakeDispatcher{}
	client, closeSrv := newTestCardClient(t, "approved", "auth-2")
	defer closeSrv()

	uc := NewSimulateTransactionUseCase(cards, txs, client, dispatcher, "http://card/webhooks/cardproc/events")

	tx, err := uc.Execute(ctx, SimulateTransactionInput{
		CardRef: card.ID, Amount: "10.00", Currency: "USD", MerchantName: "Coffee Shop", MCC: "5812", Capture: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tx.Status != domain.TransactionStatusCaptured {
		t.Fatalf("expected captured, got %q", tx.Status)
	}

	if len(dispatcher.calls) != 1 {
		t.Fatalf("expected 1 webhook scheduled, got %d", len(dispatcher.calls))
	}

	call := dispatcher.calls[0]
	if call.eventType != EventCaptured {
		t.Fatalf("expected EventCaptured, got %q", call.eventType)
	}

	payload, ok := call.payload.(CardEventPayload)
	if !ok || payload.AuthorizationID != "auth-2" {
		t.Fatalf("unexpected payload: %+v", call.payload)
	}
}

func TestSimulateTransactionDeclined(t *testing.T) {
	cards := newFakeCardRepository()
	ctx := context.Background()
	card, _ := cards.Create(ctx, "user-1", "Jane Doe", "tok", "4242", 12, 2030)

	txs := newFakeTransactionRepository()
	dispatcher := &fakeDispatcher{}
	client, closeSrv := newTestCardClient(t, "declined", "")
	defer closeSrv()

	uc := NewSimulateTransactionUseCase(cards, txs, client, dispatcher, "http://card/webhooks/cardproc/events")

	tx, err := uc.Execute(ctx, SimulateTransactionInput{
		CardRef: card.ID, Amount: "10.00", Currency: "USD", MerchantName: "Coffee Shop", Capture: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tx.Status != domain.TransactionStatusDeclined {
		t.Fatalf("expected declined, got %q", tx.Status)
	}

	if len(dispatcher.calls) != 0 {
		t.Fatalf("expected no webhook scheduled for a decline, got %d", len(dispatcher.calls))
	}
}

func TestSimulateTransactionUnknownCard(t *testing.T) {
	cards := newFakeCardRepository()
	txs := newFakeTransactionRepository()
	dispatcher := &fakeDispatcher{}
	client, closeSrv := newTestCardClient(t, "approved", "auth-3")
	defer closeSrv()

	uc := NewSimulateTransactionUseCase(cards, txs, client, dispatcher, "http://card/webhooks/cardproc/events")

	if _, err := uc.Execute(context.Background(), SimulateTransactionInput{
		CardRef: "missing", Amount: "10.00", Currency: "USD",
	}); err == nil {
		t.Fatal("expected error for unknown card")
	}
}

func TestSimulateTransactionValidatesInput(t *testing.T) {
	cards := newFakeCardRepository()
	txs := newFakeTransactionRepository()
	dispatcher := &fakeDispatcher{}
	client, closeSrv := newTestCardClient(t, "approved", "auth-4")
	defer closeSrv()

	uc := NewSimulateTransactionUseCase(cards, txs, client, dispatcher, "http://card/webhooks/cardproc/events")

	if _, err := uc.Execute(context.Background(), SimulateTransactionInput{}); err == nil {
		t.Fatal("expected validation error for empty input")
	}
}
