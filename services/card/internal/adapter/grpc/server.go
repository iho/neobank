package grpc

import (
	"context"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	apiadapter "github.com/iho/neobank/services/card/internal/adapter/api"
	"github.com/iho/neobank/services/card/internal/gen/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	neobankv1.UnimplementedCardServiceServer

	api *apiadapter.Server
}

func NewServer(apiServer *apiadapter.Server) *Server {
	return &Server{api: apiServer}
}

func (s *Server) IssueCard(ctx context.Context, req *neobankv1.IssueCardRequest) (*neobankv1.CardResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		idempotencyKey = grpcutil.IdempotencyKeyFromContext(ctx)
	}

	onlineOnly := req.GetOnlineOnly()
	body := &api.IssueCardJSONRequestBody{
		CardholderName: req.GetCardholderName(),
		OnlineOnly:     &onlineOnly,
	}
	if dailyLimit := req.GetDailyLimit(); dailyLimit != "" {
		body.DailyLimit = &dailyLimit
	}
	if walletID := req.GetWalletId(); walletID != "" {
		walletUUID, err := parseUUID(walletID)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid wallet_id: %v", err)
		}
		body.WalletId = &walletUUID
	}

	resp, err := s.api.IssueCard(ctx, api.IssueCardRequestObject{
		Params: api.IssueCardParams{
			XUserId:        userID,
			IdempotencyKey: idempotencyKey,
		},
		Body: body,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapCardResponse(resp)
}

func (s *Server) ListCards(ctx context.Context, req *neobankv1.ListCardsRequest) (*neobankv1.ListCardsResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	resp, err := s.api.ListCards(ctx, api.ListCardsRequestObject{
		Params: api.ListCardsParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapListCardsResponse(resp)
}

func (s *Server) GetCard(ctx context.Context, req *neobankv1.GetCardRequest) (*neobankv1.CardResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	cardID, err := parseUUID(req.GetCardId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid card_id: %v", err)
	}

	resp, err := s.api.GetCard(ctx, api.GetCardRequestObject{
		Id:     cardID,
		Params: api.GetCardParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapGetCardResponse(resp)
}

func (s *Server) UpdateCardControls(ctx context.Context, req *neobankv1.UpdateCardControlsRequest) (*neobankv1.CardResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	cardID, err := parseUUID(req.GetCardId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid card_id: %v", err)
	}

	resp, err := s.api.UpdateCardControls(ctx, api.UpdateCardControlsRequestObject{
		Id: cardID,
		Params: api.UpdateCardControlsParams{
			XUserId: userID,
		},
		Body: &api.UpdateCardControlsJSONRequestBody{
			DailyLimit: req.DailyLimit,
			OnlineOnly: req.OnlineOnly,
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapUpdateCardControlsResponse(resp)
}

func (s *Server) FreezeCard(ctx context.Context, req *neobankv1.CardActionRequest) (*neobankv1.CardResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	cardID, err := parseUUID(req.GetCardId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid card_id: %v", err)
	}

	resp, err := s.api.FreezeCard(ctx, api.FreezeCardRequestObject{
		Id:     cardID,
		Params: api.FreezeCardParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapFreezeCardResponse(resp)
}

func (s *Server) UnfreezeCard(ctx context.Context, req *neobankv1.CardActionRequest) (*neobankv1.CardResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	cardID, err := parseUUID(req.GetCardId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid card_id: %v", err)
	}

	resp, err := s.api.UnfreezeCard(ctx, api.UnfreezeCardRequestObject{
		Id:     cardID,
		Params: api.UnfreezeCardParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapUnfreezeCardResponse(resp)
}

func (s *Server) AuthorizeTransaction(ctx context.Context, req *neobankv1.AuthorizeTransactionRequest) (*neobankv1.AuthorizationResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	cardID, err := parseUUID(req.GetCardId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid card_id: %v", err)
	}

	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		idempotencyKey = grpcutil.IdempotencyKeyFromContext(ctx)
	}

	body := &api.AuthorizeTransactionJSONRequestBody{Amount: req.GetAmount()}
	if currency := req.GetCurrency(); currency != "" {
		body.Currency = &currency
	}
	if merchant := req.GetMerchantName(); merchant != "" {
		body.MerchantName = &merchant
	}
	if mcc := req.GetMerchantCategoryCode(); mcc != "" {
		body.MerchantCategoryCode = &mcc
	}
	if channel := req.GetChannel(); channel != "" {
		ch := api.AuthorizeRequestChannel(channel)
		body.Channel = &ch
	}

	resp, err := s.api.AuthorizeTransaction(ctx, api.AuthorizeTransactionRequestObject{
		Id: cardID,
		Params: api.AuthorizeTransactionParams{
			XUserId:        userID,
			IdempotencyKey: idempotencyKey,
		},
		Body: body,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapAuthorizeTransactionResponse(resp)
}

func (s *Server) ListAuthorizations(ctx context.Context, req *neobankv1.ListAuthorizationsRequest) (*neobankv1.ListAuthorizationsResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	params := api.ListAuthorizationsParams{XUserId: userID}
	if limit := req.GetLimit(); limit > 0 {
		l := int(limit)
		params.Limit = &l
	}

	resp, err := s.api.ListAuthorizations(ctx, api.ListAuthorizationsRequestObject{Params: params})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapListAuthorizationsResponse(resp)
}

func (s *Server) GetAuthorization(ctx context.Context, req *neobankv1.GetAuthorizationRequest) (*neobankv1.AuthorizationResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	authID, err := parseUUID(req.GetAuthorizationId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid authorization_id: %v", err)
	}

	resp, err := s.api.GetAuthorization(ctx, api.GetAuthorizationRequestObject{
		Id:     authID,
		Params: api.GetAuthorizationParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapGetAuthorizationResponse(resp)
}

func (s *Server) CaptureAuthorization(ctx context.Context, req *neobankv1.CaptureAuthorizationRequest) (*neobankv1.AuthorizationResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	authID, err := parseUUID(req.GetAuthorizationId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid authorization_id: %v", err)
	}

	resp, err := s.api.CaptureAuthorization(ctx, api.CaptureAuthorizationRequestObject{
		Id:     authID,
		Params: api.CaptureAuthorizationParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	return mapCaptureAuthorizationResponse(resp)
}

func parseUserID(userID string) (openapi_types.UUID, error) {
	return parseUUID(userID)
}

func parseUUID(id string) (openapi_types.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return openapi_types.UUID{}, err
	}
	return openapi_types.UUID(parsed), nil
}