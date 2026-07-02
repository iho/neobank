package domain

import "time"

type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "pending"
	TransferStatusCompleted TransferStatus = "completed"
	TransferStatusFailed    TransferStatus = "failed"
	TransferStatusReversed  TransferStatus = "reversed"
)

type Transfer struct {
	ID               string
	IdempotencyKey   string
	Type             string
	Status           TransferStatus
	SenderUserID     string
	RecipientUserID  string
	Amount           string
	Currency         string
	Memo             string
	LedgerTransferID string
	FailureReason    string
	CreatedAt        time.Time
	CompletedAt      *time.Time
}