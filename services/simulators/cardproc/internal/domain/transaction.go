package domain

import "time"

const (
	TransactionStatusApproved = "approved"
	TransactionStatusDeclined = "declined"
	TransactionStatusCaptured = "captured"
	TransactionStatusReversed = "reversed"
	TransactionStatusExpired  = "expired"
)

// Transaction is one simulated merchant charge against a card. AuthorizationID
// is the neobank's own authorization ID, learned from the synchronous auth
// call's response once approved, and carried in later async webhooks so the
// card service can reference the record it created rather than ours.
type Transaction struct {
	CreatedAt       time.Time
	CapturedAt      *time.Time
	ReversedAt      *time.Time
	ID              string
	CardID          string
	AuthorizationID string
	Amount          string
	Currency        string
	MerchantName    string
	MCC             string
	Status          string
	ReasonCode      string
}
