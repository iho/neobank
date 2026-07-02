package usecase

import (
	"context"

	"github.com/iho/neobank/services/card/internal/domain"
)

type ListAuthorizationsUseCase struct {
	auths AuthorizationRepository
}

func NewListAuthorizationsUseCase(auths AuthorizationRepository) *ListAuthorizationsUseCase {
	return &ListAuthorizationsUseCase{auths: auths}
}

func (uc *ListAuthorizationsUseCase) Execute(ctx context.Context, userID string, limit int) ([]domain.Authorization, error) {
	return uc.auths.ListByUser(ctx, userID, limit)
}