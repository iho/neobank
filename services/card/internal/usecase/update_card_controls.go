package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/money"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/port"
)

type UpdateCardControlsInput struct {
	UserID     string
	CardID     string
	DailyLimit *string
	OnlineOnly *bool
}

type UpdateCardControlsUseCase struct {
	cards port.CardRepository
}

func NewUpdateCardControlsUseCase(cards port.CardRepository) *UpdateCardControlsUseCase {
	return &UpdateCardControlsUseCase{cards: cards}
}

func (uc *UpdateCardControlsUseCase) Execute(ctx context.Context, in UpdateCardControlsInput) (*domain.Card, error) {
	if in.UserID == "" || in.CardID == "" {
		return nil, fmt.Errorf("user_id and card_id are required")
	}
	if in.DailyLimit == nil && in.OnlineOnly == nil {
		return nil, fmt.Errorf("at least one control must be provided")
	}
	if in.DailyLimit != nil && *in.DailyLimit != "" {
		if _, err := money.Parse(*in.DailyLimit); err != nil {
			return nil, err
		}
	}
	return uc.cards.UpdateControls(ctx, in.CardID, in.UserID, in.DailyLimit, in.OnlineOnly)
}