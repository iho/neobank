package port

import (
	"context"

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
}
