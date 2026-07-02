package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

type PasswordRepository interface {
	UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error
}

type ChangePasswordUseCase struct {
	users PasswordRepository
	lookup UserRepository
}

func NewChangePasswordUseCase(users PasswordRepository, lookup UserRepository) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{users: users, lookup: lookup}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, in ChangePasswordInput) error {
	if in.UserID == "" || in.CurrentPassword == "" || in.NewPassword == "" {
		return fmt.Errorf("user_id, current_password, and new_password are required")
	}
	if len(in.NewPassword) < 8 {
		return fmt.Errorf("new_password must be at least 8 characters")
	}

	user, err := uc.lookup.GetByID(ctx, in.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		return err
	}
	if user.Status == domain.UserStatusSuspended {
		return fmt.Errorf("account suspended")
	}
	if user.Status == domain.UserStatusMasked {
		return fmt.Errorf("account closed")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.CurrentPassword)); err != nil {
		return fmt.Errorf("invalid current password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return uc.users.UpdatePasswordHash(ctx, in.UserID, string(hash))
}