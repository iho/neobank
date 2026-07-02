package sqlcrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type KYCRepository struct {
	q sqlc.Querier
}

func NewKYCRepository(q sqlc.Querier) *KYCRepository {
	return &KYCRepository{q: q}
}

func (r *KYCRepository) UpsertProfile(ctx context.Context, userID, fullName, dob, country string) error {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return err
	}
	date, err := parseDate(dob)
	if err != nil {
		return err
	}
	return r.q.UpsertProfile(ctx, sqlc.UpsertProfileParams{
		UserID:      uid,
		FullName:    pgutil.Text(fullName),
		DateOfBirth: date,
		CountryCode: pgutil.Text(country),
	})
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

func (r *KYCRepository) ApproveCase(ctx context.Context, caseID string) error {
	id, err := pgutil.ParseUUID(caseID)
	if err != nil {
		return err
	}
	return r.q.ApproveKYCCase(ctx, id)
}

func parseDate(value string) (pgtype.Date, error) {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("invalid date_of_birth: %w", err)
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}