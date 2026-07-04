package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
	"github.com/iho/neobank/services/simulators/cardproc/internal/port"
)

// ReverseTransactionUseCase is the admin entry point for voiding a
// previously-approved, auth-only transaction without ever capturing it.
type ReverseTransactionUseCase struct {
	txs        port.TransactionRepository
	dispatcher WebhookDispatcher
	eventsURL  string
}

func NewReverseTransactionUseCase(txs port.TransactionRepository, dispatcher WebhookDispatcher, eventsURL string) *ReverseTransactionUseCase {
	return &ReverseTransactionUseCase{txs: txs, dispatcher: dispatcher, eventsURL: eventsURL}
}

func (uc *ReverseTransactionUseCase) Execute(ctx context.Context, transactionID, reason string) (domain.Transaction, error) {
	tx, err := uc.txs.GetByID(ctx, transactionID)
	if err != nil {
		return domain.Transaction{}, err
	}

	if tx == nil {
		return domain.Transaction{}, fmt.Errorf("transaction %q not found", transactionID)
	}

	if tx.Status != domain.TransactionStatusApproved {
		return domain.Transaction{}, fmt.Errorf("transaction cannot be reversed from status %q", tx.Status)
	}

	if _, err := uc.dispatcher.Enqueue(ctx, uc.eventsURL, EventAuthReversed, CardEventPayload{
		AuthorizationID: tx.AuthorizationID,
		Reason:          reason,
	}); err != nil {
		return domain.Transaction{}, fmt.Errorf("schedule reversal webhook: %w", err)
	}

	if err := uc.txs.MarkReversed(ctx, tx.ID); err != nil {
		return domain.Transaction{}, err
	}

	updated, err := uc.txs.GetByID(ctx, tx.ID)
	if err != nil {
		return domain.Transaction{}, err
	}

	return *updated, nil
}
