package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/iho/neobank/services/simulators/fx/internal/domain"
)

func TestExecuteQuoteMarksExecuted(t *testing.T) {
	ctx := context.Background()
	repo := newFakeQuoteRepository()
	quote, _ := repo.Create(ctx, domain.Quote{
		FromCurrency: "EUR", ToCurrency: "USD", Amount: "100.00", ConvertedAmount: "107.46",
		Rate: "1.0746", Status: domain.QuoteStatusPending, ExpiresAt: time.Now().Add(time.Minute),
	})

	uc := NewExecuteQuoteUseCase(repo)

	executed, err := uc.Execute(ctx, quote.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if executed.Status != domain.QuoteStatusExecuted {
		t.Fatalf("expected executed, got %q", executed.Status)
	}
}

func TestExecuteQuoteIsIdempotent(t *testing.T) {
	ctx := context.Background()
	repo := newFakeQuoteRepository()
	quote, _ := repo.Create(ctx, domain.Quote{
		FromCurrency: "EUR", ToCurrency: "USD", Amount: "100.00", ConvertedAmount: "107.46",
		Rate: "1.0746", Status: domain.QuoteStatusPending, ExpiresAt: time.Now().Add(time.Minute),
	})

	uc := NewExecuteQuoteUseCase(repo)

	if _, err := uc.Execute(ctx, quote.ID); err != nil {
		t.Fatalf("unexpected error on first execute: %v", err)
	}

	second, err := uc.Execute(ctx, quote.ID)
	if err != nil {
		t.Fatalf("expected re-executing an already-executed quote to be a no-op, got error: %v", err)
	}

	if second.Status != domain.QuoteStatusExecuted {
		t.Fatalf("expected executed, got %q", second.Status)
	}
}

func TestExecuteQuoteRejectsExpired(t *testing.T) {
	ctx := context.Background()
	repo := newFakeQuoteRepository()
	quote, _ := repo.Create(ctx, domain.Quote{
		FromCurrency: "EUR", ToCurrency: "USD", Amount: "100.00", ConvertedAmount: "107.46",
		Rate: "1.0746", Status: domain.QuoteStatusPending, ExpiresAt: time.Now().Add(-time.Second),
	})

	uc := NewExecuteQuoteUseCase(repo)

	if _, err := uc.Execute(ctx, quote.ID); err == nil {
		t.Fatal("expected error for expired quote")
	}
}

func TestExecuteQuoteRejectsUnknown(t *testing.T) {
	repo := newFakeQuoteRepository()
	uc := NewExecuteQuoteUseCase(repo)

	if _, err := uc.Execute(context.Background(), "missing"); err == nil {
		t.Fatal("expected error for unknown quote")
	}
}
