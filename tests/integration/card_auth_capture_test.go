package integration

import (
	"net/http"
	"testing"
)

func TestCardAuthorizeAndCapture(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "carduser@example.com", "+15554000001")
	kyc := h.submitKYC(t, user.UserID)

	wallet := h.getInternalWallet(t, user.UserID)
	if err := h.Ledger.CreditAccount(wallet.LedgerAccountID, "300.00"); err != nil {
		t.Fatalf("credit wallet: %v", err)
	}

	card := h.issueCard(t, user.UserID, kyc.WalletID)
	if card.Status != "active" {
		t.Fatalf("card status = %q", card.Status)
	}

	auth, status := h.authorizeCard(t, user.UserID, card.ID, "45.00", newID("auth"))
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("authorize status = %d", status)
	}
	if auth.Status != "authorized" {
		t.Fatalf("authorization status = %q", auth.Status)
	}
	if auth.LedgerHoldID == nil || *auth.LedgerHoldID == "" {
		t.Fatal("expected ledger_hold_id")
	}

	captured := h.captureAuthorization(t, user.UserID, auth.ID)
	if captured.Status != "captured" {
		t.Fatalf("captured status = %q", captured.Status)
	}
	if captured.LedgerTransferID == nil || *captured.LedgerTransferID == "" {
		t.Fatal("expected ledger_transfer_id after capture")
	}

	walletView := h.getWallet(t, user.UserID)
	if walletView.Balance != "255.00" {
		t.Fatalf("balance after capture = %q, want 255.00", walletView.Balance)
	}
}

func TestCardAuthorizeIdempotency(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "card-idem@example.com", "+15554000002")
	kyc := h.submitKYC(t, user.UserID)
	wallet := h.getInternalWallet(t, user.UserID)
	if err := h.Ledger.CreditAccount(wallet.LedgerAccountID, "100.00"); err != nil {
		t.Fatalf("credit wallet: %v", err)
	}
	card := h.issueCard(t, user.UserID, kyc.WalletID)

	idemKey := newID("auth-idem")
	first, status := h.authorizeCard(t, user.UserID, card.ID, "25.00", idemKey)
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("first authorize status = %d", status)
	}

	second, status := h.authorizeCard(t, user.UserID, card.ID, "25.00", idemKey)
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("second authorize status = %d", status)
	}
	if first.ID != second.ID {
		t.Fatalf("authorization ids differ: %s vs %s", first.ID, second.ID)
	}
}