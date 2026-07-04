package usecase

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/iho/neobank/services/simulators/rails/internal/domain"
)

type fakeInboundTransferRepository struct {
	created []domain.InboundTransfer
	seq     int
}

func (r *fakeInboundTransferRepository) Create(_ context.Context, accountID, amount, currency, senderName, reference string) (domain.InboundTransfer, error) {
	r.seq++
	t := domain.InboundTransfer{
		ID:         "transfer-" + strconv.Itoa(r.seq),
		AccountID:  accountID,
		Amount:     amount,
		Currency:   currency,
		SenderName: senderName,
		Reference:  reference,
		Status:     "received",
		CreatedAt:  time.Now().UTC(),
	}
	r.created = append(r.created, t)

	return t, nil
}

func (r *fakeInboundTransferRepository) ListInRange(_ context.Context, _, _ time.Time) ([]domain.InboundTransfer, error) {
	return r.created, nil
}

type fakeDispatcher struct {
	calls []struct {
		url       string
		eventType string
		payload   any
	}
}

func (d *fakeDispatcher) Enqueue(_ context.Context, url, eventType string, payload any) (string, error) {
	d.calls = append(d.calls, struct {
		url       string
		eventType string
		payload   any
	}{url, eventType, payload})

	return "delivery-1", nil
}

func TestInjectInboundTransferSchedulesWebhook(t *testing.T) {
	accounts := newFakeAccountRepository()
	ctx := context.Background()

	account, err := accounts.Create(ctx, "user-1", "USD", "DE00SIM123")
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}

	transfers := &fakeInboundTransferRepository{}
	dispatcher := &fakeDispatcher{}
	uc := NewInjectInboundTransferUseCase(accounts, transfers, dispatcher, "http://payment/webhooks/rails")

	transfer, err := uc.Execute(ctx, InjectInboundTransferInput{
		AccountID:  account.ID,
		Amount:     "100.00",
		Currency:   "USD",
		SenderName: "Jane Doe",
		Reference:  "rent",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(transfers.created) != 1 {
		t.Fatalf("expected 1 transfer created, got %d", len(transfers.created))
	}

	if len(dispatcher.calls) != 1 {
		t.Fatalf("expected 1 webhook scheduled, got %d", len(dispatcher.calls))
	}

	call := dispatcher.calls[0]
	if call.url != "http://payment/webhooks/rails" || call.eventType != EventTransferReceived {
		t.Fatalf("unexpected webhook target/event: %+v", call)
	}

	payload, ok := call.payload.(TransferReceivedPayload)
	if !ok {
		t.Fatalf("expected TransferReceivedPayload, got %T", call.payload)
	}

	if payload.TransferID != transfer.ID || payload.ExternalRef != "user-1" || payload.IBAN != "DE00SIM123" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestInjectInboundTransferRejectsUnknownAccount(t *testing.T) {
	accounts := newFakeAccountRepository()
	transfers := &fakeInboundTransferRepository{}
	dispatcher := &fakeDispatcher{}
	uc := NewInjectInboundTransferUseCase(accounts, transfers, dispatcher, "http://payment/webhooks/rails")

	_, err := uc.Execute(context.Background(), InjectInboundTransferInput{
		AccountID:  "missing",
		Amount:     "10.00",
		Currency:   "USD",
		SenderName: "Jane Doe",
	})
	if err == nil {
		t.Fatal("expected error for unknown account")
	}

	if len(dispatcher.calls) != 0 {
		t.Fatal("expected no webhook scheduled for a failed injection")
	}
}

func TestInjectInboundTransferValidatesInput(t *testing.T) {
	accounts := newFakeAccountRepository()
	transfers := &fakeInboundTransferRepository{}
	dispatcher := &fakeDispatcher{}
	uc := NewInjectInboundTransferUseCase(accounts, transfers, dispatcher, "http://payment/webhooks/rails")

	if _, err := uc.Execute(context.Background(), InjectInboundTransferInput{}); err == nil {
		t.Fatal("expected validation error for empty input")
	}
}
