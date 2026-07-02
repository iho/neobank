package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/services/card/internal/domain"
)

type FreezeCardUseCase struct {
	cards  CardRepository
	outbox OutboxPublisher
}

func NewFreezeCardUseCase(cards CardRepository, outbox OutboxPublisher) *FreezeCardUseCase {
	return &FreezeCardUseCase{cards: cards, outbox: outbox}
}

func (uc *FreezeCardUseCase) Execute(ctx context.Context, userID, cardID string) (*domain.Card, error) {
	card, err := uc.cards.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}
	if card.UserID != userID {
		return nil, fmt.Errorf("card not found")
	}
	if card.Status == domain.CardStatusFrozen {
		return card, nil
	}
	if card.Status != domain.CardStatusActive {
		return nil, fmt.Errorf("card cannot be frozen")
	}

	if err := uc.cards.UpdateStatus(ctx, cardID, userID, domain.CardStatusFrozen); err != nil {
		return nil, err
	}
	_ = uc.outbox.Publish(ctx, events.CardFrozen{CardID: cardID, UserID: userID})
	return uc.cards.GetByID(ctx, cardID)
}