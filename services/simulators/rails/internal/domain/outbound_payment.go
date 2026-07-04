package domain

import "time"

const (
	OutboundPaymentStatusAccepted = "accepted"
	OutboundPaymentStatusSettled  = "settled"
	OutboundPaymentStatusReturned = "returned"
	OutboundPaymentStatusFailed   = "failed"
)

// OutboundPayment is a neobank-initiated payment leaving the rail —
// initiation is synchronous ("accepted"), settlement/return/failure is
// async, delivered via webhook.
type OutboundPayment struct {
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ID               string
	AccountID        string
	Amount           string
	Currency         string
	CounterpartyIBAN string
	Reference        string
	Status           string
}
