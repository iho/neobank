package integration

import (
	"context"
	"net/http"
	"testing"
)

func TestAMLCTRThresholdDoesNotBlockTransfer(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "aml-ctr-sender@example.com", "+15557000001")
	recipient := h.registerUser(t, "aml-ctr-recipient@example.com", "+15557000002")
	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)

	senderWallet := h.getInternalWallet(t, sender.UserID)
	if err := h.Ledger.CreditAccount(senderWallet.LedgerAccountID, "20000.00"); err != nil {
		t.Fatalf("credit sender: %v", err)
	}

	transfer, status := h.createTransfer(t, sender.UserID, "+15557000002", "10000.00", newID("aml-ctr"))
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("create transfer status = %d", status)
	}
	if transfer.Status != "completed" {
		t.Fatalf("transfer status = %q, want completed", transfer.Status)
	}

	var caseCount int
	err := h.Pool().QueryRow(context.Background(), `
SELECT COUNT(*) FROM payment.aml_cases
WHERE entity_id = $1 AND case_type = 'ctr' AND status = 'open'`, transfer.ID).Scan(&caseCount)
	if err != nil {
		t.Fatalf("query aml case: %v", err)
	}
	if caseCount != 1 {
		t.Fatalf("open ctr cases = %d, want 1", caseCount)
	}
}

func TestAMLStructuringOpensSARCase(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "aml-struct-sender@example.com", "+15557100001")
	recipient := h.registerUser(t, "aml-struct-recipient@example.com", "+15557100002")
	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)

	senderWallet := h.getInternalWallet(t, sender.UserID)
	if err := h.Ledger.CreditAccount(senderWallet.LedgerAccountID, "25000.00"); err != nil {
		t.Fatalf("credit sender: %v", err)
	}

	first, status := h.createTransfer(t, sender.UserID, "+15557100002", "9000.00", newID("aml-struct-1"))
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("first transfer status = %d", status)
	}
	if first.Status != "completed" {
		t.Fatalf("first transfer status = %q", first.Status)
	}

	second, status := h.createTransfer(t, sender.UserID, "+15557100002", "9000.00", newID("aml-struct-2"))
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("second transfer status = %d", status)
	}
	if second.Status != "completed" {
		t.Fatalf("second transfer status = %q", second.Status)
	}

	var sarCount int
	err := h.Pool().QueryRow(context.Background(), `
SELECT COUNT(*) FROM payment.aml_cases
WHERE entity_id = $1 AND case_type = 'sar' AND reason_code = 'STRUCTURING'`, second.ID).Scan(&sarCount)
	if err != nil {
		t.Fatalf("query sar case: %v", err)
	}
	if sarCount != 1 {
		t.Fatalf("sar cases for second transfer = %d, want 1", sarCount)
	}
}