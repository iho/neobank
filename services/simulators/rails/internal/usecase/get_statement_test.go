package usecase

import (
	"context"
	"testing"
)

func TestGetStatementRejectsInvalidDate(t *testing.T) {
	uc := NewGetStatementUseCase(&fakeInboundTransferRepository{})

	if _, err := uc.Execute(context.Background(), "not-a-date"); err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestGetStatementReturnsTransfersInRange(t *testing.T) {
	repo := &fakeInboundTransferRepository{}
	uc := NewGetStatementUseCase(repo)

	if _, err := repo.Create(context.Background(), "acct-1", "10.00", "USD", "Jane Doe", ""); err != nil {
		t.Fatalf("seed transfer: %v", err)
	}

	got, err := uc.Execute(context.Background(), "2026-07-04")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 transfer, got %d", len(got))
	}
}
