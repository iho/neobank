package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Email          string
	Phone          string
	Password       string
	InviteCode     string
	IdempotencyKey string
}

type RegisterOutput struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type TokenIssuer interface {
	Issue(userID string) (access, refresh string, err error)
}

type RegisterUseCase struct {
	users        UserRepository
	tokens       TokenIssuer
	acceptInvite *AcceptReferralInviteUseCase
}

func NewRegisterUseCase(users UserRepository, tokens TokenIssuer, acceptInvite *AcceptReferralInviteUseCase) *RegisterUseCase {
	return &RegisterUseCase{users: users, tokens: tokens, acceptInvite: acceptInvite}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, in RegisterInput) (RegisterOutput, error) {
	if in.Email == "" || in.Password == "" {
		return RegisterOutput{}, fmt.Errorf("email and password are required")
	}

	existing, err := uc.users.GetByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return RegisterOutput{}, err
	}
	if existing != nil {
		return RegisterOutput{}, fmt.Errorf("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("hash password: %w", err)
	}

	userID := uuid.NewString()
	user := domain.User{
		ID:           userID,
		Email:        in.Email,
		Phone:        in.Phone,
		PasswordHash: string(hash),
		Status:       domain.UserStatusActive,
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return RegisterOutput{}, err
	}

	if in.InviteCode != "" && uc.acceptInvite != nil {
		_ = uc.acceptInvite.Execute(ctx, in.InviteCode, userID)
	}

	access, refresh, err := uc.tokens.Issue(userID)
	if err != nil {
		return RegisterOutput{}, err
	}

	return RegisterOutput{
		UserID:       userID,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}