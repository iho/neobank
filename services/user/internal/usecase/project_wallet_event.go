package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/walletprojection"
	"github.com/iho/neobank/services/user/internal/domain"
)

type WalletTransactionRepository interface {
	Insert(ctx context.Context, row walletprojection.Row) error
	ApplyCapture(ctx context.Context, update walletprojection.CaptureUpdate) error
	ListByUser(ctx context.Context, userID string, limit int) ([]domain.WalletTransaction, error)
}

type ConsumerInboxRepository interface {
	Exists(ctx context.Context, eventID string) (bool, error)
	Record(ctx context.Context, eventID, eventType string) error
}

type ProjectWalletEventUseCase struct {
	repo  WalletTransactionRepository
	inbox ConsumerInboxRepository
}

func NewProjectWalletEventUseCase(repo WalletTransactionRepository, inbox ConsumerInboxRepository) *ProjectWalletEventUseCase {
	return &ProjectWalletEventUseCase{repo: repo, inbox: inbox}
}

func (uc *ProjectWalletEventUseCase) Execute(ctx context.Context, envelope events.Envelope) error {
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

	rows, capture, err := walletprojection.Apply(envelope)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := uc.repo.Insert(ctx, row); err != nil {
			return fmt.Errorf("insert wallet tx: %w", err)
		}
	}
	if capture != nil {
		if err := uc.repo.ApplyCapture(ctx, *capture); err != nil {
			return fmt.Errorf("capture wallet tx: %w", err)
		}
	}

	if uc.inbox != nil {
		if err := uc.inbox.Record(ctx, envelope.EventID, envelope.EventType); err != nil {
			return fmt.Errorf("record consumer inbox: %w", err)
		}
	}

	return nil
}