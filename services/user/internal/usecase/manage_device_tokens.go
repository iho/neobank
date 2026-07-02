package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/services/user/internal/domain"
)

type DeviceTokenRepository interface {
	Upsert(ctx context.Context, userID, platform, token string) (*domain.DeviceToken, error)
	Delete(ctx context.Context, userID, tokenID string) (bool, error)
	ListByUser(ctx context.Context, userID string) ([]domain.DeviceToken, error)
}

type RegisterDeviceTokenInput struct {
	UserID   string
	Platform string
	Token    string
}

type RegisterDeviceTokenUseCase struct {
	repo DeviceTokenRepository
}

func NewRegisterDeviceTokenUseCase(repo DeviceTokenRepository) *RegisterDeviceTokenUseCase {
	return &RegisterDeviceTokenUseCase{repo: repo}
}

func (uc *RegisterDeviceTokenUseCase) Execute(ctx context.Context, in RegisterDeviceTokenInput) (*domain.DeviceToken, error) {
	if in.UserID == "" || in.Platform == "" || in.Token == "" {
		return nil, fmt.Errorf("user_id, platform, and token are required")
	}
	switch in.Platform {
	case "ios", "android", "web":
	default:
		return nil, fmt.Errorf("platform must be ios, android, or web")
	}
	return uc.repo.Upsert(ctx, in.UserID, in.Platform, in.Token)
}

type DeleteDeviceTokenUseCase struct {
	repo DeviceTokenRepository
}

func NewDeleteDeviceTokenUseCase(repo DeviceTokenRepository) *DeleteDeviceTokenUseCase {
	return &DeleteDeviceTokenUseCase{repo: repo}
}

func (uc *DeleteDeviceTokenUseCase) Execute(ctx context.Context, userID, tokenID string) error {
	if userID == "" || tokenID == "" {
		return fmt.Errorf("user_id and token_id are required")
	}
	deleted, err := uc.repo.Delete(ctx, userID, tokenID)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("device token not found")
	}
	return nil
}

type ListDeviceTokensUseCase struct {
	repo DeviceTokenRepository
}

func NewListDeviceTokensUseCase(repo DeviceTokenRepository) *ListDeviceTokensUseCase {
	return &ListDeviceTokensUseCase{repo: repo}
}

func (uc *ListDeviceTokensUseCase) Execute(ctx context.Context, userID string) ([]domain.DeviceToken, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	return uc.repo.ListByUser(ctx, userID)
}