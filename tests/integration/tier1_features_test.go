package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func (h *Harness) postJSON(t *testing.T, url, userID, idempotencyKey string, body any) ([]byte, int, string) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req.Header.Set("X-User-Id", userID)
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http post: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return raw, resp.StatusCode, string(raw)
}

func TestWalletDeposit(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "deposit-user@example.com", "+15552000001")
	h.submitKYC(t, user.UserID)

	deposit := h.depositWallet(t, user.UserID, "250.00", newID("deposit"))
	if deposit.Amount != "250.00" || deposit.Status != "completed" {
		t.Fatalf("deposit = %+v", deposit)
	}

	idemKey := newID("deposit-replay")
	first := h.depositWalletWithKey(t, user.UserID, "100.00", idemKey)
	second := h.depositWalletWithKey(t, user.UserID, "100.00", idemKey)
	if second.ID != first.ID {
		t.Fatalf("idempotent deposit id = %q, want %q", second.ID, first.ID)
	}

	wallet := h.getWallet(t, user.UserID)
	if wallet.AvailableBalance != "350" && wallet.AvailableBalance != "350.00" {
		t.Fatalf("balance after deposits = %q", wallet.AvailableBalance)
	}

	txs, status := h.listWalletTransactions(t, user.UserID)
	if status != http.StatusOK || len(txs.Transactions) == 0 {
		t.Fatalf("wallet transactions status=%d len=%d", status, len(txs.Transactions))
	}
	if txs.Transactions[0].Type != "deposit" {
		t.Fatalf("first tx type = %q", txs.Transactions[0].Type)
	}
}

func TestP2PTransferByEmail(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	sender := h.registerUser(t, "p2p-email-sender@example.com", "+15552000002")
	recipient := h.registerUser(t, "p2p-email-recipient@example.com", "+15552000003")
	h.submitKYC(t, sender.UserID)
	h.submitKYC(t, recipient.UserID)

	h.depositWallet(t, sender.UserID, "500.00", newID("fund-sender"))

	transfer := h.createTransferByEmail(t, sender.UserID, "p2p-email-recipient@example.com", "75.00", newID("p2p-email"))
	if transfer.Status != "completed" {
		t.Fatalf("transfer status = %q", transfer.Status)
	}
	if transfer.RecipientUserID != recipient.UserID {
		t.Fatalf("recipient = %q, want %q", transfer.RecipientUserID, recipient.UserID)
	}
}

func TestMarkNotificationsRead(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	user := h.registerUser(t, "notif-read@example.com", "+15552000004")
	h.submitKYC(t, user.UserID)
	h.depositWallet(t, user.UserID, "10.00", newID("notif-deposit"))

	waitUntil(t, waitTimeout, func() bool {
		list := h.listNotifications(t, user.UserID)
		return len(list.Notifications) > 0 && list.UnreadCount > 0
	})

	list := h.listNotifications(t, user.UserID)
	n := list.Notifications[0]
	marked := h.markNotificationRead(t, user.UserID, n.ID)
	if !marked.Read {
		t.Fatal("expected notification marked read")
	}

	listAfter := h.listNotifications(t, user.UserID)
	if listAfter.UnreadCount >= list.UnreadCount {
		t.Fatalf("unread_count = %d, want < %d", listAfter.UnreadCount, list.UnreadCount)
	}

	count := h.markAllNotificationsRead(t, user.UserID)
	if count < 1 {
		t.Fatalf("marked_count = %d", count)
	}
	final := h.listNotifications(t, user.UserID)
	if final.UnreadCount != 0 {
		t.Fatalf("unread_count after read-all = %d", final.UnreadCount)
	}
}

func TestChangePassword(t *testing.T) {
	h := NewHarness(t)
	h.Start()
	defer h.Cleanup()

	email := "change-pass@example.com"
	user := h.registerUser(t, email, "+15552000005")

	h.changePassword(t, user.UserID, defaultPassword, "newsecret99")

	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/auth/login", "", newID("login-old"), map[string]string{
		"email":    email,
		"password": defaultPassword,
	}, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("login with old password: status %d, want 401", status)
	}

	var login registerResponse
	status = (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/auth/login", "", newID("login-new"), map[string]string{
		"email":    email,
		"password": "newsecret99",
	}, &login)
	if status != http.StatusOK {
		t.Fatalf("login with new password: status %d", status)
	}
}

type depositResponse struct {
	ID       string `json:"id"`
	WalletID string `json:"wallet_id"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

type markAllReadResponse struct {
	MarkedCount int64 `json:"marked_count"`
}

func (h *Harness) depositWallet(t *testing.T, userID, amount, idempotencyKey string) depositResponse {
	t.Helper()
	return h.depositWalletWithKey(t, userID, amount, idempotencyKey)
}

func (h *Harness) depositWalletWithKey(t *testing.T, userID, amount, idempotencyKey string) depositResponse {
	t.Helper()
	out, status, raw := h.postJSON(t, h.UserURL+"/api/v1/wallets/deposit", userID, idempotencyKey, map[string]string{
		"amount":   amount,
		"currency": "USD",
	})
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("deposit wallet: status %d body=%s", status, raw)
	}
	var deposit depositResponse
	if err := json.Unmarshal(out, &deposit); err != nil {
		t.Fatalf("decode deposit: %v body=%s", err, raw)
	}
	return deposit
}

func (h *Harness) createTransferByEmail(t *testing.T, senderID, recipientEmail, amount, idempotencyKey string) transferResponse {
	t.Helper()
	var out transferResponse
	status := (&httpClient{base: h.PaymentURL}).do(t, http.MethodPost, "/api/v1/transfers", senderID, idempotencyKey, map[string]string{
		"recipient_email": recipientEmail,
		"amount":          amount,
		"currency":        "USD",
	}, &out)
	if status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("transfer by email: status %d", status)
	}
	return out
}

func (h *Harness) markNotificationRead(t *testing.T, userID, notificationID string) notificationResponse {
	t.Helper()
	var out notificationResponse
	path := "/api/v1/notifications/" + notificationID + "/read"
	status := (&httpClient{base: h.NotificationURL}).do(t, http.MethodPost, path, userID, "", nil, &out)
	if status != http.StatusOK {
		t.Fatalf("mark notification read: status %d", status)
	}
	return out
}

func (h *Harness) markAllNotificationsRead(t *testing.T, userID string) int64 {
	t.Helper()
	var out markAllReadResponse
	status := (&httpClient{base: h.NotificationURL}).do(t, http.MethodPost, "/api/v1/notifications/read-all", userID, "", nil, &out)
	if status != http.StatusOK {
		t.Fatalf("mark all notifications read: status %d", status)
	}
	return out.MarkedCount
}

func (h *Harness) changePassword(t *testing.T, userID, currentPassword, newPassword string) {
	t.Helper()
	_, status, raw := h.postJSON(t, h.UserURL+"/api/v1/auth/change-password", userID, newID("change-pass"), map[string]string{
		"current_password": currentPassword,
		"new_password":     newPassword,
	})
	if status != http.StatusNoContent {
		t.Fatalf("change password: status %d, want 204, body=%s", status, raw)
	}
}