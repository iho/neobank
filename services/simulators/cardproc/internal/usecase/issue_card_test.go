package usecase

import (
	"context"
	"testing"
)

func TestIssueCardCreatesCard(t *testing.T) {
	repo := newFakeCardRepository()
	uc := NewIssueCardUseCase(repo)

	card, err := uc.Execute(context.Background(), IssueCardInput{ExternalRef: "user-1", CardholderName: "Jane Doe"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if card.ExternalRef != "user-1" || card.CardholderName != "Jane Doe" {
		t.Fatalf("unexpected card: %+v", card)
	}

	if len(card.LastFour) != 4 {
		t.Fatalf("expected 4-digit last four, got %q", card.LastFour)
	}

	if card.PANToken == "" {
		t.Fatal("expected a non-empty PAN token")
	}
}

func TestIssueCardRequiresExternalRef(t *testing.T) {
	repo := newFakeCardRepository()
	uc := NewIssueCardUseCase(repo)

	if _, err := uc.Execute(context.Background(), IssueCardInput{CardholderName: "Jane Doe"}); err == nil {
		t.Fatal("expected error for missing external_ref")
	}
}
