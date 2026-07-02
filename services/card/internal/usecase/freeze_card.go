package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/port"
	"github.com/jackc/pgx/v5"
)

type FreezeCardUseCase struct {
	cards  port.CardRepository
	outbox outbox.TxPublisher
	audit  audit.Recorder
	tx     *pgutil.TxRunner
}

func NewFreezeCardUseCase(cards port.CardRepository, outboxPublisher outbox.TxPublisher, auditRecorder audit.Recorder, tx *pgutil.TxRunner) *FreezeCardUseCase {
	return &FreezeCardUseCase{cards: cards, outbox: outboxPublisher, audit: auditRecorder, tx: tx}
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

	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.cards.WithTx(tx).UpdateStatus(ctx, cardID, userID, domain.CardStatusFrozen); err != nil {
			return err
		}
		if err := uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "card",
			EntityID:   cardID,
			Action:     "frozen",
			FromStatus: string(domain.CardStatusActive),
			ToStatus:   string(domain.CardStatusFrozen),
		}); err != nil {
			return err
		}
		return uc.outbox.WithTx(tx).Publish(ctx, events.CardFrozen{CardID: cardID, UserID: userID})
	}); err != nil {
		return nil, err
	}
	return uc.cards.GetByID(ctx, cardID)
}
