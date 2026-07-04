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
	"fmt"
	"sort"
	"sync"
	"time"
)

// MemoryDeliveryStore is an in-process DeliveryStore for local development
// and tests. State is lost on restart; simulators that need durability
// across restarts should back DeliveryStore with Postgres instead.
type MemoryDeliveryStore struct {
	data map[string]Delivery
	mu   sync.Mutex
}

func NewMemoryDeliveryStore() *MemoryDeliveryStore {
	return &MemoryDeliveryStore{data: make(map[string]Delivery)}
}

func (s *MemoryDeliveryStore) Enqueue(_ context.Context, d Delivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[d.ID] = d

	return nil
}

func (s *MemoryDeliveryStore) ClaimDue(_ context.Context, now time.Time, limit int) ([]Delivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var due []Delivery

	for _, d := range s.data {
		if d.DeliveredAt == nil && !d.NextAttemptAt.After(now) {
			due = append(due, d)
		}
	}

	sort.Slice(due, func(i, j int) bool { return due[i].NextAttemptAt.Before(due[j].NextAttemptAt) })

	if limit > 0 && len(due) > limit {
		due = due[:limit]
	}

	return due, nil
}

func (s *MemoryDeliveryStore) MarkDelivered(_ context.Context, id string, deliveredAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.data[id]
	if !ok {
		return fmt.Errorf("vendorsim: delivery %q not found", id)
	}

	d.DeliveredAt = &deliveredAt
	d.Attempts++
	s.data[id] = d

	return nil
}

func (s *MemoryDeliveryStore) MarkFailed(_ context.Context, id string, nextAttemptAt time.Time, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.data[id]
	if !ok {
		return fmt.Errorf("vendorsim: delivery %q not found", id)
	}

	d.Attempts++
	d.NextAttemptAt = nextAttemptAt
	d.LastError = errMsg
	s.data[id] = d

	return nil
}

func (s *MemoryDeliveryStore) List(_ context.Context, limit int) ([]Delivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Delivery, 0, len(s.data))
	for _, d := range s.data {
		out = append(out, d)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}

	return out, nil
}

func (s *MemoryDeliveryStore) Get(_ context.Context, id string) (Delivery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.data[id]
	if !ok {
		return Delivery{}, fmt.Errorf("vendorsim: delivery %q not found", id)
	}

	return d, nil
}
