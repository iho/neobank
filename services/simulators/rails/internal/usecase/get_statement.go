package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/iho/neobank/services/simulators/rails/internal/domain"
	"github.com/iho/neobank/services/simulators/rails/internal/port"
)

// Statement is everything that moved on the rail for a given day,
// independent of whether the corresponding webhook was ever delivered
// successfully — this is the reconciliation source of truth.
type Statement struct {
	InboundTransfers []domain.InboundTransfer
	OutboundPayments []domain.OutboundPayment
}

type GetStatementUseCase struct {
	transfers port.InboundTransferRepository
	payments  port.OutboundPaymentRepository
}

func NewGetStatementUseCase(transfers port.InboundTransferRepository, payments port.OutboundPaymentRepository) *GetStatementUseCase {
	return &GetStatementUseCase{transfers: transfers, payments: payments}
}

func (uc *GetStatementUseCase) Execute(ctx context.Context, date string) (Statement, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return Statement{}, fmt.Errorf("invalid date %q, expected YYYY-MM-DD: %w", date, err)
	}

	from := day.UTC()
	to := from.Add(24 * time.Hour)

	inbound, err := uc.transfers.ListInRange(ctx, from, to)
	if err != nil {
		return Statement{}, err
	}

	outbound, err := uc.payments.ListInRange(ctx, from, to)
	if err != nil {
		return Statement{}, err
	}

	return Statement{InboundTransfers: inbound, OutboundPayments: outbound}, nil
}
