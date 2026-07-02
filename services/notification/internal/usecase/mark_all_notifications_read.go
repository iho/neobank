package usecase

import (
	"context"
	"fmt"
)

type MarkAllNotificationsReadUseCase struct {
	repo NotificationRepository
}

func NewMarkAllNotificationsReadUseCase(repo NotificationRepository) *MarkAllNotificationsReadUseCase {
	return &MarkAllNotificationsReadUseCase{repo: repo}
}

func (uc *MarkAllNotificationsReadUseCase) Execute(ctx context.Context, userID string) (int64, error) {
	if userID == "" {
		return 0, fmt.Errorf("user_id is required")
	}
	return uc.repo.MarkAllRead(ctx, userID)
}