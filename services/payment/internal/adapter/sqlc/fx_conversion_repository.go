package sqlcrepo

import (
	"context"
	"errors"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/iho/neobank/services/payment/internal/port"
	"github.com/jackc/pgx/v5"
)

type FXConversionRepository struct {
	q sqlc.Querier
}

func NewFXConversionRepository(q sqlc.Querier) *FXConversionRepository {
	return &FXConversionRepository{q: q}
}

func (r *FXConversionRepository) WithTx(tx pgx.Tx) port.FXConversionRepository {
	return &FXConversionRepository{q: withTx(r.q, tx)}
}

func (r *FXConversionRepository) Create(ctx context.Context, c domain.FXConversion) (domain.FXConversion, error) {
	uid, err := pgutil.ParseUUID(c.UserID)
	if err != nil {
		return domain.FXConversion{}, err
	}

	amount, err := pgutil.NumericFromString(c.Amount)
	if err != nil {
		return domain.FXConversion{}, err
	}

	converted, err := pgutil.NumericFromString(c.ConvertedAmount)
	if err != nil {
		return domain.FXConversion{}, err
	}

	rate, err := pgutil.NumericFromString(c.Rate)
	if err != nil {
		return domain.FXConversion{}, err
	}

	row, err := r.q.CreateFXConversion(ctx, sqlc.CreateFXConversionParams{
		QuoteID:              c.QuoteID,
		UserID:               uid,
		FromCurrency:         c.FromCurrency,
		ToCurrency:           c.ToCurrency,
		Amount:               amount,
		ConvertedAmount:      converted,
		Rate:                 rate,
		FromLedgerTransferID: c.FromLedgerTransferID,
		ToLedgerTransferID:   c.ToLedgerTransferID,
	})
	if err != nil {
		return domain.FXConversion{}, err
	}

	return domain.FXConversion{
		ID:                   row.ID.String(),
		QuoteID:              row.QuoteID,
		UserID:               row.UserID.String(),
		FromCurrency:         row.FromCurrency,
		ToCurrency:           row.ToCurrency,
		Amount:               row.Amount,
		ConvertedAmount:      row.ConvertedAmount,
		Rate:                 row.Rate,
		FromLedgerTransferID: row.FromLedgerTransferID,
		ToLedgerTransferID:   row.ToLedgerTransferID,
		Status:               row.Status,
		CreatedAt:            row.CreatedAt.Time.UTC(),
	}, nil
}

func (r *FXConversionRepository) GetByQuoteID(ctx context.Context, quoteID string) (*domain.FXConversion, error) {
	row, err := r.q.GetFXConversionByQuoteID(ctx, quoteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &domain.FXConversion{
		ID:                   row.ID.String(),
		QuoteID:              row.QuoteID,
		UserID:               row.UserID.String(),
		FromCurrency:         row.FromCurrency,
		ToCurrency:           row.ToCurrency,
		Amount:               row.Amount,
		ConvertedAmount:      row.ConvertedAmount,
		Rate:                 row.Rate,
		FromLedgerTransferID: row.FromLedgerTransferID,
		ToLedgerTransferID:   row.ToLedgerTransferID,
		Status:               row.Status,
		CreatedAt:            row.CreatedAt.Time.UTC(),
	}, nil
}
