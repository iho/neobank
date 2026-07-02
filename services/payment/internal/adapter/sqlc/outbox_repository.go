package sqlcrepo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/outbox"
	"github.com/iho/neobank/services/payment/internal/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type OutboxRepository struct {
	q sqlc.Querier
}

func NewOutboxRepository(q sqlc.Querier) *OutboxRepository {
	return &OutboxRepository{q: q}
}

func (r *OutboxRepository) Insert(ctx context.Context, record outbox.Record) error {
	id, err := uuid.Parse(record.ID)
	if err != nil {
		return err
	}
	return r.q.InsertOutboxEvent(ctx, sqlc.InsertOutboxEventParams{
		ID:            id,
		AggregateType: record.AggregateType,
		AggregateID:   record.AggregateID,
		EventType:     record.EventType,
		Payload:       record.Payload,
		CreatedAt:     pgtype.Timestamptz{Time: record.CreatedAt, Valid: true},
	})
}

func (r *OutboxRepository) FetchUnpublished(ctx context.Context, limit int) ([]outbox.Record, error) {
	rows, err := r.q.FetchUnpublishedOutboxEvents(ctx, int32(limit))
	if err != nil {
		return nil, err
	}
	records := make([]outbox.Record, 0, len(rows))
	for _, row := range rows {
		rec := outbox.Record{
			ID:            row.ID.String(),
			AggregateType: row.AggregateType,
			AggregateID:   row.AggregateID,
			EventType:     row.EventType,
			Payload:       row.Payload,
			CreatedAt:     row.CreatedAt.Time,
		}
		if row.PublishedAt.Valid {
			t := row.PublishedAt.Time
			rec.PublishedAt = &t
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *OutboxRepository) MarkPublished(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.q.MarkOutboxEventPublished(ctx, uid)
}

func (r *OutboxRepository) Publish(ctx context.Context, evt events.Event) error {
	record, err := outbox.BuildRecord(evt)
	if err != nil {
		return err
	}
	return r.Insert(ctx, record)
}

type LogProducer struct{}

func (LogProducer) Produce(_ context.Context, topic, key string, value []byte) error {
	var envelope events.Envelope
	_ = json.Unmarshal(value, &envelope)
	fmt.Printf("outbox publish topic=%s key=%s type=%s id=%s\n", topic, key, envelope.EventType, envelope.EventID)
	return nil
}