package sqlcrepo

import (
	"context"
	"errors"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
)

type SavedPayeeRepository struct {
	q sqlc.Querier
}

func NewSavedPayeeRepository(q sqlc.Querier) *SavedPayeeRepository {
	return &SavedPayeeRepository{q: q}
}

func (r *SavedPayeeRepository) ListByUser(ctx context.Context, userID string, limit int) ([]domain.SavedPayee, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListSavedPayeesByUser(ctx, sqlc.ListSavedPayeesByUserParams{
		UserID: uid,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.SavedPayee, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapSavedPayeeRow(row))
	}
	return out, nil
}

func (r *SavedPayeeRepository) Upsert(ctx context.Context, userID, payeeUserID, nickname string) (*domain.SavedPayee, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	payeeID, err := pgutil.ParseUUID(payeeUserID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.UpsertSavedPayee(ctx, sqlc.UpsertSavedPayeeParams{
		UserID:      uid,
		PayeeUserID: payeeID,
		Nickname:    pgutil.Text(nickname),
	})
	if err != nil {
		return nil, err
	}
	payee := mapUpsertPayeeRow(row)
	return &payee, nil
}

func (r *SavedPayeeRepository) Create(ctx context.Context, userID, payeeUserID, nickname string) (*domain.SavedPayee, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	payeeID, err := pgutil.ParseUUID(payeeUserID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.CreateSavedPayee(ctx, sqlc.CreateSavedPayeeParams{
		UserID:      uid,
		PayeeUserID: payeeID,
		Nickname:    pgutil.Text(nickname),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	payee := mapCreatePayeeRow(row)
	return &payee, nil
}

func (r *SavedPayeeRepository) GetByID(ctx context.Context, userID, payeeID string) (*domain.SavedPayee, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	id, err := pgutil.ParseUUID(payeeID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetSavedPayeeByID(ctx, sqlc.GetSavedPayeeByIDParams{
		ID:     id,
		UserID: uid,
	})
	if err != nil {
		return nil, err
	}
	payee := mapGetPayeeRow(row)
	return &payee, nil
}

func (r *SavedPayeeRepository) Delete(ctx context.Context, userID, payeeID string) (bool, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return false, err
	}
	id, err := pgutil.ParseUUID(payeeID)
	if err != nil {
		return false, err
	}
	n, err := r.q.DeleteSavedPayee(ctx, sqlc.DeleteSavedPayeeParams{
		ID:     id,
		UserID: uid,
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func mapSavedPayeeRow(row sqlc.ListSavedPayeesByUserRow) domain.SavedPayee {
	return domain.SavedPayee{
		ID:          row.ID.String(),
		UserID:      row.UserID.String(),
		PayeeUserID: row.PayeeUserID.String(),
		Nickname:    row.Nickname,
		PayeeEmail:  row.PayeeEmail,
		PayeePhone:  row.PayeePhone,
		LastUsedAt:  row.LastUsedAt.Time.UTC(),
		CreatedAt:   row.CreatedAt.Time.UTC(),
	}
}

func mapUpsertPayeeRow(row sqlc.UpsertSavedPayeeRow) domain.SavedPayee {
	return domain.SavedPayee{
		ID:          row.ID.String(),
		UserID:      row.UserID.String(),
		PayeeUserID: row.PayeeUserID.String(),
		Nickname:    row.Nickname,
		LastUsedAt:  row.LastUsedAt.Time.UTC(),
		CreatedAt:   row.CreatedAt.Time.UTC(),
	}
}

func mapCreatePayeeRow(row sqlc.CreateSavedPayeeRow) domain.SavedPayee {
	return domain.SavedPayee{
		ID:          row.ID.String(),
		UserID:      row.UserID.String(),
		PayeeUserID: row.PayeeUserID.String(),
		Nickname:    row.Nickname,
		LastUsedAt:  row.LastUsedAt.Time.UTC(),
		CreatedAt:   row.CreatedAt.Time.UTC(),
	}
}

func mapGetPayeeRow(row sqlc.GetSavedPayeeByIDRow) domain.SavedPayee {
	return domain.SavedPayee{
		ID:          row.ID.String(),
		UserID:      row.UserID.String(),
		PayeeUserID: row.PayeeUserID.String(),
		Nickname:    row.Nickname,
		LastUsedAt:  row.LastUsedAt.Time.UTC(),
		CreatedAt:   row.CreatedAt.Time.UTC(),
	}
}