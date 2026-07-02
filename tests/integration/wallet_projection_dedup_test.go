package integration

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
)

func TestWalletProjectionIngestDedup(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "wallet-dedup-sender@example.com", "+15559000001")
	recipient := h.registerUser(t, "wallet-dedup-recipient@example.com", "+15559000002")

	transferID := uuid.NewString()
	eventID := uuid.NewString()
	body := map[string]any{
		"event_id":       eventID,
		"event_type":     events.TypeTransferCompleted,
		"aggregate_type": "transfer",
		"aggregate_id":   transferID,
		"payload": map[string]string{
			"transfer_id":        transferID,
			"ledger_transfer_id": uuid.NewString(),
			"sender_user_id":     sender.UserID,
			"recipient_user_id":  recipient.UserID,
			"amount":             "25.00",
			"currency":           "USD",
		},
	}

	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/internal/events", "", newID("wallet-ingest-1"), body, nil)
	if status != http.StatusAccepted {
		t.Fatalf("first ingest status = %d, want 202", status)
	}

	// Different idempotency key so middleware replays do not mask consumer-inbox dedup.
	status = (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/internal/events", "", newID("wallet-ingest-2"), body, nil)
	if status != http.StatusAccepted {
		t.Fatalf("second ingest status = %d, want 202", status)
	}

	var senderRows int
	err := h.Pool().QueryRow(context.Background(), `
SELECT COUNT(*) FROM "user".wallet_transactions
WHERE user_id = $1 AND source_event_id = $2`, sender.UserID, eventID).Scan(&senderRows)
	if err != nil {
		t.Fatalf("query sender wallet tx: %v", err)
	}
	if senderRows != 1 {
		t.Fatalf("sender wallet rows = %d, want 1", senderRows)
	}

	var inboxCount int
	err = h.Pool().QueryRow(context.Background(), `
SELECT COUNT(*) FROM "user".consumer_inbox WHERE event_id = $1`, eventID).Scan(&inboxCount)
	if err != nil {
		t.Fatalf("query consumer inbox: %v", err)
	}
	if inboxCount != 1 {
		t.Fatalf("inbox rows = %d, want 1", inboxCount)
	}
}