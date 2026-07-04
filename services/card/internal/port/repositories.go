package port

import (
	"context"
	"time"

	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/jackc/pgx/v5"
)

type CardRepository interface {
	Create(ctx context.Context, c domain.Card) error
	GetByID(ctx context.Context, id string) (*domain.Card, error)
	GetByUserAndIdempotencyKey(ctx context.Context, userID, key string) (*domain.Card, error)
	// GetByProcessorRef resolves the cardproc simulator's card reference back
	// to our card, so the synchronous auth webhook can identify which card a
	// transaction is against.
	GetByProcessorRef(ctx context.Context, processorRef string) (*domain.Card, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Card, error)
	UpdateStatus(ctx context.Context, id, userID string, status domain.CardStatus) error
	UpdateControls(ctx context.Context, id, userID string, dailyLimit *string, onlineOnly *bool) (*domain.Card, error)
	MarkCancelled(ctx context.Context, id string) error
	WithTx(tx pgx.Tx) CardRepository
}

type AuthorizationRepository interface {
	Create(ctx context.Context, a domain.Authorization) error
	GetByID(ctx context.Context, id string) (*domain.Authorization, error)
	GetByCardAndIdempotencyKey(ctx context.Context, cardID, key string) (*domain.Authorization, error)
	SumTodayForCard(ctx context.Context, cardID string, dayStart time.Time) (string, error)
	ListByUser(ctx context.Context, userID string, limit int) ([]domain.Authorization, error)
	MarkHold(ctx context.Context, id, holdID string) error
	MarkFailed(ctx context.Context, id, reason string) error
	MarkCaptured(ctx context.Context, id, transferID string) error
	MarkVoided(ctx context.Context, id, reason string) error
	WithTx(tx pgx.Tx) AuthorizationRepository
}

// FraudDecisionRepository persists every fraud evaluation (allow or deny),
// not just the ones that block a transaction, so disputes and regulators can
// see what rule fired and why.
type FraudDecisionRepository interface {
	Record(ctx context.Context, entityType, entityID, userID, transactionType, amount, currency string, result fraud.Result) error
}

// DisputeRepository is keyed by the cardproc simulator's chargeback ID
// (chargeback_id UNIQUE), so a redelivered "opened" webhook is a no-op
// rather than a second provisional credit.
type DisputeRepository interface {
	Create(ctx context.Context, d domain.Dispute) (*domain.Dispute, error)
	GetByChargebackID(ctx context.Context, chargebackID string) (*domain.Dispute, error)
	MarkResolved(ctx context.Context, chargebackID, status, reversalTransferID string) (*domain.Dispute, error)
	WithTx(tx pgx.Tx) DisputeRepository
}
