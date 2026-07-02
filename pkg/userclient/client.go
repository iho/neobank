package userclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
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