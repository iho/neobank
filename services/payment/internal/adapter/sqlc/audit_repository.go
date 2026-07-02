package sqlcrepo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuditRepository struct {
	q sqlc.Querier
}

func NewAuditRepository(q sqlc.Querier) *AuditRepository {
	return &AuditRepository{q: q}
}

func (r *AuditRepository) WithTx(tx pgx.Tx) audit.Recorder {
	return &AuditRepository{q: withTx(r.q, tx)}
}

func (r *AuditRepository) Record(ctx context.Context, e audit.Entry) error {
	e = audit.Resolve(ctx, e)
	metadata, err := json.Marshal(e.Metadata)
	if err != nil {
		return err
	}
	return r.q.InsertAuditLog(ctx, sqlc.InsertAuditLogParams{
		ID:            uuid.New(),
		EntityType:    e.EntityType,
		EntityID:      e.EntityID,
		Action:        e.Action,
		FromStatus:    textOrNil(e.FromStatus),
		ToStatus:      textOrNil(e.ToStatus),
		Actor:         e.Actor,
		CorrelationID: textOrNil(e.CorrelationID),
		Metadata:      metadata,
		CreatedAt:     pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	})
}
