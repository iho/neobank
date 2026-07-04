package domain

import "time"

const (
	BankTransferOrderStatusProcessing = "processing"
	BankTransferOrderStatusSettled    = "settled"
	BankTransferOrderStatusReturned   = "returned"
	BankTransferOrderStatusFailed     = "failed"
)

// BankTransferOrder is one neobank-initiated outbound bank transfer, keyed
// by the rails simulator's payment ID so a redelivered settle/return
// webhook is a no-op.
type BankTransferOrder struct {
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ID               string
	RailsPaymentID   string
	UserID           string
	Amount           string
	Currency         string
	CounterpartyIBAN string
	Reference        string
	LedgerTransferID string
	ReturnTransferID string
	Status           string
}
