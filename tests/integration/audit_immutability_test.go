package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestUserAuditLogIsImmutable(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "audit-immut@example.com", "+15551000777")
	rowID := uuid.New()
	_, err := h.Pool().Exec(context.Background(), `
INSERT INTO "user".audit_log (id, entity_type, entity_id, action, actor, metadata, created_at)
VALUES ($1, 'user', $2, 'test', 'system', '{}', now())`, rowID, user.UserID)
	if err != nil {
		t.Fatalf("insert audit_log: %v", err)
	}

	_, err = h.Pool().Exec(context.Background(), `
UPDATE "user".audit_log SET action = 'tampered' WHERE id = $1`, rowID)
	if err == nil || !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("expected update blocked, got %v", err)
	}

	_, err = h.Pool().Exec(context.Background(), `DELETE FROM "user".audit_log WHERE id = $1`, rowID)
	if err == nil || !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("expected delete blocked, got %v", err)
	}
}

func TestPaymentFraudDecisionsAreImmutable(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "fraud-immut@example.com", "+15551000778")
	rowID := uuid.New()
	_, err := h.Pool().Exec(context.Background(), `
INSERT INTO payment.fraud_decisions (
    id, entity_type, entity_id, user_id, transaction_type, amount, currency,
    decision, reason_code, risk_score, rule_set_version, created_at
) VALUES ($1, 'transfer', 'tx-1', $2, 'p2p', 10.00, 'USD', 'allow', 'test', 0, 'mvp-1.0.0', now())`,
		rowID, user.UserID)
	if err != nil {
		t.Fatalf("insert fraud_decisions: %v", err)
	}

	_, err = h.Pool().Exec(context.Background(), `
UPDATE payment.fraud_decisions SET decision = 'deny' WHERE id = $1`, rowID)
	if err == nil || !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("expected update blocked, got %v", err)
	}
}