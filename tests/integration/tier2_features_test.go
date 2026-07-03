package integration

import (
	"fmt"
	"net/http"
	"testing"
)

func TestSavedPayeesAutoAddOnP2P(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "payee-sender@example.com", "+15553000001")
	recipient := h.registerUser(t, "payee-recipient@example.com", "+15553000002")
	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)
	h.depositWallet(t, sender.UserID, "200.00", newID("payee-fund"))

	h.createTransferByEmail(t, sender.UserID, "payee-recipient@example.com", "25.00", newID("payee-p2p"))

	waitUntil(t, waitTimeout, func() bool {
		payees := h.listPayees(t, sender.UserID)
		return len(payees.Payees) > 0 && payees.Payees[0].PayeeUserID == recipient.UserID
	})
}

func TestTransferLimitsAPI(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "limits-user@example.com", "+15553000003")
	limits := h.getLimits(t, user.UserID)
	if limits.P2P.HourlyTransferCount.Limit != "10" {
		t.Fatalf("hourly limit = %q", limits.P2P.HourlyTransferCount.Limit)
	}
	if limits.P2P.DailyTransferAmount.Limit != "10000.00" {
		t.Fatalf("daily limit = %q", limits.P2P.DailyTransferAmount.Limit)
	}
}

func TestCardSpendControls(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "card-controls@example.com", "+15553000004")
	h.submitKYC(t, user.UserID)
	h.depositWallet(t, user.UserID, "500.00", newID("card-fund"))
	wallet := h.getInternalWallet(t, user.UserID)

	card := h.issueCardWithControls(t, user.UserID, wallet.ID, "100.00", true)

	_, status := h.authorizeCardWithChannel(t, user.UserID, card.ID, "50.00", "pos", newID("pos-auth"))
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("expected pos authorization declined with 422, got %d", status)
	}

	auth, status := h.authorizeCardWithChannel(t, user.UserID, card.ID, "30.00", "online", newID("online-auth"))
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("online authorization status = %d", status)
	}
	if auth.Status != "authorized" {
		t.Fatalf("auth status = %q", auth.Status)
	}
}

func TestTransferCursorPagination(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "cursor-sender@example.com", "+15553000005")
	recipient := h.registerUser(t, "cursor-recipient@example.com", "+15553000006")
	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)
	h.depositWallet(t, sender.UserID, "300.00", newID("cursor-fund"))

	for i := 0; i < 3; i++ {
		h.createTransferByEmail(t, sender.UserID, "cursor-recipient@example.com", "10.00", newID(fmt.Sprintf("cursor-tx-%d", i)))
	}

	first := h.listTransfers(t, sender.UserID, 2, "")
	if len(first.Transfers) != 2 {
		t.Fatalf("first page len = %d", len(first.Transfers))
	}
	if first.NextCursor == "" {
		t.Fatal("expected next_cursor on first page")
	}

	second := h.listTransfers(t, sender.UserID, 2, first.NextCursor)
	if len(second.Transfers) < 1 {
		t.Fatalf("second page len = %d", len(second.Transfers))
	}
	if first.Transfers[0].ID == second.Transfers[0].ID {
		t.Fatal("cursor returned overlapping transfers")
	}
}

func TestWalletTransactionEnrichment(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "enrich-sender@example.com", "+15553000007")
	recipient := h.registerUser(t, "enrich-recipient@example.com", "+15553000008")
	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)
	h.depositWallet(t, sender.UserID, "150.00", newID("enrich-fund"))

	h.createTransferByEmailWithMemo(t, sender.UserID, "enrich-recipient@example.com", "20.00", "lunch money", newID("enrich-p2p"))

	waitUntil(t, waitTimeout, func() bool {
		txs, code := h.listWalletTransactionsFull(t, sender.UserID)
		if code != http.StatusOK {
			return false
		}
		for _, tx := range txs.Transactions {
			if tx.Type == "p2p_out" && tx.Counterparty == "enrich-recipient@example.com" && tx.Memo == "lunch money" {
				return true
			}
		}
		return false
	})
}

type payeeListResponse struct {
	Payees []struct {
		ID          string `json:"id"`
		PayeeUserID string `json:"payee_user_id"`
		PayeeEmail  string `json:"payee_email,omitempty"`
	} `json:"payees"`
}

type limitsResponse struct {
	P2P struct {
		HourlyTransferCount struct {
			Limit string `json:"limit"`
			Used  string `json:"used"`
		} `json:"hourly_transfer_count"`
		DailyTransferAmount struct {
			Limit string `json:"limit"`
		} `json:"daily_transfer_amount"`
		SingleTransferMax string `json:"single_transfer_max"`
	} `json:"p2p"`
}

type transferListResponse struct {
	Transfers  []transferResponse `json:"transfers"`
	NextCursor string             `json:"next_cursor,omitempty"`
}

type walletTransactionFullListResponse struct {
	Transactions []struct {
		ID           string `json:"id"`
		Type         string `json:"type"`
		Amount       string `json:"amount"`
		Counterparty string `json:"counterparty,omitempty"`
		Memo         string `json:"memo,omitempty"`
	} `json:"transactions"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func (h *Harness) listPayees(t *testing.T, userID string) payeeListResponse {
	t.Helper()
	var out payeeListResponse
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodGet, "/api/v1/payees?limit=20", userID, "", nil, &out)
	if status != http.StatusOK {
		t.Fatalf("list payees: status %d", status)
	}
	return out
}

func (h *Harness) getLimits(t *testing.T, userID string) limitsResponse {
	t.Helper()
	var out limitsResponse
	status := (&httpClient{base: h.PaymentURL}).do(t, http.MethodGet, "/api/v1/limits", userID, "", nil, &out)
	if status != http.StatusOK {
		t.Fatalf("get limits: status %d", status)
	}
	return out
}

func (h *Harness) issueCardWithControls(t *testing.T, userID, walletID, dailyLimit string, onlineOnly bool) cardResponse {
	t.Helper()
	var out cardResponse
	status := (&httpClient{base: h.CardURL}).do(t, http.MethodPost, "/api/v1/cards", userID, newID("card-controls"), map[string]any{
		"wallet_id":       walletID,
		"cardholder_name": "Test User",
		"daily_limit":     dailyLimit,
		"online_only":     onlineOnly,
	}, &out)
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("issue card with controls: status %d", status)
	}
	return out
}

func (h *Harness) authorizeCardWithChannel(t *testing.T, userID, cardID, amount, channel, idempotencyKey string) (authorizationResponse, int) {
	t.Helper()
	var out authorizationResponse
	path := fmt.Sprintf("/api/v1/cards/%s/authorize", cardID)
	status := (&httpClient{base: h.CardURL}).do(t, http.MethodPost, path, userID, idempotencyKey, map[string]string{
		"amount":        amount,
		"currency":      "USD",
		"merchant_name": "Test Merchant",
		"channel":       channel,
	}, &out)
	return out, status
}

func (h *Harness) listTransfers(t *testing.T, userID string, limit int, cursor string) transferListResponse {
	t.Helper()
	path := fmt.Sprintf("/api/v1/transfers?limit=%d", limit)
	if cursor != "" {
		path += "&cursor=" + cursor
	}
	var out transferListResponse
	status := (&httpClient{base: h.PaymentURL}).do(t, http.MethodGet, path, userID, "", nil, &out)
	if status != http.StatusOK {
		t.Fatalf("list transfers: status %d", status)
	}
	return out
}

func (h *Harness) createTransferByEmailWithMemo(t *testing.T, senderID, recipientEmail, amount, memo, idempotencyKey string) transferResponse {
	t.Helper()
	var out transferResponse
	status := (&httpClient{base: h.PaymentURL}).do(t, http.MethodPost, "/api/v1/transfers", senderID, idempotencyKey, map[string]string{
		"recipient_email": recipientEmail,
		"amount":          amount,
		"currency":        "USD",
		"memo":            memo,
	}, &out)
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("transfer with memo: status %d", status)
	}
	return out
}

func (h *Harness) listWalletTransactionsFull(t *testing.T, userID string) (walletTransactionFullListResponse, int) {
	t.Helper()
	var out walletTransactionFullListResponse
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodGet, "/api/v1/wallet/transactions?limit=20", userID, "", nil, &out)
	return out, status
}