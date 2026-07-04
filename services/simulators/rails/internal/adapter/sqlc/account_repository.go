package sqlcrepo

import (
	"context"
	"errors"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/simulators/rails/internal/domain"
	"github.com/iho/neobank/services/simulators/rails/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
)

type AccountRepository struct {
	q sqlc.Querier
}

func NewAccountRepository(q sqlc.Querier) *AccountRepository {
	return &AccountRepository{q: q}
}

func (r *AccountRepository) Create(ctx context.Context, externalRef, currency, iban string) (domain.Account, error) {
	row, err := r.q.CreateAccount(ctx, sqlc.CreateAccountParams{
		ExternalRef: externalRef,
		Currency:    currency,
		Iban:        iban,
	})
	if err != nil {
		return domain.Account{}, err
	}

	return domain.Account{
		ID:          row.ID.String(),
		ExternalRef: row.ExternalRef,
		Currency:    row.Currency,
		IBAN:        row.Iban,
		CreatedAt:   row.CreatedAt.Time.UTC(),
	}, nil
}

func (r *AccountRepository) GetByExternalRefAndCurrency(ctx context.Context, externalRef, currency string) (*domain.Account, error) {
	row, err := r.q.GetAccountByExternalRefAndCurrency(ctx, sqlc.GetAccountByExternalRefAndCurrencyParams{
		ExternalRef: externalRef,
		Currency:    currency,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &domain.Account{
		ID:          row.ID.String(),
		ExternalRef: row.ExternalRef,
		Currency:    row.Currency,
		IBAN:        row.Iban,
		CreatedAt:   row.CreatedAt.Time.UTC(),
	}, nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.q.GetAccountByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &domain.Account{
		ID:          row.ID.String(),
		ExternalRef: row.ExternalRef,
		Currency:    row.Currency,
		IBAN:        row.Iban,
		CreatedAt:   row.CreatedAt.Time.UTC(),
	}, nil
}
