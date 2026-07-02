package sqlcrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthorizationRepository struct {
	q sqlc.Querier
}

func NewAuthorizationRepository(q sqlc.Querier) *AuthorizationRepository {
	return &AuthorizationRepository{q: q}
}

func (r *AuthorizationRepository) Create(ctx context.Context, a domain.Authorization) error {
	id, err := pgutil.ParseUUID(a.ID)
	if err != nil {
		return err
	}
	cardID, err := pgutil.ParseUUID(a.CardID)
	if err != nil {
		return err
	}
	userID, err := pgutil.ParseUUID(a.UserID)
	if err != nil {
		return err
	}
	amount, err := pgutil.NumericFromString(a.Amount)
	if err != nil {
		return err
	}
	return r.q.CreateAuthorization(ctx, sqlc.CreateAuthorizationParams{
		ID:             id,
		CardID:         cardID,
		UserID:         userID,
		IdempotencyKey: a.IdempotencyKey,
		MerchantName:   pgutil.Text(a.MerchantName),
		Amount:         amount,
		Currency:       a.Currency,
		Status:         string(a.Status),
	})
}

func (r *AuthorizationRepository) GetByID(ctx context.Context, id string) (*domain.Authorization, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetAuthorizationByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return mapAuthorizationByIDRow(row), nil
}

func (r *AuthorizationRepository) GetByCardAndIdempotencyKey(ctx context.Context, cardID, key string) (*domain.Authorization, error) {
	cid, err := pgutil.ParseUUID(cardID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetAuthorizationByCardAndIdempotencyKey(ctx, sqlc.GetAuthorizationByCardAndIdempotencyKeyParams{
		CardID:         cid,
		IdempotencyKey: key,
	})
	if err != nil {
		return nil, err
	}
	return mapAuthorizationByKeyRow(row), nil
}

func (r *AuthorizationRepository) ListByUser(ctx context.Context, userID string, limit int) ([]domain.Authorization, error) {
	if limit <= 0 {
		limit = 20
	}
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListAuthorizationsByUser(ctx, sqlc.ListAuthorizationsByUserParams{
		UserID: uid,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Authorization, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapAuthorizationListRow(row))
	}
	return out, nil
}

func (r *AuthorizationRepository) MarkHold(ctx context.Context, id, holdID string) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}
	return r.q.MarkAuthorizationHold(ctx, sqlc.MarkAuthorizationHoldParams{
		ID:           uid,
		LedgerHoldID: pgutil.Text(holdID),
	})
}

func (r *AuthorizationRepository) MarkFailed(ctx context.Context, id, reason string) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}
	return r.q.MarkAuthorizationFailed(ctx, sqlc.MarkAuthorizationFailedParams{
		ID:            uid,
		FailureReason: pgutil.Text(reason),
	})
}

func (r *AuthorizationRepository) MarkCaptured(ctx context.Context, id, transferID string) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}
	return r.q.MarkAuthorizationCaptured(ctx, sqlc.MarkAuthorizationCapturedParams{
		ID:               uid,
		LedgerTransferID: pgutil.Text(transferID),
	})
}

func mapAuthorizationByIDRow(row sqlc.GetAuthorizationByIDRow) *domain.Authorization {
	return mapAuthorization(
		row.ID, row.CardID, row.UserID, row.IdempotencyKey, row.MerchantName,
		row.Amount, row.Currency, row.Status, row.LedgerHoldID, row.LedgerTransferID,
		row.FailureReason, row.CreatedAt, row.CapturedAt,
	)
}

func mapAuthorizationByKeyRow(row sqlc.GetAuthorizationByCardAndIdempotencyKeyRow) *domain.Authorization {
	return mapAuthorization(
		row.ID, row.CardID, row.UserID, row.IdempotencyKey, row.MerchantName,
		row.Amount, row.Currency, row.Status, row.LedgerHoldID, row.LedgerTransferID,
		row.FailureReason, row.CreatedAt, row.CapturedAt,
	)
}

func mapAuthorizationListRow(row sqlc.ListAuthorizationsByUserRow) *domain.Authorization {
	return mapAuthorization(
		row.ID, row.CardID, row.UserID, row.IdempotencyKey, row.MerchantName,
		row.Amount, row.Currency, row.Status, row.LedgerHoldID, row.LedgerTransferID,
		row.FailureReason, row.CreatedAt, row.CapturedAt,
	)
}

func mapAuthorization(
	id, cardID, userID uuid.UUID,
	idempotencyKey, merchantName, amount, currency, status, ledgerHoldID, ledgerTransferID, failureReason string,
	createdAt, capturedAt pgtype.Timestamptz,
) *domain.Authorization {
	a := &domain.Authorization{
		ID:               id.String(),
		CardID:           cardID.String(),
		UserID:           userID.String(),
		IdempotencyKey:   idempotencyKey,
		MerchantName:     merchantName,
		Amount:           amount,
		Currency:         currency,
		Status:           domain.AuthStatus(status),
		LedgerHoldID:     ledgerHoldID,
		LedgerTransferID: ledgerTransferID,
		FailureReason:    failureReason,
	}
	if createdAt.Valid {
		a.CreatedAt = createdAt.Time.UTC()
	}
	if capturedAt.Valid {
		t := capturedAt.Time.UTC()
		a.CapturedAt = &t
	}
	return a
}