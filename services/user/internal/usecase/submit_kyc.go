package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/port"
	"github.com/jackc/pgx/v5"
)

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
	kyc       port.KYCRepository
	provision *ProvisionWalletUseCase
	outbox    outbox.TxPublisher
	audit     audit.Recorder
	tx        *pgutil.TxRunner
}

func NewSubmitKYCUseCase(kyc port.KYCRepository, provision *ProvisionWalletUseCase, outboxPublisher outbox.TxPublisher, auditRecorder audit.Recorder, tx *pgutil.TxRunner) *SubmitKYCUseCase {
	return &SubmitKYCUseCase{kyc: kyc, provision: provision, outbox: outboxPublisher, audit: auditRecorder, tx: tx}
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
		wallet, wErr := uc.provision.Execute(ctx, ProvisionWalletInput{
			UserID:         in.UserID,
			Currency:       "USD",
			IdempotencyKey: walletProvisionKey(in.IdempotencyKey, in.UserID),
		})
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
	if err := uc.audit.Record(ctx, audit.Entry{
		EntityType: "kyc_case",
		EntityID:   caseID,
		Action:     "submitted",
		ToStatus:   string(domain.KYCStatusPending),
		Metadata:   map[string]any{"user_id": in.UserID, "country_code": in.CountryCode},
	}); err != nil {
		return SubmitKYCOutput{}, err
	}

	if err := uc.tx.Run(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := uc.kyc.WithTx(tx).ApproveCase(ctx, caseID); err != nil {
			return err
		}
		return uc.audit.WithTx(tx).Record(ctx, audit.Entry{
			EntityType: "kyc_case",
			EntityID:   caseID,
			Action:     "approved",
			FromStatus: string(domain.KYCStatusPending),
			ToStatus:   string(domain.KYCStatusApproved),
		})
	}); err != nil {
		return SubmitKYCOutput{}, err
	}

	wallet, err := uc.provision.Execute(ctx, ProvisionWalletInput{
		UserID:         in.UserID,
		Currency:       "USD",
		IdempotencyKey: walletProvisionKey(in.IdempotencyKey, in.UserID),
	})
	if err != nil {
		return SubmitKYCOutput{}, fmt.Errorf("provision wallet: %w", err)
	}

	if err := uc.outbox.Publish(ctx, events.KYCApproved{
		UserID:    in.UserID,
		KYCCaseID: kycCase.ID,
	}); err != nil {
		return SubmitKYCOutput{}, fmt.Errorf("publish kyc approved event: %w", err)
	}

	return SubmitKYCOutput{
		KYCCaseID: kycCase.ID,
		Status:    domain.KYCStatusApproved,
		WalletID:  wallet.WalletID,
	}, nil
}

func walletProvisionKey(idempotencyKey, userID string) string {
	if idempotencyKey != "" {
		return idempotencyKey + ":wallet"
	}
	return fmt.Sprintf("wallet:%s:USD", userID)
}

type GetKYCStatusUseCase struct {
	kyc port.KYCRepository
}

func NewGetKYCStatusUseCase(kyc port.KYCRepository) *GetKYCStatusUseCase {
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
