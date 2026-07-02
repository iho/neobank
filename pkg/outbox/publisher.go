package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
)

// Store persists and retrieves outbox records.
type Store interface {
	Insert(ctx context.Context, record Record) error
	FetchUnpublished(ctx context.Context, limit int) ([]Record, error)
	MarkPublished(ctx context.Context, id string) error
}

// Worker flushes unpublished outbox events to the producer.
type Worker struct {
	store    Store
	producer Producer
	topic    string
	interval time.Duration
	batch    int
}

func NewWorker(store Store, producer Producer, topic string) *Worker {
	return &Worker{
		store:    store,
		producer: producer,
		topic:    topic,
		interval: 100 * time.Millisecond,
		batch:    100,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.flush(ctx); err != nil {
				// Log at call site; worker keeps running.
				_ = err
			}
		}
	}
}

func (w *Worker) flush(ctx context.Context) error {
	records, err := w.store.FetchUnpublished(ctx, w.batch)
	if err != nil {
		return err
	}
	for _, rec := range records {
		envelope := events.Envelope{
			EventID:       rec.ID,
			EventType:     rec.EventType,
			EventVersion:  1,
			OccurredAt:    rec.CreatedAt,
			AggregateType: rec.AggregateType,
			AggregateID:   rec.AggregateID,
			Payload:       rec.Payload,
		}
		data, err := json.Marshal(envelope)
		if err != nil {
			return err
		}
		if err := w.producer.Produce(ctx, w.topic, rec.AggregateID, data); err != nil {
			return err
		}
		if err := w.store.MarkPublished(ctx, rec.ID); err != nil {
			return err
		}
	}
	return nil
}

// BuildRecord creates an outbox record from a domain event.
func BuildRecord(evt events.Event) (Record, error) {
	payload, err := events.MarshalPayload(evt)
	if err != nil {
		return Record{}, fmt.Errorf("marshal event payload: %w", err)
	}
	return Record{
		ID:            uuid.NewString(),
		AggregateType: evt.AggregateType(),
		AggregateID:   evt.AggregateID(),
		EventType:     evt.EventType(),
		Payload:       payload,
		CreatedAt:     time.Now().UTC(),
	}, nil
}