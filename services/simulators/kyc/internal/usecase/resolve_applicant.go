package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/services/simulators/kyc/internal/domain"
	"github.com/iho/neobank/services/simulators/kyc/internal/port"
)

type ResolveApplicantInput struct {
	ApplicantID string
	// Verdict must be "approved" or "rejected" — a human reviewer resolving
	// a manual_review case.
	Verdict string
	Reason  string
}

// ResolveApplicantUseCase is the admin entry point mimicking a human
// reviewer clearing a manual_review applicant.
type ResolveApplicantUseCase struct {
	applicants port.ApplicantRepository
	dispatcher WebhookDispatcher
	eventsURL  string
}

func NewResolveApplicantUseCase(applicants port.ApplicantRepository, dispatcher WebhookDispatcher, eventsURL string) *ResolveApplicantUseCase {
	return &ResolveApplicantUseCase{applicants: applicants, dispatcher: dispatcher, eventsURL: eventsURL}
}

func (uc *ResolveApplicantUseCase) Execute(ctx context.Context, in ResolveApplicantInput) (domain.Applicant, error) {
	if in.Verdict != domain.StatusApproved && in.Verdict != domain.StatusRejected {
		return domain.Applicant{}, fmt.Errorf("verdict must be %q or %q", domain.StatusApproved, domain.StatusRejected)
	}

	applicant, err := uc.applicants.GetByID(ctx, in.ApplicantID)
	if err != nil {
		return domain.Applicant{}, err
	}

	if applicant == nil {
		return domain.Applicant{}, fmt.Errorf("applicant %q not found", in.ApplicantID)
	}

	if applicant.Status != domain.StatusManualReview {
		return domain.Applicant{}, fmt.Errorf("applicant is not in manual_review (status %q)", applicant.Status)
	}

	resolved, err := uc.applicants.Resolve(ctx, applicant.ID, in.Verdict, in.Reason)
	if err != nil {
		return domain.Applicant{}, err
	}

	if _, err := uc.dispatcher.Enqueue(ctx, uc.eventsURL, EventCheckCompleted, CheckCompletedPayload{
		ApplicantID: resolved.ID,
		ExternalRef: resolved.ExternalRef,
		Verdict:     resolved.Status,
		Reason:      resolved.Reason,
	}); err != nil {
		return domain.Applicant{}, fmt.Errorf("schedule verdict webhook: %w", err)
	}

	return resolved, nil
}
