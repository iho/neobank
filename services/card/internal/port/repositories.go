package port

import (
	"context"

	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/jackc/pgx/v5"
)

type CardRepository interface {
	Create(ctx context.Context, c domain.Card) error
	GetByID(ctx context.Context, id string) (*domain.Card, error)
	GetByUserAndIdempotencyKey(ctx context.Context, userID, key string) (*domain.Card, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Card, error)
	UpdateStatus(ctx context.Context, id, userID string, status domain.CardStatus) error
	MarkCancelled(ctx context.Context, id string) error
	WithTx(tx pgx.Tx) CardRepository
}

type AuthorizationRepository interface {
	Create(ctx context.Context, a domain.Authorization) error
	GetByID(ctx context.Context, id string) (*domain.Authorization, error)
	GetByCardAndIdempotencyKey(ctx context.Context, cardID, key string) (*domain.Authorization, error)
	ListByUser(ctx context.Context, userID string, limit int) ([]domain.Authorization, error)
	MarkHold(ctx context.Context, id, holdID string) error
	MarkFailed(ctx context.Context, id, reason string) error
	MarkCaptured(ctx context.Context, id, transferID string) error
	WithTx(tx pgx.Tx) AuthorizationRepository
}

// FraudDecisionRepository persists every fraud evaluation (allow or deny),
// not just the ones that block a transaction, so disputes and regulators can
// see what rule fired and why.
type FraudDecisionRepository interface {
	Record(ctx context.Context, entityType, entityID, userID, transactionType, amount, currency string, result fraud.Result) error
}