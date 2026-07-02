//
// Copyright (c) 2026 Sumicare
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

package events

import (
	"encoding/json"
	"time"
)

// Envelope is the standard Kafka event wrapper.
type Envelope struct {
	OccurredAt    time.Time       `json:"occurred_at"`
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	CorrelationID string          `json:"correlation_id"`
	CausationID   string          `json:"causation_id,omitempty"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	Payload       json.RawMessage `json:"payload"`
	EventVersion  int             `json:"event_version"`
}

// Event is implemented by all domain events published via outbox.
type Event interface {
	EventType() string
	AggregateType() string
	AggregateID() string
	Version() int
}
