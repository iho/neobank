package domain

import "time"

const (
	DisputeStatusOpen = "open"
	DisputeStatusWon  = "won"
	DisputeStatusLost = "lost"
)

// Dispute tracks a chargeback against a captured authorization, from the
// cardproc simulator's "opened" webhook through to its resolution. The
// cardholder gets a provisional credit as soon as the dispute opens; if
// it's later lost, that credit is clawed back.
type Dispute struct {
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	ID                          string
	ChargebackID                string
	AuthorizationID             string
	CardID                      string
	UserID                      string
	Amount                      string
	Currency                    string
	Reason                      string
	Status                      string
	ProvisionalCreditTransferID string
	ReversalTransferID          string
}
