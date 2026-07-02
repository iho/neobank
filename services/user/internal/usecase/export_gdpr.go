package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/port"
	"github.com/jackc/pgx/v5"
)

type GDPRExport struct {
	UserID                  string
	ExportedAt              time.Time
	Profile                 *domain.Profile
	KYCSubmissions          []port.KYCSubmission
	Wallets                 []domain.Wallet
	WalletTransactionCount  int64
}

type ExportGDPRUseCase struct {
	profiles ProfileReader
	gdpr     port.GDPRRepository
}

func NewExportGDPRUseCase(profiles ProfileReader, gdpr port.GDPRRepository) *ExportGDPRUseCase {
	return &ExportGDPRUseCase{profiles: profiles, gdpr: gdpr}
}

func (uc *ExportGDPRUseCase) Execute(ctx context.Context, userID string) (GDPRExport, error) {
	if userID == "" {
		return GDPRExport{}, fmt.Errorf("user_id is required")
	}

	profile, err := uc.profiles.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GDPRExport{}, fmt.Errorf("user not found")
		}
		return GDPRExport{}, err
	}

	submissions, err := uc.gdpr.ListKYCSubmissionsByUser(ctx, userID)
	if err != nil {
		return GDPRExport{}, err
	}
	wallets, err := uc.gdpr.ListWalletsByUser(ctx, userID)
	if err != nil {
		return GDPRExport{}, err
	}
	txCount, err := uc.gdpr.CountWalletTransactions(ctx, userID)
	if err != nil {
		return GDPRExport{}, err
	}
	if err := uc.gdpr.RecordRequest(ctx, userID, "export"); err != nil {
		return GDPRExport{}, err
	}

	return GDPRExport{
		UserID:                 userID,
		ExportedAt:             time.Now().UTC(),
		Profile:                profile,
		KYCSubmissions:         submissions,
		Wallets:                wallets,
		WalletTransactionCount: txCount,
	}, nil
}