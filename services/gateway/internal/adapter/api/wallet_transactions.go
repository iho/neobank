package api

import (
	"sort"
	"time"

	"github.com/iho/neobank/services/gateway/internal/client"
	"github.com/iho/neobank/services/gateway/internal/gen/api"
)

func buildWalletTransactions(userID string, transfers client.TransferList, auths client.AuthorizationList, limit int) []api.WalletTransaction {
	transactions := make([]api.WalletTransaction, 0, len(transfers.Transfers)+len(auths.Authorizations))

	for _, t := range transfers.Transfers {
		if t.Status != "completed" && t.Status != "failed" {
			continue
		}
		txType := "p2p_out"
		direction := "debit"
		counterparty := t.RecipientUserID
		if t.RecipientUserID == userID {
			txType = "p2p_in"
			direction = "credit"
			counterparty = t.SenderUserID
		}
		createdAt := parseTimeOrNow(t.CreatedAt)
		tx := api.WalletTransaction{
			Id:        t.ID,
			Type:      txType,
			Amount:    t.Amount,
			Currency:  t.Currency,
			Direction: direction,
			Status:    t.Status,
			CreatedAt: createdAt,
		}
		if counterparty != "" {
			tx.Counterparty = &counterparty
		}
		if t.ID != "" {
			tx.ReferenceId = &t.ID
		}
		if t.Memo != "" {
			tx.Memo = &t.Memo
		}
		transactions = append(transactions, tx)
	}

	for _, a := range auths.Authorizations {
		if a.Status == "declined" || a.Status == "voided" {
			continue
		}
		txType := "card_hold"
		if a.Status == "captured" {
			txType = "card_purchase"
		}
		tx := api.WalletTransaction{
			Id:        a.ID,
			Type:      txType,
			Amount:    a.Amount,
			Currency:  a.Currency,
			Direction: "debit",
			Status:    a.Status,
			CreatedAt: parseTimeOrNow(a.CreatedAt),
		}
		if a.ID != "" {
			tx.ReferenceId = &a.ID
		}
		if a.MerchantName != "" {
			tx.Counterparty = &a.MerchantName
		}
		transactions = append(transactions, tx)
	}

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt.After(transactions[j].CreatedAt)
	})

	if limit > 0 && len(transactions) > limit {
		transactions = transactions[:limit]
	}
	return transactions
}

func parseTimeOrNow(value string) time.Time {
	if value == "" {
		return time.Now().UTC()
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	return time.Now().UTC()
}

func parseTimePtr(value string) *time.Time {
	if value == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return &t
	}
	return nil
}