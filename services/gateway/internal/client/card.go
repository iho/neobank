package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CardClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewCardClient(baseURL string) *CardClient {
	return &CardClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type IssueCardRequest struct {
	WalletID       string `json:"wallet_id,omitempty"`
	CardholderName string `json:"cardholder_name"`
}

type CardView struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	WalletID    string `json:"wallet_id"`
	LastFour    string `json:"last_four"`
	Status      string `json:"status"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

type CardList struct {
	Cards []CardView `json:"cards"`
}

func (c *CardClient) IssueCard(ctx context.Context, userID, idempotencyKey string, req IssueCardRequest) (CardView, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return CardView{}, 0, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/cards", bytes.NewReader(body))
	if err != nil {
		return CardView{}, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return CardView{}, 0, fmt.Errorf("card service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CardView{}, 0, err
	}
	if resp.StatusCode >= 400 {
		return CardView{}, resp.StatusCode, fmt.Errorf("card service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out CardView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return CardView{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *CardClient) ListCards(ctx context.Context, userID string) (CardList, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/cards", nil)
	if err != nil {
		return CardList{}, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return CardList{}, fmt.Errorf("card service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CardList{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return CardList{}, fmt.Errorf("card service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out CardList
	if err := json.Unmarshal(respBody, &out); err != nil {
		return CardList{}, err
	}
	return out, nil
}

func (c *CardClient) GetCard(ctx context.Context, userID, cardID string) (CardView, int, error) {
	url := fmt.Sprintf("%s/api/v1/cards/%s", c.baseURL, cardID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return CardView{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return CardView{}, 0, fmt.Errorf("card service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CardView{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return CardView{}, resp.StatusCode, fmt.Errorf("card service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out CardView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return CardView{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *CardClient) FreezeCard(ctx context.Context, userID, cardID string) (CardView, int, error) {
	return c.cardAction(ctx, userID, cardID, "freeze")
}

func (c *CardClient) UnfreezeCard(ctx context.Context, userID, cardID string) (CardView, int, error) {
	return c.cardAction(ctx, userID, cardID, "unfreeze")
}

type AuthorizeRequest struct {
	Amount       string `json:"amount"`
	Currency     string `json:"currency,omitempty"`
	MerchantName string `json:"merchant_name,omitempty"`
}

type AuthorizationView struct {
	ID               string `json:"id"`
	CardID           string `json:"card_id"`
	UserID           string `json:"user_id"`
	MerchantName     string `json:"merchant_name,omitempty"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	Status           string `json:"status"`
	LedgerHoldID     string `json:"ledger_hold_id,omitempty"`
	LedgerTransferID string `json:"ledger_transfer_id,omitempty"`
	FailureReason    string `json:"failure_reason,omitempty"`
}

type AuthorizationList struct {
	Authorizations []AuthorizationView `json:"authorizations"`
}

func (c *CardClient) AuthorizeTransaction(ctx context.Context, userID, cardID, idempotencyKey string, req AuthorizeRequest) (AuthorizationView, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return AuthorizationView{}, 0, err
	}

	url := fmt.Sprintf("%s/api/v1/cards/%s/authorize", c.baseURL, cardID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return AuthorizationView{}, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return AuthorizationView{}, 0, fmt.Errorf("card service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return AuthorizationView{}, 0, err
	}

	var out AuthorizationView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return AuthorizationView{}, resp.StatusCode, fmt.Errorf("card service status %d: %s", resp.StatusCode, string(respBody))
	}
	if resp.StatusCode >= 400 {
		return out, resp.StatusCode, fmt.Errorf("card service status %d", resp.StatusCode)
	}
	return out, resp.StatusCode, nil
}

func (c *CardClient) ListAuthorizations(ctx context.Context, userID string, limit int) (AuthorizationList, error) {
	url := fmt.Sprintf("%s/api/v1/authorizations?limit=%d", c.baseURL, limit)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return AuthorizationList{}, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return AuthorizationList{}, fmt.Errorf("card service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return AuthorizationList{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return AuthorizationList{}, fmt.Errorf("card service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out AuthorizationList
	if err := json.Unmarshal(respBody, &out); err != nil {
		return AuthorizationList{}, err
	}
	return out, nil
}

func (c *CardClient) CaptureAuthorization(ctx context.Context, userID, authID string) (AuthorizationView, int, error) {
	url := fmt.Sprintf("%s/api/v1/authorizations/%s/capture", c.baseURL, authID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return AuthorizationView{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return AuthorizationView{}, 0, fmt.Errorf("card service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return AuthorizationView{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return AuthorizationView{}, resp.StatusCode, fmt.Errorf("card service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out AuthorizationView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return AuthorizationView{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *CardClient) cardAction(ctx context.Context, userID, cardID, action string) (CardView, int, error) {
	url := fmt.Sprintf("%s/api/v1/cards/%s/%s", c.baseURL, cardID, action)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return CardView{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return CardView{}, 0, fmt.Errorf("card service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return CardView{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return CardView{}, resp.StatusCode, fmt.Errorf("card service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out CardView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return CardView{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}