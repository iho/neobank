package sqlcrepo

import (
	"context"
	"errors"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
)

type BankAccountRepository struct {
	q sqlc.Querier
}

func NewBankAccountRepository(q sqlc.Querier) *BankAccountRepository {
	return &BankAccountRepository{q: q}
}

func (r *BankAccountRepository) Create(ctx context.Context, userID, currency, railsAccountID, iban string) (domain.BankAccount, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return domain.BankAccount{}, err
	}

	row, err := r.q.CreateBankAccount(ctx, sqlc.CreateBankAccountParams{
		UserID:         uid,
		Currency:       currency,
		RailsAccountID: railsAccountID,
		Iban:           iban,
	})
	if err != nil {
		return domain.BankAccount{}, err
	}

	return domain.BankAccount{
		ID:             row.ID.String(),
		UserID:         row.UserID.String(),
		Currency:       row.Currency,
		RailsAccountID: row.RailsAccountID,
		IBAN:           row.Iban,
		CreatedAt:      row.CreatedAt.Time.UTC(),
	}, nil
}

func (r *BankAccountRepository) GetByUserAndCurrency(ctx context.Context, userID, currency string) (*domain.BankAccount, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}

	row, err := r.q.GetBankAccountByUserAndCurrency(ctx, sqlc.GetBankAccountByUserAndCurrencyParams{
		UserID:   uid,
		Currency: currency,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &domain.BankAccount{
		ID:             row.ID.String(),
		UserID:         row.UserID.String(),
		Currency:       row.Currency,
		RailsAccountID: row.RailsAccountID,
		IBAN:           row.Iban,
		CreatedAt:      row.CreatedAt.Time.UTC(),
	}, nil
}
