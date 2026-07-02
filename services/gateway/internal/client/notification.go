package client

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

type NotificationClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Transport: otel.OutboundTransport(nil)},
	}
}

type NotificationView struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	EventType string `json:"event_type"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}

type NotificationList struct {
	Notifications []NotificationView `json:"notifications"`
	UnreadCount   int64              `json:"unread_count"`
	NextCursor    string             `json:"next_cursor,omitempty"`
}

func (c *NotificationClient) ListNotifications(ctx context.Context, userID string, limit int, cursor string) (NotificationList, error) {
	reqURL := fmt.Sprintf("%s/api/v1/notifications?limit=%d", c.baseURL, limit)
	if cursor != "" {
		reqURL += "&cursor=" + url.QueryEscape(cursor)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return NotificationList{}, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return NotificationList{}, fmt.Errorf("notification service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NotificationList{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return NotificationList{}, fmt.Errorf("notification service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out NotificationList
	if err := json.Unmarshal(respBody, &out); err != nil {
		return NotificationList{}, err
	}
	return out, nil
}

func (c *NotificationClient) MarkNotificationRead(ctx context.Context, userID, notificationID string) (NotificationView, int, error) {
	url := fmt.Sprintf("%s/api/v1/notifications/%s/read", c.baseURL, notificationID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return NotificationView{}, 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return NotificationView{}, 0, fmt.Errorf("notification service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NotificationView{}, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return NotificationView{}, resp.StatusCode, fmt.Errorf("notification service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out NotificationView
	if err := json.Unmarshal(respBody, &out); err != nil {
		return NotificationView{}, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func (c *NotificationClient) MarkAllNotificationsRead(ctx context.Context, userID string) (int64, error) {
	url := c.baseURL + "/api/v1/notifications/read-all"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("X-User-Id", userID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("notification service request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("notification service status %d: %s", resp.StatusCode, string(respBody))
	}

	var out struct {
		MarkedCount int64 `json:"marked_count"`
	}
	if err := json.Unmarshal(respBody, &out); err != nil {
		return 0, err
	}
	return out.MarkedCount, nil
}

type NotificationPreferences struct {
	Transfers bool `json:"transfers"`
	Cards     bool `json:"cards"`
	KYC       bool `json:"kyc"`
	Push      bool `json:"push"`
	Email     bool `json:"email"`
}

type UpdateNotificationPreferencesRequest struct {
	Transfers *bool `json:"transfers,omitempty"`
	Cards     *bool `json:"cards,omitempty"`
	KYC       *bool `json:"kyc,omitempty"`
	Push      *bool `json:"push,omitempty"`
	Email     *bool `json:"email,omitempty"`
}

func (c *NotificationClient) GetNotificationPreferences(ctx context.Context, userID string) (NotificationPreferences, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/notification-preferences", nil)
	if err != nil {
		return NotificationPreferences{}, err
	}
	httpReq.Header.Set("X-User-Id", userID)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return NotificationPreferences{}, fmt.Errorf("notification service request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NotificationPreferences{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return NotificationPreferences{}, fmt.Errorf("notification service status %d: %s", resp.StatusCode, string(respBody))
	}
	var out NotificationPreferences
	if err := json.Unmarshal(respBody, &out); err != nil {
		return NotificationPreferences{}, err
	}
	return out, nil
}

func (c *NotificationClient) UpdateNotificationPreferences(ctx context.Context, userID string, req UpdateNotificationPreferencesRequest) (NotificationPreferences, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return NotificationPreferences{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+"/api/v1/notification-preferences", bytes.NewReader(body))
	if err != nil {
		return NotificationPreferences{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-Id", userID)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return NotificationPreferences{}, fmt.Errorf("notification service request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return NotificationPreferences{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return NotificationPreferences{}, fmt.Errorf("notification service status %d: %s", resp.StatusCode, string(respBody))
	}
	var out NotificationPreferences
	if err := json.Unmarshal(respBody, &out); err != nil {
		return NotificationPreferences{}, err
	}
	return out, nil
}
