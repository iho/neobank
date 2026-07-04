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

package vendorsim

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Delivery is one attempted (or pending) webhook delivery.
type Delivery struct {
	CreatedAt     time.Time
	NextAttemptAt time.Time
	DeliveredAt   *time.Time
	ID            string
	URL           string
	EventType     string
	LastError     string
	Payload       json.RawMessage
	Attempts      int
}

// DeliveryStore persists webhook deliveries so Dispatcher can retry them and
// a simulator's admin API can list or replay them.
type DeliveryStore interface {
	Enqueue(ctx context.Context, d Delivery) error
	// ClaimDue returns pending deliveries whose NextAttemptAt is due,
	// oldest-due first, up to limit (0 = no limit).
	ClaimDue(ctx context.Context, now time.Time, limit int) ([]Delivery, error)
	MarkDelivered(ctx context.Context, id string, deliveredAt time.Time) error
	MarkFailed(ctx context.Context, id string, nextAttemptAt time.Time, errMsg string) error
	// List returns the most recently created deliveries, up to limit (0 = no limit).
	List(ctx context.Context, limit int) ([]Delivery, error)
	Get(ctx context.Context, id string) (Delivery, error)
}

// NewDelivery builds a pending Delivery ready to Enqueue, due immediately.
// Callers apply chaos delay by advancing NextAttemptAt before enqueueing.
func NewDelivery(url, eventType string, payload json.RawMessage) Delivery {
	now := time.Now().UTC()

	return Delivery{
		ID:            uuid.NewString(),
		URL:           url,
		EventType:     eventType,
		Payload:       payload,
		CreatedAt:     now,
		NextAttemptAt: now,
	}
}
