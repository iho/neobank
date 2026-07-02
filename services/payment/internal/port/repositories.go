package port

import (
	"context"

	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/jackc/pgx/v5"
)

type TransferRepository interface {
	Create(ctx context.Context, t domain.Transfer) error
	GetBySenderAndIdempotencyKey(ctx context.Context, senderUserID, key string) (*domain.Transfer, error)
	GetByID(ctx context.Context, id string) (*domain.Transfer, error)
	MarkCompleted(ctx context.Context, id, ledgerTransferID string) error
	MarkFailed(ctx context.Context, id, reason string) error
	ListByUser(ctx context.Context, userID string, limit int) ([]domain.Transfer, error)
	WithTx(tx pgx.Tx) TransferRepository
}

// FraudDecisionRepository persists every fraud evaluation (allow or deny),
// not just the ones that block a transaction, so disputes and regulators can
// see what rule fired and why.
type FraudDecisionRepository interface {
	Record(ctx context.Context, entityType, entityID, userID, transactionType, amount, currency string, result fraud.Result) error
}