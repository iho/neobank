package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/audit"
)

func TestPIIAccessAuditOnProfileRead(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "pii-audit@example.com", "+15551000999")
	correlationID := uuid.NewString()
	actorID := user.UserID

	status := h.getProfileWithHeaders(t, user.UserID, actorID, correlationID)
	if status != http.StatusOK {
		t.Fatalf("get profile: status %d", status)
	}

	rows, err := h.pool.Query(h.ctx, `
		SELECT resource, actor, correlation_id
		FROM "user".pii_access_log
		WHERE subject_user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, user.UserID)
	if err != nil {
		t.Fatalf("query pii_access_log: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("expected pii_access_log row")
	}
	var resource, actor, corrID string
	if err := rows.Scan(&resource, &actor, &corrID); err != nil {
		t.Fatalf("scan: %v", err)
	}
	if resource != audit.PIIResourceProfile {
		t.Fatalf("resource = %q, want %q", resource, audit.PIIResourceProfile)
	}
	if actor != actorID {
		t.Fatalf("actor = %q, want %q", actor, actorID)
	}
	if corrID != correlationID {
		t.Fatalf("correlation_id = %q, want %q", corrID, correlationID)
	}
}

func TestPIIAccessAuditNotRecordedOnProfileMiss(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	missingID := uuid.NewString()
	status := h.getProfileWithHeaders(t, missingID, missingID, uuid.NewString())
	if status != http.StatusNotFound {
		t.Fatalf("get profile: status %d, want 404", status)
	}

	var count int
	err := h.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM "user".pii_access_log WHERE subject_user_id = $1`, missingID).Scan(&count)
	if err != nil {
		t.Fatalf("count pii_access_log: %v", err)
	}
	if count != 0 {
		t.Fatalf("pii_access_log count = %d, want 0", count)
	}
}

func (h *Harness) getProfileWithHeaders(t *testing.T, userID, actorID, correlationID string) int {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, h.UserURL+"/api/v1/me", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("X-User-Id", userID)
	req.Header.Set("X-Actor-Id", actorID)
	req.Header.Set("X-Correlation-Id", correlationID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http get profile: %v", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}