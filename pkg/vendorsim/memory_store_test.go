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
	"testing"
	"time"
)

func TestMemoryDeliveryStoreClaimDueOrdersOldestFirst(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryDeliveryStore()
	now := time.Now().UTC()

	late := NewDelivery("http://a", "evt", nil)
	late.NextAttemptAt = now.Add(-time.Second)

	early := NewDelivery("http://b", "evt", nil)
	early.NextAttemptAt = now.Add(-time.Minute)

	future := NewDelivery("http://c", "evt", nil)
	future.NextAttemptAt = now.Add(time.Hour)

	for _, d := range []Delivery{late, early, future} {
		if err := store.Enqueue(ctx, d); err != nil {
			t.Fatalf("enqueue: %v", err)
		}
	}

	due, err := store.ClaimDue(ctx, now, 0)
	if err != nil {
		t.Fatalf("claim due: %v", err)
	}

	if len(due) != 2 {
		t.Fatalf("expected 2 due deliveries, got %d", len(due))
	}

	if due[0].ID != early.ID || due[1].ID != late.ID {
		t.Fatalf("expected oldest-due-first ordering, got %v", due)
	}
}

func TestMemoryDeliveryStoreClaimDueRespectsLimit(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryDeliveryStore()
	now := time.Now().UTC()

	for range 5 {
		d := NewDelivery("http://a", "evt", nil)
		d.NextAttemptAt = now.Add(-time.Second)

		if err := store.Enqueue(ctx, d); err != nil {
			t.Fatalf("enqueue: %v", err)
		}
	}

	due, err := store.ClaimDue(ctx, now, 2)
	if err != nil {
		t.Fatalf("claim due: %v", err)
	}

	if len(due) != 2 {
		t.Fatalf("expected limit of 2, got %d", len(due))
	}
}

func TestMemoryDeliveryStoreMarkDeliveredExcludesFromClaimDue(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryDeliveryStore()
	now := time.Now().UTC()

	d := NewDelivery("http://a", "evt", nil)
	if err := store.Enqueue(ctx, d); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	if err := store.MarkDelivered(ctx, d.ID, now); err != nil {
		t.Fatalf("mark delivered: %v", err)
	}

	due, err := store.ClaimDue(ctx, now, 0)
	if err != nil {
		t.Fatalf("claim due: %v", err)
	}

	if len(due) != 0 {
		t.Fatalf("expected delivered delivery excluded from claim due, got %d", len(due))
	}

	got, err := store.Get(ctx, d.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.DeliveredAt == nil || got.Attempts != 1 {
		t.Fatalf("expected delivered=true attempts=1, got %+v", got)
	}
}

func TestMemoryDeliveryStoreMarkFailedReschedules(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryDeliveryStore()
	now := time.Now().UTC()

	d := NewDelivery("http://a", "evt", nil)
	if err := store.Enqueue(ctx, d); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	next := now.Add(time.Hour)
	if err := store.MarkFailed(ctx, d.ID, next, "boom"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	due, err := store.ClaimDue(ctx, now, 0)
	if err != nil {
		t.Fatalf("claim due: %v", err)
	}

	if len(due) != 0 {
		t.Fatalf("expected rescheduled delivery excluded from claim due, got %d", len(due))
	}

	got, err := store.Get(ctx, d.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.Attempts != 1 || got.LastError != "boom" || !got.NextAttemptAt.Equal(next) {
		t.Fatalf("unexpected delivery state: %+v", got)
	}
}

func TestMemoryDeliveryStoreGetNotFound(t *testing.T) {
	store := NewMemoryDeliveryStore()

	if _, err := store.Get(context.Background(), "missing"); err == nil {
		t.Fatal("expected error for missing delivery")
	}
}
