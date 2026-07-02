package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/services/user/internal/domain"
)

type ProfileReader interface {
	GetProfile(ctx context.Context, userID string) (*domain.Profile, error)
}

type GetProfileUseCase struct {
	profiles ProfileReader
}

func NewGetProfileUseCase(profiles ProfileReader) *GetProfileUseCase {
	return &GetProfileUseCase{profiles: profiles}
}

func (uc *GetProfileUseCase) Execute(ctx context.Context, userID string) (*domain.Profile, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	return uc.profiles.GetProfile(ctx, userID)
}