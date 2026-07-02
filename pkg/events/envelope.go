package events

import (
	"encoding/json"
	"time"
)

// Envelope is the standard Kafka event wrapper.
type Envelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	EventVersion  int             `json:"event_version"`
	OccurredAt    time.Time       `json:"occurred_at"`
	CorrelationID string          `json:"correlation_id"`
	CausationID   string          `json:"causation_id,omitempty"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
}

// Event is implemented by all domain events published via outbox.
type Event interface {
	EventType() string
	AggregateType() string
	AggregateID() string
	Version() int
}