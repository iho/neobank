package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/mcc"
	"github.com/iho/neobank/services/notification/internal/domain"
)

type IngestEventUseCase struct {
	repo     NotificationRepository
	inbox    ConsumerInboxRepository
	prefs    PreferencesRepository
	delivery *DeliveryService
}

func NewIngestEventUseCase(
	repo NotificationRepository,
	inbox ConsumerInboxRepository,
	prefs PreferencesRepository,
	delivery *DeliveryService,
) *IngestEventUseCase {
	return &IngestEventUseCase{repo: repo, inbox: inbox, prefs: prefs, delivery: delivery}
}

func (uc *IngestEventUseCase) Execute(ctx context.Context, envelope events.Envelope) error {
	if envelope.EventID == "" {
		return fmt.Errorf("event_id is required")
	}
	if uc.inbox != nil {
		exists, err := uc.inbox.Exists(ctx, envelope.EventID)
		if err != nil {
			return fmt.Errorf("check consumer inbox: %w", err)
		}
		if exists {
			return nil
		}
	}

	if err := uc.dispatch(ctx, envelope); err != nil {
		return err
	}

	if uc.inbox != nil {
		if err := uc.inbox.Record(ctx, envelope.EventID, envelope.EventType); err != nil {
			return fmt.Errorf("record consumer inbox: %w", err)
		}
	}

	return nil
}

func (uc *IngestEventUseCase) dispatch(ctx context.Context, envelope events.Envelope) error {
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
	case events.TypeKYCApproved:
		return uc.handleKYCApproved(ctx, envelope)
	case events.TypeKYCRejected:
		return uc.handleKYCRejected(ctx, envelope)
	case events.TypeWalletProvisioned:
		return uc.handleWalletProvisioned(ctx, envelope)
	case events.TypeDepositCompleted:
		return uc.handleDepositCompleted(ctx, envelope)
	default:
		return nil
	}
}

func (uc *IngestEventUseCase) shouldCreate(ctx context.Context, userID, eventType string) (bool, error) {
	if uc.prefs == nil {
		return true, nil
	}
	prefs, err := uc.prefs.Get(ctx, userID)
	if err != nil {
		return false, err
	}
	category := CategoryForEventType(eventType)
	if category == "" {
		return true, nil
	}
	return AllowsCategory(prefs, category), nil
}

func (uc *IngestEventUseCase) createNotification(ctx context.Context, n domain.Notification, eventID string) error {
	ok, err := uc.shouldCreate(ctx, n.UserID, n.EventType)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if err := uc.repo.Create(ctx, n, eventID); err != nil {
		return err
	}
	if uc.delivery != nil {
		_ = uc.delivery.Deliver(ctx, n.UserID, n.EventType, n.Title, n.Body)
	}
	return nil
}

func (uc *IngestEventUseCase) handleKYCApproved(ctx context.Context, envelope events.Envelope) error {
	var payload events.KYCApproved
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse kyc approved payload: %w", err)
	}
	return uc.createNotification(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.UserID,
		EventType: envelope.EventType,
		Title:     "KYC approved",
		Body:      "Your identity verification is complete. Your wallet is ready to use.",
	}, envelope.EventID)
}

func (uc *IngestEventUseCase) handleKYCRejected(ctx context.Context, envelope events.Envelope) error {
	var payload events.KYCRejected
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse kyc rejected payload: %w", err)
	}
	body := "Your identity verification was not approved."
	if payload.RejectionReason != "" {
		body = fmt.Sprintf("Your identity verification was not approved: %s.", payload.RejectionReason)
	}
	return uc.createNotification(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.UserID,
		EventType: envelope.EventType,
		Title:     "KYC rejected",
		Body:      body,
	}, envelope.EventID)
}

func (uc *IngestEventUseCase) handleWalletProvisioned(ctx context.Context, envelope events.Envelope) error {
	var payload events.WalletProvisioned
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse wallet provisioned payload: %w", err)
	}
	return uc.createNotification(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.UserID,
		EventType: envelope.EventType,
		Title:     "Wallet ready",
		Body:      fmt.Sprintf("Your %s wallet is active and ready for transfers.", payload.Currency),
	}, envelope.EventID)
}

func (uc *IngestEventUseCase) handleDepositCompleted(ctx context.Context, envelope events.Envelope) error {
	var payload events.DepositCompleted
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse deposit completed payload: %w", err)
	}
	return uc.createNotification(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.UserID,
		EventType: envelope.EventType,
		Title:     "Deposit received",
		Body:      fmt.Sprintf("%s %s was added to your wallet.", payload.Amount, payload.Currency),
	}, envelope.EventID)
}

func (uc *IngestEventUseCase) handleCardIssued(ctx context.Context, envelope events.Envelope) error {
	var payload events.CardIssued
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse card issued payload: %w", err)
	}
	return uc.createNotification(ctx, domain.Notification{
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
	category := mcc.CategoryLabel(payload.MerchantCategoryCode)
	return uc.createNotification(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.UserID,
		EventType: envelope.EventType,
		Title:     "Card purchase authorized",
		Body:      fmt.Sprintf("%s %s pending at %s (%s).", payload.Amount, payload.Currency, merchant, category),
	}, envelope.EventID)
}

func (uc *IngestEventUseCase) handleCardAuthCaptured(ctx context.Context, envelope events.Envelope) error {
	var payload events.CardAuthCaptured
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("parse card auth captured payload: %w", err)
	}
	return uc.createNotification(ctx, domain.Notification{
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
	return uc.createNotification(ctx, domain.Notification{
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

	if err := uc.createNotification(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.SenderUserID,
		EventType: envelope.EventType,
		Title:     senderTitle,
		Body:      senderBody,
	}, envelope.EventID); err != nil {
		return err
	}

	return uc.createNotification(ctx, domain.Notification{
		ID:        uuid.NewString(),
		UserID:    payload.RecipientUserID,
		EventType: envelope.EventType,
		Title:     recipientTitle,
		Body:      recipientBody,
	}, envelope.EventID)
}