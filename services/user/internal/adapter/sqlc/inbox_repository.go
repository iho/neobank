package sqlcrepo

import (
	"context"
	"time"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type ConsumerInboxRepository struct {
	q sqlc.Querier
}

func NewConsumerInboxRepository(q sqlc.Querier) *ConsumerInboxRepository {
	return &ConsumerInboxRepository{q: q}
}

func (r *ConsumerInboxRepository) Exists(ctx context.Context, eventID string) (bool, error) {
	id, err := pgutil.ParseUUID(eventID)
	if err != nil {
		return false, err
	}
	return r.q.ConsumerInboxExists(ctx, id)
}

func (r *ConsumerInboxRepository) Record(ctx context.Context, eventID, eventType string) error {
	id, err := pgutil.ParseUUID(eventID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return r.q.InsertConsumerInbox(ctx, sqlc.InsertConsumerInboxParams{
		EventID:     id,
		EventType:   eventType,
		ProcessedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
}