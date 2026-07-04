package sqlcrepo

import (
	"context"
	"errors"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
	"github.com/iho/neobank/services/simulators/cardproc/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
)

type ChargebackRepository struct {
	q sqlc.Querier
}

func NewChargebackRepository(q sqlc.Querier) *ChargebackRepository {
	return &ChargebackRepository{q: q}
}

func (r *ChargebackRepository) Create(ctx context.Context, transactionID, authorizationID, amount, currency, reason string) (domain.Chargeback, error) {
	txID, err := pgutil.ParseUUID(transactionID)
	if err != nil {
		return domain.Chargeback{}, err
	}

	numAmount, err := pgutil.NumericFromString(amount)
	if err != nil {
		return domain.Chargeback{}, err
	}

	row, err := r.q.CreateChargeback(ctx, sqlc.CreateChargebackParams{
		TransactionID:   txID,
		AuthorizationID: authorizationID,
		Amount:          numAmount,
		Currency:        currency,
		Reason:          reason,
	})
	if err != nil {
		return domain.Chargeback{}, err
	}

	return domain.Chargeback{
		ID:              row.ID.String(),
		TransactionID:   row.TransactionID.String(),
		AuthorizationID: row.AuthorizationID,
		Amount:          row.Amount,
		Currency:        row.Currency,
		Reason:          row.Reason,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt.Time.UTC(),
		UpdatedAt:       row.UpdatedAt.Time.UTC(),
	}, nil
}

func (r *ChargebackRepository) GetByID(ctx context.Context, id string) (*domain.Chargeback, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.q.GetChargebackByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &domain.Chargeback{
		ID:              row.ID.String(),
		TransactionID:   row.TransactionID.String(),
		AuthorizationID: row.AuthorizationID,
		Amount:          row.Amount,
		Currency:        row.Currency,
		Reason:          row.Reason,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt.Time.UTC(),
		UpdatedAt:       row.UpdatedAt.Time.UTC(),
	}, nil
}

func (r *ChargebackRepository) SetStatus(ctx context.Context, id, status string) (domain.Chargeback, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return domain.Chargeback{}, err
	}

	row, err := r.q.SetChargebackStatus(ctx, sqlc.SetChargebackStatusParams{
		ID:     uid,
		Status: status,
	})
	if err != nil {
		return domain.Chargeback{}, err
	}

	return domain.Chargeback{
		ID:              row.ID.String(),
		TransactionID:   row.TransactionID.String(),
		AuthorizationID: row.AuthorizationID,
		Amount:          row.Amount,
		Currency:        row.Currency,
		Reason:          row.Reason,
		Status:          row.Status,
		CreatedAt:       row.CreatedAt.Time.UTC(),
		UpdatedAt:       row.UpdatedAt.Time.UTC(),
	}, nil
}
