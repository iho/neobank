package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
)

type SavedPayeeRepository interface {
	ListByUser(ctx context.Context, userID string, limit int) ([]domain.SavedPayee, error)
	Upsert(ctx context.Context, userID, payeeUserID, nickname string) (*domain.SavedPayee, error)
	Create(ctx context.Context, userID, payeeUserID, nickname string) (*domain.SavedPayee, error)
	GetByID(ctx context.Context, userID, payeeID string) (*domain.SavedPayee, error)
	Delete(ctx context.Context, userID, payeeID string) (bool, error)
}

type ListPayeesUseCase struct {
	repo SavedPayeeRepository
}

func NewListPayeesUseCase(repo SavedPayeeRepository) *ListPayeesUseCase {
	return &ListPayeesUseCase{repo: repo}
}

func (uc *ListPayeesUseCase) Execute(ctx context.Context, userID string, limit int) ([]domain.SavedPayee, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if limit <= 0 {
		limit = 20
	}
	return uc.repo.ListByUser(ctx, userID, limit)
}

type CreatePayeeInput struct {
	UserID      string
	PayeeUserID string
	Nickname    string
}

type CreatePayeeUseCase struct {
	repo SavedPayeeRepository
}

func NewCreatePayeeUseCase(repo SavedPayeeRepository) *CreatePayeeUseCase {
	return &CreatePayeeUseCase{repo: repo}
}

func (uc *CreatePayeeUseCase) Execute(ctx context.Context, in CreatePayeeInput) (*domain.SavedPayee, error) {
	if in.UserID == "" || in.PayeeUserID == "" {
		return nil, fmt.Errorf("user_id and payee_user_id are required")
	}
	if in.UserID == in.PayeeUserID {
		return nil, fmt.Errorf("cannot save yourself as a payee")
	}
	payee, err := uc.repo.Create(ctx, in.UserID, in.PayeeUserID, in.Nickname)
	if err != nil {
		return nil, err
	}
	if payee == nil {
		return uc.repo.Upsert(ctx, in.UserID, in.PayeeUserID, in.Nickname)
	}
	return payee, nil
}

type UpsertPayeeUseCase struct {
	repo SavedPayeeRepository
}

func NewUpsertPayeeUseCase(repo SavedPayeeRepository) *UpsertPayeeUseCase {
	return &UpsertPayeeUseCase{repo: repo}
}

func (uc *UpsertPayeeUseCase) Execute(ctx context.Context, userID, payeeUserID, nickname string) (*domain.SavedPayee, error) {
	if userID == "" || payeeUserID == "" {
		return nil, fmt.Errorf("user_id and payee_user_id are required")
	}
	if userID == payeeUserID {
		return nil, fmt.Errorf("cannot save yourself as a payee")
	}
	return uc.repo.Upsert(ctx, userID, payeeUserID, nickname)
}

type DeletePayeeUseCase struct {
	repo SavedPayeeRepository
}

func NewDeletePayeeUseCase(repo SavedPayeeRepository) *DeletePayeeUseCase {
	return &DeletePayeeUseCase{repo: repo}
}

func (uc *DeletePayeeUseCase) Execute(ctx context.Context, userID, payeeID string) error {
	if userID == "" || payeeID == "" {
		return fmt.Errorf("user_id and payee_id are required")
	}
	deleted, err := uc.repo.Delete(ctx, userID, payeeID)
	if err != nil {
		return err
	}
	if !deleted {
		return pgx.ErrNoRows
	}
	return nil
}

func IsPayeeNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}