package client

import (
	"context"
	"fmt"

	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	"google.golang.org/grpc"
)

type NotificationClient struct {
	conn *grpc.ClientConn
	rpc  neobankv1.NotificationServiceClient
}

func NewNotificationClient(ctx context.Context, cfg Config) (*NotificationClient, error) {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:50055"
	}
	conn, err := grpcutil.Dial(ctx, cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("dial notification service: %w", err)
	}
	return &NotificationClient{
		conn: conn,
		rpc:  neobankv1.NewNotificationServiceClient(conn),
	}, nil
}

func (c *NotificationClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
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
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.ListNotifications(ctx, &neobankv1.ListNotificationsRequest{
		UserId: userID,
		Limit:  int32(limit),
		Cursor: cursor,
	})
	if err != nil {
		return NotificationList{}, dialError("notification", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return NotificationList{}, statusError("notification", status, resp.GetError())
	}
	out := NotificationList{
		UnreadCount: resp.GetUnreadCount(),
		NextCursor:  resp.GetNextCursor(),
	}
	for _, n := range resp.GetNotifications() {
		out.Notifications = append(out.Notifications, toNotificationView(n))
	}
	return out, nil
}

func (c *NotificationClient) MarkNotificationRead(ctx context.Context, userID, notificationID string) (NotificationView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.MarkNotificationRead(ctx, &neobankv1.MarkNotificationReadRequest{
		UserId:         userID,
		NotificationId: notificationID,
	})
	if err != nil {
		return NotificationView{}, 0, dialError("notification", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return NotificationView{}, status, statusError("notification", status, resp.GetError())
	}
	return toNotificationView(resp.GetNotification()), status, nil
}

func (c *NotificationClient) MarkAllNotificationsRead(ctx context.Context, userID string) (int64, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.MarkAllNotificationsRead(ctx, &neobankv1.MarkAllNotificationsReadRequest{UserId: userID})
	if err != nil {
		return 0, dialError("notification", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return 0, statusError("notification", status, resp.GetError())
	}
	return resp.GetMarkedCount(), nil
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
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.GetNotificationPreferences(ctx, &neobankv1.GetNotificationPreferencesRequest{UserId: userID})
	if err != nil {
		return NotificationPreferences{}, dialError("notification", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return NotificationPreferences{}, statusError("notification", status, resp.GetError())
	}
	return toNotificationPreferences(resp.GetPreferences()), nil
}

func (c *NotificationClient) UpdateNotificationPreferences(ctx context.Context, userID string, req UpdateNotificationPreferencesRequest) (NotificationPreferences, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	protoReq := &neobankv1.UpdateNotificationPreferencesRequest{UserId: userID}
	if req.Transfers != nil {
		protoReq.Transfers = req.Transfers
	}
	if req.Cards != nil {
		protoReq.Cards = req.Cards
	}
	if req.KYC != nil {
		protoReq.Kyc = req.KYC
	}
	if req.Push != nil {
		protoReq.Push = req.Push
	}
	if req.Email != nil {
		protoReq.Email = req.Email
	}
	resp, err := c.rpc.UpdateNotificationPreferences(ctx, protoReq)
	if err != nil {
		return NotificationPreferences{}, dialError("notification", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return NotificationPreferences{}, statusError("notification", status, resp.GetError())
	}
	return toNotificationPreferences(resp.GetPreferences()), nil
}

func toNotificationView(n *neobankv1.Notification) NotificationView {
	if n == nil {
		return NotificationView{}
	}
	return NotificationView{
		ID:        n.GetId(),
		UserID:    n.GetUserId(),
		EventType: n.GetEventType(),
		Title:     n.GetTitle(),
		Body:      n.GetBody(),
		Read:      n.GetRead(),
		CreatedAt: n.GetCreatedAt(),
	}
}

func toNotificationPreferences(p *neobankv1.NotificationPreferences) NotificationPreferences {
	if p == nil {
		return NotificationPreferences{}
	}
	return NotificationPreferences{
		Transfers: p.GetTransfers(),
		Cards:     p.GetCards(),
		KYC:       p.GetKyc(),
		Push:      p.GetPush(),
		Email:     p.GetEmail(),
	}
}