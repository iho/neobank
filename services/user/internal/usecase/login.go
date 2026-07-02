package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

type LoginUseCase struct {
	users  UserRepository
	tokens TokenIssuer
}

func NewLoginUseCase(users UserRepository, tokens TokenIssuer) *LoginUseCase {
	return &LoginUseCase{users: users, tokens: tokens}
}

func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (LoginOutput, error) {
	if in.Email == "" || in.Password == "" {
		return LoginOutput{}, fmt.Errorf("email and password are required")
	}

	user, err := uc.users.GetByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LoginOutput{}, fmt.Errorf("invalid credentials")
		}
		return LoginOutput{}, err
	}
	if user.Status != domain.UserStatusActive {
		return LoginOutput{}, fmt.Errorf("account suspended")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return LoginOutput{}, fmt.Errorf("invalid credentials")
	}

	access, refresh, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return LoginOutput{}, err
	}

	return LoginOutput{
		UserID:       user.ID,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}