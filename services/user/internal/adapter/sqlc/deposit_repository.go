package sqlcrepo

import (
	"context"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/iho/neobank/services/user/internal/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type DepositRepository struct {
	q sqlc.Querier
}

func NewDepositRepository(q sqlc.Querier) *DepositRepository {
	return &DepositRepository{q: q}
}

func (r *DepositRepository) WithTx(tx pgx.Tx) port.DepositRepository {
	return &DepositRepository{q: withTx(r.q, tx)}
}

func (r *DepositRepository) GetByUserAndIdempotencyKey(ctx context.Context, userID, key string) (*domain.Deposit, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetDepositByUserAndIdempotencyKey(ctx, sqlc.GetDepositByUserAndIdempotencyKeyParams{
		UserID:         uid,
		IdempotencyKey: key,
	})
	if err != nil {
		return nil, err
	}
	return rowToDeposit(row), nil
}

func (r *DepositRepository) Insert(ctx context.Context, deposit domain.Deposit) error {
	id, err := pgutil.ParseUUID(deposit.ID)
	if err != nil {
		return err
	}
	userID, err := pgutil.ParseUUID(deposit.UserID)
	if err != nil {
		return err
	}
	walletID, err := pgutil.ParseUUID(deposit.WalletID)
	if err != nil {
		return err
	}
	amount, err := pgutil.NumericFromString(deposit.Amount)
	if err != nil {
		return err
	}
	var completedAt pgtype.Timestamptz
	if deposit.CompletedAt != nil {
		completedAt = pgtype.Timestamptz{Time: *deposit.CompletedAt, Valid: true}
	}
	return r.q.InsertDeposit(ctx, sqlc.InsertDepositParams{
		ID:               id,
		UserID:           userID,
		WalletID:         walletID,
		Amount:           amount,
		Currency:         deposit.Currency,
		LedgerTransferID: pgutil.Text(deposit.LedgerTransferID),
		Status:           string(deposit.Status),
		IdempotencyKey:   deposit.IdempotencyKey,
		CompletedAt:      completedAt,
	})
}

func rowToDeposit(row sqlc.GetDepositByUserAndIdempotencyKeyRow) *domain.Deposit {
	out := &domain.Deposit{
		ID:               row.ID.String(),
		UserID:           row.UserID.String(),
		WalletID:         row.WalletID.String(),
		Amount:           row.Amount,
		Currency:         row.Currency,
		LedgerTransferID: row.LedgerTransferID,
		Status:           domain.DepositStatus(row.Status),
		IdempotencyKey:   row.IdempotencyKey,
	}
	if row.CreatedAt.Valid {
		out.CreatedAt = row.CreatedAt.Time.UTC()
	}
	if row.CompletedAt.Valid {
		t := row.CompletedAt.Time.UTC()
		out.CompletedAt = &t
	}
	return out
}