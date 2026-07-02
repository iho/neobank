package api

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/services/notification/internal/gen/api"
	"github.com/iho/neobank/services/notification/internal/usecase"
	"github.com/jackc/pgx/v5"
)

type Server struct {
	ingest    *usecase.IngestEventUseCase
	list      *usecase.ListNotificationsUseCase
	markRead  *usecase.MarkNotificationReadUseCase
	markAll   *usecase.MarkAllNotificationsReadUseCase
}

func NewServer(
	ingest *usecase.IngestEventUseCase,
	list *usecase.ListNotificationsUseCase,
	markRead *usecase.MarkNotificationReadUseCase,
	markAll *usecase.MarkAllNotificationsReadUseCase,
) *Server {
	return &Server{ingest: ingest, list: list, markRead: markRead, markAll: markAll}
}

func (s *Server) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: "ok", Service: "notification"}, nil
}

func (s *Server) IngestEvent(ctx context.Context, req api.IngestEventRequestObject) (api.IngestEventResponseObject, error) {
	if req.Body == nil {
		return api.IngestEvent400JSONResponse{Error: "invalid_json"}, nil
	}

	payload, err := json.Marshal(req.Body.Payload)
	if err != nil {
		return api.IngestEvent400JSONResponse{Error: "invalid_payload"}, nil
	}

	envelope := events.Envelope{
		EventID:       req.Body.EventId,
		EventType:     req.Body.EventType,
		AggregateType: req.Body.AggregateType,
		AggregateID:   req.Body.AggregateId,
		Payload:       payload,
	}
	if req.Body.EventVersion != nil {
		envelope.EventVersion = *req.Body.EventVersion
	}
	if req.Body.OccurredAt != nil {
		envelope.OccurredAt = *req.Body.OccurredAt
	}

	if err := s.ingest.Execute(ctx, envelope); err != nil {
		return api.IngestEvent400JSONResponse{Error: err.Error()}, nil
	}
	return api.IngestEvent202Response{}, nil
}

func (s *Server) ListNotifications(ctx context.Context, req api.ListNotificationsRequestObject) (api.ListNotificationsResponseObject, error) {
	limit := 20
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	cursor := ""
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	out, err := s.list.Execute(ctx, req.Params.XUserId.String(), limit, cursor)
	if err != nil {
		return nil, err
	}

	views := make([]api.Notification, 0, len(out.Notifications))
	for _, n := range out.Notifications {
		id, _ := uuid.Parse(n.ID)
		userID, _ := uuid.Parse(n.UserID)
		views = append(views, api.Notification{
			Id:        openapi_types.UUID(id),
			UserId:    openapi_types.UUID(userID),
			EventType: n.EventType,
			Title:     n.Title,
			Body:      n.Body,
			Read:      n.Read,
			CreatedAt: n.CreatedAt.UTC(),
		})
	}
	resp := api.ListNotifications200JSONResponse{
		Notifications: views,
		UnreadCount:   out.UnreadCount,
	}
	if out.NextCursor != "" {
		resp.NextCursor = &out.NextCursor
	}
	return resp, nil
}

func (s *Server) MarkNotificationRead(ctx context.Context, req api.MarkNotificationReadRequestObject) (api.MarkNotificationReadResponseObject, error) {
	n, err := s.markRead.Execute(ctx, req.Params.XUserId.String(), req.Id.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.MarkNotificationRead404JSONResponse{Error: "notification_not_found"}, nil
		}
		return nil, err
	}
	id, _ := uuid.Parse(n.ID)
	userID, _ := uuid.Parse(n.UserID)
	return api.MarkNotificationRead200JSONResponse{
		Id:        openapi_types.UUID(id),
		UserId:    openapi_types.UUID(userID),
		EventType: n.EventType,
		Title:     n.Title,
		Body:      n.Body,
		Read:      n.Read,
		CreatedAt: n.CreatedAt.UTC(),
	}, nil
}

func (s *Server) MarkAllNotificationsRead(ctx context.Context, req api.MarkAllNotificationsReadRequestObject) (api.MarkAllNotificationsReadResponseObject, error) {
	count, err := s.markAll.Execute(ctx, req.Params.XUserId.String())
	if err != nil {
		return nil, err
	}
	return api.MarkAllNotificationsRead200JSONResponse{MarkedCount: count}, nil
}