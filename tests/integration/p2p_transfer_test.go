package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/iho/neobank/pkg/events"
)

func TestP2PTransferHappyPath(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "sender@example.com", "+15552000001")
	recipient := h.registerUser(t, "recipient@example.com", "+15552000002")

	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)

	senderWallet := h.getInternalWallet(t, sender.UserID)
	if err := h.Ledger.CreditAccount(senderWallet.LedgerAccountID, "1000.00"); err != nil {
		t.Fatalf("credit sender: %v", err)
	}

	transfer, status := h.createTransfer(t, sender.UserID, "+15552000002", "125.50", newID("p2p"))
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("create transfer status = %d", status)
	}
	if transfer.Status != "completed" {
		t.Fatalf("transfer status = %q", transfer.Status)
	}
	if transfer.LedgerTransferID == nil || *transfer.LedgerTransferID == "" {
		t.Fatal("expected ledger_transfer_id")
	}
	if h.Ledger.CreateTransferCallCount() != 1 {
		t.Fatalf("ledger create transfer calls = %d, want 1", h.Ledger.CreateTransferCallCount())
	}

	waitUntil(t, waitTimeout, func() bool {
		notifs := h.listNotifications(t, recipient.UserID)
		for _, n := range notifs.Notifications {
			if n.EventType == events.TypeTransferCompleted {
				return true
			}
		}
		return false
	})
}

func TestP2PTransferIdempotencyNoDoubleLedgerCall(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "idem-sender@example.com", "+15553000001")
	recipient := h.registerUser(t, "idem-recipient@example.com", "+15553000002")
	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)

	senderWallet := h.getInternalWallet(t, sender.UserID)
	if err := h.Ledger.CreditAccount(senderWallet.LedgerAccountID, "200.00"); err != nil {
		t.Fatalf("credit sender: %v", err)
	}

	idemKey := newID("p2p-idem")
	first, status := h.createTransfer(t, sender.UserID, "+15553000002", "50.00", idemKey)
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("first transfer status = %d", status)
	}

	second, status := h.createTransfer(t, sender.UserID, "+15553000002", "50.00", idemKey)
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("second transfer status = %d", status)
	}
	if first.ID != second.ID {
		t.Fatalf("transfer ids differ: %s vs %s", first.ID, second.ID)
	}
	if h.Ledger.CreateTransferCallCount() != 1 {
		t.Fatalf("ledger create transfer calls = %d, want 1", h.Ledger.CreateTransferCallCount())
	}
}

const waitTimeout = 10 * time.Second