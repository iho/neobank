package sqlcrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
)

type DeviceTokenRepository struct {
	q sqlc.Querier
}

func NewDeviceTokenRepository(q sqlc.Querier) *DeviceTokenRepository {
	return &DeviceTokenRepository{q: q}
}

func (r *DeviceTokenRepository) Upsert(ctx context.Context, userID, platform, token string) (*domain.DeviceToken, error) {
	row, err := r.q.UpsertDeviceToken(ctx, sqlc.UpsertDeviceTokenParams{
		UserID:   uuid.MustParse(userID),
		Platform: platform,
		Token:    token,
	})
	if err != nil {
		return nil, err
	}
	return deviceTokenFromRow(row), nil
}

func (r *DeviceTokenRepository) Delete(ctx context.Context, userID, tokenID string) (bool, error) {
	n, err := r.q.DeleteDeviceToken(ctx, sqlc.DeleteDeviceTokenParams{
		UserID: uuid.MustParse(userID),
		ID:     uuid.MustParse(tokenID),
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *DeviceTokenRepository) ListByUser(ctx context.Context, userID string) ([]domain.DeviceToken, error) {
	rows, err := r.q.ListDeviceTokensByUser(ctx, uuid.MustParse(userID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.DeviceToken, 0, len(rows))
	for _, row := range rows {
		out = append(out, *deviceTokenFromRow(row))
	}
	return out, nil
}

func deviceTokenFromRow(row sqlc.UserDeviceToken) *domain.DeviceToken {
	return &domain.DeviceToken{
		ID:        row.ID.String(),
		UserID:    row.UserID.String(),
		Platform:  row.Platform,
		Token:     row.Token,
		CreatedAt: row.CreatedAt.Time.UTC(),
		UpdatedAt: row.UpdatedAt.Time.UTC(),
	}
}