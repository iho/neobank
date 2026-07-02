package usecase

import (
	"context"
	"time"

	"github.com/iho/neobank/pkg/pagination"
	"github.com/iho/neobank/services/user/internal/domain"
)

type WalletTransactionListResult struct {
	Transactions []domain.WalletTransaction
	NextCursor   string
}

type ListWalletTransactionsUseCase struct {
	repo WalletTransactionRepository
}

func NewListWalletTransactionsUseCase(repo WalletTransactionRepository) *ListWalletTransactionsUseCase {
	return &ListWalletTransactionsUseCase{repo: repo}
}

func (uc *ListWalletTransactionsUseCase) Execute(ctx context.Context, userID string, limit int, cursor string) (WalletTransactionListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	pageSize := limit + 1
	var cursorAt *time.Time
	var cursorID string
	if cursor != "" {
		decoded, err := pagination.Decode(cursor)
		if err != nil {
			return WalletTransactionListResult{}, err
		}
		at := decoded.CreatedAt
		cursorAt = &at
		cursorID = decoded.ID
	}
	rows, err := uc.repo.ListByUser(ctx, userID, pageSize, cursorAt, cursorID)
	if err != nil {
		return WalletTransactionListResult{}, err
	}
	items, next := pagination.Trim(rows, limit, func(t domain.WalletTransaction) pagination.Cursor {
		return pagination.Cursor{CreatedAt: t.CreatedAt, ID: t.ID}
	})
	return WalletTransactionListResult{Transactions: items, NextCursor: next}, nil
}