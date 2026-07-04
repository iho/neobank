package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
	"github.com/iho/neobank/services/simulators/cardproc/internal/port"
)

// CaptureTransactionUseCase is the admin entry point for settling a
// previously-approved, auth-only transaction: it schedules the same
// EventCaptured webhook SimulateTransactionUseCase would have sent
// immediately if Capture had been requested up front.
type CaptureTransactionUseCase struct {
	txs        port.TransactionRepository
	dispatcher WebhookDispatcher
	eventsURL  string
}

func NewCaptureTransactionUseCase(txs port.TransactionRepository, dispatcher WebhookDispatcher, eventsURL string) *CaptureTransactionUseCase {
	return &CaptureTransactionUseCase{txs: txs, dispatcher: dispatcher, eventsURL: eventsURL}
}

func (uc *CaptureTransactionUseCase) Execute(ctx context.Context, transactionID string) (domain.Transaction, error) {
	tx, err := uc.txs.GetByID(ctx, transactionID)
	if err != nil {
		return domain.Transaction{}, err
	}

	if tx == nil {
		return domain.Transaction{}, fmt.Errorf("transaction %q not found", transactionID)
	}

	if tx.Status != domain.TransactionStatusApproved {
		return domain.Transaction{}, fmt.Errorf("transaction cannot be captured from status %q", tx.Status)
	}

	if _, err := uc.dispatcher.Enqueue(ctx, uc.eventsURL, EventCaptured, CardEventPayload{
		AuthorizationID: tx.AuthorizationID,
	}); err != nil {
		return domain.Transaction{}, fmt.Errorf("schedule capture webhook: %w", err)
	}

	if err := uc.txs.MarkCaptured(ctx, tx.ID); err != nil {
		return domain.Transaction{}, err
	}

	updated, err := uc.txs.GetByID(ctx, tx.ID)
	if err != nil {
		return domain.Transaction{}, err
	}

	return *updated, nil
}
