package domain

import "time"

// Account is a virtual bank account this simulator issued for a neobank
// wallet, identified externally by an IBAN.
type Account struct {
	CreatedAt   time.Time
	ID          string
	ExternalRef string
	Currency    string
	IBAN        string
}
