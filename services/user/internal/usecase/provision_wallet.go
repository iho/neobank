package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet domain.Wallet) error
	GetByUserAndCurrency(ctx context.Context, userID, currency string) (*domain.Wallet, error)
}

type LedgerClient interface {
	CreateAccount(ctx context.Context, in ledgerclient.CreateAccountInput) (*goledgerv1.Account, error)
}

type ProvisionWalletInput struct {
	UserID   string
	Currency string
}

type ProvisionWalletOutput struct {
	WalletID        string
	LedgerAccountID string
}

type ProvisionWalletUseCase struct {
	wallets WalletRepository
	ledger  LedgerClient
}

func NewProvisionWalletUseCase(wallets WalletRepository, ledger LedgerClient) *ProvisionWalletUseCase {
	return &ProvisionWalletUseCase{wallets: wallets, ledger: ledger}
}

func (uc *ProvisionWalletUseCase) Execute(ctx context.Context, in ProvisionWalletInput) (ProvisionWalletOutput, error) {
	if in.UserID == "" {
		return ProvisionWalletOutput{}, fmt.Errorf("user_id is required")
	}
	if in.Currency == "" {
		in.Currency = "USD"
	}

	existing, err := uc.wallets.GetByUserAndCurrency(ctx, in.UserID, in.Currency)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ProvisionWalletOutput{}, err
	}
	if existing != nil {
		return ProvisionWalletOutput{
			WalletID:        existing.ID,
			LedgerAccountID: existing.LedgerAccountID,
		}, nil
	}

	account, err := uc.ledger.CreateAccount(ctx, ledgerclient.CreateAccountInput{
		Name:                 fmt.Sprintf("CUSTOMER:%s:%s", in.UserID, in.Currency),
		Currency:             in.Currency,
		AllowNegativeBalance: false,
		AllowPositiveBalance: true,
	})
	if err != nil {
		return ProvisionWalletOutput{}, fmt.Errorf("create ledger account: %w", err)
	}

	walletID := uuid.NewString()
	wallet := domain.Wallet{
		ID:              walletID,
		UserID:          in.UserID,
		Currency:        in.Currency,
		LedgerAccountID: account.Id,
		Status:          "active",
	}
	if err := uc.wallets.Create(ctx, wallet); err != nil {
		return ProvisionWalletOutput{}, err
	}

	return ProvisionWalletOutput{
		WalletID:        walletID,
		LedgerAccountID: account.Id,
	}, nil
}