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
	p2p *usecase.P2PTransferUseCase
}

func NewServer(p2p *usecase.P2PTransferUseCase) *Server {
	return &Server{p2p: p2p}
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

	transfer, err := s.p2p.Execute(ctx, usecase.P2PTransferInput{
		SenderUserID:   req.Params.XUserId.String(),
		RecipientPhone: req.Body.RecipientPhone,
		Amount:         req.Body.Amount,
		Currency:       currency,
		Memo:           memo,
		IdempotencyKey: req.Params.IdempotencyKey,
	})
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

	transfers, err := s.p2p.List(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	views := make([]api.Transfer, 0, len(transfers))
	for _, t := range transfers {
		views = append(views, toTransfer(t))
	}
	return api.ListTransfers200JSONResponse{Transfers: views}, nil
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