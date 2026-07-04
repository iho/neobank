package usecase

import (
	"context"
	"fmt"

	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/port"
)

type ProcessKYCVerdictInput struct {
	ApplicantID string
	// Verdict is one of "approved", "rejected", "manual_review" — the KYC
	// vendor simulator's magic-value conventions (see
	// docs/vendor-simulators-plan.md Phase 3).
	Verdict string
	Reason  string
}

// ProcessKYCVerdictUseCase advances a KYC case from the vendor's async
// verdict webhook (or its manual-review resolution): approved cases
// provision a wallet, same as the old synchronous auto-approve path did.
type ProcessKYCVerdictUseCase struct {
	kyc       port.KYCRepository
	provision *ProvisionWalletUseCase
	outbox    outbox.TxPublisher
	audit     audit.Recorder
}

func NewProcessKYCVerdictUseCase(
	kyc port.KYCRepository,
	provision *ProvisionWalletUseCase,
	outboxPublisher outbox.TxPublisher,
	auditRecorder audit.Recorder,
) *ProcessKYCVerdictUseCase {
	return &ProcessKYCVerdictUseCase{kyc: kyc, provision: provision, outbox: outboxPublisher, audit: auditRecorder}
}

func (uc *ProcessKYCVerdictUseCase) Execute(ctx context.Context, in ProcessKYCVerdictInput) (domain.KYCCase, error) {
	if in.ApplicantID == "" {
		return domain.KYCCase{}, fmt.Errorf("applicant_id is required")
	}

	kycCase, err := uc.kyc.GetByVendorApplicant(ctx, in.ApplicantID)
	if err != nil {
		return domain.KYCCase{}, err
	}

	if kycCase.Status != domain.KYCStatusPending && kycCase.Status != domain.KYCStatusManualReview {
		// Already decided — a redelivered or duplicated webhook is a no-op.
		return *kycCase, nil
	}

	switch in.Verdict {
	case "approved":
		return uc.approve(ctx, *kycCase)
	case "rejected":
		return uc.reject(ctx, *kycCase, in.Reason)
	case "manual_review":
		if err := uc.kyc.MarkManualReview(ctx, kycCase.ID); err != nil {
			return domain.KYCCase{}, err
		}
		kycCase.Status = domain.KYCStatusManualReview
		return *kycCase, nil
	default:
		return domain.KYCCase{}, fmt.Errorf("unknown verdict %q", in.Verdict)
	}
}

func (uc *ProcessKYCVerdictUseCase) approve(ctx context.Context, kycCase domain.KYCCase) (domain.KYCCase, error) {
	decidedBy := "kyc-vendor-simulator"
	if err := uc.kyc.ApproveCase(ctx, kycCase.ID, decidedBy); err != nil {
		return domain.KYCCase{}, err
	}

	if err := uc.audit.Record(ctx, audit.Entry{
		EntityType: "kyc_case",
		EntityID:   kycCase.ID,
		Action:     "approved",
		FromStatus: string(kycCase.Status),
		ToStatus:   string(domain.KYCStatusApproved),
		Metadata:   map[string]any{"decided_by": decidedBy},
	}); err != nil {
		return domain.KYCCase{}, err
	}

	if _, err := uc.provision.Execute(ctx, ProvisionWalletInput{
		UserID:         kycCase.UserID,
		Currency:       "USD",
		IdempotencyKey: walletProvisionKey("", kycCase.UserID),
	}); err != nil {
		return domain.KYCCase{}, fmt.Errorf("provision wallet: %w", err)
	}

	if err := uc.outbox.Publish(ctx, events.KYCApproved{
		UserID:    kycCase.UserID,
		KYCCaseID: kycCase.ID,
	}); err != nil {
		return domain.KYCCase{}, fmt.Errorf("publish kyc approved event: %w", err)
	}

	kycCase.Status = domain.KYCStatusApproved

	return kycCase, nil
}

func (uc *ProcessKYCVerdictUseCase) reject(ctx context.Context, kycCase domain.KYCCase, reason string) (domain.KYCCase, error) {
	if reason == "" {
		reason = "vendor_rejected"
	}

	if err := uc.kyc.RejectCase(ctx, kycCase.ID, reason, "kyc-vendor-simulator"); err != nil {
		return domain.KYCCase{}, err
	}

	if err := uc.audit.Record(ctx, audit.Entry{
		EntityType: "kyc_case",
		EntityID:   kycCase.ID,
		Action:     "rejected",
		FromStatus: string(kycCase.Status),
		ToStatus:   string(domain.KYCStatusRejected),
		Metadata:   map[string]any{"reason": reason, "decided_by": "kyc-vendor-simulator"},
	}); err != nil {
		return domain.KYCCase{}, err
	}

	if err := uc.outbox.Publish(ctx, events.KYCRejected{
		UserID:          kycCase.UserID,
		KYCCaseID:       kycCase.ID,
		RejectionReason: reason,
	}); err != nil {
		return domain.KYCCase{}, fmt.Errorf("publish kyc rejected event: %w", err)
	}

	kycCase.Status = domain.KYCStatusRejected
	kycCase.RejectionReason = reason

	return kycCase, nil
}
