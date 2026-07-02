package usecase

import (
	"context"
	"errors"
	"fmt"

	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/iho/neobank/pkg/ledgerclient"
	"github.com/jackc/pgx/v5"
)

type LedgerAccountReader interface {
	GetAccount(ctx context.Context, id string) (*goledgerv1.Account, error)
}

type GetWalletBalanceInput struct {
	UserID   string
	Currency string
}

type WalletBalance struct {
	WalletID          string
	LedgerAccountID   string
	Currency          string
	Balance           string
	EncumberedBalance string
	AvailableBalance  string
}

type GetWalletBalanceUseCase struct {
	wallets WalletRepository
	ledger  LedgerAccountReader
}

func NewGetWalletBalanceUseCase(wallets WalletRepository, ledger LedgerAccountReader) *GetWalletBalanceUseCase {
	return &GetWalletBalanceUseCase{wallets: wallets, ledger: ledger}
}

func (uc *GetWalletBalanceUseCase) Execute(ctx context.Context, in GetWalletBalanceInput) (WalletBalance, error) {
	if in.UserID == "" {
		return WalletBalance{}, fmt.Errorf("user_id is required")
	}
	if in.Currency == "" {
		in.Currency = "USD"
	}

	wallet, err := uc.wallets.GetByUserAndCurrency(ctx, in.UserID, in.Currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WalletBalance{}, fmt.Errorf("wallet not found")
		}
		return WalletBalance{}, err
	}
	if uc.ledger == nil {
		return WalletBalance{}, ledgerclient.ErrUnavailable
	}

	account, err := uc.ledger.GetAccount(ctx, wallet.LedgerAccountID)
	if err != nil {
		return WalletBalance{}, fmt.Errorf("ledger account: %w", err)
	}

	return WalletBalance{
		WalletID:          wallet.ID,
		LedgerAccountID:   wallet.LedgerAccountID,
		Currency:          wallet.Currency,
		Balance:           account.Balance,
		EncumberedBalance: account.EncumberedBalance,
		AvailableBalance:  account.Balance, // MVP: ledger exposes spendable balance in Balance field
	}, nil
}