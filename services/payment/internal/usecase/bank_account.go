package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/iho/neobank/services/payment/internal/adapter/railsclient"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/port"
)

type GetOrCreateBankAccountInput struct {
	UserID   string
	Currency string
}

// GetOrCreateBankAccountUseCase returns the user's virtual IBAN for topping
// up by bank transfer, minting one via the rails simulator on first request
// and caching the mapping locally thereafter.
type GetOrCreateBankAccountUseCase struct {
	bankAccounts port.BankAccountRepository
	rails        *railsclient.Client
}

func NewGetOrCreateBankAccountUseCase(bankAccounts port.BankAccountRepository, rails *railsclient.Client) *GetOrCreateBankAccountUseCase {
	return &GetOrCreateBankAccountUseCase{bankAccounts: bankAccounts, rails: rails}
}

func (uc *GetOrCreateBankAccountUseCase) Execute(ctx context.Context, in GetOrCreateBankAccountInput) (domain.BankAccount, error) {
	if in.UserID == "" {
		return domain.BankAccount{}, fmt.Errorf("user_id is required")
	}

	currency := strings.ToUpper(in.Currency)
	if currency == "" {
		currency = "USD"
	}

	existing, err := uc.bankAccounts.GetByUserAndCurrency(ctx, in.UserID, currency)
	if err != nil {
		return domain.BankAccount{}, err
	}

	if existing != nil {
		return *existing, nil
	}

	if uc.rails == nil {
		return domain.BankAccount{}, fmt.Errorf("rails simulator unavailable")
	}

	railsAccount, err := uc.rails.CreateAccount(ctx, in.UserID, currency)
	if err != nil {
		return domain.BankAccount{}, fmt.Errorf("issue virtual account: %w", err)
	}

	return uc.bankAccounts.Create(ctx, in.UserID, currency, railsAccount.ID, railsAccount.IBAN)
}
