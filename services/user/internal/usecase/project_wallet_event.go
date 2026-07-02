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

type ProjectWalletEventUseCase struct {
	repo WalletTransactionRepository
}

func NewProjectWalletEventUseCase(repo WalletTransactionRepository) *ProjectWalletEventUseCase {
	return &ProjectWalletEventUseCase{repo: repo}
}

func (uc *ProjectWalletEventUseCase) Execute(ctx context.Context, envelope events.Envelope) error {
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
	return nil
}