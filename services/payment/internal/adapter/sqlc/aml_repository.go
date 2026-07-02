package sqlcrepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/amlmonitor"
	"github.com/iho/neobank/pkg/money"
	"github.com/iho/neobank/pkg/reqctx"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type AMLRepository struct {
	q sqlc.Querier
}

func NewAMLRepository(q sqlc.Querier) *AMLRepository {
	return &AMLRepository{q: q}
}

func (r *AMLRepository) RecordEvaluation(ctx context.Context, entityType, entityID, userID, transactionType, amount, currency string, result amlmonitor.Result) (string, error) {
	amt, err := money.Parse(amount)
	if err != nil {
		return "", err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}
	var numeric pgtype.Numeric
	if err := numeric.Scan(amt.String()); err != nil {
		return "", err
	}
	evalID := uuid.New()
	now := time.Now().UTC()
	id, err := r.q.InsertAMLEvaluation(ctx, sqlc.InsertAMLEvaluationParams{
		ID:              evalID,
		EntityType:      entityType,
		EntityID:        entityID,
		UserID:          uid,
		TransactionType: transactionType,
		Column6:         numeric,
		Currency:        currency,
		Disposition:     dispositionLabel(result.Disposition),
		ReasonCode:      result.ReasonCode,
		RiskScore:       int32(result.RiskScore),
		RuleSetVersion:  amlRuleSetVersion(result),
		CreatedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		CorrelationID:   textOrNil(reqctx.CorrelationID(ctx)),
	})
	if err != nil {
		return "", err
	}

	return id.String(), nil
}

func (r *AMLRepository) OpenCase(ctx context.Context, evaluationID, userID, entityType, entityID, caseType, reasonCode, correlationID string) error {
	evalUUID, err := uuid.Parse(evaluationID)
	if err != nil {
		return err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	return r.q.InsertAMLCase(ctx, sqlc.InsertAMLCaseParams{
		ID:            uuid.New(),
		EvaluationID:  evalUUID,
		UserID:        uid,
		EntityType:    entityType,
		EntityID:      entityID,
		CaseType:      caseType,
		ReasonCode:    reasonCode,
		CorrelationID: textOrNil(correlationID),
		CreatedAt:     pgtype.Timestamptz{Time: now, Valid: true},
	})
}

func dispositionLabel(d amlmonitor.Disposition) string {
	switch d {
	case amlmonitor.DispositionClear:
		return "clear"
	case amlmonitor.DispositionReview:
		return "review"
	case amlmonitor.DispositionReport:
		return "report"
	default:
		return "unknown"
	}
}

func amlRuleSetVersion(result amlmonitor.Result) string {
	if result.RuleSetVersion != "" {
		return result.RuleSetVersion
	}
	return amlmonitor.DefaultRuleSetVersion
}