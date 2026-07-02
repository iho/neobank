package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type NotificationClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
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
}

func (c *NotificationClient) ListNotifications(ctx context.Context, userID string, limit int) (NotificationList, error) {
	url := fmt.Sprintf("%s/api/v1/notifications?limit=%d", c.baseURL, limit)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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