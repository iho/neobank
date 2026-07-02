package integration

import (
	"net/http"
	"testing"
)

func TestWalletTransactionsFromProjection(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "wallet-tx-sender@example.com", "+15554000001")
	recipient := h.registerUser(t, "wallet-tx-recipient@example.com", "+15554000002")
	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)

	senderWallet := h.getInternalWallet(t, sender.UserID)
	if err := h.Ledger.CreditAccount(senderWallet.LedgerAccountID, "500.00"); err != nil {
		t.Fatalf("credit sender: %v", err)
	}

	transfer, status := h.createTransfer(t, sender.UserID, "+15554000002", "42.00", newID("wallet-tx"))
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("create transfer status = %d", status)
	}
	if transfer.Status != "completed" {
		t.Fatalf("transfer status = %q", transfer.Status)
	}

	waitUntil(t, waitTimeout, func() bool {
		txs, code := h.listWalletTransactions(t, sender.UserID)
		if code != http.StatusOK {
			return false
		}
		for _, tx := range txs.Transactions {
			if tx.ID == transfer.ID && tx.Type == "p2p_out" && tx.Amount == "42.00" {
				return true
			}
		}
		return false
	})

	recipientTxs, code := h.listWalletTransactions(t, recipient.UserID)
	if code != http.StatusOK {
		t.Fatalf("recipient list status = %d", code)
	}
	foundIn := false
	for _, tx := range recipientTxs.Transactions {
		if tx.ID == transfer.ID && tx.Type == "p2p_in" {
			foundIn = true
			break
		}
	}
	if !foundIn {
		t.Fatal("expected p2p_in transaction for recipient")
	}
}