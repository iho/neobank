package usecase

import (
	"context"
	"testing"
)

func TestGetStatementRejectsInvalidDate(t *testing.T) {
	uc := NewGetStatementUseCase(&fakeInboundTransferRepository{}, &fakeOutboundPaymentRepository{})

	if _, err := uc.Execute(context.Background(), "not-a-date"); err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestGetStatementReturnsTransfersAndPaymentsInRange(t *testing.T) {
	transfers := &fakeInboundTransferRepository{}
	payments := &fakeOutboundPaymentRepository{}
	uc := NewGetStatementUseCase(transfers, payments)

	if _, err := transfers.Create(context.Background(), "acct-1", "10.00", "USD", "Jane Doe", ""); err != nil {
		t.Fatalf("seed transfer: %v", err)
	}

	if _, err := payments.Create(context.Background(), "acct-1", "5.00", "USD", "DE00OTHER", ""); err != nil {
		t.Fatalf("seed payment: %v", err)
	}

	got, err := uc.Execute(context.Background(), "2026-07-04")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got.InboundTransfers) != 1 {
		t.Fatalf("expected 1 inbound transfer, got %d", len(got.InboundTransfers))
	}

	if len(got.OutboundPayments) != 1 {
		t.Fatalf("expected 1 outbound payment, got %d", len(got.OutboundPayments))
	}
}
