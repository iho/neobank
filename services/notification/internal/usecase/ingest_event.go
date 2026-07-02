package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/services/notification/internal/domain"
)

type IngestEventUseCase struct {
	repo NotificationRepository
}

func NewIngestEventUseCase(repo NotificationRepository) *IngestEventUseCase {
	return &IngestEventUseCase{repo: repo}
}

func (uc *IngestEventUseCase) Execute(ctx context.Context, envelope events.Envelope) error {
	switch envelope.EventType {
	case events.TypeTransferCompleted:
		return uc.handleTransferCompleted(ctx, envelope)
	case events.TypeCardIssued:
		return uc.handleCardIssued(ctx, envelope)
	case events.TypeCardFrozen:
		return uc.handleCardStatus(ctx, envelope, "Card frozen", "Your card has been frozen.")
	case events.TypeCardUnfrozen:
		return uc.handleCardStatus(ctx, envelope, "Card unfrozen", "Your card is active again.")
	case events.TypeCardAuthApproved:
		return uc.handleCardAuthApproved(ctx, envelope)
	case events.TypeCardAuthCaptured:
		return uc.handleCardAuthCaptured(ctx, envelope)
	default:
		return nil
	}
}

func (uc *IngestEventUseCase) handleCardIssued(ctx context.Context, envelope events.Envelope) error {
	var payload events.CardIssued
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse card issued payload: %w", err)
	}
	return uc.repo.Create(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.UserID,
		EventType: envelope.EventType,
		Title:     "Virtual card issued",
		Body:      fmt.Sprintf("Your new card ending in %s is ready to use.", payload.LastFour),
	}, envelope.EventID)
}

func (uc *IngestEventUseCase) handleCardAuthApproved(ctx context.Context, envelope events.Envelope) error {
	var payload events.CardAuthApproved
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse card auth approved payload: %w", err)
	}
	merchant := payload.MerchantName
	if merchant == "" {
		merchant = "a merchant"
	}
	return uc.repo.Create(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.UserID,
		EventType: envelope.EventType,
		Title:     "Card purchase authorized",
		Body:      fmt.Sprintf("%s %s pending at %s.", payload.Amount, payload.Currency, merchant),
	}, envelope.EventID)
}

func (uc *IngestEventUseCase) handleCardAuthCaptured(ctx context.Context, envelope events.Envelope) error {
	var payload events.CardAuthCaptured
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse card auth captured payload: %w", err)
	}
	return uc.repo.Create(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.UserID,
		EventType: envelope.EventType,
		Title:     "Card purchase completed",
		Body:      fmt.Sprintf("%s %s was charged to your card.", payload.Amount, payload.Currency),
	}, envelope.EventID)
}

func (uc *IngestEventUseCase) handleCardStatus(ctx context.Context, envelope events.Envelope, title, body string) error {
	var payload events.CardFrozen
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse card status payload: %w", err)
	}
	return uc.repo.Create(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.UserID,
		EventType: envelope.EventType,
		Title:     title,
		Body:      body,
	}, envelope.EventID)
}

func (uc *IngestEventUseCase) handleTransferCompleted(ctx context.Context, envelope events.Envelope) error {
	var payload events.TransferCompleted
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse transfer completed payload: %w", err)
	}

	senderTitle := "Transfer sent"
	senderBody := fmt.Sprintf("You sent %s %s", payload.Amount, payload.Currency)
	recipientTitle := "Transfer received"
	recipientBody := fmt.Sprintf("You received %s %s", payload.Amount, payload.Currency)

	if err := uc.repo.Create(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.SenderUserID,
		EventType: envelope.EventType,
		Title:     senderTitle,
		Body:      senderBody,
	}, envelope.EventID); err != nil {
		return err
	}

	return uc.repo.Create(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.RecipientUserID,
		EventType: envelope.EventType,
		Title:     recipientTitle,
		Body:      recipientBody,
	}, envelope.EventID)
}