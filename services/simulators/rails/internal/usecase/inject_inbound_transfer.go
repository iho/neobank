package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/iho/neobank/pkg/vendorsim"
	"github.com/iho/neobank/services/simulators/rails/internal/domain"
	"github.com/iho/neobank/services/simulators/rails/internal/port"
)

// EventTransferReceived is delivered when simulated funds arrive on a rails account.
const EventTransferReceived = "rails.transfer.received"

// WebhookDispatcher schedules a signed webhook delivery; satisfied by
// pkg/vendorsim.Dispatcher.
type WebhookDispatcher interface {
	Enqueue(ctx context.Context, url, eventType string, payload any) (string, error)
	EnqueueAfter(ctx context.Context, url, eventType string, payload any, minDelay time.Duration) (string, error)
}

// TransferReceivedPayload is the webhook body delivered on EventTransferReceived.
type TransferReceivedPayload struct {
	TransferID  string `json:"transfer_id"`
	AccountID   string `json:"account_id"`
	ExternalRef string `json:"external_ref"`
	IBAN        string `json:"iban"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	SenderName  string `json:"sender_name"`
	Reference   string `json:"reference"`
}

type InjectInboundTransferInput struct {
	AccountID  string
	Amount     string
	Currency   string
	SenderName string
	Reference  string
}

// InjectInboundTransferUseCase is the admin/test entry point that simulates
// money arriving on a rail: it records the transfer and schedules a webhook
// to the neobank, per docs/vendor-simulators-plan.md Phase 1.
type InjectInboundTransferUseCase struct {
	accounts   port.AccountRepository
	transfers  port.InboundTransferRepository
	dispatcher WebhookDispatcher
	webhookURL string
}

func NewInjectInboundTransferUseCase(
	accounts port.AccountRepository,
	transfers port.InboundTransferRepository,
	dispatcher WebhookDispatcher,
	webhookURL string,
) *InjectInboundTransferUseCase {
	return &InjectInboundTransferUseCase{
		accounts:   accounts,
		transfers:  transfers,
		dispatcher: dispatcher,
		webhookURL: webhookURL,
	}
}

func (uc *InjectInboundTransferUseCase) Execute(ctx context.Context, in InjectInboundTransferInput) (domain.InboundTransfer, error) {
	if in.AccountID == "" || in.Amount == "" || in.Currency == "" || in.SenderName == "" {
		return domain.InboundTransfer{}, fmt.Errorf("account_id, amount, currency, and sender_name are required")
	}

	account, err := uc.accounts.GetByID(ctx, in.AccountID)
	if err != nil {
		return domain.InboundTransfer{}, err
	}

	if account == nil {
		return domain.InboundTransfer{}, fmt.Errorf("account %q not found", in.AccountID)
	}

	transfer, err := uc.transfers.Create(ctx, account.ID, in.Amount, in.Currency, in.SenderName, in.Reference)
	if err != nil {
		return domain.InboundTransfer{}, err
	}

	_, err = uc.dispatcher.Enqueue(ctx, uc.webhookURL, EventTransferReceived, TransferReceivedPayload{
		TransferID:  transfer.ID,
		AccountID:   account.ID,
		ExternalRef: account.ExternalRef,
		IBAN:        account.IBAN,
		Amount:      transfer.Amount,
		Currency:    transfer.Currency,
		SenderName:  transfer.SenderName,
		Reference:   transfer.Reference,
	})
	if err != nil {
		return domain.InboundTransfer{}, fmt.Errorf("schedule webhook: %w", err)
	}

	return transfer, nil
}

var _ WebhookDispatcher = (*vendorsim.Dispatcher)(nil)
