package sqlcrepo

import (
	"context"
	"errors"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
	"github.com/iho/neobank/services/simulators/cardproc/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
)

type TransactionRepository struct {
	q sqlc.Querier
}

func NewTransactionRepository(q sqlc.Querier) *TransactionRepository {
	return &TransactionRepository{q: q}
}

func (r *TransactionRepository) Create(ctx context.Context, cardID, amount, currency, merchantName, mcc string) (domain.Transaction, error) {
	cid, err := pgutil.ParseUUID(cardID)
	if err != nil {
		return domain.Transaction{}, err
	}

	numAmount, err := pgutil.NumericFromString(amount)
	if err != nil {
		return domain.Transaction{}, err
	}

	row, err := r.q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		CardID:       cid,
		Amount:       numAmount,
		Currency:     currency,
		MerchantName: merchantName,
		Mcc:          mcc,
	})
	if err != nil {
		return domain.Transaction{}, err
	}

	return domain.Transaction{
		ID:              row.ID.String(),
		CardID:          row.CardID.String(),
		AuthorizationID: row.AuthorizationID,
		Amount:          row.Amount,
		Currency:        row.Currency,
		MerchantName:    row.MerchantName,
		MCC:             row.Mcc,
		Status:          row.Status,
		ReasonCode:      row.ReasonCode,
		CreatedAt:       row.CreatedAt.Time.UTC(),
	}, nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.q.GetTransactionByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	tx := domain.Transaction{
		ID:              row.ID.String(),
		CardID:          row.CardID.String(),
		AuthorizationID: row.AuthorizationID,
		Amount:          row.Amount,
		Currency:        row.Currency,
		MerchantName:    row.MerchantName,
		MCC:             row.Mcc,
		Status:          row.Status,
		ReasonCode:      row.ReasonCode,
		CreatedAt:       row.CreatedAt.Time.UTC(),
	}
	if row.CapturedAt.Valid {
		t := row.CapturedAt.Time.UTC()
		tx.CapturedAt = &t
	}
	if row.ReversedAt.Valid {
		t := row.ReversedAt.Time.UTC()
		tx.ReversedAt = &t
	}

	return &tx, nil
}

func (r *TransactionRepository) SetAuthResult(ctx context.Context, id, status, authorizationID, reasonCode string) (domain.Transaction, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return domain.Transaction{}, err
	}

	row, err := r.q.SetTransactionAuthResult(ctx, sqlc.SetTransactionAuthResultParams{
		ID:              uid,
		Status:          status,
		AuthorizationID: authorizationID,
		ReasonCode:      reasonCode,
	})
	if err != nil {
		return domain.Transaction{}, err
	}

	return domain.Transaction{
		ID:              row.ID.String(),
		CardID:          row.CardID.String(),
		AuthorizationID: row.AuthorizationID,
		Amount:          row.Amount,
		Currency:        row.Currency,
		MerchantName:    row.MerchantName,
		MCC:             row.Mcc,
		Status:          row.Status,
		ReasonCode:      row.ReasonCode,
		CreatedAt:       row.CreatedAt.Time.UTC(),
	}, nil
}

func (r *TransactionRepository) MarkCaptured(ctx context.Context, id string) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}

	return r.q.MarkTransactionCaptured(ctx, uid)
}

func (r *TransactionRepository) MarkReversed(ctx context.Context, id string) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}

	return r.q.MarkTransactionReversed(ctx, uid)
}
