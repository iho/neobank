package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
)

func TestNotificationIngestDedup(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "notif-dedup@example.com", "+15558000001")
	eventID := uuid.NewString()

	body := map[string]any{
		"event_id":       eventID,
		"event_type":     events.TypeKYCApproved,
		"aggregate_type": "user",
		"aggregate_id":   user.UserID,
		"payload": map[string]string{
			"user_id": user.UserID,
		},
	}

	status := (&httpClient{base: h.NotificationURL}).do(t, http.MethodPost, "/api/v1/internal/events", "", "", body, nil)
	if status != http.StatusAccepted {
		t.Fatalf("first ingest status = %d, want 202", status)
	}

	status = (&httpClient{base: h.NotificationURL}).do(t, http.MethodPost, "/api/v1/internal/events", "", "", body, nil)
	if status != http.StatusAccepted {
		t.Fatalf("second ingest status = %d, want 202", status)
	}

	notifs := h.listNotifications(t, user.UserID)
	kycCount := 0
	for _, n := range notifs.Notifications {
		if n.EventType == events.TypeKYCApproved {
			kycCount++
		}
	}
	if kycCount != 1 {
		t.Fatalf("kyc notifications = %d, want 1", kycCount)
	}

	var inboxCount int
	err := h.Pool().QueryRow(context.Background(), `
SELECT COUNT(*) FROM notification.consumer_inbox WHERE event_id = $1`, eventID).Scan(&inboxCount)
	if err != nil {
		t.Fatalf("query consumer inbox: %v", err)
	}
	if inboxCount != 1 {
		t.Fatalf("inbox rows = %d, want 1", inboxCount)
	}
}