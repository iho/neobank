package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/gdpr"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/port"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type MaskGDPRUseCase struct {
	users UserRepository
	gdpr  port.GDPRRepository
	audit audit.Recorder
	tx    *pgutil.TxRunner
}

func NewMaskGDPRUseCase(users UserRepository, gdpr port.GDPRRepository, auditRecorder audit.Recorder, tx *pgutil.TxRunner) *MaskGDPRUseCase {
	return &MaskGDPRUseCase{users: users, gdpr: gdpr, audit: auditRecorder, tx: tx}
}

func (uc *MaskGDPRUseCase) Execute(ctx context.Context, userID string) error {
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

	maskedHash, err := bcrypt.GenerateFromPassword([]byte("gdpr-masked"), bcrypt.MinCost)
	if err != nil {
		return err
	}

	return uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		gdprRepo := uc.gdpr.WithTx(tx)
		if err := gdprRepo.MaskUserPII(ctx, userID, gdpr.MaskedEmail(userID), string(maskedHash)); err != nil {
			return err
		}
		if err := gdprRepo.RecordRequest(ctx, userID, "mask"); err != nil {
			return err
		}
		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "user",
			EntityID:   userID,
			Action:     "mask_pii",
			FromStatus: string(user.Status),
			ToStatus:   gdpr.UserStatusMasked,
			Metadata: map[string]any{
				"reason": "gdpr_erasure",
			},
		})
	})
}