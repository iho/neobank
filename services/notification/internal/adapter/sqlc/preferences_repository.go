package sqlcrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/iho/neobank/services/notification/internal/domain"
	"github.com/iho/neobank/services/notification/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
)

type PreferencesRepository struct {
	q sqlc.Querier
}

func NewPreferencesRepository(q sqlc.Querier) *PreferencesRepository {
	return &PreferencesRepository{q: q}
}

func (r *PreferencesRepository) Get(ctx context.Context, userID string) (domain.NotificationPreferences, error) {
	row, err := r.q.GetNotificationPreferences(ctx, uuid.MustParse(userID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return defaultPreferences(userID), nil
		}
		return domain.NotificationPreferences{}, err
	}
	return preferencesFromRow(row), nil
}

func (r *PreferencesRepository) Upsert(ctx context.Context, prefs domain.NotificationPreferences) (domain.NotificationPreferences, error) {
	row, err := r.q.UpsertNotificationPreferences(ctx, sqlc.UpsertNotificationPreferencesParams{
		UserID:    uuid.MustParse(prefs.UserID),
		Transfers: prefs.Transfers,
		Cards:     prefs.Cards,
		Kyc:       prefs.KYC,
		Push:      prefs.Push,
		Email:     prefs.Email,
	})
	if err != nil {
		return domain.NotificationPreferences{}, err
	}
	return preferencesFromRow(row), nil
}

func defaultPreferences(userID string) domain.NotificationPreferences {
	return domain.NotificationPreferences{
		UserID:    userID,
		Transfers: true,
		Cards:     true,
		KYC:       true,
		Push:      true,
		Email:     true,
	}
}

func preferencesFromRow(row sqlc.NotificationNotificationPreference) domain.NotificationPreferences {
	return domain.NotificationPreferences{
		UserID:    row.UserID.String(),
		Transfers: row.Transfers,
		Cards:     row.Cards,
		KYC:       row.Kyc,
		Push:      row.Push,
		Email:     row.Email,
		UpdatedAt: row.UpdatedAt.Time.UTC(),
	}
}