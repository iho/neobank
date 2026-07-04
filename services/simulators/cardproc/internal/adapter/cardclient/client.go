// Package cardclient calls the card service's synchronous, real-time
// authorization webhook and waits for an approve/decline decision — the one
// place in the vendor-simulator design that isn't fire-and-forget, matching
// how a real card network's stand-in authorization flow behaves.
package cardclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/iho/neobank/pkg/vendorsim"
)

type Config struct {
	AuthorizeURL string
	Secret       []byte
	Timeout      time.Duration
}

type Client struct {
	cfg  Config
	http *http.Client
}

func New(cfg Config) *Client {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}

	return &Client{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

type AuthorizeRequest struct {
	TransactionID        string `json:"transaction_id"`
	CardRef              string `json:"card_ref"`
	Amount               string `json:"amount"`
	Currency             string `json:"currency"`
	MerchantName         string `json:"merchant_name"`
	MerchantCategoryCode string `json:"mcc"`
}

type AuthorizeResult struct {
	Decision        string `json:"decision"`
	AuthorizationID string `json:"authorization_id,omitempty"`
	ReasonCode      string `json:"reason_code,omitempty"`
}

// Authorize calls the card service synchronously; a network error or
// timeout is treated as a decline (real card networks "stand in" and
// decline when the issuer doesn't answer in time), never as an approval.
func (c *Client) Authorize(ctx context.Context, req AuthorizeRequest) (AuthorizeResult, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return AuthorizeResult{}, fmt.Errorf("marshal request: %w", err)
	}

	ts := time.Now().Unix()
	sig := vendorsim.Sign(c.cfg.Secret, ts, body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.AuthorizeURL, bytes.NewReader(body))
	if err != nil {
		return AuthorizeResult{}, fmt.Errorf("build request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(vendorsim.HeaderTimestamp, fmt.Sprintf("%d", ts))
	httpReq.Header.Set(vendorsim.HeaderSignature, sig)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return AuthorizeResult{Decision: "declined", ReasonCode: "issuer_unreachable"}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AuthorizeResult{Decision: "declined", ReasonCode: "issuer_error"}, nil
	}

	var result AuthorizeResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return AuthorizeResult{Decision: "declined", ReasonCode: "invalid_issuer_response"}, nil
	}

	return result, nil
}
