package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/iho/neobank/pkg/vendorsim"
	"github.com/iho/neobank/services/simulators/kyc/internal/domain"
	"github.com/iho/neobank/services/simulators/kyc/internal/port"
)

// EventCheckCompleted is delivered when an applicant's identity check
// reaches a final verdict — on submission for immediate approve/reject
// outcomes, or later via ResolveApplicantUseCase for manual-review cases.
const EventCheckCompleted = "kyc.check.completed"

// CheckCompletedPayload is the webhook body delivered on EventCheckCompleted.
type CheckCompletedPayload struct {
	ApplicantID string `json:"applicant_id"`
	ExternalRef string `json:"external_ref"`
	Verdict     string `json:"verdict"`
	Reason      string `json:"reason,omitempty"`
}

type SubmitApplicantInput struct {
	ExternalRef string
	FullName    string
	DateOfBirth string
	CountryCode string
}

// SubmitApplicantUseCase is the entry point the user service calls during
// KYC submission. The verdict is decided here deterministically by magic
// value (see docs/vendor-simulators-plan.md Phase 3) but delivered async via
// webhook — approved/rejected fire immediately, manual_review waits for
// ResolveApplicantUseCase, mimicking a human reviewer.
type SubmitApplicantUseCase struct {
	applicants port.ApplicantRepository
	dispatcher WebhookDispatcher
	eventsURL  string
}

func NewSubmitApplicantUseCase(applicants port.ApplicantRepository, dispatcher WebhookDispatcher, eventsURL string) *SubmitApplicantUseCase {
	return &SubmitApplicantUseCase{applicants: applicants, dispatcher: dispatcher, eventsURL: eventsURL}
}

func (uc *SubmitApplicantUseCase) Execute(ctx context.Context, in SubmitApplicantInput) (domain.Applicant, error) {
	if in.ExternalRef == "" || in.FullName == "" || in.DateOfBirth == "" || in.CountryCode == "" {
		return domain.Applicant{}, fmt.Errorf("external_ref, full_name, date_of_birth, and country_code are required")
	}

	status, reason, err := decideVerdict(in.FullName, in.DateOfBirth)
	if err != nil {
		return domain.Applicant{}, err
	}

	applicant, err := uc.applicants.Create(ctx, in.ExternalRef, in.FullName, in.DateOfBirth, in.CountryCode, status, reason)
	if err != nil {
		return domain.Applicant{}, err
	}

	if status == domain.StatusApproved || status == domain.StatusRejected {
		if err := uc.scheduleVerdictWebhook(ctx, applicant); err != nil {
			return domain.Applicant{}, err
		}
	}

	return applicant, nil
}

func (uc *SubmitApplicantUseCase) scheduleVerdictWebhook(ctx context.Context, applicant domain.Applicant) error {
	_, err := uc.dispatcher.Enqueue(ctx, uc.eventsURL, EventCheckCompleted, CheckCompletedPayload{
		ApplicantID: applicant.ID,
		ExternalRef: applicant.ExternalRef,
		Verdict:     applicant.Status,
		Reason:      applicant.Reason,
	})
	if err != nil {
		return fmt.Errorf("schedule verdict webhook: %w", err)
	}

	return nil
}

// decideVerdict applies the shared magic-value conventions: a name
// containing REJECT/REVIEW forces that outcome; otherwise an applicant
// under 18 is rejected; anything else is approved.
func decideVerdict(fullName, dateOfBirth string) (status, reason string, err error) {
	if vendorsim.ContainsToken(fullName, "REJECT") {
		return domain.StatusRejected, "vendor_rejected_magic_value", nil
	}

	if vendorsim.ContainsToken(fullName, "REVIEW") {
		return domain.StatusManualReview, "", nil
	}

	dob, err := time.Parse("2006-01-02", dateOfBirth)
	if err != nil {
		return "", "", fmt.Errorf("invalid date_of_birth %q: %w", dateOfBirth, err)
	}

	if age(dob) < 18 {
		return domain.StatusRejected, "underage", nil
	}

	return domain.StatusApproved, "", nil
}

func age(dob time.Time) int {
	now := time.Now().UTC()
	years := now.Year() - dob.Year()

	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		years--
	}

	return years
}
