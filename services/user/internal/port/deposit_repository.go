package port

import (
	"context"

	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
)

type DepositRepository interface {
	GetByUserAndIdempotencyKey(ctx context.Context, userID, key string) (*domain.Deposit, error)
	Insert(ctx context.Context, deposit domain.Deposit) error
	WithTx(tx pgx.Tx) DepositRepository
}