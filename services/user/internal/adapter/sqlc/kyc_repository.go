package sqlcrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/piicrypto"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/iho/neobank/services/user/internal/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type KYCRepository struct {
	q   sqlc.Querier
	pii piicrypto.Protector
}

func NewKYCRepository(q sqlc.Querier, pii piicrypto.Protector) *KYCRepository {
	if pii == nil {
		pii = piicrypto.NewNoop()
	}
	return &KYCRepository{q: q, pii: pii}
}

func (r *KYCRepository) WithTx(tx pgx.Tx) port.KYCRepository {
	return &KYCRepository{q: withTx(r.q, tx)}
}

func (r *KYCRepository) UpsertProfile(ctx context.Context, userID, fullName, dob, country string) error {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return err
	}
	params := sqlc.UpsertProfileParams{
		UserID:      uid,
		FullName:    pgutil.Text(fullName),
		CountryCode: pgutil.Text(country),
	}
	if r.pii.Enabled() {
		encDOB, err := piicrypto.Store(ctx, r.pii, dob)
		if err != nil {
			return err
		}
		params.DateOfBirthEncrypted = textOrNil(encDOB)
	} else {
		plainDOB, err := parseDate(dob)
		if err != nil {
			return err
		}
		params.DateOfBirth = plainDOB
	}
	return r.q.UpsertProfile(ctx, params)
}

func (r *KYCRepository) CreateCase(ctx context.Context, id, userID, status string) (domain.KYCCase, error) {
	caseID, err := pgutil.ParseUUID(id)
	if err != nil {
		return domain.KYCCase{}, err
	}
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return domain.KYCCase{}, err
	}
	row, err := r.q.CreateKYCCase(ctx, sqlc.CreateKYCCaseParams{
		ID:     caseID,
		UserID: uid,
		Status: status,
	})
	if err != nil {
		return domain.KYCCase{}, err
	}
	return domain.KYCCase{
		ID:     row.ID.String(),
		UserID: row.UserID.String(),
		Status: domain.KYCStatus(row.Status),
	}, nil
}

func (r *KYCRepository) GetLatestByUser(ctx context.Context, userID string) (*domain.KYCCase, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetLatestKYCCaseByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	return &domain.KYCCase{
		ID:              row.ID.String(),
		UserID:          row.UserID.String(),
		Status:          domain.KYCStatus(row.Status),
		RejectionReason: row.RejectionReason,
	}, nil
}

func (r *KYCRepository) GetByID(ctx context.Context, caseID string) (*domain.KYCCase, error) {
	id, err := pgutil.ParseUUID(caseID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetKYCCaseByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.KYCCase{
		ID:                row.ID.String(),
		UserID:            row.UserID.String(),
		Status:            domain.KYCStatus(row.Status),
		RejectionReason:   row.RejectionReason,
		VendorApplicantID: row.VendorApplicantID,
	}, nil
}

func (r *KYCRepository) GetByVendorApplicant(ctx context.Context, applicantID string) (*domain.KYCCase, error) {
	row, err := r.q.GetKYCCaseByVendorApplicant(ctx, pgutil.Text(applicantID))
	if err != nil {
		return nil, err
	}
	return &domain.KYCCase{
		ID:                row.ID.String(),
		UserID:            row.UserID.String(),
		Status:            domain.KYCStatus(row.Status),
		RejectionReason:   row.RejectionReason,
		VendorApplicantID: row.VendorApplicantID,
	}, nil
}

func (r *KYCRepository) SetVendorApplicant(ctx context.Context, caseID, applicantID string) error {
	id, err := pgutil.ParseUUID(caseID)
	if err != nil {
		return err
	}
	return r.q.SetKYCCaseVendorApplicant(ctx, sqlc.SetKYCCaseVendorApplicantParams{
		ID:                id,
		VendorApplicantID: pgutil.Text(applicantID),
	})
}

func (r *KYCRepository) MarkManualReview(ctx context.Context, caseID string) error {
	id, err := pgutil.ParseUUID(caseID)
	if err != nil {
		return err
	}
	return r.q.MarkKYCCaseManualReview(ctx, id)
}

func (r *KYCRepository) ApproveCase(ctx context.Context, caseID, decidedBy string) error {
	id, err := pgutil.ParseUUID(caseID)
	if err != nil {
		return err
	}
	return r.q.ApproveKYCCase(ctx, sqlc.ApproveKYCCaseParams{
		ID:        id,
		DecidedBy: pgutil.Text(decidedBy),
	})
}

func (r *KYCRepository) RejectCase(ctx context.Context, caseID, reason, decidedBy string) error {
	id, err := pgutil.ParseUUID(caseID)
	if err != nil {
		return err
	}
	return r.q.RejectKYCCase(ctx, sqlc.RejectKYCCaseParams{
		ID:              id,
		RejectionReason: pgutil.Text(reason),
		DecidedBy:       pgutil.Text(decidedBy),
	})
}

func parseDate(value string) (pgtype.Date, error) {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("invalid date_of_birth: %w", err)
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}
