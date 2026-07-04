package port

import (
	"context"
	"time"

	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet domain.Wallet) error
	DeleteByID(ctx context.Context, walletID string) error
	GetByUserAndCurrency(ctx context.Context, userID, currency string) (*domain.Wallet, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Wallet, error)
	WithTx(tx pgx.Tx) WalletRepository
}

type KYCSubmission struct {
	ID                string
	KYCCaseID         string
	UserID            string
	DocumentType      string
	DocumentNumber    string
	Provider          string
	ProviderReference string
	ProviderResponse  []byte
	ScreeningDecision string
	ScreeningReason   string
	CorrelationID     string
	CreatedAt         time.Time
}

type ScreeningCheck struct {
	ID                string
	CheckType         string
	SubjectUserID     string
	EntityType        string
	EntityID          string
	Decision          string
	ReasonCode        string
	Provider          string
	ProviderReference string
	RawResponse       []byte
	CorrelationID     string
}

type GDPRRepository interface {
	ListKYCSubmissionsByUser(ctx context.Context, userID string) ([]KYCSubmission, error)
	ListWalletsByUser(ctx context.Context, userID string) ([]domain.Wallet, error)
	CountWalletTransactions(ctx context.Context, userID string) (int64, error)
	RecordRequest(ctx context.Context, userID, requestType string) error
	MaskUserPII(ctx context.Context, userID, maskedEmail, passwordHash string) error
	WithTx(tx pgx.Tx) GDPRRepository
}

type KYCRepository interface {
	UpsertProfile(ctx context.Context, userID, fullName, dob, country string) error
	CreateCase(ctx context.Context, id, userID, status string) (domain.KYCCase, error)
	GetLatestByUser(ctx context.Context, userID string) (*domain.KYCCase, error)
	GetByID(ctx context.Context, caseID string) (*domain.KYCCase, error)
	// GetByVendorApplicant resolves the KYC vendor's applicant ID back to a
	// case, so the async verdict webhook can find which case to update.
	GetByVendorApplicant(ctx context.Context, applicantID string) (*domain.KYCCase, error)
	SetVendorApplicant(ctx context.Context, caseID, applicantID string) error
	ApproveCase(ctx context.Context, caseID, decidedBy string) error
	RejectCase(ctx context.Context, caseID, reason, decidedBy string) error
	MarkManualReview(ctx context.Context, caseID string) error
	CreateSubmission(ctx context.Context, sub KYCSubmission) error
	RecordScreeningCheck(ctx context.Context, check ScreeningCheck) error
	WithTx(tx pgx.Tx) KYCRepository
}
