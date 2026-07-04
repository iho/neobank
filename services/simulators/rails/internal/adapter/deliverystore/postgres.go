// Package deliverystore backs pkg/vendorsim.DeliveryStore with the
// rails.webhook_deliveries table, so delivery state survives simulator
// restarts (see docs/vendor-simulators-plan.md, Phase 0 "not yet done").
package deliverystore

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/vendorsim"
	"github.com/iho/neobank/services/simulators/rails/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type Postgres struct {
	q sqlc.Querier
}

func NewPostgres(q sqlc.Querier) *Postgres {
	return &Postgres{q: q}
}

var _ vendorsim.DeliveryStore = (*Postgres)(nil)

func (s *Postgres) Enqueue(ctx context.Context, d vendorsim.Delivery) error {
	id, err := pgutil.ParseUUID(d.ID)
	if err != nil {
		return err
	}

	return s.q.EnqueueDelivery(ctx, sqlc.EnqueueDeliveryParams{
		ID:            id,
		Url:           d.URL,
		EventType:     d.EventType,
		Payload:       d.Payload,
		NextAttemptAt: pgtype.Timestamptz{Time: d.NextAttemptAt.UTC(), Valid: true},
	})
}

func (s *Postgres) ClaimDue(ctx context.Context, now time.Time, limit int) ([]vendorsim.Delivery, error) {
	rows, err := s.q.ClaimDueDeliveries(ctx, sqlc.ClaimDueDeliveriesParams{
		Now:      pgtype.Timestamptz{Time: now.UTC(), Valid: true},
		LimitVal: limitOrMax(limit),
	})
	if err != nil {
		return nil, err
	}

	out := make([]vendorsim.Delivery, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDelivery(row))
	}

	return out, nil
}

func (s *Postgres) MarkDelivered(ctx context.Context, id string, deliveredAt time.Time) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}

	return s.q.MarkDeliveryDelivered(ctx, sqlc.MarkDeliveryDeliveredParams{
		ID:          uid,
		DeliveredAt: pgtype.Timestamptz{Time: deliveredAt.UTC(), Valid: true},
	})
}

func (s *Postgres) MarkFailed(ctx context.Context, id string, nextAttemptAt time.Time, errMsg string) error {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return err
	}

	return s.q.MarkDeliveryFailed(ctx, sqlc.MarkDeliveryFailedParams{
		ID:            uid,
		NextAttemptAt: pgtype.Timestamptz{Time: nextAttemptAt.UTC(), Valid: true},
		LastError:     errMsg,
	})
}

func (s *Postgres) List(ctx context.Context, limit int) ([]vendorsim.Delivery, error) {
	rows, err := s.q.ListDeliveries(ctx, limitOrMax(limit))
	if err != nil {
		return nil, err
	}

	out := make([]vendorsim.Delivery, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDelivery(row))
	}

	return out, nil
}

func (s *Postgres) Get(ctx context.Context, id string) (vendorsim.Delivery, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return vendorsim.Delivery{}, err
	}

	row, err := s.q.GetDelivery(ctx, uid)
	if err != nil {
		return vendorsim.Delivery{}, err
	}

	return toDelivery(row), nil
}

func toDelivery(row sqlc.RailsWebhookDelivery) vendorsim.Delivery {
	d := vendorsim.Delivery{
		ID:            row.ID.String(),
		URL:           row.Url,
		EventType:     row.EventType,
		LastError:     row.LastError,
		Payload:       json.RawMessage(row.Payload),
		Attempts:      int(row.Attempts),
		CreatedAt:     row.CreatedAt.Time.UTC(),
		NextAttemptAt: row.NextAttemptAt.Time.UTC(),
	}

	if row.DeliveredAt.Valid {
		t := row.DeliveredAt.Time.UTC()
		d.DeliveredAt = &t
	}

	return d
}

func limitOrMax(limit int) int32 {
	if limit <= 0 || limit > math.MaxInt32 {
		return math.MaxInt32
	}

	return int32(limit)
}
