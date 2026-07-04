package domain

import "time"

// InboundTransfer is money that arrived on the simulated rail against one of
// this simulator's accounts, delivered to the neobank as a webhook.
type InboundTransfer struct {
	CreatedAt  time.Time
	ID         string
	AccountID  string
	Amount     string
	Currency   string
	SenderName string
	Reference  string
	Status     string
}
