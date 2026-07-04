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
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type payload struct {
	Amount string `json:"amount"`
}

func TestDispatcherDeliversSignedPayload(t *testing.T) {
	var (
		gotCalls    int32
		gotEvent    string
		gotBody     []byte
		gotDelivery string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&gotCalls, 1)

		gotEvent = r.Header.Get(HeaderEventType)
		gotDelivery = r.Header.Get(HeaderDeliveryID)
		body, _ := io.ReadAll(r.Body)
		gotBody = body

		secret := []byte("test-secret")
		if err := VerifySignature(secret, r.Header.Get(HeaderTimestamp), r.Header.Get(HeaderSignature), body, time.Minute); err != nil {
			t.Errorf("consumer-side signature verification failed: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store := NewMemoryDeliveryStore()
	d := NewDispatcher(store, []byte("test-secret"), nil)

	id, err := d.Enqueue(context.Background(), srv.URL, "rails.transfer.received", payload{Amount: "10.00"})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	d.tick(context.Background())

	if atomic.LoadInt32(&gotCalls) != 1 {
		t.Fatalf("expected 1 delivery call, got %d", gotCalls)
	}

	if gotEvent != "rails.transfer.received" {
		t.Fatalf("unexpected event header: %q", gotEvent)
	}

	if gotDelivery != id {
		t.Fatalf("expected delivery id header %q, got %q", id, gotDelivery)
	}

	var got payload
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("unmarshal delivered body: %v", err)
	}

	if got.Amount != "10.00" {
		t.Fatalf("unexpected delivered payload: %+v", got)
	}

	stored, err := store.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if stored.DeliveredAt == nil {
		t.Fatal("expected delivery marked delivered")
	}
}

func TestDispatcherRetriesOnFailureAndBacksOff(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	store := NewMemoryDeliveryStore()
	d := NewDispatcher(store, []byte("test-secret"), nil)
	d.Backoff = BackoffConfig{Base: time.Minute, Max: time.Hour, MaxRetries: 10}

	id, err := d.Enqueue(context.Background(), srv.URL, "evt", payload{Amount: "1.00"})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	d.tick(context.Background())

	stored, err := store.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if stored.DeliveredAt != nil {
		t.Fatal("expected delivery not marked delivered after consumer error")
	}

	if stored.Attempts != 1 {
		t.Fatalf("expected 1 attempt recorded, got %d", stored.Attempts)
	}

	if !stored.NextAttemptAt.After(time.Now().UTC().Add(30 * time.Second)) {
		t.Fatalf("expected next attempt backed off into the future, got %v", stored.NextAttemptAt)
	}

	due, err := store.ClaimDue(context.Background(), time.Now().UTC(), 0)
	if err != nil {
		t.Fatalf("claim due: %v", err)
	}

	if len(due) != 0 {
		t.Fatal("expected delivery not immediately due again after backoff")
	}
}

func TestDispatcherGivesUpAfterMaxRetries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	store := NewMemoryDeliveryStore()
	d := NewDispatcher(store, []byte("test-secret"), nil)
	d.Backoff = BackoffConfig{Base: time.Millisecond, Max: time.Millisecond, MaxRetries: 1}

	id, err := d.Enqueue(context.Background(), srv.URL, "evt", payload{Amount: "1.00"})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	d.tick(context.Background())

	stored, err := store.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if stored.LastError == "" {
		t.Fatal("expected last error recorded")
	}
}

func TestDispatcherChaosDuplicateSendsTwice(t *testing.T) {
	var calls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store := NewMemoryDeliveryStore()
	d := NewDispatcher(store, []byte("test-secret"), nil)
	d.Chaos = ChaosConfig{DuplicateProb: 1}

	if _, err := d.Enqueue(context.Background(), srv.URL, "evt", payload{Amount: "1.00"}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	d.tick(context.Background())

	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 delivery attempts with DuplicateProb=1, got %d", calls)
	}
}
