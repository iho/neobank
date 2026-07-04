package sqlcrepo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/simulators/rails/internal/domain"
	"github.com/iho/neobank/services/simulators/rails/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type OutboundPaymentRepository struct {
	q sqlc.Querier
}

func NewOutboundPaymentRepository(q sqlc.Querier) *OutboundPaymentRepository {
	return &OutboundPaymentRepository{q: q}
}

func (r *OutboundPaymentRepository) Create(ctx context.Context, accountID, amount, currency, counterpartyIBAN, reference string) (domain.OutboundPayment, error) {
	acctID, err := pgutil.ParseUUID(accountID)
	if err != nil {
		return domain.OutboundPayment{}, err
	}

	numAmount, err := pgutil.NumericFromString(amount)
	if err != nil {
		return domain.OutboundPayment{}, err
	}

	row, err := r.q.CreateOutboundPayment(ctx, sqlc.CreateOutboundPaymentParams{
		AccountID:        acctID,
		Amount:           numAmount,
		Currency:         currency,
		CounterpartyIban: counterpartyIBAN,
		Reference:        reference,
	})
	if err != nil {
		return domain.OutboundPayment{}, err
	}

	return toOutboundPayment(
		row.ID, row.AccountID, row.Amount, row.Currency, row.CounterpartyIban,
		row.Reference, row.Status, row.CreatedAt, row.UpdatedAt,
	), nil
}

func (r *OutboundPaymentRepository) GetByID(ctx context.Context, id string) (*domain.OutboundPayment, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.q.GetOutboundPaymentByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	payment := toOutboundPayment(
		row.ID, row.AccountID, row.Amount, row.Currency, row.CounterpartyIban,
		row.Reference, row.Status, row.CreatedAt, row.UpdatedAt,
	)

	return &payment, nil
}

func (r *OutboundPaymentRepository) SetStatus(ctx context.Context, id, status string) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}

	return r.q.SetOutboundPaymentStatus(ctx, sqlc.SetOutboundPaymentStatusParams{ID: uid, Status: status})
}

func (r *OutboundPaymentRepository) ListInRange(ctx context.Context, from, to time.Time) ([]domain.OutboundPayment, error) {
	rows, err := r.q.ListOutboundPaymentsInRange(ctx, sqlc.ListOutboundPaymentsInRangeParams{
		FromTs: pgtype.Timestamptz{Time: from.UTC(), Valid: true},
		ToTs:   pgtype.Timestamptz{Time: to.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	out := make([]domain.OutboundPayment, 0, len(rows))
	for _, row := range rows {
		out = append(out, toOutboundPayment(
			row.ID, row.AccountID, row.Amount, row.Currency, row.CounterpartyIban,
			row.Reference, row.Status, row.CreatedAt, row.UpdatedAt,
		))
	}

	return out, nil
}

func toOutboundPayment(
	id, accountID uuid.UUID,
	amount, currency, counterpartyIBAN, reference, status string,
	createdAt, updatedAt pgtype.Timestamptz,
) domain.OutboundPayment {
	return domain.OutboundPayment{
		ID:               id.String(),
		AccountID:        accountID.String(),
		Amount:           amount,
		Currency:         currency,
		CounterpartyIBAN: counterpartyIBAN,
		Reference:        reference,
		Status:           status,
		CreatedAt:        createdAt.Time.UTC(),
		UpdatedAt:        updatedAt.Time.UTC(),
	}
}
