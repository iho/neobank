package sqlcrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/iho/neobank/services/payment/internal/port"
	"github.com/jackc/pgx/v5/pgtype"
)

type ScreeningRepository struct {
	q sqlc.Querier
}

func NewScreeningRepository(q sqlc.Querier) *ScreeningRepository {
	return &ScreeningRepository{q: q}
}

func (r *ScreeningRepository) Record(ctx context.Context, check port.ScreeningCheck) error {
	id := uuid.New()
	subjectID, err := pgutil.ParseUUID(check.SubjectUserID)
	if err != nil {
		return err
	}
	var relatedID pgtype.UUID
	if check.RelatedUserID != "" {
		parsed, err := pgutil.ParseUUID(check.RelatedUserID)
		if err != nil {
			return err
		}
		relatedID = pgtype.UUID{Bytes: parsed, Valid: true}
	}
	return r.q.InsertScreeningCheck(ctx, sqlc.InsertScreeningCheckParams{
		ID:                id,
		CheckType:         check.CheckType,
		SubjectUserID:     subjectID,
		RelatedUserID:     relatedID,
		EntityType:        check.EntityType,
		EntityID:          check.EntityID,
		Decision:          check.Decision,
		ReasonCode:        check.ReasonCode,
		Provider:          check.Provider,
		ProviderReference: pgutil.Text(check.ProviderReference),
		RawResponse:       check.RawResponse,
		CorrelationID:     pgutil.Text(check.CorrelationID),
	})
}