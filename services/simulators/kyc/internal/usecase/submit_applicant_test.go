package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/iho/neobank/services/simulators/kyc/internal/domain"
)

func TestSubmitApplicantApprovesByDefault(t *testing.T) {
	repo := newFakeApplicantRepository()
	dispatcher := &fakeDispatcher{}
	uc := NewSubmitApplicantUseCase(repo, dispatcher, "http://user/webhooks/kyc/events")

	applicant, err := uc.Execute(context.Background(), SubmitApplicantInput{
		ExternalRef: "user-1", FullName: "Jane Doe", DateOfBirth: "1990-01-01", CountryCode: "US",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if applicant.Status != domain.StatusApproved {
		t.Fatalf("expected approved, got %q", applicant.Status)
	}

	if len(dispatcher.calls) != 1 || dispatcher.calls[0].eventType != EventCheckCompleted {
		t.Fatalf("unexpected dispatcher calls: %+v", dispatcher.calls)
	}

	payload, ok := dispatcher.calls[0].payload.(CheckCompletedPayload)
	if !ok || payload.Verdict != domain.StatusApproved || payload.ExternalRef != "user-1" {
		t.Fatalf("unexpected payload: %+v", dispatcher.calls[0].payload)
	}
}

func TestSubmitApplicantRejectsOnMagicValue(t *testing.T) {
	repo := newFakeApplicantRepository()
	dispatcher := &fakeDispatcher{}
	uc := NewSubmitApplicantUseCase(repo, dispatcher, "http://user/webhooks/kyc/events")

	applicant, err := uc.Execute(context.Background(), SubmitApplicantInput{
		ExternalRef: "user-1", FullName: "Jane REJECT Doe", DateOfBirth: "1990-01-01", CountryCode: "US",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if applicant.Status != domain.StatusRejected {
		t.Fatalf("expected rejected, got %q", applicant.Status)
	}

	if len(dispatcher.calls) != 1 {
		t.Fatalf("expected 1 webhook scheduled, got %d", len(dispatcher.calls))
	}
}

func TestSubmitApplicantRejectsUnderage(t *testing.T) {
	repo := newFakeApplicantRepository()
	dispatcher := &fakeDispatcher{}
	uc := NewSubmitApplicantUseCase(repo, dispatcher, "http://user/webhooks/kyc/events")

	dob := time.Now().AddDate(-16, 0, 0).Format("2006-01-02")

	applicant, err := uc.Execute(context.Background(), SubmitApplicantInput{
		ExternalRef: "user-1", FullName: "Young Person", DateOfBirth: dob, CountryCode: "US",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if applicant.Status != domain.StatusRejected || applicant.Reason != "underage" {
		t.Fatalf("expected rejected/underage, got %q/%q", applicant.Status, applicant.Reason)
	}
}

func TestSubmitApplicantManualReviewDoesNotScheduleWebhook(t *testing.T) {
	repo := newFakeApplicantRepository()
	dispatcher := &fakeDispatcher{}
	uc := NewSubmitApplicantUseCase(repo, dispatcher, "http://user/webhooks/kyc/events")

	applicant, err := uc.Execute(context.Background(), SubmitApplicantInput{
		ExternalRef: "user-1", FullName: "Jane REVIEW Doe", DateOfBirth: "1990-01-01", CountryCode: "US",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if applicant.Status != domain.StatusManualReview {
		t.Fatalf("expected manual_review, got %q", applicant.Status)
	}

	if len(dispatcher.calls) != 0 {
		t.Fatalf("expected no webhook scheduled for manual_review, got %d", len(dispatcher.calls))
	}
}

func TestSubmitApplicantValidatesInput(t *testing.T) {
	repo := newFakeApplicantRepository()
	dispatcher := &fakeDispatcher{}
	uc := NewSubmitApplicantUseCase(repo, dispatcher, "http://user/webhooks/kyc/events")

	if _, err := uc.Execute(context.Background(), SubmitApplicantInput{}); err == nil {
		t.Fatal("expected validation error for empty input")
	}
}
