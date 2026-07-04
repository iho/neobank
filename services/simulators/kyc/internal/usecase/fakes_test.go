package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/iho/neobank/services/simulators/kyc/internal/domain"
)

type fakeApplicantRepository struct {
	applicants map[string]domain.Applicant
	seq        int
}

func newFakeApplicantRepository() *fakeApplicantRepository {
	return &fakeApplicantRepository{applicants: map[string]domain.Applicant{}}
}

func (r *fakeApplicantRepository) Create(_ context.Context, externalRef, fullName, dateOfBirth, countryCode, status, reason string) (domain.Applicant, error) {
	r.seq++
	a := domain.Applicant{
		ID:          "applicant-" + strconv.Itoa(r.seq),
		ExternalRef: externalRef,
		FullName:    fullName,
		DateOfBirth: dateOfBirth,
		CountryCode: countryCode,
		Status:      status,
		Reason:      reason,
		CreatedAt:   time.Now().UTC(),
	}
	r.applicants[a.ID] = a

	return a, nil
}

func (r *fakeApplicantRepository) GetByID(_ context.Context, id string) (*domain.Applicant, error) {
	a, ok := r.applicants[id]
	if !ok {
		return nil, nil
	}

	return &a, nil
}

func (r *fakeApplicantRepository) Resolve(_ context.Context, id, status, reason string) (domain.Applicant, error) {
	a, ok := r.applicants[id]
	if !ok {
		return domain.Applicant{}, fmt.Errorf("applicant %q not found", id)
	}

	a.Status = status
	a.Reason = reason
	r.applicants[id] = a

	return a, nil
}

type fakeDispatcher struct {
	calls []struct {
		url       string
		eventType string
		payload   any
	}
}

func (d *fakeDispatcher) Enqueue(_ context.Context, url, eventType string, payload any) (string, error) {
	d.calls = append(d.calls, struct {
		url       string
		eventType string
		payload   any
	}{url, eventType, payload})

	return "delivery-1", nil
}
