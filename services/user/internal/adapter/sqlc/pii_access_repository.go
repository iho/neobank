package sqlcrepo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type PIIAccessRepository struct {
	q sqlc.Querier
}

func NewPIIAccessRepository(q sqlc.Querier) *PIIAccessRepository {
	return &PIIAccessRepository{q: q}
}

func (r *PIIAccessRepository) RecordPIIAccess(ctx context.Context, e audit.PIIAccessEntry) error {
	e = audit.ResolvePIIAccess(ctx, e)
	metadata, err := json.Marshal(e.Metadata)
	if err != nil {
		return err
	}
	subjectID, err := uuid.Parse(e.SubjectUserID)
	if err != nil {
		return err
	}
	return r.q.InsertPIIAccessLog(ctx, sqlc.InsertPIIAccessLogParams{
		ID:            uuid.New(),
		SubjectUserID: subjectID,
		Resource:      e.Resource,
		Actor:         e.Actor,
		CorrelationID: textOrNil(e.CorrelationID),
		Metadata:      metadata,
		CreatedAt:     pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	})
}