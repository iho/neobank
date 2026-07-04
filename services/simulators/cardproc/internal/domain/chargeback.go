package domain

import "time"

const (
	ChargebackStatusOpened = "opened"
	ChargebackStatusWon    = "won"
	ChargebackStatusLost   = "lost"
)

// Chargeback is a simulated dispute against a captured transaction. It is
// this simulator's own bookkeeping, separate from the card service's
// dispute record (created from the "opened" webhook) — the two are linked
// by ID (this Chargeback's ID is the card service's dispute_id).
type Chargeback struct {
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ID              string
	TransactionID   string
	AuthorizationID string
	Amount          string
	Currency        string
	Reason          string
	Status          string
}
