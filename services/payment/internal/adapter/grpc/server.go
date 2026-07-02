package grpc

import (
	"context"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	apiadapter "github.com/iho/neobank/services/payment/internal/adapter/api"
	"github.com/iho/neobank/services/payment/internal/gen/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	neobankv1.UnimplementedPaymentServiceServer

	api *apiadapter.Server
}

func NewServer(apiServer *apiadapter.Server) *Server {
	return &Server{api: apiServer}
}

func (s *Server) CreateP2PTransfer(ctx context.Context, req *neobankv1.CreateP2PTransferRequest) (*neobankv1.TransferResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		idempotencyKey = grpcutil.IdempotencyKeyFromContext(ctx)
	}

	body := &api.CreateP2PTransferJSONRequestBody{Amount: req.GetAmount()}
	if currency := req.GetCurrency(); currency != "" {
		body.Currency = &currency
	}
	if memo := req.GetMemo(); memo != "" {
		body.Memo = &memo
	}
	if phone := req.GetRecipientPhone(); phone != "" {
		body.RecipientPhone = &phone
	}
	if email := req.GetRecipientEmail(); email != "" {
		addr := openapi_types.Email(email)
		body.RecipientEmail = &addr
	}
	if recipientUserID := req.GetRecipientUserId(); recipientUserID != "" {
		recipientID, err := parseUUID(recipientUserID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid recipient_user_id: %v", err)
		}
		body.RecipientUserId = &recipientID
	}

	resp, err := s.api.CreateP2PTransfer(ctx, api.CreateP2PTransferRequestObject{
		Params: api.CreateP2PTransferParams{
			XUserId:        userID,
			IdempotencyKey: idempotencyKey,
		},
		Body: body,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapCreateP2PTransferResponse(resp)
}

func (s *Server) GetTransfer(ctx context.Context, req *neobankv1.GetTransferRequest) (*neobankv1.TransferResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	transferID, err := parseUUID(req.GetTransferId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid transfer_id: %v", err)
	}

	resp, err := s.api.GetTransfer(ctx, api.GetTransferRequestObject{
		Id: transferID,
		Params: api.GetTransferParams{
			XUserId: userID,
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapGetTransferResponse(resp)
}

func (s *Server) ListTransfers(ctx context.Context, req *neobankv1.ListTransfersRequest) (*neobankv1.ListTransfersResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	params := api.ListTransfersParams{XUserId: userID}
	if limit := req.GetLimit(); limit > 0 {
		l := int(limit)
		params.Limit = &l
	}
	if cursor := req.GetCursor(); cursor != "" {
		params.Cursor = &cursor
	}

	resp, err := s.api.ListTransfers(ctx, api.ListTransfersRequestObject{Params: params})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapListTransfersResponse(resp)
}

func (s *Server) GetLimits(ctx context.Context, req *neobankv1.GetLimitsRequest) (*neobankv1.LimitsResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	resp, err := s.api.GetLimits(ctx, api.GetLimitsRequestObject{
		Params: api.GetLimitsParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapGetLimitsResponse(resp)
}

func parseUUID(id string) (openapi_types.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return openapi_types.UUID{}, err
	}
	return openapi_types.UUID(parsed), nil
}