package domain

import "time"

type WalletTransaction struct {
	ID           string
	Type         string
	Amount       string
	Currency     string
	Direction    string
	Status       string
	Counterparty string
	Memo         string
	CreatedAt    time.Time
}