package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/iho/neobank/services/simulators/rails/internal/domain"
	"github.com/iho/neobank/services/simulators/rails/internal/port"
)

// GetStatementUseCase returns everything that moved on the rail for a given
// day, independent of whether the corresponding webhook was ever delivered
// successfully — this is the reconciliation source of truth.
type GetStatementUseCase struct {
	transfers port.InboundTransferRepository
}

func NewGetStatementUseCase(transfers port.InboundTransferRepository) *GetStatementUseCase {
	return &GetStatementUseCase{transfers: transfers}
}

func (uc *GetStatementUseCase) Execute(ctx context.Context, date string) ([]domain.InboundTransfer, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q, expected YYYY-MM-DD: %w", date, err)
	}

	from := day.UTC()
	to := from.Add(24 * time.Hour)

	return uc.transfers.ListInRange(ctx, from, to)
}
