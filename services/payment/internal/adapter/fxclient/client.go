// Package fxclient is a thin HTTP client for the fx rates simulator
// (services/simulators/fx); a real FX-provider integration would implement
// the same usecase interfaces behind this package.
package fxclient

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

type Quote struct {
	ExpiresAt       time.Time `json:"expires_at"`
	ID              string    `json:"id"`
	FromCurrency    string    `json:"from_currency"`
	ToCurrency      string    `json:"to_currency"`
	Amount          string    `json:"amount"`
	ConvertedAmount string    `json:"converted_amount"`
	Rate            string    `json:"rate"`
	Status          string    `json:"status"`
	SpreadBps       int       `json:"spread_bps"`
}

// CreateQuote prices a conversion; the caller must ExecuteQuote before
// Quote.ExpiresAt to lock in the rate.
func (c *Client) CreateQuote(ctx context.Context, fromCurrency, toCurrency, amount string) (Quote, error) {
	body, err := json.Marshal(map[string]string{
		"from_currency": fromCurrency,
		"to_currency":   toCurrency,
		"amount":        amount,
	})
	if err != nil {
		return Quote{}, fmt.Errorf("marshal request: %w", err)
	}

	return c.do(ctx, http.MethodPost, c.baseURL+"/v1/quotes", body, http.StatusCreated)
}

// ExecuteQuote locks in a previously priced quote. Calling it again on an
// already-executed quote is a no-op that returns the same result.
func (c *Client) ExecuteQuote(ctx context.Context, quoteID string) (Quote, error) {
	return c.do(ctx, http.MethodPost, c.baseURL+"/v1/quotes/"+quoteID+"/execute", nil, http.StatusOK)
}

func (c *Client) do(ctx context.Context, method, url string, body []byte, wantStatus int) (Quote, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return Quote{}, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Quote{}, fmt.Errorf("call fx simulator: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		respBody, _ := io.ReadAll(resp.Body)
		return Quote{}, fmt.Errorf("fx simulator returned status %d: %s", resp.StatusCode, respBody)
	}

	var quote Quote
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return Quote{}, fmt.Errorf("decode response: %w", err)
	}

	return quote, nil
}
