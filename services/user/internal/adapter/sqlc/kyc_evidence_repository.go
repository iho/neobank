package sqlcrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/piicrypto"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/iho/neobank/services/user/internal/port"
)

func (r *KYCRepository) CreateSubmission(ctx context.Context, sub port.KYCSubmission) error {
	id, err := parseOrNewUUID(sub.ID)
	if err != nil {
		return err
	}
	caseID, err := pgutil.ParseUUID(sub.KYCCaseID)
	if err != nil {
		return err
	}
	userID, err := pgutil.ParseUUID(sub.UserID)
	if err != nil {
		return err
	}
	docNumber, err := piicrypto.Store(ctx, r.pii, sub.DocumentNumber)
	if err != nil {
		return err
	}
	return r.q.InsertKYCSubmission(ctx, sqlc.InsertKYCSubmissionParams{
		ID:                id,
		KycCaseID:         caseID,
		UserID:            userID,
		DocumentType:      pgutil.Text(sub.DocumentType),
		DocumentNumber:    pgutil.Text(docNumber),
		Provider:          sub.Provider,
		ProviderReference: pgutil.Text(sub.ProviderReference),
		ProviderResponse:  sub.ProviderResponse,
		ScreeningDecision: sub.ScreeningDecision,
		ScreeningReason:   pgutil.Text(sub.ScreeningReason),
		CorrelationID:     pgutil.Text(sub.CorrelationID),
	})
}

func (r *KYCRepository) RecordScreeningCheck(ctx context.Context, check port.ScreeningCheck) error {
	id, err := parseOrNewUUID(check.ID)
	if err != nil {
		return err
	}
	subjectID, err := pgutil.ParseUUID(check.SubjectUserID)
	if err != nil {
		return err
	}
	return r.q.InsertScreeningCheck(ctx, sqlc.InsertScreeningCheckParams{
		ID:                id,
		CheckType:         check.CheckType,
		SubjectUserID:     subjectID,
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

func parseOrNewUUID(id string) (uuid.UUID, error) {
	if id == "" {
		return uuid.New(), nil
	}
	return pgutil.ParseUUID(id)
}