package sqlcrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/iho/neobank/services/payment/internal/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type BankTransferOrderRepository struct {
	q sqlc.Querier
}

func NewBankTransferOrderRepository(q sqlc.Querier) *BankTransferOrderRepository {
	return &BankTransferOrderRepository{q: q}
}

func (r *BankTransferOrderRepository) WithTx(tx pgx.Tx) port.BankTransferOrderRepository {
	return &BankTransferOrderRepository{q: withTx(r.q, tx)}
}

func (r *BankTransferOrderRepository) Create(ctx context.Context, o domain.BankTransferOrder) (domain.BankTransferOrder, error) {
	uid, err := pgutil.ParseUUID(o.UserID)
	if err != nil {
		return domain.BankTransferOrder{}, err
	}

	amount, err := pgutil.NumericFromString(o.Amount)
	if err != nil {
		return domain.BankTransferOrder{}, err
	}

	row, err := r.q.CreateBankTransferOrder(ctx, sqlc.CreateBankTransferOrderParams{
		RailsPaymentID:   o.RailsPaymentID,
		UserID:           uid,
		Amount:           amount,
		Currency:         o.Currency,
		CounterpartyIban: o.CounterpartyIBAN,
		Reference:        o.Reference,
		LedgerTransferID: o.LedgerTransferID,
	})
	if err != nil {
		return domain.BankTransferOrder{}, err
	}

	return toBankTransferOrder(
		row.ID, row.RailsPaymentID, row.UserID, row.Amount, row.Currency, row.CounterpartyIban,
		row.Reference, row.LedgerTransferID, row.ReturnTransferID, row.Status, row.CreatedAt, row.UpdatedAt,
	), nil
}

func (r *BankTransferOrderRepository) GetByRailsPaymentID(ctx context.Context, railsPaymentID string) (*domain.BankTransferOrder, error) {
	row, err := r.q.GetBankTransferOrderByRailsPaymentID(ctx, railsPaymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	order := toBankTransferOrder(
		row.ID, row.RailsPaymentID, row.UserID, row.Amount, row.Currency, row.CounterpartyIban,
		row.Reference, row.LedgerTransferID, row.ReturnTransferID, row.Status, row.CreatedAt, row.UpdatedAt,
	)

	return &order, nil
}

func (r *BankTransferOrderRepository) MarkSettled(ctx context.Context, railsPaymentID string) error {
	return r.q.MarkBankTransferOrderSettled(ctx, railsPaymentID)
}

func (r *BankTransferOrderRepository) MarkReturned(ctx context.Context, railsPaymentID, status, returnTransferID string) error {
	return r.q.MarkBankTransferOrderReturned(ctx, sqlc.MarkBankTransferOrderReturnedParams{
		RailsPaymentID:   railsPaymentID,
		Status:           status,
		ReturnTransferID: pgutil.Text(returnTransferID),
	})
}

func toBankTransferOrder(
	id uuid.UUID,
	railsPaymentID string,
	userID uuid.UUID,
	amount, currency, counterpartyIBAN, reference, ledgerTransferID, returnTransferID, status string,
	createdAt, updatedAt pgtype.Timestamptz,
) domain.BankTransferOrder {
	return domain.BankTransferOrder{
		ID:               id.String(),
		RailsPaymentID:   railsPaymentID,
		UserID:           userID.String(),
		Amount:           amount,
		Currency:         currency,
		CounterpartyIBAN: counterpartyIBAN,
		Reference:        reference,
		LedgerTransferID: ledgerTransferID,
		ReturnTransferID: returnTransferID,
		Status:           status,
		CreatedAt:        createdAt.Time.UTC(),
		UpdatedAt:        updatedAt.Time.UTC(),
	}
}
