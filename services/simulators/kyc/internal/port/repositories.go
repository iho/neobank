package port

import (
	"context"

	"github.com/iho/neobank/services/simulators/kyc/internal/domain"
)

type ApplicantRepository interface {
	Create(ctx context.Context, externalRef, fullName, dateOfBirth, countryCode, status, reason string) (domain.Applicant, error)
	GetByID(ctx context.Context, id string) (*domain.Applicant, error)
	Resolve(ctx context.Context, id, status, reason string) (domain.Applicant, error)
}
