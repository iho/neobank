package domain

import "time"

type Notification struct {
	ID        string
	UserID    string
	EventType string
	Title     string
	Body      string
	Read      bool
	CreatedAt time.Time
}