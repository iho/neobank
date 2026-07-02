package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type PaymentClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPaymentClient(baseURL string) *PaymentClient {
	return &PaymentClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type CreateTransferRequest struct {
	RecipientPhone string `json:"recipient_phone"`
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	Memo           string `json:"memo"`
}

type TransferView struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	SenderUserID     string `json:"sender_user_id"`
	RecipientUserID  string `json:"recipient_user_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	LedgerTransferID string `json:"ledger_transfer_id,omitempty"`
	FailureReason    string `json:"failure_reason,omitempty"`
	Memo             string `json:"memo,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	CompletedAt      string `json:"completed_at,omitempty"`
}

type TransferList struct {
	Transfers []TransferView `json:"transfers"`
}

func (c *PaymentClient) ListTransfers(ctx context.Context, userID string, limit int) (TransferList, int, error) {
	url := fmt.Sprintf("%s/api/v1/transfers?limit=%d", c.baseURL, limit)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return TransferList{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return TransferList{}, 0, fmt.Errorf("payment service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TransferList{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return TransferList{}, resp.StatusCode, fmt.Errorf("payment service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out TransferList
	if err := json.Unmarshal(respBody, &out); err != nil {
		return TransferList{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *PaymentClient) GetTransfer(ctx context.Context, userID, transferID string) (TransferView, int, error) {
	url := fmt.Sprintf("%s/api/v1/transfers/%s", c.baseURL, transferID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return TransferView{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return TransferView{}, 0, fmt.Errorf("payment service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TransferView{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return TransferView{}, resp.StatusCode, fmt.Errorf("payment service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out TransferView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return TransferView{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *PaymentClient) CreateP2PTransfer(ctx context.Context, userID, idempotencyKey string, req CreateTransferRequest) (TransferView, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return TransferView{}, 0, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/transfers", bytes.NewReader(body))
	if err != nil {
		return TransferView{}, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return TransferView{}, 0, fmt.Errorf("payment service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TransferView{}, 0, err
	}

	var out TransferView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return TransferView{}, resp.StatusCode, fmt.Errorf("payment service status %d: %s", resp.StatusCode, string(respBody))
	}
	if resp.StatusCode >= 400 {
		return out, resp.StatusCode, fmt.Errorf("payment service status %d", resp.StatusCode)
	}
	return out, resp.StatusCode, nil
}