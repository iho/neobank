package usecase

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/iho/neobank/services/simulators/rails/internal/domain"
)

type fakeOutboundPaymentRepository struct {
	payments map[string]domain.OutboundPayment
	seq      int
}

func (r *fakeOutboundPaymentRepository) Create(_ context.Context, accountID, amount, currency, counterpartyIBAN, reference string) (domain.OutboundPayment, error) {
	if r.payments == nil {
		r.payments = map[string]domain.OutboundPayment{}
	}

	r.seq++
	p := domain.OutboundPayment{
		ID:               "payment-" + strconv.Itoa(r.seq),
		AccountID:        accountID,
		Amount:           amount,
		Currency:         currency,
		CounterpartyIBAN: counterpartyIBAN,
		Reference:        reference,
		Status:           domain.OutboundPaymentStatusAccepted,
		CreatedAt:        time.Now().UTC(),
	}
	r.payments[p.ID] = p

	return p, nil
}

func (r *fakeOutboundPaymentRepository) GetByID(_ context.Context, id string) (*domain.OutboundPayment, error) {
	p, ok := r.payments[id]
	if !ok {
		return nil, nil
	}

	return &p, nil
}

func (r *fakeOutboundPaymentRepository) SetStatus(_ context.Context, id, status string) error {
	p, ok := r.payments[id]
	if !ok {
		return fmt.Errorf("payment %q not found", id)
	}

	p.Status = status
	r.payments[id] = p

	return nil
}

func (r *fakeOutboundPaymentRepository) ListInRange(_ context.Context, _, _ time.Time) ([]domain.OutboundPayment, error) {
	out := make([]domain.OutboundPayment, 0, len(r.payments))
	for _, p := range r.payments {
		out = append(out, p)
	}

	return out, nil
}

func TestInitiateOutboundPaymentSettlesByDefault(t *testing.T) {
	accounts := newFakeAccountRepository()
	ctx := context.Background()
	account, _ := accounts.Create(ctx, "user-1", "USD", "DE00SIM123")

	payments := &fakeOutboundPaymentRepository{}
	dispatcher := &fakeDispatcher{}
	uc := NewInitiateOutboundPaymentUseCase(accounts, payments, dispatcher, "http://payment/webhooks/rails")

	payment, err := uc.Execute(ctx, InitiateOutboundPaymentInput{
		AccountID: account.ID, Amount: "50.00", Currency: "USD", CounterpartyIBAN: "DE00OTHER",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payment.Status != domain.OutboundPaymentStatusAccepted {
		t.Fatalf("expected accepted, got %q", payment.Status)
	}

	if len(dispatcher.calls) != 1 || dispatcher.calls[0].eventType != EventPaymentSettled {
		t.Fatalf("expected 1 settled webhook scheduled, got %+v", dispatcher.calls)
	}
}

func TestInitiateOutboundPaymentReturnMagicValueSchedulesBoth(t *testing.T) {
	accounts := newFakeAccountRepository()
	ctx := context.Background()
	account, _ := accounts.Create(ctx, "user-1", "USD", "DE00SIM123")

	payments := &fakeOutboundPaymentRepository{}
	dispatcher := &fakeDispatcher{}
	uc := NewInitiateOutboundPaymentUseCase(accounts, payments, dispatcher, "http://payment/webhooks/rails")

	_, err := uc.Execute(ctx, InitiateOutboundPaymentInput{
		AccountID: account.ID, Amount: "50.00", Currency: "USD", CounterpartyIBAN: "DE00OTHER", Reference: "please RETURN this",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dispatcher.calls) != 2 {
		t.Fatalf("expected settled+returned webhooks scheduled, got %+v", dispatcher.calls)
	}

	if dispatcher.calls[0].eventType != EventPaymentSettled || dispatcher.calls[1].eventType != EventPaymentReturned {
		t.Fatalf("expected settled then returned, got %+v", dispatcher.calls)
	}
}

func TestInitiateOutboundPaymentAmountMagicValueFails(t *testing.T) {
	accounts := newFakeAccountRepository()
	ctx := context.Background()
	account, _ := accounts.Create(ctx, "user-1", "USD", "DE00SIM123")

	payments := &fakeOutboundPaymentRepository{}
	dispatcher := &fakeDispatcher{}
	uc := NewInitiateOutboundPaymentUseCase(accounts, payments, dispatcher, "http://payment/webhooks/rails")

	_, err := uc.Execute(ctx, InitiateOutboundPaymentInput{
		AccountID: account.ID, Amount: "50.99", Currency: "USD", CounterpartyIBAN: "DE00OTHER",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dispatcher.calls) != 1 || dispatcher.calls[0].eventType != EventPaymentFailed {
		t.Fatalf("expected 1 failed webhook scheduled, got %+v", dispatcher.calls)
	}
}

func TestInitiateOutboundPaymentRejectsUnknownAccount(t *testing.T) {
	accounts := newFakeAccountRepository()
	payments := &fakeOutboundPaymentRepository{}
	dispatcher := &fakeDispatcher{}
	uc := NewInitiateOutboundPaymentUseCase(accounts, payments, dispatcher, "http://payment/webhooks/rails")

	_, err := uc.Execute(context.Background(), InitiateOutboundPaymentInput{
		AccountID: "missing", Amount: "10.00", Currency: "USD", CounterpartyIBAN: "DE00OTHER",
	})
	if err == nil {
		t.Fatal("expected error for unknown account")
	}
}

func TestInitiateOutboundPaymentValidatesInput(t *testing.T) {
	accounts := newFakeAccountRepository()
	payments := &fakeOutboundPaymentRepository{}
	dispatcher := &fakeDispatcher{}
	uc := NewInitiateOutboundPaymentUseCase(accounts, payments, dispatcher, "http://payment/webhooks/rails")

	if _, err := uc.Execute(context.Background(), InitiateOutboundPaymentInput{}); err == nil {
		t.Fatal("expected validation error for empty input")
	}
}
