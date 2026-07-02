package domain

import "time"

type NotificationPreferences struct {
	UserID    string
	Transfers bool
	Cards     bool
	KYC       bool
	Push      bool
	Email     bool
	UpdatedAt time.Time
}