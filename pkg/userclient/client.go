//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package userclient

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

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Transport: otel.OutboundTransport(nil)},
	}
}

type User struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

type Wallet struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	Currency        string `json:"currency"`
	LedgerAccountID string `json:"ledger_account_id"`
	Status          string `json:"status"`
}

func (c *Client) GetByPhone(ctx context.Context, phone string) (User, error) {
	path := fmt.Sprintf("%s/api/v1/internal/users/by-phone/%s", c.baseURL, url.PathEscape(phone))
	return c.getUser(ctx, path)
}

func (c *Client) GetByEmail(ctx context.Context, email string) (User, error) {
	path := fmt.Sprintf("%s/api/v1/internal/users/by-email/%s", c.baseURL, url.PathEscape(email))
	return c.getUser(ctx, path)
}

func (c *Client) GetByID(ctx context.Context, userID string) (User, error) {
	path := fmt.Sprintf("%s/api/v1/internal/users/%s", c.baseURL, url.PathEscape(userID))
	return c.getUser(ctx, path)
}

func (c *Client) GetWallet(ctx context.Context, userID, currency string) (Wallet, error) {
	u, err := url.Parse(c.baseURL + "/api/v1/internal/wallets")
	if err != nil {
		return Wallet{}, err
	}

	q := u.Query()
	q.Set("user_id", userID)
	q.Set("currency", currency)

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return Wallet{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Wallet{}, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Wallet{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return Wallet{}, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(body))
	}

	var wallet Wallet
	if err := json.Unmarshal(body, &wallet); err != nil {
		return Wallet{}, err
	}

	return wallet, nil
}

type DeviceToken struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

func (c *Client) ListDeviceTokens(ctx context.Context, userID string) ([]DeviceToken, error) {
	path := fmt.Sprintf("%s/api/v1/internal/users/%s/device-tokens", c.baseURL, url.PathEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		DeviceTokens []DeviceToken `json:"device_tokens"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out.DeviceTokens, nil
}

func (c *Client) UpsertPayee(ctx context.Context, userID, payeeUserID, nickname, idempotencyKey string) error {
	body, err := json.Marshal(map[string]string{
		"user_id":       userID,
		"payee_user_id": payeeUserID,
		"nickname":      nickname,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/internal/payees", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("user service status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (c *Client) getUser(ctx context.Context, path string) (User, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return User{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("user service request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return User{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("user service status %d: %s", resp.StatusCode, string(body))
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, err
	}

	return user, nil
}
