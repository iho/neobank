package sqlcrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/gen/sqlc"
	"github.com/iho/neobank/services/card/internal/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type DisputeRepository struct {
	q sqlc.Querier
}

func NewDisputeRepository(q sqlc.Querier) *DisputeRepository {
	return &DisputeRepository{q: q}
}

func (r *DisputeRepository) WithTx(tx pgx.Tx) port.DisputeRepository {
	return &DisputeRepository{q: withTx(r.q, tx)}
}

func (r *DisputeRepository) Create(ctx context.Context, d domain.Dispute) (*domain.Dispute, error) {
	authID, err := pgutil.ParseUUID(d.AuthorizationID)
	if err != nil {
		return nil, err
	}
	cardID, err := pgutil.ParseUUID(d.CardID)
	if err != nil {
		return nil, err
	}
	userID, err := pgutil.ParseUUID(d.UserID)
	if err != nil {
		return nil, err
	}
	amount, err := pgutil.NumericFromString(d.Amount)
	if err != nil {
		return nil, err
	}

	row, err := r.q.CreateDispute(ctx, sqlc.CreateDisputeParams{
		ChargebackID:                d.ChargebackID,
		AuthorizationID:             authID,
		CardID:                      cardID,
		UserID:                      userID,
		Amount:                      amount,
		Currency:                    d.Currency,
		Reason:                      d.Reason,
		ProvisionalCreditTransferID: d.ProvisionalCreditTransferID,
	})
	if err != nil {
		return nil, err
	}

	return mapDispute(
		row.ID, row.ChargebackID, row.AuthorizationID, row.CardID, row.UserID, row.Amount, row.Currency,
		row.Reason, row.Status, row.ProvisionalCreditTransferID, row.ReversalTransferID, row.CreatedAt, row.UpdatedAt,
	), nil
}

func (r *DisputeRepository) GetByChargebackID(ctx context.Context, chargebackID string) (*domain.Dispute, error) {
	row, err := r.q.GetDisputeByChargebackID(ctx, chargebackID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return mapDispute(
		row.ID, row.ChargebackID, row.AuthorizationID, row.CardID, row.UserID, row.Amount, row.Currency,
		row.Reason, row.Status, row.ProvisionalCreditTransferID, row.ReversalTransferID, row.CreatedAt, row.UpdatedAt,
	), nil
}

func (r *DisputeRepository) MarkResolved(ctx context.Context, chargebackID, status, reversalTransferID string) (*domain.Dispute, error) {
	row, err := r.q.MarkDisputeResolved(ctx, sqlc.MarkDisputeResolvedParams{
		ChargebackID:       chargebackID,
		Status:             status,
		ReversalTransferID: reversalTransferID,
	})
	if err != nil {
		return nil, err
	}

	return mapDispute(
		row.ID, row.ChargebackID, row.AuthorizationID, row.CardID, row.UserID, row.Amount, row.Currency,
		row.Reason, row.Status, row.ProvisionalCreditTransferID, row.ReversalTransferID, row.CreatedAt, row.UpdatedAt,
	), nil
}

func mapDispute(
	id uuid.UUID, chargebackID string, authorizationID, cardID, userID uuid.UUID,
	amount, currency, reason, status, provisionalCreditTransferID, reversalTransferID string,
	createdAt, updatedAt pgtype.Timestamptz,
) *domain.Dispute {
	d := &domain.Dispute{
		ID:                          id.String(),
		ChargebackID:                chargebackID,
		AuthorizationID:             authorizationID.String(),
		CardID:                      cardID.String(),
		UserID:                      userID.String(),
		Amount:                      amount,
		Currency:                    currency,
		Reason:                      reason,
		Status:                      status,
		ProvisionalCreditTransferID: provisionalCreditTransferID,
		ReversalTransferID:          reversalTransferID,
	}
	if createdAt.Valid {
		d.CreatedAt = createdAt.Time.UTC()
	}
	if updatedAt.Valid {
		d.UpdatedAt = updatedAt.Time.UTC()
	}

	return d
}
