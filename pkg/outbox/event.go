package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/iho/neobank/pkg/events"
	"github.com/jackc/pgx/v5"
)

// Record is a row in the outbox_events table.
type Record struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	EventVersion  int
	Payload       json.RawMessage
	CorrelationID string
	CausationID   string
	CreatedAt     time.Time
	PublishedAt   *time.Time
}

// Publisher writes events to the outbox table.
type Publisher interface {
	Publish(ctx context.Context, evt events.Event) error
}

// TxPublisher supports publishing inside an open database transaction.
type TxPublisher interface {
	Publisher
	WithTx(tx pgx.Tx) TxPublisher
}

// Producer delivers serialized events to the message bus.
type Producer interface {
	Produce(ctx context.Context, topic, key string, value []byte) error
}
