package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/services/card/internal/domain"
)

type UnfreezeCardUseCase struct {
	cards  CardRepository
	outbox OutboxPublisher
}

func NewUnfreezeCardUseCase(cards CardRepository, outbox OutboxPublisher) *UnfreezeCardUseCase {
	return &UnfreezeCardUseCase{cards: cards, outbox: outbox}
}

func (uc *UnfreezeCardUseCase) Execute(ctx context.Context, userID, cardID string) (*domain.Card, error) {
	card, err := uc.cards.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if card.UserID != userID {
		return nil, fmt.Errorf("card not found")
	}
	if card.Status == domain.CardStatusActive {
		return card, nil
	}
	if card.Status != domain.CardStatusFrozen {
		return nil, fmt.Errorf("card cannot be unfrozen")
	}

	if err := uc.cards.UpdateStatus(ctx, cardID, userID, domain.CardStatusActive); err != nil {
		return nil, err
	}
	_ = uc.outbox.Publish(ctx, events.CardUnfrozen{CardID: cardID, UserID: userID})
	return uc.cards.GetByID(ctx, cardID)
}