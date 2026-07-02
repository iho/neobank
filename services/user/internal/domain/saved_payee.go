package domain

import "time"

type SavedPayee struct {
	ID          string
	UserID      string
	PayeeUserID string
	Nickname    string
	PayeeEmail  string
	PayeePhone  string
	LastUsedAt  time.Time
	CreatedAt   time.Time
}