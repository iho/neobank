package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/iho/neobank/pkg/events"
)

// Record is a row in the outbox_events table.
type Record struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       json.RawMessage
	CreatedAt     time.Time
	PublishedAt   *time.Time
}

// Publisher writes events to the outbox within a transaction.
type Publisher interface {
	Publish(ctx context.Context, evt events.Event) error
}

// Producer delivers serialized events to the message bus.
type Producer interface {
	Produce(ctx context.Context, topic, key string, value []byte) error
}