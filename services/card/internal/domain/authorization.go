package domain

import "time"

type AuthStatus string

const (
	AuthStatusAuthorized AuthStatus = "authorized"
	AuthStatusCaptured   AuthStatus = "captured"
	AuthStatusDeclined   AuthStatus = "declined"
	AuthStatusVoided     AuthStatus = "voided"
)

type Authorization struct {
	ID               string
	CardID           string
	UserID           string
	IdempotencyKey   string
	MerchantName     string
	Amount           string
	Currency         string
	Status           AuthStatus
	LedgerHoldID     string
	LedgerTransferID string
	FailureReason    string
	CreatedAt        time.Time
	CapturedAt       *time.Time
}