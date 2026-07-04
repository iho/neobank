package sqlcrepo

import (
	"context"
	"errors"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/simulators/cardproc/internal/domain"
	"github.com/iho/neobank/services/simulators/cardproc/internal/gen/sqlc"
	"github.com/jackc/pgx/v5"
)

type CardRepository struct {
	q sqlc.Querier
}

func NewCardRepository(q sqlc.Querier) *CardRepository {
	return &CardRepository{q: q}
}

func (r *CardRepository) Create(ctx context.Context, externalRef, cardholderName, panToken, lastFour string, expiryMonth, expiryYear int) (domain.Card, error) {
	row, err := r.q.CreateCard(ctx, sqlc.CreateCardParams{
		ExternalRef:    externalRef,
		CardholderName: cardholderName,
		PanToken:       panToken,
		LastFour:       lastFour,
		ExpiryMonth:    int16(expiryMonth),
		ExpiryYear:     int16(expiryYear),
	})
	if err != nil {
		return domain.Card{}, err
	}

	return toCard(row), nil
}

func (r *CardRepository) GetByID(ctx context.Context, id string) (*domain.Card, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.q.GetCardByID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	card := toCard(row)

	return &card, nil
}

func (r *CardRepository) Cancel(ctx context.Context, id string) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}

	return r.q.CancelCard(ctx, uid)
}

func toCard(row sqlc.CardprocCard) domain.Card {
	return domain.Card{
		ID:             row.ID.String(),
		ExternalRef:    row.ExternalRef,
		CardholderName: row.CardholderName,
		PANToken:       row.PanToken,
		LastFour:       row.LastFour,
		ExpiryMonth:    int(row.ExpiryMonth),
		ExpiryYear:     int(row.ExpiryYear),
		Status:         row.Status,
		CreatedAt:      row.CreatedAt.Time.UTC(),
	}
}
