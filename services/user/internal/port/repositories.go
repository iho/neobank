package port

import (
	"context"

	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet domain.Wallet) error
	DeleteByID(ctx context.Context, walletID string) error
	GetByUserAndCurrency(ctx context.Context, userID, currency string) (*domain.Wallet, error)
	WithTx(tx pgx.Tx) WalletRepository
}

type KYCRepository interface {
	UpsertProfile(ctx context.Context, userID, fullName, dob, country string) error
	CreateCase(ctx context.Context, id, userID, status string) (domain.KYCCase, error)
	GetLatestByUser(ctx context.Context, userID string) (*domain.KYCCase, error)
	ApproveCase(ctx context.Context, caseID string) error
	WithTx(tx pgx.Tx) KYCRepository
}