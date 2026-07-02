package integration

import (
	"net/http"
	"testing"
)

func TestKYCWalletProvisioning(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "alice-kyc@example.com", "+15551000001")
	kyc := h.submitKYC(t, user.UserID)
	if kyc.WalletID == "" {
		t.Fatal("expected wallet_id after kyc")
	}

	wallet := h.getWallet(t, user.UserID)
	if wallet.LedgerAccountID == "" {
		t.Fatal("expected ledger_account_id on wallet")
	}
	if wallet.Currency != "USD" {
		t.Fatalf("currency = %q", wallet.Currency)
	}

	if err := h.Ledger.CreditAccount(wallet.LedgerAccountID, "500.00"); err != nil {
		t.Fatalf("credit account: %v", err)
	}

	wallet = h.getWallet(t, user.UserID)
	if wallet.Balance != "500.00" {
		t.Fatalf("balance = %q, want 500.00", wallet.Balance)
	}
	if wallet.AvailableBalance != "500.00" {
		t.Fatalf("available = %q", wallet.AvailableBalance)
	}
}

func TestRegisterRequiresIdempotencyKey(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/auth/register", "", "", map[string]string{
		"email":    "no-key@example.com",
		"password": defaultPassword,
	}, nil)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
}