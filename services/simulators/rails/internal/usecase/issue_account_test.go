package usecase

import (
	"context"
	"strconv"
	"testing"

	"github.com/iho/neobank/services/simulators/rails/internal/domain"
)

type fakeAccountRepository struct {
	byID  map[string]domain.Account
	byRef map[string]domain.Account
	seq   int
}

func newFakeAccountRepository() *fakeAccountRepository {
	return &fakeAccountRepository{byID: map[string]domain.Account{}, byRef: map[string]domain.Account{}}
}

func (r *fakeAccountRepository) Create(_ context.Context, externalRef, currency, iban string) (domain.Account, error) {
	r.seq++
	acct := domain.Account{
		ID:          "acct-" + strconv.Itoa(r.seq),
		ExternalRef: externalRef,
		Currency:    currency,
		IBAN:        iban,
	}
	r.byID[acct.ID] = acct
	r.byRef[externalRef+"|"+currency] = acct

	return acct, nil
}

func (r *fakeAccountRepository) GetByExternalRefAndCurrency(_ context.Context, externalRef, currency string) (*domain.Account, error) {
	acct, ok := r.byRef[externalRef+"|"+currency]
	if !ok {
		return nil, nil
	}

	return &acct, nil
}

func (r *fakeAccountRepository) GetByID(_ context.Context, id string) (*domain.Account, error) {
	acct, ok := r.byID[id]
	if !ok {
		return nil, nil
	}

	return &acct, nil
}

func TestIssueAccountCreatesOnFirstCall(t *testing.T) {
	repo := newFakeAccountRepository()
	uc := NewIssueAccountUseCase(repo, "DE", "SIM")

	acct, err := uc.Execute(context.Background(), IssueAccountInput{ExternalRef: "user-1", Currency: "usd"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acct.Currency != "USD" {
		t.Fatalf("expected currency normalized to USD, got %q", acct.Currency)
	}

	if acct.IBAN == "" {
		t.Fatal("expected a non-empty IBAN")
	}
}

func TestIssueAccountIsIdempotentPerRefAndCurrency(t *testing.T) {
	repo := newFakeAccountRepository()
	uc := NewIssueAccountUseCase(repo, "DE", "SIM")
	ctx := context.Background()

	first, err := uc.Execute(ctx, IssueAccountInput{ExternalRef: "user-1", Currency: "USD"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := uc.Execute(ctx, IssueAccountInput{ExternalRef: "user-1", Currency: "USD"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first.ID != second.ID || first.IBAN != second.IBAN {
		t.Fatalf("expected same account on repeat request, got %+v and %+v", first, second)
	}
}

func TestIssueAccountValidatesInput(t *testing.T) {
	repo := newFakeAccountRepository()
	uc := NewIssueAccountUseCase(repo, "DE", "SIM")

	if _, err := uc.Execute(context.Background(), IssueAccountInput{Currency: "USD"}); err == nil {
		t.Fatal("expected error for missing external_ref")
	}

	if _, err := uc.Execute(context.Background(), IssueAccountInput{ExternalRef: "user-1"}); err == nil {
		t.Fatal("expected error for missing currency")
	}
}
