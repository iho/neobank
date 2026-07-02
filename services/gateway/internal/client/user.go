package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/iho/neobank/pkg/otel"
)

type UserClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewUserClient(baseURL string) *UserClient {
	return &UserClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Transport: otel.OutboundTransport(nil)},
	}
}

type RegisterRequest struct {
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	InviteCode string `json:"invite_code,omitempty"`
}

type RegisterResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SubmitKYCRequest struct {
	FullName       string `json:"full_name"`
	DateOfBirth    string `json:"date_of_birth"`
	CountryCode    string `json:"country_code"`
	DocumentType   string `json:"document_type,omitempty"`
	DocumentNumber string `json:"document_number,omitempty"`
}

type SubmitKYCResponse struct {
	KYCCaseID       string `json:"kyc_case_id"`
	Status          string `json:"status"`
	WalletID        string `json:"wallet_id,omitempty"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type KYCStatusResponse struct {
	Status          string `json:"status"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type WalletBalance struct {
	WalletID          string `json:"wallet_id"`
	LedgerAccountID   string `json:"ledger_account_id,omitempty"`
	Currency          string `json:"currency"`
	Balance           string `json:"balance"`
	EncumberedBalance string `json:"encumbered_balance,omitempty"`
	AvailableBalance  string `json:"available_balance"`
}

type ProvisionWalletResponse struct {
	WalletID        string `json:"wallet_id"`
	LedgerAccountID string `json:"ledger_account_id"`
}

type WalletTransactionView struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	Direction    string `json:"direction"`
	Status       string `json:"status"`
	Counterparty string `json:"counterparty,omitempty"`
	Memo         string `json:"memo,omitempty"`
	ReferenceID  string `json:"reference_id,omitempty"`
	CreatedAt    string `json:"created_at"`
}

type WalletTransactionList struct {
	Transactions []WalletTransactionView `json:"transactions"`
	NextCursor   string                  `json:"next_cursor,omitempty"`
}

type PayeeView struct {
	ID          string `json:"id"`
	PayeeUserID string `json:"payee_user_id"`
	Nickname    string `json:"nickname,omitempty"`
	PayeeEmail  string `json:"payee_email,omitempty"`
	PayeePhone  string `json:"payee_phone,omitempty"`
	LastUsedAt  string `json:"last_used_at"`
	CreatedAt   string `json:"created_at"`
}

type PayeeList struct {
	Payees []PayeeView `json:"payees"`
}

type ProfileView struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Status      string `json:"status"`
	FullName    string `json:"full_name,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	KYCStatus   string `json:"kyc_status"`
	CreatedAt   string `json:"created_at"`
}

func (c *UserClient) RefreshToken(ctx context.Context, refreshToken string) (LoginResponse, error) {
	body, err := json.Marshal(map[string]string{"refresh_token": refreshToken})
	if err != nil {
		return LoginResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/auth/refresh", bytes.NewReader(body))
	if err != nil {
		return LoginResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return LoginResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return LoginResponse{}, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out LoginResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return LoginResponse{}, err
	}
	return out, nil
}

func (c *UserClient) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return LoginResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/auth/login", bytes.NewReader(body))
	if err != nil {
		return LoginResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return LoginResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return LoginResponse{}, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out LoginResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return LoginResponse{}, err
	}
	return out, nil
}

func (c *UserClient) Register(ctx context.Context, idempotencyKey string, req RegisterRequest) (RegisterResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return RegisterResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/auth/register", bytes.NewReader(body))
	if err != nil {
		return RegisterResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return RegisterResponse{}, err
	}

	if resp.StatusCode != http.StatusCreated {
		return RegisterResponse{}, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out RegisterResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return RegisterResponse{}, err
	}
	return out, nil
}

func (c *UserClient) SubmitKYC(ctx context.Context, userID, idempotencyKey string, req SubmitKYCRequest) (SubmitKYCResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return SubmitKYCResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/kyc", bytes.NewReader(body))
	if err != nil {
		return SubmitKYCResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return SubmitKYCResponse{}, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return SubmitKYCResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return SubmitKYCResponse{}, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out SubmitKYCResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return SubmitKYCResponse{}, err
	}
	return out, nil
}

func (c *UserClient) GetProfile(ctx context.Context, userID string) (ProfileView, int, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/me", nil)
	if err != nil {
		return ProfileView{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ProfileView{}, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ProfileView{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return ProfileView{}, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out ProfileView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return ProfileView{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *UserClient) GetKYCStatus(ctx context.Context, userID string) (KYCStatusResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/kyc/status", nil)
	if err != nil {
		return KYCStatusResponse{}, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return KYCStatusResponse{}, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return KYCStatusResponse{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return KYCStatusResponse{}, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out KYCStatusResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return KYCStatusResponse{}, err
	}
	return out, nil
}

func (c *UserClient) GetWalletBalance(ctx context.Context, userID, currency string) (WalletBalance, int, error) {
	url := c.baseURL + "/api/v1/wallets/balance?currency=" + currency
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return WalletBalance{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return WalletBalance{}, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return WalletBalance{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return WalletBalance{}, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out WalletBalance
	if err := json.Unmarshal(respBody, &out); err != nil {
		return WalletBalance{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *UserClient) ListWalletTransactions(ctx context.Context, userID string, limit int, cursor string) (WalletTransactionList, int, error) {
	reqURL := fmt.Sprintf("%s/api/v1/wallet/transactions?limit=%d", c.baseURL, limit)
	if cursor != "" {
		reqURL += "&cursor=" + url.QueryEscape(cursor)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return WalletTransactionList{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return WalletTransactionList{}, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return WalletTransactionList{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return WalletTransactionList{}, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out WalletTransactionList
	if err := json.Unmarshal(respBody, &out); err != nil {
		return WalletTransactionList{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *UserClient) ProvisionWallet(ctx context.Context, userID, idempotencyKey, currency string) (ProvisionWalletResponse, error) {
	body, _ := json.Marshal(map[string]string{"user_id": userID, "currency": currency})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/wallets", bytes.NewReader(body))
	if err != nil {
		return ProvisionWalletResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ProvisionWalletResponse{}, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ProvisionWalletResponse{}, err
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return ProvisionWalletResponse{}, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out ProvisionWalletResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return ProvisionWalletResponse{}, err
	}
	return out, nil
}

type DepositWalletRequest struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency,omitempty"`
}

type DepositWalletResponse struct {
	ID               string `json:"id"`
	WalletID         string `json:"wallet_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	LedgerTransferID string `json:"ledger_transfer_id,omitempty"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at,omitempty"`
}

func (c *UserClient) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) (int, error) {
	body, err := json.Marshal(map[string]string{
		"current_password": currentPassword,
		"new_password":     newPassword,
	})
	if err != nil {
		return 0, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/auth/change-password", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusNoContent {
		return resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}
	return resp.StatusCode, nil
}

func (c *UserClient) ListPayees(ctx context.Context, userID string, limit int) (PayeeList, int, error) {
	url := fmt.Sprintf("%s/api/v1/payees?limit=%d", c.baseURL, limit)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return PayeeList{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return PayeeList{}, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return PayeeList{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return PayeeList{}, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}
	var out PayeeList
	if err := json.Unmarshal(respBody, &out); err != nil {
		return PayeeList{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *UserClient) CreatePayee(ctx context.Context, userID, payeeUserID, nickname string) (PayeeView, int, error) {
	body, err := json.Marshal(map[string]string{
		"payee_user_id": payeeUserID,
		"nickname":      nickname,
	})
	if err != nil {
		return PayeeView{}, 0, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/payees", bytes.NewReader(body))
	if err != nil {
		return PayeeView{}, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-Id", userID)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return PayeeView{}, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return PayeeView{}, 0, err
	}
	if resp.StatusCode != http.StatusCreated {
		return PayeeView{}, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}
	var out PayeeView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return PayeeView{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *UserClient) DeletePayee(ctx context.Context, userID, payeeID string) (int, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/v1/payees/"+payeeID, nil)
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusNoContent {
		return resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}
	return resp.StatusCode, nil
}

type DeviceTokenView struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Platform  string `json:"platform"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

type RegisterDeviceTokenRequest struct {
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

func (c *UserClient) RegisterDeviceToken(ctx context.Context, userID, idempotencyKey string, req RegisterDeviceTokenRequest) (DeviceTokenView, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return DeviceTokenView{}, 0, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/devices", bytes.NewReader(body))
	if err != nil {
		return DeviceTokenView{}, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-Id", userID)
	if idempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return DeviceTokenView{}, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return DeviceTokenView{}, 0, err
	}
	if resp.StatusCode != http.StatusCreated {
		return DeviceTokenView{}, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}
	var out DeviceTokenView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return DeviceTokenView{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *UserClient) DeleteDeviceToken(ctx context.Context, userID, tokenID, idempotencyKey string) (int, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/v1/devices/"+tokenID, nil)
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)
	if idempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusNoContent {
		return resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}
	return resp.StatusCode, nil
}

func (c *UserClient) CloseAccount(ctx context.Context, userID, idempotencyKey string) (int, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/account/close", nil)
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusNoContent {
		return resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}
	return resp.StatusCode, nil
}

type WalletBalanceList struct {
	Wallets []WalletBalance `json:"wallets"`
}

type ReferralInviteView struct {
	ID            string `json:"id"`
	InviterUserID string `json:"inviter_user_id"`
	InviteCode    string `json:"invite_code"`
	InviteeUserID string `json:"invitee_user_id,omitempty"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	AcceptedAt    string `json:"accepted_at,omitempty"`
}

type ReferralInviteList struct {
	Invites []ReferralInviteView `json:"invites"`
}

func (c *UserClient) ExportWalletTransactions(ctx context.Context, userID, format string, from, to string) ([]byte, int, error) {
	reqURL := fmt.Sprintf("%s/api/v1/wallet/transactions/export?format=%s&from=%s&to=%s",
		c.baseURL, url.QueryEscape(format), url.QueryEscape(from), url.QueryEscape(to))
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, resp.StatusCode, nil
}

func (c *UserClient) ListWallets(ctx context.Context, userID string) (WalletBalanceList, int, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/wallets", nil)
	if err != nil {
		return WalletBalanceList{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return WalletBalanceList{}, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return WalletBalanceList{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return WalletBalanceList{}, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out WalletBalanceList
	if err := json.Unmarshal(respBody, &out); err != nil {
		return WalletBalanceList{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *UserClient) CreateReferralInvite(ctx context.Context, userID, idempotencyKey string) (ReferralInviteView, int, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/invites", nil)
	if err != nil {
		return ReferralInviteView{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)
	if idempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ReferralInviteView{}, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReferralInviteView{}, 0, err
	}
	if resp.StatusCode != http.StatusCreated {
		return ReferralInviteView{}, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out ReferralInviteView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return ReferralInviteView{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *UserClient) ListReferralInvites(ctx context.Context, userID string, limit int) (ReferralInviteList, int, error) {
	reqURL := fmt.Sprintf("%s/api/v1/invites?limit=%d", c.baseURL, limit)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return ReferralInviteList{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ReferralInviteList{}, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ReferralInviteList{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return ReferralInviteList{}, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out ReferralInviteList
	if err := json.Unmarshal(respBody, &out); err != nil {
		return ReferralInviteList{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *UserClient) DepositWallet(ctx context.Context, userID, idempotencyKey string, req DepositWalletRequest) (DepositWalletResponse, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return DepositWalletResponse{}, 0, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/wallets/deposit", bytes.NewReader(body))
	if err != nil {
		return DepositWalletResponse{}, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return DepositWalletResponse{}, 0, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return DepositWalletResponse{}, 0, err
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return DepositWalletResponse{}, resp.StatusCode, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out DepositWalletResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return DepositWalletResponse{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}
