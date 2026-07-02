package sqlcrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/iho/neobank/services/user/internal/usecase"
	"github.com/jackc/pgx/v5/pgtype"
)

type ReferralInviteRepository struct {
	q sqlc.Querier
}

func NewReferralInviteRepository(q sqlc.Querier) *ReferralInviteRepository {
	return &ReferralInviteRepository{q: q}
}

func (r *ReferralInviteRepository) Create(ctx context.Context, inviterUserID, inviteCode string) (*usecase.ReferralInvite, error) {
	uid, err := pgutil.ParseUUID(inviterUserID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.CreateReferralInvite(ctx, sqlc.CreateReferralInviteParams{
		InviterUserID: uid,
		InviteCode:    inviteCode,
	})
	if err != nil {
		return nil, err
	}
	return mapReferralInvite(row), nil
}

func (r *ReferralInviteRepository) ListByInviter(ctx context.Context, inviterUserID string, limit int) ([]usecase.ReferralInvite, error) {
	uid, err := pgutil.ParseUUID(inviterUserID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListReferralInvitesByInviter(ctx, sqlc.ListReferralInvitesByInviterParams{
		InviterUserID: uid,
		Limit:         int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]usecase.ReferralInvite, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapReferralInvite(row))
	}
	return out, nil
}

func (r *ReferralInviteRepository) GetByCode(ctx context.Context, code string) (*usecase.ReferralInvite, error) {
	row, err := r.q.GetReferralInviteByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return mapReferralInvite(row), nil
}

func (r *ReferralInviteRepository) Accept(ctx context.Context, code, inviteeUserID string) (*usecase.ReferralInvite, error) {
	uid, err := pgutil.ParseUUID(inviteeUserID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.AcceptReferralInvite(ctx, sqlc.AcceptReferralInviteParams{
		InviteCode:    code,
		InviteeUserID: pgtype.UUID{Bytes: uid, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapReferralInvite(row), nil
}

func mapReferralInvite(row sqlc.UserReferralInvite) *usecase.ReferralInvite {
	invite := &usecase.ReferralInvite{
		ID:            row.ID.String(),
		InviterUserID: row.InviterUserID.String(),
		InviteCode:    row.InviteCode,
		Status:        row.Status,
		CreatedAt:     row.CreatedAt.Time.UTC(),
	}
	if row.InviteeUserID.Valid {
		invite.InviteeUserID = uuid.UUID(row.InviteeUserID.Bytes).String()
	}
	if row.AcceptedAt.Valid {
		t := row.AcceptedAt.Time.UTC()
		invite.AcceptedAt = &t
	}
	return invite
}