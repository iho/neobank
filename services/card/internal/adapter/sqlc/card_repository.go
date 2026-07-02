package sqlcrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/gen/sqlc"
	"github.com/iho/neobank/services/card/internal/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CardRepository struct {
	q sqlc.Querier
}

func NewCardRepository(q sqlc.Querier) *CardRepository {
	return &CardRepository{q: q}
}

func (r *CardRepository) WithTx(tx pgx.Tx) port.CardRepository {
	return &CardRepository{q: withTx(r.q, tx)}
}

func (r *CardRepository) Create(ctx context.Context, c domain.Card) error {
	id, err := pgutil.ParseUUID(c.ID)
	if err != nil {
		return err
	}
	userID, err := pgutil.ParseUUID(c.UserID)
	if err != nil {
		return err
	}
	walletID, err := pgutil.ParseUUID(c.WalletID)
	if err != nil {
		return err
	}
	return r.q.CreateCard(ctx, sqlc.CreateCardParams{
		ID:             id,
		UserID:         userID,
		WalletID:       walletID,
		ProcessorRef:   pgutil.Text(c.ProcessorRef),
		PanToken:       c.PANToken,
		LastFour:       c.LastFour,
		ExpiryMonth:    int16(c.ExpiryMonth),
		ExpiryYear:     int16(c.ExpiryYear),
		Status:         string(c.Status),
		IdempotencyKey: c.IdempotencyKey,
	})
}

func (r *CardRepository) GetByID(ctx context.Context, id string) (*domain.Card, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetCardByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return mapCardRow(
		row.ID, row.UserID, row.WalletID, row.ProcessorRef, row.PanToken,
		row.LastFour, row.ExpiryMonth, row.ExpiryYear, row.Status,
		row.IdempotencyKey, row.CreatedAt,
	), nil
}

func (r *CardRepository) GetByUserAndIdempotencyKey(ctx context.Context, userID, key string) (*domain.Card, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetCardByUserAndIdempotencyKey(ctx, sqlc.GetCardByUserAndIdempotencyKeyParams{
		UserID:         uid,
		IdempotencyKey: key,
	})
	if err != nil {
		return nil, err
	}
	return mapCardRow(
		row.ID, row.UserID, row.WalletID, row.ProcessorRef, row.PanToken,
		row.LastFour, row.ExpiryMonth, row.ExpiryYear, row.Status,
		row.IdempotencyKey, row.CreatedAt,
	), nil
}

func (r *CardRepository) ListByUser(ctx context.Context, userID string) ([]domain.Card, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListCardsByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Card, 0, len(rows))
	for _, row := range rows {
		card := mapCardRow(
			row.ID, row.UserID, row.WalletID, row.ProcessorRef, row.PanToken,
			row.LastFour, row.ExpiryMonth, row.ExpiryYear, row.Status,
			row.IdempotencyKey, row.CreatedAt,
		)
		out = append(out, *card)
	}
	return out, nil
}

func (r *CardRepository) UpdateStatus(ctx context.Context, id, userID string, status domain.CardStatus) error {
	cid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return err
	}
	return r.q.UpdateCardStatus(ctx, sqlc.UpdateCardStatusParams{
		ID:     cid,
		Status: string(status),
		UserID: uid,
	})
}

func (r *CardRepository) MarkCancelled(ctx context.Context, id string) error {
	cid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}
	return r.q.MarkCardCancelled(ctx, cid)
}

func mapCardRow(
	id, userID, walletID uuid.UUID,
	processorRef, panToken, lastFour string,
	expiryMonth, expiryYear int16,
	status, idempotencyKey string,
	createdAt pgtype.Timestamptz,
) *domain.Card {
	c := &domain.Card{
		ID:             id.String(),
		UserID:         userID.String(),
		WalletID:       walletID.String(),
		ProcessorRef:   processorRef,
		PANToken:       panToken,
		LastFour:       lastFour,
		ExpiryMonth:    int(expiryMonth),
		ExpiryYear:     int(expiryYear),
		Status:         domain.CardStatus(status),
		IdempotencyKey: idempotencyKey,
	}
	if createdAt.Valid {
		c.CreatedAt = createdAt.Time.UTC()
	}
	return c
}
