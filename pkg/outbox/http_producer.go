package outbox

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

// HTTPProducer forwards outbox events to a notification service ingest endpoint.
type HTTPProducer struct {
	url        string
	httpClient *http.Client
}

func NewHTTPProducer(url string) *HTTPProducer {
	return &HTTPProducer{
		url:        url,
		httpClient: &http.Client{},
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

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notification ingest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notification ingest status %d", resp.StatusCode)
	}
	return nil
}