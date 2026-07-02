package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/services/user/internal/port"
)

type ListWalletsUseCase struct {
	wallets port.WalletRepository
	ledger  LedgerAccountReader
}

func NewListWalletsUseCase(wallets port.WalletRepository, ledger LedgerAccountReader) *ListWalletsUseCase {
	return &ListWalletsUseCase{wallets: wallets, ledger: ledger}
}

func (uc *ListWalletsUseCase) Execute(ctx context.Context, userID string) ([]WalletBalance, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	wallets, err := uc.wallets.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]WalletBalance, 0, len(wallets))
	for _, wallet := range wallets {
		balance := WalletBalance{
			WalletID:        wallet.ID,
			LedgerAccountID: wallet.LedgerAccountID,
			Currency:        wallet.Currency,
		}
		if uc.ledger != nil {
			account, err := uc.ledger.GetAccount(ctx, wallet.LedgerAccountID)
			if err == nil {
				balance.Balance = account.Balance
				balance.EncumberedBalance = account.EncumberedBalance
				balance.AvailableBalance = account.Balance
			}
		}
		out = append(out, balance)
	}
	return out, nil
}