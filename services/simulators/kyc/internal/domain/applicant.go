package domain

import "time"

const (
	StatusPending      = "pending"
	StatusManualReview = "manual_review"
	StatusApproved     = "approved"
	StatusRejected     = "rejected"
)

// Applicant is one identity-verification case this simulator is tracking on
// behalf of the user service.
type Applicant struct {
	CreatedAt   time.Time
	DecidedAt   *time.Time
	ID          string
	ExternalRef string
	FullName    string
	DateOfBirth string
	CountryCode string
	Status      string
	Reason      string
}
