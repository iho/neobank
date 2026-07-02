package sqlcrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/iho/neobank/services/user/internal/port"
	"github.com/jackc/pgx/v5"
)

type GDPRRepository struct {
	q sqlc.Querier
}

func NewGDPRRepository(q sqlc.Querier) *GDPRRepository {
	return &GDPRRepository{q: q}
}

func (r *GDPRRepository) WithTx(tx pgx.Tx) port.GDPRRepository {
	return &GDPRRepository{q: withTx(r.q, tx)}
}

func (r *GDPRRepository) ListKYCSubmissionsByUser(ctx context.Context, userID string) ([]port.KYCSubmission, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListKYCSubmissionsByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]port.KYCSubmission, 0, len(rows))
	for _, row := range rows {
		docType := ""
		if row.DocumentType.Valid {
			docType = row.DocumentType.String
		}
		out = append(out, port.KYCSubmission{
			ID:                row.ID.String(),
			KYCCaseID:         row.KycCaseID.String(),
			UserID:            row.UserID.String(),
			DocumentType:      docType,
			DocumentNumber:    row.DocumentNumber,
			Provider:          row.Provider,
			ProviderReference: row.ProviderReference,
			ProviderResponse:  row.ProviderResponse,
			ScreeningDecision: row.ScreeningDecision,
			ScreeningReason:   row.ScreeningReason,
			CorrelationID:     row.CorrelationID,
			CreatedAt:         row.CreatedAt.Time.UTC(),
		})
	}
	return out, nil
}

func (r *GDPRRepository) ListWalletsByUser(ctx context.Context, userID string) ([]domain.Wallet, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListWalletsByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Wallet, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Wallet{
			ID:              row.ID.String(),
			UserID:          row.UserID.String(),
			Currency:        row.Currency,
			LedgerAccountID: row.LedgerAccountID,
			Status:          row.Status,
		})
	}
	return out, nil
}

func (r *GDPRRepository) CountWalletTransactions(ctx context.Context, userID string) (int64, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return 0, err
	}
	return r.q.CountWalletTransactionsByUser(ctx, uid)
}

func (r *GDPRRepository) RecordRequest(ctx context.Context, userID, requestType string) error {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return err
	}
	entry := audit.Resolve(ctx, audit.Entry{})
	return r.q.InsertGDPRRequest(ctx, sqlc.InsertGDPRRequestParams{
		ID:            uuid.New(),
		UserID:        uid,
		RequestType:   requestType,
		Actor:         entry.Actor,
		CorrelationID: textOrNil(entry.CorrelationID),
	})
}

func (r *GDPRRepository) MaskUserPII(ctx context.Context, userID, maskedEmail, passwordHash string) error {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return err
	}
	if err := r.q.MaskUserAccount(ctx, sqlc.MaskUserAccountParams{
		ID:           uid,
		Email:        maskedEmail,
		PasswordHash: passwordHash,
	}); err != nil {
		return err
	}
	if err := r.q.MaskUserProfile(ctx, uid); err != nil {
		return err
	}
	return r.q.MaskKYCSubmissionsByUser(ctx, uid)
}