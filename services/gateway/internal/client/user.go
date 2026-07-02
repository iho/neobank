package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type UserClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewUserClient(baseURL string) *UserClient {
	return &UserClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
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
	FullName    string `json:"full_name"`
	DateOfBirth string `json:"date_of_birth"`
	CountryCode string `json:"country_code"`
}

type SubmitKYCResponse struct {
	KYCCaseID string `json:"kyc_case_id"`
	Status    string `json:"status"`
	WalletID  string `json:"wallet_id"`
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