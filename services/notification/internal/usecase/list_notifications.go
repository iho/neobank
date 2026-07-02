package usecase

import (
	"context"

	"github.com/iho/neobank/services/notification/internal/domain"
)

type ListNotificationsUseCase struct {
	repo NotificationRepository
}

func NewListNotificationsUseCase(repo NotificationRepository) *ListNotificationsUseCase {
	return &ListNotificationsUseCase{repo: repo}
}

func (uc *ListNotificationsUseCase) Execute(ctx context.Context, userID string, limit int) ([]domain.Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.repo.ListByUser(ctx, userID, limit)
}