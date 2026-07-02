//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/pkg/reqctx"
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
	logger   *slog.Logger
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
		logger:   slog.Default(),
	}
}

// WithLogger overrides the logger used to report flush failures (defaults to slog.Default()).
func (w *Worker) WithLogger(logger *slog.Logger) *Worker {
	w.logger = logger
	return w
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := w.flush(ctx)
			if err != nil {
				w.logger.Error("outbox flush failed", "error", err, "topic", w.topic)
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
		version := rec.EventVersion
		if version == 0 {
			version = 1
		}

		envelope := events.Envelope{
			EventID:       rec.ID,
			EventType:     rec.EventType,
			EventVersion:  version,
			OccurredAt:    rec.CreatedAt,
			CorrelationID: rec.CorrelationID,
			CausationID:   rec.CausationID,
			AggregateType: rec.AggregateType,
			AggregateID:   rec.AggregateID,
			Payload:       rec.Payload,
		}

		data, err := json.Marshal(envelope)
		if err != nil {
			return err
		}
		// Carry the event's own correlation ID (set at write time, not the
		// worker loop's ambient context) so producers using reqctx.Transport
		// tag the delivery request with the ID that produced the event.
		produceCtx := reqctx.WithCorrelationID(ctx, rec.CorrelationID)

		produceCtx = reqctx.WithCausationID(produceCtx, rec.CausationID)
		if err := w.producer.Produce(produceCtx, w.topic, rec.AggregateID, data); err != nil {
			return err
		}

		if err := w.store.MarkPublished(ctx, rec.ID); err != nil {
			return err
		}
	}

	return nil
}

// BuildRecord creates an outbox record from a domain event, tagging it with
// the correlation/causation IDs from ctx (see pkg/reqctx) so the full chain
// from API request to published event can be reconstructed later.
func BuildRecord(ctx context.Context, evt events.Event) (Record, error) {
	payload, err := events.MarshalPayload(evt)
	if err != nil {
		return Record{}, fmt.Errorf("marshal event payload: %w", err)
	}

	return Record{
		ID:            uuid.NewString(),
		AggregateType: evt.AggregateType(),
		AggregateID:   evt.AggregateID(),
		EventType:     evt.EventType(),
		EventVersion:  evt.Version(),
		Payload:       payload,
		CorrelationID: reqctx.CorrelationID(ctx),
		CausationID:   reqctx.CausationID(ctx),
		CreatedAt:     time.Now().UTC(),
	}, nil
}
