// Package railsclient is a thin HTTP client for the rails simulator
// (services/simulators/rails); a real payment-rail integration would
// implement the same usecase.RailsClient interface behind this package.
package railsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Config struct {
	BaseURL string
}

type Client struct {
	baseURL string
	http    *http.Client
}

func New(cfg Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type Account struct {
	ID          string `json:"id"`
	ExternalRef string `json:"external_ref"`
	Currency    string `json:"currency"`
	IBAN        string `json:"iban"`
}

// CreateAccount asks the rails simulator to get-or-create the virtual IBAN
// for (externalRef, currency); the simulator makes this idempotent.
func (c *Client) CreateAccount(ctx context.Context, externalRef, currency string) (Account, error) {
	body, err := json.Marshal(map[string]string{"external_ref": externalRef, "currency": currency})
	if err != nil {
		return Account{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/accounts", bytes.NewReader(body))
	if err != nil {
		return Account{}, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Account{}, fmt.Errorf("call rails simulator: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return Account{}, fmt.Errorf("rails simulator returned status %d: %s", resp.StatusCode, respBody)
	}

	var account Account
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return Account{}, fmt.Errorf("decode response: %w", err)
	}

	return account, nil
}

type OutboundPayment struct {
	ID               string `json:"id"`
	AccountID        string `json:"account_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	CounterpartyIBAN string `json:"counterparty_iban"`
	Reference        string `json:"reference"`
	Status           string `json:"status"`
}

// CreatePayment initiates an outbound payment; the outcome (settled,
// returned, or failed) arrives later via webhook, not in this response.
func (c *Client) CreatePayment(ctx context.Context, accountID, amount, currency, counterpartyIBAN, reference string) (OutboundPayment, error) {
	body, err := json.Marshal(map[string]string{
		"account_id":        accountID,
		"amount":            amount,
		"currency":          currency,
		"counterparty_iban": counterpartyIBAN,
		"reference":         reference,
	})
	if err != nil {
		return OutboundPayment{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/payments", bytes.NewReader(body))
	if err != nil {
		return OutboundPayment{}, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return OutboundPayment{}, fmt.Errorf("call rails simulator: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return OutboundPayment{}, fmt.Errorf("rails simulator returned status %d: %s", resp.StatusCode, respBody)
	}

	var payment OutboundPayment
	if err := json.NewDecoder(resp.Body).Decode(&payment); err != nil {
		return OutboundPayment{}, fmt.Errorf("decode response: %w", err)
	}

	return payment, nil
}
