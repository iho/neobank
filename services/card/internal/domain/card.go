package domain

import "time"

type CardStatus string

const (
	CardStatusActive    CardStatus = "active"
	CardStatusFrozen    CardStatus = "frozen"
	CardStatusCancelled CardStatus = "cancelled"
)

type Card struct {
	ID             string
	UserID         string
	WalletID       string
	ProcessorRef   string
	PANToken       string
	LastFour       string
	ExpiryMonth    int
	ExpiryYear     int
	Status         CardStatus
	IdempotencyKey string
	DailyLimit     string
	OnlineOnly     bool
	CreatedAt      time.Time
}