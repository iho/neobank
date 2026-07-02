package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
)

type CloseAccountUseCase struct {
	mask  *MaskGDPRUseCase
	users UserRepository
	audit audit.Recorder
}

func NewCloseAccountUseCase(mask *MaskGDPRUseCase, users UserRepository, auditRecorder audit.Recorder) *CloseAccountUseCase {
	return &CloseAccountUseCase{mask: mask, users: users, audit: auditRecorder}
}

func (uc *CloseAccountUseCase) Execute(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("user not found")
		}
		return err
	}
	if user.Status == domain.UserStatusMasked {
		return nil
	}
	if user.Status != domain.UserStatusActive {
		return fmt.Errorf("account cannot be closed")
	}

	if err := uc.mask.Execute(ctx, userID); err != nil {
		return err
	}

	return uc.audit.Record(ctx, audit.Entry{
		EntityType: "user",
		EntityID:   userID,
		Action:     "account_closed",
		FromStatus: string(user.Status),
		ToStatus:   string(domain.UserStatusMasked),
		Metadata: map[string]any{
			"reason": "user_requested_closure",
		},
	})
}