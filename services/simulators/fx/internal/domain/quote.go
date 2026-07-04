package domain

import "time"

const (
	QuoteStatusPending  = "pending"
	QuoteStatusExecuted = "executed"
	QuoteStatusExpired  = "expired"
)

// Quote is one FX quote this simulator issued, whether or not it was ever
// executed — the audit trail an FX vendor would show a regulator.
type Quote struct {
	CreatedAt       time.Time
	ExpiresAt       time.Time
	ExecutedAt      *time.Time
	ID              string
	FromCurrency    string
	ToCurrency      string
	Amount          string
	ConvertedAmount string
	Rate            string
	Status          string
	SpreadBps       int
}
