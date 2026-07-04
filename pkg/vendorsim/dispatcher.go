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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// BackoffConfig controls retry spacing for failed webhook deliveries.
type BackoffConfig struct {
	Base       time.Duration
	Max        time.Duration
	MaxRetries int
}

// DefaultBackoff retries at 2s, 4s, 8s, ... capped at 5m, giving up after 10 attempts.
var DefaultBackoff = BackoffConfig{Base: 2 * time.Second, Max: 5 * time.Minute, MaxRetries: 10}

func (b BackoffConfig) delay(attempt int) time.Duration {
	d := b.Base
	for range attempt {
		d *= 2
		if d > b.Max {
			return b.Max
		}
	}

	return d
}

// Dispatcher delivers signed webhooks to consumer services with retries and
// optional chaos (delay/duplicate/reorder) for integration testing.
type Dispatcher struct {
	Store   DeliveryStore
	Client  *http.Client
	Logger  *slog.Logger
	Secret  []byte
	Chaos   ChaosConfig
	Backoff BackoffConfig
}

// NewDispatcher builds a Dispatcher with a default HTTP client and backoff.
func NewDispatcher(store DeliveryStore, secret []byte, logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		Store:   store,
		Client:  &http.Client{Timeout: 10 * time.Second},
		Logger:  logger,
		Secret:  secret,
		Backoff: DefaultBackoff,
	}
}

// Enqueue schedules a webhook delivery, applying the configured chaos delay.
// Returns the delivery ID.
func (d *Dispatcher) Enqueue(ctx context.Context, url, eventType string, payload any) (string, error) {
	return d.EnqueueAfter(ctx, url, eventType, payload, 0)
}

// EnqueueAfter schedules a webhook delivery no earlier than minDelay from
// now, plus the configured chaos delay on top — for simulators that need a
// deterministic minimum gap between related events (e.g. a "settled"
// webhook followed by a later "returned" webhook for the same payment),
// not just chaos jitter.
func (d *Dispatcher) EnqueueAfter(ctx context.Context, url, eventType string, payload any, minDelay time.Duration) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("vendorsim: marshal payload: %w", err)
	}

	delivery := NewDelivery(url, eventType, body)
	delivery.NextAttemptAt = delivery.NextAttemptAt.Add(minDelay).Add(d.Chaos.Delay())

	if err := d.Store.Enqueue(ctx, delivery); err != nil {
		return "", err
	}

	return delivery.ID, nil
}

// Run polls the store for due deliveries and attempts them every tick, until
// ctx is canceled.
func (d *Dispatcher) Run(ctx context.Context, tick time.Duration) {
	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.tick(ctx)
		}
	}
}

func (d *Dispatcher) tick(ctx context.Context) {
	due, err := d.Store.ClaimDue(ctx, time.Now().UTC(), 50)
	if err != nil {
		if d.Logger != nil {
			d.Logger.Error("vendorsim: claim due deliveries", "error", err)
		}

		return
	}

	for _, delivery := range due {
		d.attempt(ctx, delivery)

		if d.Chaos.ShouldDuplicate() {
			d.attempt(ctx, delivery)
		}
	}
}

func (d *Dispatcher) attempt(ctx context.Context, delivery Delivery) {
	err := d.deliver(ctx, delivery)
	now := time.Now().UTC()

	if err == nil {
		if merr := d.Store.MarkDelivered(ctx, delivery.ID, now); merr != nil && d.Logger != nil {
			d.Logger.Error("vendorsim: mark delivered", "delivery_id", delivery.ID, "error", merr)
		}

		return
	}

	if d.Logger != nil {
		d.Logger.Warn("vendorsim: delivery attempt failed", "delivery_id", delivery.ID, "attempt", delivery.Attempts, "error", err)
	}

	backoff := d.Backoff
	if backoff.MaxRetries == 0 && backoff.Base == 0 {
		backoff = DefaultBackoff
	}

	if delivery.Attempts+1 >= backoff.MaxRetries {
		if merr := d.Store.MarkFailed(ctx, delivery.ID, now.Add(backoff.Max), "giving up: "+err.Error()); merr != nil && d.Logger != nil {
			d.Logger.Error("vendorsim: mark failed", "delivery_id", delivery.ID, "error", merr)
		}

		return
	}

	next := now.Add(backoff.delay(delivery.Attempts))
	if merr := d.Store.MarkFailed(ctx, delivery.ID, next, err.Error()); merr != nil && d.Logger != nil {
		d.Logger.Error("vendorsim: mark failed", "delivery_id", delivery.ID, "error", merr)
	}
}

func (d *Dispatcher) deliver(ctx context.Context, delivery Delivery) error {
	ts := time.Now().Unix()
	sig := Sign(d.Secret, ts, delivery.Payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, delivery.URL, bytes.NewReader(delivery.Payload))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderEventType, delivery.EventType)
	req.Header.Set(HeaderTimestamp, strconv.FormatInt(ts, 10))
	req.Header.Set(HeaderSignature, sig)
	req.Header.Set(HeaderDeliveryID, delivery.ID)

	resp, err := d.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("consumer returned status %d", resp.StatusCode)
	}

	return nil
}
