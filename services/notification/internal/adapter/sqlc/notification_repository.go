package sqlcrepo

import (
	"context"
	"time"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/notification/internal/domain"
	"github.com/iho/neobank/services/notification/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *NotificationRepository) ListByUser(ctx context.Context, userID string, limit int, cursorCreatedAt *time.Time, cursorID string) ([]domain.Notification, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	var cursorAt pgtype.Timestamptz
	if cursorCreatedAt != nil {
		cursorAt = pgtype.Timestamptz{Time: cursorCreatedAt.UTC(), Valid: true}
	}
	var cursorUUID pgtype.UUID
	if cursorID != "" {
		parsed, parseErr := pgutil.ParseUUID(cursorID)
		if parseErr != nil {
			return nil, parseErr
		}
		cursorUUID = pgtype.UUID{Bytes: parsed, Valid: true}
	}
	rows, err := r.q.ListNotificationsByUser(ctx, sqlc.ListNotificationsByUserParams{
		UserID:          uid,
		LimitVal:        int32(limit),
		CursorCreatedAt: cursorAt,
		CursorID:        cursorUUID,
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

func (r *NotificationRepository) CountUnread(ctx context.Context, userID string) (int64, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return 0, err
	}
	return r.q.CountUnreadNotificationsByUser(ctx, uid)
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID, notificationID string) (domain.Notification, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return domain.Notification{}, err
	}
	nid, err := pgutil.ParseUUID(notificationID)
	if err != nil {
		return domain.Notification{}, err
	}
	row, err := r.q.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{
		ID:     nid,
		UserID: uid,
	})
	if err != nil {
		return domain.Notification{}, err
	}
	return domain.Notification{
		ID:        row.ID.String(),
		UserID:    row.UserID.String(),
		EventType: row.EventType,
		Title:     row.Title,
		Body:      row.Body,
		Read:      row.Read,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) (int64, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return 0, err
	}
	return r.q.MarkAllNotificationsRead(ctx, uid)
}