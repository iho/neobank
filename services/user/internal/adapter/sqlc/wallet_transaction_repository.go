package sqlcrepo

import (
	"context"
	"time"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/walletprojection"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type WalletTransactionRepository struct {
	q sqlc.Querier
}

func NewWalletTransactionRepository(q sqlc.Querier) *WalletTransactionRepository {
	return &WalletTransactionRepository{q: q}
}

func (r *WalletTransactionRepository) Insert(ctx context.Context, row walletprojection.Row) error {
	userID, err := pgutil.ParseUUID(row.UserID)
	if err != nil {
		return err
	}
	eventID, err := pgutil.ParseUUID(row.SourceEventID)
	if err != nil {
		return err
	}
	return r.q.InsertWalletTransaction(ctx, sqlc.InsertWalletTransactionParams{
		UserID:        userID,
		ID:            row.ID,
		SourceEventID: eventID,
		TxType:        row.Type,
		Amount:        row.Amount,
		Currency:      row.Currency,
		Direction:     row.Direction,
		Status:        row.Status,
		Counterparty:  pgutil.Text(row.Counterparty),
		Memo:          pgutil.Text(row.Memo),
		CreatedAt:     pgtype.Timestamptz{Time: row.CreatedAt, Valid: true},
	})
}

func (r *WalletTransactionRepository) ApplyCapture(ctx context.Context, update walletprojection.CaptureUpdate) error {
	userID, err := pgutil.ParseUUID(update.UserID)
	if err != nil {
		return err
	}
	eventID, err := pgutil.ParseUUID(update.SourceEventID)
	if err != nil {
		return err
	}
	return r.q.UpsertWalletTransactionCapture(ctx, sqlc.UpsertWalletTransactionCaptureParams{
		UserID:        userID,
		ID:            update.ID,
		SourceEventID: eventID,
		TxType:        update.Type,
		Amount:        update.Amount,
		Currency:      update.Currency,
		Status:        update.Status,
		CreatedAt:     pgtype.Timestamptz{Time: update.CreatedAt, Valid: true},
	})
}

func (r *WalletTransactionRepository) ListByUser(ctx context.Context, userID string, limit int, cursorCreatedAt *time.Time, cursorID string) ([]domain.WalletTransaction, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	var cursorAt pgtype.Timestamptz
	if cursorCreatedAt != nil {
		cursorAt = pgtype.Timestamptz{Time: cursorCreatedAt.UTC(), Valid: true}
	}
	rows, err := r.q.ListWalletTransactionsByUser(ctx, sqlc.ListWalletTransactionsByUserParams{
		UserID:          uid,
		LimitVal:        int32(limit),
		CursorCreatedAt: cursorAt,
		CursorID:        pgutil.Text(cursorID),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.WalletTransaction, 0, len(rows))
	for _, row := range rows {
		tx := domain.WalletTransaction{
			ID:        row.ID,
			Type:      row.TxType,
			Amount:    row.Amount,
			Currency:  row.Currency,
			Direction: row.Direction,
			Status:    row.Status,
			CreatedAt: row.CreatedAt.Time,
		}
		if row.Counterparty.Valid {
			tx.Counterparty = row.Counterparty.String
		}
		if row.Memo.Valid {
			tx.Memo = row.Memo.String
		}
		out = append(out, tx)
	}
	return out, nil
}