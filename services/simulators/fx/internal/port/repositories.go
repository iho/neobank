package port

import (
	"context"

	"github.com/iho/neobank/services/simulators/fx/internal/domain"
)

type QuoteRepository interface {
	Create(ctx context.Context, q domain.Quote) (domain.Quote, error)
	GetByID(ctx context.Context, id string) (*domain.Quote, error)
	MarkExecuted(ctx context.Context, id string) (domain.Quote, error)
}
