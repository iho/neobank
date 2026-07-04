package domain

import "time"

// FXConversion is one executed currency conversion, keyed by the fx
// simulator's quote ID so a retried execute call is a no-op, not a second
// conversion.
type FXConversion struct {
	CreatedAt            time.Time
	ID                   string
	QuoteID              string
	UserID               string
	FromCurrency         string
	ToCurrency           string
	Amount               string
	ConvertedAmount      string
	Rate                 string
	FromLedgerTransferID string
	ToLedgerTransferID   string
	Status               string
}
