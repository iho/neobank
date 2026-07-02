package usecase

import (
	"context"
	"time"

	"github.com/iho/neobank/pkg/pagination"
	"github.com/iho/neobank/services/notification/internal/domain"
)

type NotificationList struct {
	Notifications []domain.Notification
	UnreadCount   int64
	NextCursor    string
}

type ListNotificationsUseCase struct {
	repo NotificationRepository
}

func NewListNotificationsUseCase(repo NotificationRepository) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{repo: repo}
}

func (uc *ListNotificationsUseCase) Execute(ctx context.Context, userID string, limit int, cursor string) (NotificationList, error) {
	if limit <= 0 {
		limit = 20
	}
	pageSize := limit + 1
	var cursorAt *time.Time
	var cursorID string
	if cursor != "" {
		decoded, err := pagination.Decode(cursor)
		if err != nil {
			return NotificationList{}, err
		}
		at := decoded.CreatedAt
		cursorAt = &at
		cursorID = decoded.ID
	}
	notifications, err := uc.repo.ListByUser(ctx, userID, pageSize, cursorAt, cursorID)
	if err != nil {
		return NotificationList{}, err
	}
	items, next := pagination.Trim(notifications, limit, func(n domain.Notification) pagination.Cursor {
		return pagination.Cursor{CreatedAt: n.CreatedAt, ID: n.ID}
	})
	unread, err := uc.repo.CountUnread(ctx, userID)
	if err != nil {
		return NotificationList{}, err
	}
	return NotificationList{Notifications: items, UnreadCount: unread, NextCursor: next}, nil
}