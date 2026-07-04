package usecase

import (
	"context"
	"testing"

	"github.com/iho/neobank/services/simulators/kyc/internal/domain"
)

func TestResolveApplicantApprovesAndSchedulesWebhook(t *testing.T) {
	repo := newFakeApplicantRepository()
	ctx := context.Background()
	applicant, _ := repo.Create(ctx, "user-1", "Jane REVIEW Doe", "1990-01-01", "US", domain.StatusManualReview, "")

	dispatcher := &fakeDispatcher{}
	uc := NewResolveApplicantUseCase(repo, dispatcher, "http://user/webhooks/kyc/events")

	resolved, err := uc.Execute(ctx, ResolveApplicantInput{ApplicantID: applicant.ID, Verdict: domain.StatusApproved})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resolved.Status != domain.StatusApproved {
		t.Fatalf("expected approved, got %q", resolved.Status)
	}

	if len(dispatcher.calls) != 1 || dispatcher.calls[0].eventType != EventCheckCompleted {
		t.Fatalf("unexpected dispatcher calls: %+v", dispatcher.calls)
	}
}

func TestResolveApplicantRejectsNonManualReview(t *testing.T) {
	repo := newFakeApplicantRepository()
	ctx := context.Background()
	applicant, _ := repo.Create(ctx, "user-1", "Jane Doe", "1990-01-01", "US", domain.StatusApproved, "")

	dispatcher := &fakeDispatcher{}
	uc := NewResolveApplicantUseCase(repo, dispatcher, "http://user/webhooks/kyc/events")

	if _, err := uc.Execute(ctx, ResolveApplicantInput{ApplicantID: applicant.ID, Verdict: domain.StatusApproved}); err == nil {
		t.Fatal("expected error resolving a non-manual-review applicant")
	}
}

func TestResolveApplicantRejectsInvalidVerdict(t *testing.T) {
	repo := newFakeApplicantRepository()
	ctx := context.Background()
	applicant, _ := repo.Create(ctx, "user-1", "Jane REVIEW Doe", "1990-01-01", "US", domain.StatusManualReview, "")

	dispatcher := &fakeDispatcher{}
	uc := NewResolveApplicantUseCase(repo, dispatcher, "http://user/webhooks/kyc/events")

	if _, err := uc.Execute(ctx, ResolveApplicantInput{ApplicantID: applicant.ID, Verdict: "manual_review"}); err == nil {
		t.Fatal("expected error for invalid verdict")
	}
}

func TestResolveApplicantRejectsUnknown(t *testing.T) {
	repo := newFakeApplicantRepository()
	dispatcher := &fakeDispatcher{}
	uc := NewResolveApplicantUseCase(repo, dispatcher, "http://user/webhooks/kyc/events")

	if _, err := uc.Execute(context.Background(), ResolveApplicantInput{ApplicantID: "missing", Verdict: domain.StatusApproved}); err == nil {
		t.Fatal("expected error for unknown applicant")
	}
}
