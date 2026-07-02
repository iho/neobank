package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
)

type KYCRepository interface {
	UpsertProfile(ctx context.Context, userID, fullName, dob, country string) error
	CreateCase(ctx context.Context, id, userID, status string) (domain.KYCCase, error)
	GetLatestByUser(ctx context.Context, userID string) (*domain.KYCCase, error)
	ApproveCase(ctx context.Context, caseID string) error
}

type SubmitKYCInput struct {
	UserID         string
	FullName       string
	DateOfBirth    string
	CountryCode    string
	IdempotencyKey string
}

type SubmitKYCOutput struct {
	KYCCaseID string
	Status    domain.KYCStatus
	WalletID  string
}

type SubmitKYCUseCase struct {
	kyc       KYCRepository
	provision *ProvisionWalletUseCase
}

func NewSubmitKYCUseCase(kyc KYCRepository, provision *ProvisionWalletUseCase) *SubmitKYCUseCase {
	return &SubmitKYCUseCase{kyc: kyc, provision: provision}
}

func (uc *SubmitKYCUseCase) Execute(ctx context.Context, in SubmitKYCInput) (SubmitKYCOutput, error) {
	if in.UserID == "" || in.FullName == "" || in.DateOfBirth == "" || in.CountryCode == "" {
		return SubmitKYCOutput{}, fmt.Errorf("user_id, full_name, date_of_birth, and country_code are required")
	}

	existing, err := uc.kyc.GetLatestByUser(ctx, in.UserID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return SubmitKYCOutput{}, err
	}
	if existing != nil && existing.Status == domain.KYCStatusApproved {
		wallet, wErr := uc.provision.Execute(ctx, ProvisionWalletInput{UserID: in.UserID, Currency: "USD"})
		if wErr != nil {
			return SubmitKYCOutput{}, wErr
		}
		return SubmitKYCOutput{
			KYCCaseID: existing.ID,
			Status:    domain.KYCStatusApproved,
			WalletID:  wallet.WalletID,
		}, nil
	}

	if err := uc.kyc.UpsertProfile(ctx, in.UserID, in.FullName, in.DateOfBirth, in.CountryCode); err != nil {
		return SubmitKYCOutput{}, err
	}

	caseID := uuid.NewString()
	kycCase, err := uc.kyc.CreateCase(ctx, caseID, in.UserID, string(domain.KYCStatusPending))
	if err != nil {
		return SubmitKYCOutput{}, err
	}

	if err := uc.kyc.ApproveCase(ctx, caseID); err != nil {
		return SubmitKYCOutput{}, err
	}

	wallet, err := uc.provision.Execute(ctx, ProvisionWalletInput{
		UserID:   in.UserID,
		Currency: "USD",
	})
	if err != nil {
		return SubmitKYCOutput{}, fmt.Errorf("provision wallet: %w", err)
	}

	return SubmitKYCOutput{
		KYCCaseID: kycCase.ID,
		Status:    domain.KYCStatusApproved,
		WalletID:  wallet.WalletID,
	}, nil
}

type GetKYCStatusUseCase struct {
	kyc KYCRepository
}

func NewGetKYCStatusUseCase(kyc KYCRepository) *GetKYCStatusUseCase {
	return &GetKYCStatusUseCase{kyc: kyc}
}

func (uc *GetKYCStatusUseCase) Execute(ctx context.Context, userID string) (domain.KYCCase, error) {
	kycCase, err := uc.kyc.GetLatestByUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.KYCCase{UserID: userID, Status: domain.KYCStatusPending}, nil
		}
		return domain.KYCCase{}, err
	}
	return *kycCase, nil
}