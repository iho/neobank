package port

import (
	"context"

	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
)

type WalletRepository interface {
	Create(ctx context.Context, wallet domain.Wallet) error
	DeleteByID(ctx context.Context, walletID string) error
	GetByUserAndCurrency(ctx context.Context, userID, currency string) (*domain.Wallet, error)
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

type KYCRepository interface {
	UpsertProfile(ctx context.Context, userID, fullName, dob, country string) error
	CreateCase(ctx context.Context, id, userID, status string) (domain.KYCCase, error)
	GetLatestByUser(ctx context.Context, userID string) (*domain.KYCCase, error)
	ApproveCase(ctx context.Context, caseID, decidedBy string) error
	RejectCase(ctx context.Context, caseID, reason, decidedBy string) error
	CreateSubmission(ctx context.Context, sub KYCSubmission) error
	RecordScreeningCheck(ctx context.Context, check ScreeningCheck) error
	WithTx(tx pgx.Tx) KYCRepository
}