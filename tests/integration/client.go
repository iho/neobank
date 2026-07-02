package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

type httpClient struct {
	base string
}

func (c *httpClient) do(t *testing.T, method, path string, userID, idempotencyKey string, body any, out any) int {
	t.Helper()
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if userID != "" {
		req.Header.Set("X-User-Id", userID)
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			t.Fatalf("decode response (%d): %v body=%s", resp.StatusCode, err, string(respBody))
		}
	}
	return resp.StatusCode
}

type registerResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type submitKYCResponse struct {
	KYCCaseID string `json:"kyc_case_id"`
	Status    string `json:"status"`
	WalletID  string `json:"wallet_id"`
}

type walletResponse struct {
	WalletID          string `json:"wallet_id"`
	LedgerAccountID   string `json:"ledger_account_id"`
	Currency          string `json:"currency"`
	Balance           string `json:"balance"`
	AvailableBalance  string `json:"available_balance"`
	EncumberedBalance string `json:"encumbered_balance"`
}

type internalWalletResponse struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	Currency        string `json:"currency"`
	LedgerAccountID string `json:"ledger_account_id"`
	Status          string `json:"status"`
}

type transferResponse struct {
	ID              string  `json:"id"`
	Status          string  `json:"status"`
	SenderUserID    string  `json:"sender_user_id"`
	RecipientUserID string  `json:"recipient_user_id"`
	Amount          string  `json:"amount"`
	Currency        string  `json:"currency"`
	LedgerTransferID *string `json:"ledger_transfer_id,omitempty"`
}

type cardResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	WalletID string `json:"wallet_id"`
	Status   string `json:"status"`
	LastFour string `json:"last_four"`
}

type authorizationResponse struct {
	ID               string  `json:"id"`
	Status           string  `json:"status"`
	Amount           string  `json:"amount"`
	Currency         string  `json:"currency"`
	LedgerHoldID     *string `json:"ledger_hold_id,omitempty"`
	LedgerTransferID *string `json:"ledger_transfer_id,omitempty"`
}

type walletTransactionListResponse struct {
	Transactions []struct {
		ID        string `json:"id"`
		Type      string `json:"type"`
		Amount    string `json:"amount"`
		Currency  string `json:"currency"`
		Direction string `json:"direction"`
		Status    string `json:"status"`
	} `json:"transactions"`
}

type notificationResponse struct {
	ID        string `json:"id"`
	EventType string `json:"event_type"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Read      bool   `json:"read"`
}

type notificationListResponse struct {
	Notifications []notificationResponse `json:"notifications"`
	UnreadCount   int64                  `json:"unread_count"`
}

func (h *Harness) registerUser(t *testing.T, email, phone string) registerResponse {
	t.Helper()
	var out registerResponse
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/auth/register", "", newID("reg"), map[string]string{
		"email":    email,
		"phone":    phone,
		"password": defaultPassword,
	}, &out)
	if status != http.StatusCreated {
		t.Fatalf("register %s: status %d", email, status)
	}
	return out
}

func (h *Harness) submitKYC(t *testing.T, userID string) submitKYCResponse {
	return h.submitKYCWithName(t, userID, "Test User", "US")
}

func (h *Harness) submitKYCWithName(t *testing.T, userID, fullName, country string) submitKYCResponse {
	t.Helper()
	var out submitKYCResponse
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodPost, "/api/v1/kyc", userID, newID("kyc"), map[string]string{
		"full_name":     fullName,
		"date_of_birth": "1990-01-15",
		"country_code":  country,
	}, &out)
	if status != http.StatusOK {
		t.Fatalf("submit kyc: status %d", status)
	}
	if out.Status != "approved" {
		t.Fatalf("kyc status = %q", out.Status)
	}
	return out
}

func (h *Harness) getWallet(t *testing.T, userID string) walletResponse {
	t.Helper()
	var out walletResponse
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodGet, "/api/v1/wallets/balance?currency=USD", userID, "", nil, &out)
	if status != http.StatusOK {
		t.Fatalf("get wallet: status %d", status)
	}
	return out
}

func (h *Harness) getInternalWallet(t *testing.T, userID string) internalWalletResponse {
	t.Helper()
	var out internalWalletResponse
	path := fmt.Sprintf("/api/v1/internal/wallets?user_id=%s&currency=USD", userID)
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodGet, path, "", "", nil, &out)
	if status != http.StatusOK {
		t.Fatalf("internal wallet: status %d", status)
	}
	return out
}

func (h *Harness) createTransfer(t *testing.T, senderID, recipientPhone, amount, idempotencyKey string) (transferResponse, int) {
	t.Helper()
	var out transferResponse
	status := (&httpClient{base: h.PaymentURL}).do(t, http.MethodPost, "/api/v1/transfers", senderID, idempotencyKey, map[string]string{
		"recipient_phone": recipientPhone,
		"amount":          amount,
		"currency":        "USD",
	}, &out)
	return out, status
}

func (h *Harness) issueCard(t *testing.T, userID, walletID string) cardResponse {
	t.Helper()
	var out cardResponse
	status := (&httpClient{base: h.CardURL}).do(t, http.MethodPost, "/api/v1/cards", userID, newID("card"), map[string]string{
		"wallet_id":       walletID,
		"cardholder_name": "Test User",
	}, &out)
	if status != http.StatusCreated && status != http.StatusOK {
		t.Fatalf("issue card: status %d", status)
	}
	return out
}

func (h *Harness) authorizeCard(t *testing.T, userID, cardID, amount, idempotencyKey string) (authorizationResponse, int) {
	t.Helper()
	var out authorizationResponse
	path := fmt.Sprintf("/api/v1/cards/%s/authorize", cardID)
	status := (&httpClient{base: h.CardURL}).do(t, http.MethodPost, path, userID, idempotencyKey, map[string]string{
		"amount":        amount,
		"currency":      "USD",
		"merchant_name": "Test Merchant",
	}, &out)
	return out, status
}

func (h *Harness) captureAuthorization(t *testing.T, userID, authID string) authorizationResponse {
	t.Helper()
	var out authorizationResponse
	path := fmt.Sprintf("/api/v1/authorizations/%s/capture", authID)
	status := (&httpClient{base: h.CardURL}).do(t, http.MethodPost, path, userID, newID("capture"), nil, &out)
	if status != http.StatusOK {
		t.Fatalf("capture authorization: status %d", status)
	}
	return out
}

func (h *Harness) listWalletTransactions(t *testing.T, userID string) (walletTransactionListResponse, int) {
	t.Helper()
	var out walletTransactionListResponse
	status := (&httpClient{base: h.UserURL}).do(t, http.MethodGet, "/api/v1/wallet/transactions?limit=20", userID, "", nil, &out)
	return out, status
}

func (h *Harness) listNotifications(t *testing.T, userID string) notificationListResponse {
	t.Helper()
	var out notificationListResponse
	status := (&httpClient{base: h.NotificationURL}).do(t, http.MethodGet, "/api/v1/notifications?limit=20", userID, "", nil, &out)
	if status != http.StatusOK {
		t.Fatalf("list notifications: status %d", status)
	}
	return out
}

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, timeNowUnixNano())
}

func timeNowUnixNano() int64 {
	return timeNow().UnixNano()
}

var timeNow = func() time.Time { return time.Now().UTC() }