package port

import (
	"context"
	"time"

	"github.com/iho/neobank/services/simulators/rails/internal/domain"
)

type AccountRepository interface {
	Create(ctx context.Context, externalRef, currency, iban string) (domain.Account, error)
	GetByExternalRefAndCurrency(ctx context.Context, externalRef, currency string) (*domain.Account, error)
	GetByID(ctx context.Context, id string) (*domain.Account, error)
}

type InboundTransferRepository interface {
	Create(ctx context.Context, accountID, amount, currency, senderName, reference string) (domain.InboundTransfer, error)
	ListInRange(ctx context.Context, from, to time.Time) ([]domain.InboundTransfer, error)
}

type OutboundPaymentRepository interface {
	Create(ctx context.Context, accountID, amount, currency, counterpartyIBAN, reference string) (domain.OutboundPayment, error)
	GetByID(ctx context.Context, id string) (*domain.OutboundPayment, error)
	SetStatus(ctx context.Context, id, status string) error
	ListInRange(ctx context.Context, from, to time.Time) ([]domain.OutboundPayment, error)
}
