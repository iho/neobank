package usecase

import (
	"context"

	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/port"
)

type ListAuthorizationsUseCase struct {
	auths port.AuthorizationRepository
}

func NewListAuthorizationsUseCase(auths port.AuthorizationRepository) *ListAuthorizationsUseCase {
	return &ListAuthorizationsUseCase{auths: auths}
}

func (uc *ListAuthorizationsUseCase) Execute(ctx context.Context, userID string, limit int) ([]domain.Authorization, error) {
	return uc.auths.ListByUser(ctx, userID, limit)
}
