package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/services/notification/internal/domain"
)

type PreferencesRepository interface {
	Get(ctx context.Context, userID string) (domain.NotificationPreferences, error)
	Upsert(ctx context.Context, prefs domain.NotificationPreferences) (domain.NotificationPreferences, error)
}

type GetNotificationPreferencesUseCase struct {
	repo PreferencesRepository
}

func NewGetNotificationPreferencesUseCase(repo PreferencesRepository) *GetNotificationPreferencesUseCase {
	return &GetNotificationPreferencesUseCase{repo: repo}
}

func (uc *GetNotificationPreferencesUseCase) Execute(ctx context.Context, userID string) (domain.NotificationPreferences, error) {
	if userID == "" {
		return domain.NotificationPreferences{}, fmt.Errorf("user_id is required")
	}
	return uc.repo.Get(ctx, userID)
}

type UpdateNotificationPreferencesInput struct {
	UserID    string
	Transfers *bool
	Cards     *bool
	KYC       *bool
	Push      *bool
	Email     *bool
}

type UpdateNotificationPreferencesUseCase struct {
	repo PreferencesRepository
}

func NewUpdateNotificationPreferencesUseCase(repo PreferencesRepository) *UpdateNotificationPreferencesUseCase {
	return &UpdateNotificationPreferencesUseCase{repo: repo}
}

func (uc *UpdateNotificationPreferencesUseCase) Execute(ctx context.Context, in UpdateNotificationPreferencesInput) (domain.NotificationPreferences, error) {
	if in.UserID == "" {
		return domain.NotificationPreferences{}, fmt.Errorf("user_id is required")
	}
	current, err := uc.repo.Get(ctx, in.UserID)
	if err != nil {
		return domain.NotificationPreferences{}, err
	}
	if in.Transfers != nil {
		current.Transfers = *in.Transfers
	}
	if in.Cards != nil {
		current.Cards = *in.Cards
	}
	if in.KYC != nil {
		current.KYC = *in.KYC
	}
	if in.Push != nil {
		current.Push = *in.Push
	}
	if in.Email != nil {
		current.Email = *in.Email
	}
	return uc.repo.Upsert(ctx, current)
}

func CategoryForEventType(eventType string) string {
	switch eventType {
	case events.TypeTransferCompleted:
		return "transfers"
	case events.TypeCardIssued, events.TypeCardFrozen, events.TypeCardUnfrozen,
		events.TypeCardAuthApproved, events.TypeCardAuthCaptured:
		return "cards"
	case events.TypeKYCApproved, events.TypeKYCRejected,
		events.TypeWalletProvisioned, events.TypeDepositCompleted:
		return "kyc"
	default:
		return ""
	}
}

func AllowsCategory(prefs domain.NotificationPreferences, category string) bool {
	switch category {
	case "transfers":
		return prefs.Transfers
	case "cards":
		return prefs.Cards
	case "kyc":
		return prefs.KYC
	default:
		return true
	}
}