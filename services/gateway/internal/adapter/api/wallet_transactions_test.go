package api

import (
	"testing"

	"github.com/iho/neobank/services/gateway/internal/client"
)

func TestBuildWalletTransactionsMergesAndSorts(t *testing.T) {
	userID := "user-1"
	transfers := client.TransferList{
		Transfers: []client.TransferView{
			{
				ID: "t1", Status: "completed", SenderUserID: userID, RecipientUserID: "user-2",
				Amount: "10.00", Currency: "USD", CreatedAt: "2026-01-02T10:00:00Z",
			},
			{
				ID: "t2", Status: "completed", SenderUserID: "user-3", RecipientUserID: userID,
				Amount: "5.00", Currency: "USD", CreatedAt: "2026-01-03T10:00:00Z",
			},
		},
	}
	auths := client.AuthorizationList{
		Authorizations: []client.AuthorizationView{
			{
				ID: "a1", Status: "captured", UserID: userID, Amount: "2.00", Currency: "USD",
				MerchantName: "Coffee Shop", CreatedAt: "2026-01-04T10:00:00Z",
			},
		},
	}

	txs := buildWalletTransactions(userID, transfers, auths, 10)
	if len(txs) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(txs))
	}
	if txs[0].Id != "a1" || txs[0].Type != "card_purchase" {
		t.Fatalf("expected newest card purchase first, got %+v", txs[0])
	}
	if txs[1].Type != "p2p_in" {
		t.Fatalf("expected p2p_in second, got %s", txs[1].Type)
	}
	if txs[2].Type != "p2p_out" {
		t.Fatalf("expected p2p_out third, got %s", txs[2].Type)
	}
}