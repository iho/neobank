//
// Copyright (c) 2026 Sumicare
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
