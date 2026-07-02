package api

import (
	"context"
	"errors"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/iho/neobank/services/payment/internal/domain"
	"github.com/iho/neobank/services/payment/internal/gen/api"
	"github.com/iho/neobank/services/payment/internal/usecase"
	"github.com/jackc/pgx/v5"
)

type Server struct {
	p2p    *usecase.P2PTransferUseCase
	limits *usecase.GetLimitsUseCase
}

func NewServer(p2p *usecase.P2PTransferUseCase, limits *usecase.GetLimitsUseCase) *Server {
	return &Server{p2p: p2p, limits: limits}
}

func (s *Server) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: "ok", Service: "payment"}, nil
}

func (s *Server) CreateP2PTransfer(ctx context.Context, req api.CreateP2PTransferRequestObject) (api.CreateP2PTransferResponseObject, error) {
	if req.Body == nil {
		return api.CreateP2PTransfer400JSONResponse{Error: "invalid_json"}, nil
	}

	currency := "USD"
	if req.Body.Currency != nil {
		currency = *req.Body.Currency
	}
	memo := ""
	if req.Body.Memo != nil {
		memo = *req.Body.Memo
	}

	in := usecase.P2PTransferInput{
		SenderUserID:   req.Params.XUserId.String(),
		Amount:         req.Body.Amount,
		Currency:       currency,
		Memo:           memo,
		IdempotencyKey: req.Params.IdempotencyKey,
	}
	if req.Body.RecipientPhone != nil {
		in.RecipientPhone = *req.Body.RecipientPhone
	}
	if req.Body.RecipientEmail != nil {
		in.RecipientEmail = string(*req.Body.RecipientEmail)
	}
	if req.Body.RecipientUserId != nil {
		in.RecipientUserID = req.Body.RecipientUserId.String()
	}

	transfer, err := s.p2p.Execute(ctx, in)
	if err != nil {
		return api.CreateP2PTransfer400JSONResponse{Error: err.Error()}, nil
	}

	view := toTransfer(*transfer)
	switch transfer.Status {
	case domain.TransferStatusFailed:
		return api.CreateP2PTransfer422JSONResponse(view), nil
	case domain.TransferStatusCompleted:
		return api.CreateP2PTransfer200JSONResponse(view), nil
	default:
		return api.CreateP2PTransfer201JSONResponse(view), nil
	}
}

func (s *Server) GetTransfer(ctx context.Context, req api.GetTransferRequestObject) (api.GetTransferResponseObject, error) {
	transfer, err := s.p2p.GetByID(ctx, req.Id.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetTransfer404JSONResponse{Error: "transfer_not_found"}, nil
		}
		return nil, err
	}
	userID := req.Params.XUserId.String()
	if transfer.SenderUserID != userID && transfer.RecipientUserID != userID {
		return api.GetTransfer404JSONResponse{Error: "transfer_not_found"}, nil
	}
	view := toTransfer(*transfer)
	return api.GetTransfer200JSONResponse(view), nil
}

func (s *Server) ListTransfers(ctx context.Context, req api.ListTransfersRequestObject) (api.ListTransfersResponseObject, error) {
	userID := req.Params.XUserId.String()
	if req.Params.UserId != nil {
		userID = req.Params.UserId.String()
	}
	limit := 20
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	cursor := ""
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}

	result, err := s.p2p.List(ctx, userID, limit, cursor)
	if err != nil {
		return nil, err
	}

	views := make([]api.Transfer, 0, len(result.Transfers))
	for _, t := range result.Transfers {
		views = append(views, toTransfer(t))
	}
	resp := api.ListTransfers200JSONResponse{Transfers: views}
	if result.NextCursor != "" {
		resp.NextCursor = &result.NextCursor
	}
	return resp, nil
}

func (s *Server) GetLimits(ctx context.Context, req api.GetLimitsRequestObject) (api.GetLimitsResponseObject, error) {
	limits, err := s.limits.Execute(ctx, req.Params.XUserId.String())
	if err != nil {
		return nil, err
	}
	return api.GetLimits200JSONResponse{
		P2p: api.TransferLimits{
			HourlyTransferCount: api.LimitGauge{
				Limit:     limits.P2P.HourlyTransferCount.Limit,
				Used:      limits.P2P.HourlyTransferCount.Used,
				Remaining: limits.P2P.HourlyTransferCount.Remaining,
			},
			DailyTransferAmount: api.LimitGauge{
				Limit:     limits.P2P.DailyTransferAmount.Limit,
				Used:      limits.P2P.DailyTransferAmount.Used,
				Remaining: limits.P2P.DailyTransferAmount.Remaining,
			},
			SingleTransferMax: limits.P2P.SingleTransferMax,
		},
	}, nil
}

func toTransfer(t domain.Transfer) api.Transfer {
	senderID, _ := uuid.Parse(t.SenderUserID)
	recipientID, _ := uuid.Parse(t.RecipientUserID)
	id, _ := uuid.Parse(t.ID)

	out := api.Transfer{
		Id:              openapi_types.UUID(id),
		Status:          string(t.Status),
		SenderUserId:    openapi_types.UUID(senderID),
		RecipientUserId: openapi_types.UUID(recipientID),
		Amount:          t.Amount,
		Currency:        t.Currency,
	}
	if t.LedgerTransferID != "" {
		out.LedgerTransferId = &t.LedgerTransferID
	}
	if t.FailureReason != "" {
		out.FailureReason = &t.FailureReason
	}
	if t.Memo != "" {
		out.Memo = &t.Memo
	}
	if !t.CreatedAt.IsZero() {
		createdAt := t.CreatedAt.UTC()
		out.CreatedAt = &createdAt
	}
	if t.CompletedAt != nil {
		completedAt := t.CompletedAt.UTC()
		out.CompletedAt = &completedAt
	}
	return out
}