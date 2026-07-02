package integration

import (
	"net/http"
	"testing"
)

func TestKYCScreeningRejectsSanctionedName(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "sanctioned-kyc@example.com", "+15555000001")
	var out submitKYCResponse
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/kyc", user.UserID, newID("kyc-sanctioned"), map[string]string{
		"full_name":     "SANCTIONED User",
		"date_of_birth": "1990-01-15",
		"country_code":  "US",
	}, &out)
	if status != http.StatusOK {
		t.Fatalf("expected 200 for sanctioned kyc rejection, got %d", status)
	}
	if out.Status != "rejected" {
		t.Fatalf("kyc status = %q, want rejected", out.Status)
	}
}

func TestP2PBlockedByCounterpartyScreening(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "screen-sender@example.com", "+15556000001")
	recipient := h.registerUser(t, "screen-recipient@example.com", "+1555SANCTIONED99")
	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)

	senderWallet := h.getInternalWallet(t, sender.UserID)
	if err := h.Ledger.CreditAccount(senderWallet.LedgerAccountID, "100.00"); err != nil {
		t.Fatalf("credit sender: %v", err)
	}

	_, status := h.createTransfer(t, sender.UserID, "+1555SANCTIONED99", "10.00", newID("screen-block"))
	if status == http.StatusOK || status == http.StatusCreated {
		t.Fatalf("expected transfer blocked by screening, got %d", status)
	}
}