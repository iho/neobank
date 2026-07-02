package domain

import "time"

type DepositStatus string

const (
	DepositStatusCompleted DepositStatus = "completed"
	DepositStatusFailed    DepositStatus = "failed"
)

type Deposit struct {
	ID               string
	UserID           string
	WalletID         string
	Amount           string
	Currency         string
	LedgerTransferID string
	Status           DepositStatus
	IdempotencyKey   string
	CreatedAt        time.Time
	CompletedAt      *time.Time
}