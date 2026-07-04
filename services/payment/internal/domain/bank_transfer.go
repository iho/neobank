package domain

import "time"

// BankAccount is the local mirror of a virtual IBAN issued by the rails
// simulator (or, later, a real payment rail) for a user's wallet.
type BankAccount struct {
	CreatedAt      time.Time
	ID             string
	UserID         string
	Currency       string
	RailsAccountID string
	IBAN           string
}

// BankTransfer is one inbound rails transfer processed into a ledger credit.
type BankTransfer struct {
	CreatedAt        time.Time
	ID               string
	RailsTransferID  string
	UserID           string
	Amount           string
	Currency         string
	SenderName       string
	Reference        string
	LedgerTransferID string
	Status           string
}
