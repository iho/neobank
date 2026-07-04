package sqlcrepo

import (
	"context"
	"time"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/simulators/rails/internal/domain"
	"github.com/iho/neobank/services/simulators/rails/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type InboundTransferRepository struct {
	q sqlc.Querier
}

func NewInboundTransferRepository(q sqlc.Querier) *InboundTransferRepository {
	return &InboundTransferRepository{q: q}
}

func (r *InboundTransferRepository) Create(ctx context.Context, accountID, amount, currency, senderName, reference string) (domain.InboundTransfer, error) {
	acctID, err := pgutil.ParseUUID(accountID)
	if err != nil {
		return domain.InboundTransfer{}, err
	}

	numAmount, err := pgutil.NumericFromString(amount)
	if err != nil {
		return domain.InboundTransfer{}, err
	}

	row, err := r.q.CreateInboundTransfer(ctx, sqlc.CreateInboundTransferParams{
		AccountID:  acctID,
		Amount:     numAmount,
		Currency:   currency,
		SenderName: senderName,
		Reference:  reference,
	})
	if err != nil {
		return domain.InboundTransfer{}, err
	}

	return domain.InboundTransfer{
		ID:         row.ID.String(),
		AccountID:  row.AccountID.String(),
		Amount:     row.Amount,
		Currency:   row.Currency,
		SenderName: row.SenderName,
		Reference:  row.Reference,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt.Time.UTC(),
	}, nil
}

func (r *InboundTransferRepository) ListInRange(ctx context.Context, from, to time.Time) ([]domain.InboundTransfer, error) {
	rows, err := r.q.ListInboundTransfersInRange(ctx, sqlc.ListInboundTransfersInRangeParams{
		FromTs: pgtype.Timestamptz{Time: from.UTC(), Valid: true},
		ToTs:   pgtype.Timestamptz{Time: to.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	out := make([]domain.InboundTransfer, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.InboundTransfer{
			ID:         row.ID.String(),
			AccountID:  row.AccountID.String(),
			Amount:     row.Amount,
			Currency:   row.Currency,
			SenderName: row.SenderName,
			Reference:  row.Reference,
			Status:     row.Status,
			CreatedAt:  row.CreatedAt.Time.UTC(),
		})
	}

	return out, nil
}
