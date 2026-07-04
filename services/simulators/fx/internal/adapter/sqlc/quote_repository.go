package sqlcrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/simulators/fx/internal/domain"
	"github.com/iho/neobank/services/simulators/fx/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type QuoteRepository struct {
	q sqlc.Querier
}

func NewQuoteRepository(q sqlc.Querier) *QuoteRepository {
	return &QuoteRepository{q: q}
}

func (r *QuoteRepository) Create(ctx context.Context, quote domain.Quote) (domain.Quote, error) {
	amount, err := pgutil.NumericFromString(quote.Amount)
	if err != nil {
		return domain.Quote{}, err
	}

	converted, err := pgutil.NumericFromString(quote.ConvertedAmount)
	if err != nil {
		return domain.Quote{}, err
	}

	rate, err := pgutil.NumericFromString(quote.Rate)
	if err != nil {
		return domain.Quote{}, err
	}

	row, err := r.q.CreateQuote(ctx, sqlc.CreateQuoteParams{
		FromCurrency:    quote.FromCurrency,
		ToCurrency:      quote.ToCurrency,
		Amount:          amount,
		ConvertedAmount: converted,
		Rate:            rate,
		SpreadBps:       int32(quote.SpreadBps),
		ExpiresAt:       pgtype.Timestamptz{Time: quote.ExpiresAt.UTC(), Valid: true},
	})
	if err != nil {
		return domain.Quote{}, err
	}

	return toQuote(
		row.ID, row.FromCurrency, row.ToCurrency, row.Amount, row.ConvertedAmount,
		row.Rate, row.SpreadBps, row.Status, row.CreatedAt, row.ExpiresAt, row.ExecutedAt,
	), nil
}

func (r *QuoteRepository) GetByID(ctx context.Context, id string) (*domain.Quote, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.q.GetQuoteByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	quote := toQuote(
		row.ID, row.FromCurrency, row.ToCurrency, row.Amount, row.ConvertedAmount,
		row.Rate, row.SpreadBps, row.Status, row.CreatedAt, row.ExpiresAt, row.ExecutedAt,
	)

	return &quote, nil
}

func (r *QuoteRepository) MarkExecuted(ctx context.Context, id string) (domain.Quote, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return domain.Quote{}, err
	}

	row, err := r.q.MarkQuoteExecuted(ctx, uid)
	if err != nil {
		return domain.Quote{}, err
	}

	return toQuote(
		row.ID, row.FromCurrency, row.ToCurrency, row.Amount, row.ConvertedAmount,
		row.Rate, row.SpreadBps, row.Status, row.CreatedAt, row.ExpiresAt, row.ExecutedAt,
	), nil
}

func toQuote(
	id uuid.UUID,
	fromCurrency, toCurrency, amount, convertedAmount, rate string,
	spreadBps int32,
	status string,
	createdAt, expiresAt, executedAt pgtype.Timestamptz,
) domain.Quote {
	q := domain.Quote{
		ID:              id.String(),
		FromCurrency:    fromCurrency,
		ToCurrency:      toCurrency,
		Amount:          amount,
		ConvertedAmount: convertedAmount,
		Rate:            rate,
		SpreadBps:       int(spreadBps),
		Status:          status,
		CreatedAt:       createdAt.Time.UTC(),
		ExpiresAt:       expiresAt.Time.UTC(),
	}
	if executedAt.Valid {
		t := executedAt.Time.UTC()
		q.ExecutedAt = &t
	}

	return q
}
