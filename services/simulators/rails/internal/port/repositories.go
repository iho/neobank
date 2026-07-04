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
