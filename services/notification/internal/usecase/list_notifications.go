package usecase

import (
	"context"

	"github.com/iho/neobank/services/notification/internal/domain"
)

type NotificationList struct {
	Notifications []domain.Notification
	UnreadCount   int64
}

type ListNotificationsUseCase struct {
	repo NotificationRepository
}

func NewListNotificationsUseCase(repo NotificationRepository) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{repo: repo}
}

func (uc *ListNotificationsUseCase) Execute(ctx context.Context, userID string, limit int) (NotificationList, error) {
	if limit <= 0 {
		limit = 20
	}
	notifications, err := uc.repo.ListByUser(ctx, userID, limit)
	if err != nil {
		return NotificationList{}, err
	}
	unread, err := uc.repo.CountUnread(ctx, userID)
	if err != nil {
		return NotificationList{}, err
	}
	return NotificationList{Notifications: notifications, UnreadCount: unread}, nil
}