package port

import (
	"context"
	"time"

	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
)

type CardRepository interface {
	Create(ctx context.Context, externalRef, cardholderName, panToken, lastFour string, expiryMonth, expiryYear int) (domain.Card, error)
	GetByID(ctx context.Context, id string) (*domain.Card, error)
	Cancel(ctx context.Context, id string) error
}

type TransactionRepository interface {
	Create(ctx context.Context, cardID, amount, currency, merchantName, mcc string) (domain.Transaction, error)
	GetByID(ctx context.Context, id string) (*domain.Transaction, error)
	SetAuthResult(ctx context.Context, id, status, authorizationID, reasonCode string) (domain.Transaction, error)
	MarkCaptured(ctx context.Context, id string) error
	MarkReversed(ctx context.Context, id string) error
	MarkExpired(ctx context.Context, id string) error
	ListExpiredApproved(ctx context.Context, cutoff time.Time, limit int32) ([]domain.Transaction, error)
}

type ChargebackRepository interface {
	Create(ctx context.Context, transactionID, authorizationID, amount, currency, reason string) (domain.Chargeback, error)
	GetByID(ctx context.Context, id string) (*domain.Chargeback, error)
	SetStatus(ctx context.Context, id, status string) (domain.Chargeback, error)
}
