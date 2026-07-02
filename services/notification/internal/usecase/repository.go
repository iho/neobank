package usecase

import (
	"context"

	"github.com/iho/neobank/services/notification/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, n domain.Notification, eventID string) error
	ListByUser(ctx context.Context, userID string, limit int) ([]domain.Notification, error)
	CountUnread(ctx context.Context, userID string) (int64, error)
	MarkRead(ctx context.Context, userID, notificationID string) (domain.Notification, error)
	MarkAllRead(ctx context.Context, userID string) (int64, error)
}

type ConsumerInboxRepository interface {
	Exists(ctx context.Context, eventID string) (bool, error)
	Record(ctx context.Context, eventID, eventType string) error
}