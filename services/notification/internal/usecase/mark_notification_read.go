package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/iho/neobank/services/notification/internal/domain"
	"github.com/jackc/pgx/v5"
)

type MarkNotificationReadUseCase struct {
	repo NotificationRepository
}

func NewMarkNotificationReadUseCase(repo NotificationRepository) *MarkNotificationReadUseCase {
	return &MarkNotificationReadUseCase{repo: repo}
}

func (uc *MarkNotificationReadUseCase) Execute(ctx context.Context, userID, notificationID string) (domain.Notification, error) {
	if userID == "" || notificationID == "" {
		return domain.Notification{}, fmt.Errorf("user_id and notification_id are required")
	}
	n, err := uc.repo.MarkRead(ctx, userID, notificationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Notification{}, pgx.ErrNoRows
		}
		return domain.Notification{}, err
	}
	return n, nil
}