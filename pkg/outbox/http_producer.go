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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/iho/neobank/pkg/otel"
)

// HTTPProducer forwards outbox events to a notification service ingest endpoint.
type HTTPProducer struct {
	httpClient *http.Client
	url        string
}

func NewHTTPProducer(url string) *HTTPProducer {
	return &HTTPProducer{
		url:        url,
		httpClient: &http.Client{Transport: otel.OutboundTransport(nil)},
	}
}

func (p *HTTPProducer) Produce(ctx context.Context, topic, key string, value []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(value))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-Topic", topic)
	req.Header.Set("X-Event-Key", key)
	req.Header.Set("Idempotency-Key", eventIDFromEnvelope(value, key))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notification ingest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http ingest %s status %d: %s", p.url, resp.StatusCode, string(body))
	}

	return nil
}

func eventIDFromEnvelope(value []byte, fallback string) string {
	var envelope struct {
		EventID string `json:"event_id"`
	}
	err := json.Unmarshal(value, &envelope)
	if err == nil && envelope.EventID != "" {
		return envelope.EventID
	}

	return fallback
}
