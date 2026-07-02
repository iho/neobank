package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/iho/neobank/pkg/auth"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
)

type RefreshTokenUseCase struct {
	jwt   *auth.JWT
	users UserRepository
}

func NewRefreshTokenUseCase(jwt *auth.JWT, users UserRepository) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{jwt: jwt, users: users}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, refreshToken string) (LoginOutput, error) {
	if refreshToken == "" {
		return LoginOutput{}, fmt.Errorf("refresh_token is required")
	}

	access, refresh, userID, err := uc.jwt.Refresh(refreshToken)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("invalid refresh token")
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginOutput{}, fmt.Errorf("invalid refresh token")
		}
		return LoginOutput{}, err
	}
	if user.Status != domain.UserStatusActive {
		return LoginOutput{}, fmt.Errorf("account suspended")
	}

	return LoginOutput{
		UserID:       userID,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}