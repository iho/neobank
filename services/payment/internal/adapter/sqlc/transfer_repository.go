package sqlcrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/iho/neobank/services/payment/internal/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type TransferRepository struct {
	q sqlc.Querier
}

func NewTransferRepository(q sqlc.Querier) *TransferRepository {
	return &TransferRepository{q: q}
}

func (r *TransferRepository) WithTx(tx pgx.Tx) port.TransferRepository {
	return &TransferRepository{q: withTx(r.q, tx)}
}

func (r *TransferRepository) Create(ctx context.Context, t domain.Transfer) error {
	id, err := pgutil.ParseUUID(t.ID)
	if err != nil {
		return err
	}
	senderID, err := pgutil.ParseUUID(t.SenderUserID)
	if err != nil {
		return err
	}
	recipientID, err := pgutil.ParseUUID(t.RecipientUserID)
	if err != nil {
		return err
	}
	amount, err := pgutil.NumericFromString(t.Amount)
	if err != nil {
		return err
	}

	return r.q.CreateTransfer(ctx, sqlc.CreateTransferParams{
		ID:              id,
		IdempotencyKey:  t.IdempotencyKey,
		Type:            t.Type,
		Status:          string(t.Status),
		SenderUserID:    senderID,
		RecipientUserID: recipientID,
		Amount:          amount,
		Currency:        t.Currency,
		Memo:            pgutil.Text(t.Memo),
	})
}

func (r *TransferRepository) GetBySenderAndIdempotencyKey(ctx context.Context, senderUserID, key string) (*domain.Transfer, error) {
	senderID, err := pgutil.ParseUUID(senderUserID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetTransferBySenderAndIdempotencyKey(ctx, sqlc.GetTransferBySenderAndIdempotencyKeyParams{
		SenderUserID:   senderID,
		IdempotencyKey: key,
	})
	if err != nil {
		return nil, err
	}
	return mapTransferByKeyRow(row), nil
}

func (r *TransferRepository) GetByID(ctx context.Context, id string) (*domain.Transfer, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetTransferByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return mapTransferByIDRow(row), nil
}

func (r *TransferRepository) MarkCompleted(ctx context.Context, id, ledgerTransferID string) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}
	return r.q.MarkTransferCompleted(ctx, sqlc.MarkTransferCompletedParams{
		ID:               uid,
		LedgerTransferID: pgutil.Text(ledgerTransferID),
	})
}

func (r *TransferRepository) MarkFailed(ctx context.Context, id, reason string) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}
	return r.q.MarkTransferFailed(ctx, sqlc.MarkTransferFailedParams{
		ID:            uid,
		FailureReason: pgutil.Text(reason),
	})
}

func (r *TransferRepository) ListByUser(ctx context.Context, userID string, limit int) ([]domain.Transfer, error) {
	if limit <= 0 {
		limit = 20
	}
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListTransfersByUser(ctx, sqlc.ListTransfersByUserParams{
		SenderUserID: uid,
		Limit:        int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Transfer, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapTransferListRow(row))
	}
	return out, nil
}

func mapTransferByIDRow(row sqlc.GetTransferByIDRow) *domain.Transfer {
	return mapTransfer(
		row.ID, row.IdempotencyKey, row.Type, row.Status,
		row.SenderUserID, row.RecipientUserID,
		row.Amount, row.Currency, row.Memo, row.LedgerTransferID, row.FailureReason,
		row.CompletedAt,
	)
}

func mapTransferByKeyRow(row sqlc.GetTransferBySenderAndIdempotencyKeyRow) *domain.Transfer {
	return mapTransfer(
		row.ID, row.IdempotencyKey, row.Type, row.Status,
		row.SenderUserID, row.RecipientUserID,
		row.Amount, row.Currency, row.Memo, row.LedgerTransferID, row.FailureReason,
		row.CompletedAt,
	)
}

func mapTransferListRow(row sqlc.ListTransfersByUserRow) *domain.Transfer {
	return mapTransfer(
		row.ID, row.IdempotencyKey, row.Type, row.Status,
		row.SenderUserID, row.RecipientUserID,
		row.Amount, row.Currency, row.Memo, row.LedgerTransferID, row.FailureReason,
		row.CompletedAt,
	)
}

func mapTransfer(
	id uuid.UUID,
	idempotencyKey, typ, status string,
	senderID, recipientID uuid.UUID,
	amount, currency, memo, ledgerTransferID, failureReason string,
	completedAt pgtype.Timestamptz,
) *domain.Transfer {
	t := &domain.Transfer{
		ID:               id.String(),
		IdempotencyKey:   idempotencyKey,
		Type:             typ,
		Status:           domain.TransferStatus(status),
		SenderUserID:     senderID.String(),
		RecipientUserID:  recipientID.String(),
		Amount:           amount,
		Currency:         currency,
		Memo:             memo,
		LedgerTransferID: ledgerTransferID,
		FailureReason:    failureReason,
	}
	if completedAt.Valid {
		ts := completedAt.Time
		t.CompletedAt = &ts
	}
	return t
}
