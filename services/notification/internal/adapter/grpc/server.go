package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	apiadapter "github.com/iho/neobank/services/notification/internal/adapter/api"
	genapi "github.com/iho/neobank/services/notification/internal/gen/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	neobankv1.UnimplementedNotificationServiceServer
	api *apiadapter.Server
}

func NewServer(api *apiadapter.Server) *Server {
	return &Server{api: api}
}

func (s *Server) ListNotifications(ctx context.Context, req *neobankv1.ListNotificationsRequest) (*neobankv1.ListNotificationsResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	params := genapi.ListNotificationsParams{XUserId: userID}
	if limit := req.GetLimit(); limit > 0 {
		l := int(limit)
		params.Limit = &l
	}
	if cursor := req.GetCursor(); cursor != "" {
		params.Cursor = &cursor
	}
	resp, err := s.api.ListNotifications(ctx, genapi.ListNotificationsRequestObject{Params: params})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list notifications: %v", err)
	}
	return mapListNotificationsResponse(resp)
}

func (s *Server) MarkNotificationRead(ctx context.Context, req *neobankv1.MarkNotificationReadRequest) (*neobankv1.NotificationResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	notificationID, err := parseUUID(req.GetNotificationId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid notification_id: %v", err)
	}
	resp, err := s.api.MarkNotificationRead(ctx, genapi.MarkNotificationReadRequestObject{
		Id:     notificationID,
		Params: genapi.MarkNotificationReadParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "mark notification read: %v", err)
	}
	return mapNotificationResponse(resp)
}

func (s *Server) MarkAllNotificationsRead(ctx context.Context, req *neobankv1.MarkAllNotificationsReadRequest) (*neobankv1.MarkAllReadResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	resp, err := s.api.MarkAllNotificationsRead(ctx, genapi.MarkAllNotificationsReadRequestObject{
		Params: genapi.MarkAllNotificationsReadParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "mark all notifications read: %v", err)
	}
	return mapMarkAllReadResponse(resp)
}

func (s *Server) GetNotificationPreferences(ctx context.Context, req *neobankv1.GetNotificationPreferencesRequest) (*neobankv1.NotificationPreferencesResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	resp, err := s.api.GetNotificationPreferences(ctx, genapi.GetNotificationPreferencesRequestObject{
		Params: genapi.GetNotificationPreferencesParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get notification preferences: %v", err)
	}
	return mapNotificationPreferencesResponse(resp)
}

func (s *Server) UpdateNotificationPreferences(ctx context.Context, req *neobankv1.UpdateNotificationPreferencesRequest) (*neobankv1.NotificationPreferencesResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	body := &genapi.UpdateNotificationPreferencesJSONRequestBody{}
	if req.Transfers != nil {
		body.Transfers = req.Transfers
	}
	if req.Cards != nil {
		body.Cards = req.Cards
	}
	if req.Kyc != nil {
		body.Kyc = req.Kyc
	}
	if req.Push != nil {
		body.Push = req.Push
	}
	if req.Email != nil {
		body.Email = req.Email
	}
	resp, err := s.api.UpdateNotificationPreferences(ctx, genapi.UpdateNotificationPreferencesRequestObject{
		Params: genapi.UpdateNotificationPreferencesParams{XUserId: userID},
		Body:   body,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update notification preferences: %v", err)
	}
	return mapUpdateNotificationPreferencesResponse(resp)
}

func parseXUserID(userID string) (genapi.XUserId, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return genapi.XUserId{}, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	return id, nil
}

func parseUUID(value string) (openapi_types.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return openapi_types.UUID{}, err
	}
	return openapi_types.UUID(id), nil
}

func mapListNotificationsResponse(resp genapi.ListNotificationsResponseObject) (*neobankv1.ListNotificationsResponse, error) {
	switch r := resp.(type) {
	case genapi.ListNotifications200JSONResponse:
		notifications := make([]*neobankv1.Notification, 0, len(r.Notifications))
		for _, n := range r.Notifications {
			item := toProtoNotification(n)
			notifications = append(notifications, &item)
		}
		out := &neobankv1.ListNotificationsResponse{
			Notifications: notifications,
			UnreadCount:   r.UnreadCount,
			HttpStatus:    200,
		}
		if r.NextCursor != nil {
			out.NextCursor = *r.NextCursor
		}
		return out, nil
	case genapi.ListNotifications401JSONResponse:
		return &neobankv1.ListNotificationsResponse{HttpStatus: 401, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected list notifications response")
	}
}

func mapNotificationResponse(resp genapi.MarkNotificationReadResponseObject) (*neobankv1.NotificationResponse, error) {
	switch r := resp.(type) {
	case genapi.MarkNotificationRead200JSONResponse:
		n := toProtoNotification(genapi.Notification(r))
		return &neobankv1.NotificationResponse{Notification: &n, HttpStatus: 200}, nil
	case genapi.MarkNotificationRead404JSONResponse:
		return &neobankv1.NotificationResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected notification response")
	}
}

func mapMarkAllReadResponse(resp genapi.MarkAllNotificationsReadResponseObject) (*neobankv1.MarkAllReadResponse, error) {
	switch r := resp.(type) {
	case genapi.MarkAllNotificationsRead200JSONResponse:
		return &neobankv1.MarkAllReadResponse{MarkedCount: r.MarkedCount, HttpStatus: 200}, nil
	case genapi.MarkAllNotificationsRead401JSONResponse:
		return &neobankv1.MarkAllReadResponse{HttpStatus: 401, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected mark all read response")
	}
}

func mapNotificationPreferencesResponse(resp genapi.GetNotificationPreferencesResponseObject) (*neobankv1.NotificationPreferencesResponse, error) {
	switch r := resp.(type) {
	case genapi.GetNotificationPreferences200JSONResponse:
		prefs := toProtoNotificationPreferences(genapi.NotificationPreferences(r))
		return &neobankv1.NotificationPreferencesResponse{Preferences: &prefs, HttpStatus: 200}, nil
	case genapi.GetNotificationPreferences401JSONResponse:
		return &neobankv1.NotificationPreferencesResponse{HttpStatus: 401, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected notification preferences response")
	}
}

func mapUpdateNotificationPreferencesResponse(resp genapi.UpdateNotificationPreferencesResponseObject) (*neobankv1.NotificationPreferencesResponse, error) {
	switch r := resp.(type) {
	case genapi.UpdateNotificationPreferences200JSONResponse:
		prefs := toProtoNotificationPreferences(genapi.NotificationPreferences(r))
		return &neobankv1.NotificationPreferencesResponse{Preferences: &prefs, HttpStatus: 200}, nil
	case genapi.UpdateNotificationPreferences400JSONResponse:
		return &neobankv1.NotificationPreferencesResponse{HttpStatus: 400, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected update notification preferences response")
	}
}

func toProtoNotification(n genapi.Notification) neobankv1.Notification {
	return neobankv1.Notification{
		Id:        n.Id.String(),
		UserId:    n.UserId.String(),
		EventType: n.EventType,
		Title:     n.Title,
		Body:      n.Body,
		Read:      n.Read,
		CreatedAt: n.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toProtoNotificationPreferences(p genapi.NotificationPreferences) neobankv1.NotificationPreferences {
	return neobankv1.NotificationPreferences{
		Transfers: p.Transfers,
		Cards:     p.Cards,
		Kyc:       p.Kyc,
		Push:      p.Push,
		Email:     p.Email,
	}
}