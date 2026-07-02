package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPaymentOutboxEventsAreImmutable(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	eventID := uuid.New()
	_, err := h.Pool().Exec(context.Background(), `
INSERT INTO payment.outbox_events (
    id, aggregate_type, aggregate_id, event_type, event_version, payload, created_at
) VALUES ($1, 'transfer', 'tx-1', 'payment.transfer.completed', 1, '{}', now())`, eventID)
	if err != nil {
		t.Fatalf("insert outbox event: %v", err)
	}

	_, err = h.Pool().Exec(context.Background(), `
UPDATE payment.outbox_events SET payload = '{"tampered":true}' WHERE id = $1`, eventID)
	if err == nil || !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("expected update blocked, got %v", err)
	}

	_, err = h.Pool().Exec(context.Background(), `DELETE FROM payment.outbox_events WHERE id = $1`, eventID)
	if err == nil || !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("expected delete blocked, got %v", err)
	}

	_, err = h.Pool().Exec(context.Background(), `
INSERT INTO payment.outbox_publications (event_id, published_at) VALUES ($1, now())`, eventID)
	if err != nil {
		t.Fatalf("record publication: %v", err)
	}
}