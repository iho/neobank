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
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/iho/neobank/pkg/events"
)

// Record is a row in the outbox_events table.
type Record struct {
	CreatedAt     time.Time
	PublishedAt   *time.Time
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	CorrelationID string
	CausationID   string
	Payload       json.RawMessage
	EventVersion  int
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
