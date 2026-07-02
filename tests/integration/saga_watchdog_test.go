package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/sagawatchdog"
)

func TestSagaWatchdogDetectsStuckInstance(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	ctx := context.Background()
	sagaID := uuid.New()
	idemKey := "stuck-" + uuid.New().String()
	staleAt := time.Now().UTC().Add(-30 * time.Minute)

	_, err := h.Pool().Exec(ctx, `
INSERT INTO payment.saga_instances (id, saga_type, idempotency_key, status, completed_steps, context, updated_at)
VALUES ($1, 'p2p_transfer', $2, 'running', '{}', '{}', $3)`,
		sagaID, idemKey, staleAt,
	)
	if err != nil {
		t.Fatalf("insert stuck saga: %v", err)
	}

	scanner := sagawatchdog.New(h.Pool())
	result, err := scanner.Scan(ctx, "payment", time.Second)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if result.StuckFound != 1 {
		t.Fatalf("stuck_found = %d, want 1", result.StuckFound)
	}
	if result.AlertsOpen != 1 {
		t.Fatalf("alerts_open = %d, want 1", result.AlertsOpen)
	}

	alerts, err := scanner.ListOpenAlerts(ctx, "payment", 10)
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("open alerts = %d, want 1", len(alerts))
	}
	if alerts[0].SagaInstanceID != sagaID {
		t.Fatalf("alert saga_instance_id = %s, want %s", alerts[0].SagaInstanceID, sagaID)
	}
	if alerts[0].SagaType != "p2p_transfer" {
		t.Fatalf("alert saga_type = %q, want p2p_transfer", alerts[0].SagaType)
	}
}

func TestSagaWatchdogAutoResolvesTerminalSaga(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	ctx := context.Background()
	sagaID := uuid.New()
	idemKey := "resolved-" + uuid.New().String()
	staleAt := time.Now().UTC().Add(-30 * time.Minute)

	_, err := h.Pool().Exec(ctx, `
INSERT INTO payment.saga_instances (id, saga_type, idempotency_key, status, completed_steps, context, updated_at)
VALUES ($1, 'p2p_transfer', $2, 'running', '{}', '{}', $3)`,
		sagaID, idemKey, staleAt,
	)
	if err != nil {
		t.Fatalf("insert stuck saga: %v", err)
	}

	scanner := sagawatchdog.New(h.Pool())
	if _, err := scanner.Scan(ctx, "payment", time.Second); err != nil {
		t.Fatalf("first scan: %v", err)
	}

	_, err = h.Pool().Exec(ctx, `
UPDATE payment.saga_instances SET status = 'completed', updated_at = now() WHERE id = $1`, sagaID)
	if err != nil {
		t.Fatalf("complete saga: %v", err)
	}

	result, err := scanner.Scan(ctx, "payment", time.Second)
	if err != nil {
		t.Fatalf("second scan: %v", err)
	}
	if result.AutoResolved != 1 {
		t.Fatalf("auto_resolved = %d, want 1", result.AutoResolved)
	}
	if result.AlertsOpen != 0 {
		t.Fatalf("alerts_open = %d, want 0", result.AlertsOpen)
	}
}