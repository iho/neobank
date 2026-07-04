package sqlcrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/simulators/kyc/internal/domain"
	"github.com/iho/neobank/services/simulators/kyc/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const dateLayout = "2006-01-02"

type ApplicantRepository struct {
	q sqlc.Querier
}

func NewApplicantRepository(q sqlc.Querier) *ApplicantRepository {
	return &ApplicantRepository{q: q}
}

func (r *ApplicantRepository) Create(ctx context.Context, externalRef, fullName, dateOfBirth, countryCode, status, reason string) (domain.Applicant, error) {
	dob, err := parseDate(dateOfBirth)
	if err != nil {
		return domain.Applicant{}, err
	}

	row, err := r.q.CreateApplicant(ctx, sqlc.CreateApplicantParams{
		ExternalRef: externalRef,
		FullName:    fullName,
		DateOfBirth: dob,
		CountryCode: countryCode,
		Status:      status,
		Reason:      reason,
	})
	if err != nil {
		return domain.Applicant{}, err
	}

	return toApplicant(row), nil
}

func (r *ApplicantRepository) GetByID(ctx context.Context, id string) (*domain.Applicant, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.q.GetApplicantByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	applicant := toApplicant(row)

	return &applicant, nil
}

func (r *ApplicantRepository) Resolve(ctx context.Context, id, status, reason string) (domain.Applicant, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return domain.Applicant{}, err
	}

	row, err := r.q.ResolveApplicant(ctx, sqlc.ResolveApplicantParams{
		ID:     uid,
		Status: status,
		Reason: reason,
	})
	if err != nil {
		return domain.Applicant{}, err
	}

	return toApplicant(row), nil
}

func toApplicant(row sqlc.KycApplicant) domain.Applicant {
	a := domain.Applicant{
		ID:          row.ID.String(),
		ExternalRef: row.ExternalRef,
		FullName:    row.FullName,
		DateOfBirth: row.DateOfBirth.Time.Format(dateLayout),
		CountryCode: row.CountryCode,
		Status:      row.Status,
		Reason:      row.Reason,
		CreatedAt:   row.CreatedAt.Time.UTC(),
	}
	if row.DecidedAt.Valid {
		t := row.DecidedAt.Time.UTC()
		a.DecidedAt = &t
	}

	return a
}

func parseDate(value string) (pgtype.Date, error) {
	t, err := time.Parse(dateLayout, value)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("invalid date_of_birth %q: %w", value, err)
	}

	return pgtype.Date{Time: t, Valid: true}, nil
}
