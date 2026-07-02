package sqlcrepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/fraud"
	"github.com/iho/neobank/pkg/money"
	"github.com/iho/neobank/pkg/reqctx"
	"github.com/iho/neobank/services/card/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type FraudDecisionRepository struct {
	q sqlc.Querier
}

func NewFraudDecisionRepository(q sqlc.Querier) *FraudDecisionRepository {
	return &FraudDecisionRepository{q: q}
}

func (r *FraudDecisionRepository) Record(ctx context.Context, entityType, entityID, userID, transactionType, amount, currency string, result fraud.Result) error {
	amt, err := money.Parse(amount)
	if err != nil {
		return err
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	var numeric pgtype.Numeric
	if err := numeric.Scan(amt.String()); err != nil {
		return err
	}
	return r.q.InsertFraudDecision(ctx, sqlc.InsertFraudDecisionParams{
		ID:              uuid.New(),
		EntityType:      entityType,
		EntityID:        entityID,
		UserID:          uid,
		TransactionType: transactionType,
		Column6:         numeric,
		Currency:        currency,
		Decision:        decisionLabel(result.Decision),
		ReasonCode:      result.ReasonCode,
		RiskScore:       int32(result.RiskScore),
		CorrelationID:   textOrNil(reqctx.CorrelationID(ctx)),
		CreatedAt:       pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	})
}

func decisionLabel(d fraud.Decision) string {
	switch d {
	case fraud.DecisionAllow:
		return "allow"
	case fraud.DecisionReview:
		return "review"
	case fraud.DecisionDeny:
		return "deny"
	default:
		return "unknown"
	}
}
