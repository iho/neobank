package usecase

import (
	"context"
	"testing"

	"github.com/iho/neobank/services/simulators/fx/internal/domain"
)

func TestGetQuotePricesConversion(t *testing.T) {
	repo := newFakeQuoteRepository()
	uc := NewGetQuoteUseCase(repo)

	quote, err := uc.Execute(context.Background(), GetQuoteInput{
		FromCurrency: "EUR", ToCurrency: "USD", Amount: "100.00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if quote.Status != domain.QuoteStatusPending {
		t.Fatalf("expected pending, got %q", quote.Status)
	}

	if quote.SpreadBps != DefaultSpreadBps {
		t.Fatalf("expected spread %d, got %d", DefaultSpreadBps, quote.SpreadBps)
	}

	if quote.ExpiresAt.Before(quote.CreatedAt) {
		t.Fatal("expected expires_at after created_at")
	}

	if quote.ConvertedAmount == "" || quote.Rate == "" {
		t.Fatalf("expected populated converted_amount/rate, got %+v", quote)
	}
}

func TestGetQuoteRejectsUnsupportedPair(t *testing.T) {
	repo := newFakeQuoteRepository()
	uc := NewGetQuoteUseCase(repo)

	if _, err := uc.Execute(context.Background(), GetQuoteInput{
		FromCurrency: "USD", ToCurrency: "JPY", Amount: "100.00",
	}); err == nil {
		t.Fatal("expected error for unsupported pair")
	}
}

func TestGetQuoteValidatesInput(t *testing.T) {
	repo := newFakeQuoteRepository()
	uc := NewGetQuoteUseCase(repo)

	cases := []GetQuoteInput{
		{ToCurrency: "USD", Amount: "1.00"},
		{FromCurrency: "EUR", Amount: "1.00"},
		{FromCurrency: "EUR", ToCurrency: "USD"},
		{FromCurrency: "EUR", ToCurrency: "USD", Amount: "0"},
		{FromCurrency: "EUR", ToCurrency: "USD", Amount: "-5"},
		{FromCurrency: "EUR", ToCurrency: "USD", Amount: "not-a-number"},
	}

	for _, in := range cases {
		if _, err := uc.Execute(context.Background(), in); err == nil {
			t.Fatalf("expected validation error for input %+v", in)
		}
	}
}
