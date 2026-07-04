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

type BankTransferRepository struct {
	q sqlc.Querier
}

func NewBankTransferRepository(q sqlc.Querier) *BankTransferRepository {
	return &BankTransferRepository{q: q}
}

func (r *BankTransferRepository) WithTx(tx pgx.Tx) port.BankTransferRepository {
	return &BankTransferRepository{q: withTx(r.q, tx)}
}

func (r *BankTransferRepository) Create(ctx context.Context, t domain.BankTransfer) (domain.BankTransfer, error) {
	uid, err := pgutil.ParseUUID(t.UserID)
	if err != nil {
		return domain.BankTransfer{}, err
	}

	amount, err := pgutil.NumericFromString(t.Amount)
	if err != nil {
		return domain.BankTransfer{}, err
	}

	row, err := r.q.CreateBankTransfer(ctx, sqlc.CreateBankTransferParams{
		RailsTransferID:  t.RailsTransferID,
		UserID:           uid,
		Amount:           amount,
		Currency:         t.Currency,
		SenderName:       t.SenderName,
		Reference:        t.Reference,
		LedgerTransferID: t.LedgerTransferID,
	})
	if err != nil {
		return domain.BankTransfer{}, err
	}

	return domain.BankTransfer{
		ID:               row.ID.String(),
		RailsTransferID:  row.RailsTransferID,
		UserID:           row.UserID.String(),
		Amount:           row.Amount,
		Currency:         row.Currency,
		SenderName:       row.SenderName,
		Reference:        row.Reference,
		LedgerTransferID: row.LedgerTransferID,
		Status:           row.Status,
		CreatedAt:        row.CreatedAt.Time.UTC(),
	}, nil
}

func (r *BankTransferRepository) GetByRailsTransferID(ctx context.Context, railsTransferID string) (*domain.BankTransfer, error) {
	row, err := r.q.GetBankTransferByRailsTransferID(ctx, railsTransferID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &domain.BankTransfer{
		ID:               row.ID.String(),
		RailsTransferID:  row.RailsTransferID,
		UserID:           row.UserID.String(),
		Amount:           row.Amount,
		Currency:         row.Currency,
		SenderName:       row.SenderName,
		Reference:        row.Reference,
		LedgerTransferID: row.LedgerTransferID,
		Status:           row.Status,
		CreatedAt:        row.CreatedAt.Time.UTC(),
	}, nil
}
