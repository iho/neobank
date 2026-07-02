package domain

import "time"

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusMasked    UserStatus = "masked"
)

type KYCStatus string

const (
	KYCStatusPending      KYCStatus = "pending"
	KYCStatusApproved     KYCStatus = "approved"
	KYCStatusRejected     KYCStatus = "rejected"
	KYCStatusManualReview KYCStatus = "manual_review"
)

type User struct {
	ID           string
	Email        string
	Phone        string
	PasswordHash string
	Status       UserStatus
	CreatedAt    time.Time
}

type Profile struct {
	UserID      string
	Email       string
	Phone       string
	Status      string
	FullName    string
	DateOfBirth string
	CountryCode string
	KYCStatus   string
	CreatedAt   time.Time
}

type KYCCase struct {
	ID              string
	UserID          string
	Status          KYCStatus
	RejectionReason string
}

func (s KYCStatus) String() string { return string(s) }

type Wallet struct {
	ID              string
	UserID          string
	Currency        string
	LedgerAccountID string
	Status          string
}