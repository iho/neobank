package usecase

import (
	"context"

	"github.com/iho/neobank/services/user/internal/domain"
)

type ListWalletTransactionsUseCase struct {
	repo WalletTransactionRepository
}

func NewListWalletTransactionsUseCase(repo WalletTransactionRepository) *ListWalletTransactionsUseCase {
	return &ListWalletTransactionsUseCase{repo: repo}
}

func (uc *ListWalletTransactionsUseCase) Execute(ctx context.Context, userID string, limit int) ([]domain.WalletTransaction, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.repo.ListByUser(ctx, userID, limit)
}