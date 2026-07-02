package sqlcrepo

import (
	"context"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/notification/internal/domain"
	"github.com/iho/neobank/services/notification/internal/gen/sqlc"
)

type NotificationRepository struct {
	q sqlc.Querier
}

func NewNotificationRepository(q sqlc.Querier) *NotificationRepository {
	return &NotificationRepository{q: q}
}

func (r *NotificationRepository) Create(ctx context.Context, n domain.Notification, eventID string) error {
	id, err := pgutil.ParseUUID(n.ID)
	if err != nil {
		return err
	}
	userID, err := pgutil.ParseUUID(n.UserID)
	if err != nil {
		return err
	}
	eid, err := pgutil.ParseUUID(eventID)
	if err != nil {
		return err
	}
	return r.q.InsertNotification(ctx, sqlc.InsertNotificationParams{
		ID:        id,
		UserID:    userID,
		EventID:   eid,
		EventType: n.EventType,
		Title:     n.Title,
		Body:      n.Body,
	})
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID string, limit int) ([]domain.Notification, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListNotificationsByUser(ctx, sqlc.ListNotificationsByUserParams{
		UserID: uid,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Notification{
			ID:        row.ID.String(),
			UserID:    row.UserID.String(),
			EventType: row.EventType,
			Title:     row.Title,
			Body:      row.Body,
			Read:      row.Read,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return out, nil
}