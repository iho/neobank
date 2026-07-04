package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient is a Processor backed by the cardproc simulator
// (services/simulators/cardproc); a real card-network integration would
// implement the same Processor interface behind its own package.
type HTTPClient struct {
	baseURL string
	http    *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type createCardResponse struct {
	Ref         string `json:"ref"`
	PANToken    string `json:"pan_token"`
	LastFour    string `json:"last_four"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

func (c *HTTPClient) CreateVirtualCard(ctx context.Context, userID, cardholderName string) (VirtualCard, error) {
	body, err := json.Marshal(map[string]string{"external_ref": userID, "cardholder_name": cardholderName})
	if err != nil {
		return VirtualCard{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/cards", bytes.NewReader(body))
	if err != nil {
		return VirtualCard{}, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return VirtualCard{}, fmt.Errorf("call cardproc simulator: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return VirtualCard{}, fmt.Errorf("cardproc simulator returned status %d: %s", resp.StatusCode, respBody)
	}

	var card createCardResponse
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return VirtualCard{}, fmt.Errorf("decode response: %w", err)
	}

	return VirtualCard(card), nil
}

func (c *HTTPClient) CancelCard(ctx context.Context, ref string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/cards/"+ref+"/cancel", nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call cardproc simulator: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cardproc simulator returned status %d: %s", resp.StatusCode, respBody)
	}

	return nil
}
