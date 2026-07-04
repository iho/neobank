package usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/iho/neobank/services/simulators/rails/internal/domain"
	"github.com/iho/neobank/services/simulators/rails/internal/port"
)

type IssueAccountInput struct {
	ExternalRef string
	Currency    string
}

// IssueAccountUseCase gets or creates the virtual IBAN for (external_ref,
// currency), so requesting it twice for the same wallet is a no-op.
type IssueAccountUseCase struct {
	accounts     port.AccountRepository
	ibanCountry  string
	ibanBankCode string
}

func NewIssueAccountUseCase(accounts port.AccountRepository, ibanCountry, ibanBankCode string) *IssueAccountUseCase {
	return &IssueAccountUseCase{accounts: accounts, ibanCountry: ibanCountry, ibanBankCode: ibanBankCode}
}

func (uc *IssueAccountUseCase) Execute(ctx context.Context, in IssueAccountInput) (domain.Account, error) {
	if in.ExternalRef == "" || in.Currency == "" {
		return domain.Account{}, fmt.Errorf("external_ref and currency are required")
	}

	currency := strings.ToUpper(in.Currency)

	existing, err := uc.accounts.GetByExternalRefAndCurrency(ctx, in.ExternalRef, currency)
	if err != nil {
		return domain.Account{}, err
	}

	if existing != nil {
		return *existing, nil
	}

	iban, err := generateIBAN(uc.ibanCountry, uc.ibanBankCode)
	if err != nil {
		return domain.Account{}, err
	}

	return uc.accounts.Create(ctx, in.ExternalRef, currency, iban)
}

// generateIBAN produces a plausible-looking but non-checksummed IBAN for
// simulation purposes: country + fixed "00" check digits + bank code +
// a random account number. It is not a valid IBAN under ISO 7064.
func generateIBAN(country, bankCode string) (string, error) {
	buf := make([]byte, 10)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate account number: %w", err)
	}

	acctNum := fmt.Sprintf("%x", buf)[:16]

	return fmt.Sprintf("%s00%s%s", strings.ToUpper(country), strings.ToUpper(bankCode), acctNum), nil
}
