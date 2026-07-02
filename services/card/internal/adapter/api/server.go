package api

import (
	"context"
	"errors"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/iho/neobank/pkg/saga"
	"github.com/iho/neobank/services/card/internal/domain"
	"github.com/iho/neobank/services/card/internal/gen/api"
	"github.com/iho/neobank/services/card/internal/usecase"
	"github.com/jackc/pgx/v5"
)

type Server struct {
	issue      *usecase.IssueCardUseCase
	freeze     *usecase.FreezeCardUseCase
	unfreeze   *usecase.UnfreezeCardUseCase
	authorize  *usecase.AuthorizeTransactionUseCase
	capture    *usecase.CaptureAuthorizationUseCase
	listAuths  *usecase.ListAuthorizationsUseCase
}

func NewServer(
	issue *usecase.IssueCardUseCase,
	freeze *usecase.FreezeCardUseCase,
	unfreeze *usecase.UnfreezeCardUseCase,
	authorize *usecase.AuthorizeTransactionUseCase,
	capture *usecase.CaptureAuthorizationUseCase,
	listAuths *usecase.ListAuthorizationsUseCase,
) *Server {
	return &Server{
		issue: issue, freeze: freeze, unfreeze: unfreeze,
		authorize: authorize, capture: capture, listAuths: listAuths,
	}
}

func (s *Server) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: "ok", Service: "card"}, nil
}

func (s *Server) IssueCard(ctx context.Context, req api.IssueCardRequestObject) (api.IssueCardResponseObject, error) {
	if req.Body == nil {
		return api.IssueCard400JSONResponse{Error: "invalid_json"}, nil
	}

	walletID := ""
	if req.Body.WalletId != nil {
		walletID = req.Body.WalletId.String()
	}

	out, err := s.issue.Execute(ctx, usecase.IssueCardInput{
		UserID:         req.Params.XUserId.String(),
		WalletID:       walletID,
		CardholderName: req.Body.CardholderName,
		IdempotencyKey: req.Params.IdempotencyKey,
	})
	if err != nil {
		if _, ok := err.(*saga.BusinessError); ok {
			return api.IssueCard422JSONResponse{Error: err.Error()}, nil
		}
		return api.IssueCard400JSONResponse{Error: err.Error()}, nil
	}

	view := toCard(*out.Card)
	if out.Replayed {
		return api.IssueCard200JSONResponse(view), nil
	}
	return api.IssueCard201JSONResponse(view), nil
}

func (s *Server) ListCards(ctx context.Context, req api.ListCardsRequestObject) (api.ListCardsResponseObject, error) {
	cards, err := s.issue.List(ctx, req.Params.XUserId.String())
	if err != nil {
		return nil, err
	}
	views := make([]api.Card, 0, len(cards))
	for _, c := range cards {
		views = append(views, toCard(c))
	}
	return api.ListCards200JSONResponse{Cards: views}, nil
}

func (s *Server) GetCard(ctx context.Context, req api.GetCardRequestObject) (api.GetCardResponseObject, error) {
	card, err := s.issue.GetByID(ctx, req.Id.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetCard404JSONResponse{Error: "card_not_found"}, nil
		}
		return nil, err
	}
	if card.UserID != req.Params.XUserId.String() {
		return api.GetCard404JSONResponse{Error: "card_not_found"}, nil
	}
	return api.GetCard200JSONResponse(toCard(*card)), nil
}

func (s *Server) FreezeCard(ctx context.Context, req api.FreezeCardRequestObject) (api.FreezeCardResponseObject, error) {
	card, err := s.freeze.Execute(ctx, req.Params.XUserId.String(), req.Id.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "card not found" {
			return api.FreezeCard404JSONResponse{Error: "card_not_found"}, nil
		}
		return api.FreezeCard404JSONResponse{Error: err.Error()}, nil
	}
	return api.FreezeCard200JSONResponse(toCard(*card)), nil
}

func (s *Server) UnfreezeCard(ctx context.Context, req api.UnfreezeCardRequestObject) (api.UnfreezeCardResponseObject, error) {
	card, err := s.unfreeze.Execute(ctx, req.Params.XUserId.String(), req.Id.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || err.Error() == "card not found" {
			return api.UnfreezeCard404JSONResponse{Error: "card_not_found"}, nil
		}
		return api.UnfreezeCard404JSONResponse{Error: err.Error()}, nil
	}
	return api.UnfreezeCard200JSONResponse(toCard(*card)), nil
}

func (s *Server) AuthorizeTransaction(ctx context.Context, req api.AuthorizeTransactionRequestObject) (api.AuthorizeTransactionResponseObject, error) {
	if req.Body == nil {
		return api.AuthorizeTransaction400JSONResponse{Error: "invalid_json"}, nil
	}

	currency := "USD"
	if req.Body.Currency != nil {
		currency = *req.Body.Currency
	}
	merchant := ""
	if req.Body.MerchantName != nil {
		merchant = *req.Body.MerchantName
	}

	out, err := s.authorize.Execute(ctx, usecase.AuthorizeTransactionInput{
		UserID:         req.Params.XUserId.String(),
		CardID:         req.Id.String(),
		Amount:         req.Body.Amount,
		Currency:       currency,
		MerchantName:   merchant,
		IdempotencyKey: req.Params.IdempotencyKey,
	})
	if err != nil {
		return api.AuthorizeTransaction400JSONResponse{Error: err.Error()}, nil
	}

	view := toAuthorization(*out.Authorization)
	if out.Authorization.Status == domain.AuthStatusDeclined {
		return api.AuthorizeTransaction422JSONResponse(view), nil
	}
	if out.Replayed {
		return api.AuthorizeTransaction200JSONResponse(view), nil
	}
	return api.AuthorizeTransaction201JSONResponse(view), nil
}

func (s *Server) ListAuthorizations(ctx context.Context, req api.ListAuthorizationsRequestObject) (api.ListAuthorizationsResponseObject, error) {
	limit := 20
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	auths, err := s.listAuths.Execute(ctx, req.Params.XUserId.String(), limit)
	if err != nil {
		return nil, err
	}
	views := make([]api.Authorization, 0, len(auths))
	for _, a := range auths {
		views = append(views, toAuthorization(a))
	}
	return api.ListAuthorizations200JSONResponse{Authorizations: views}, nil
}

func (s *Server) GetAuthorization(ctx context.Context, req api.GetAuthorizationRequestObject) (api.GetAuthorizationResponseObject, error) {
	auth, err := s.authorize.GetByID(ctx, req.Id.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetAuthorization404JSONResponse{Error: "authorization_not_found"}, nil
		}
		return nil, err
	}
	if auth.UserID != req.Params.XUserId.String() {
		return api.GetAuthorization404JSONResponse{Error: "authorization_not_found"}, nil
	}
	return api.GetAuthorization200JSONResponse(toAuthorization(*auth)), nil
}

func (s *Server) CaptureAuthorization(ctx context.Context, req api.CaptureAuthorizationRequestObject) (api.CaptureAuthorizationResponseObject, error) {
	auth, err := s.capture.Execute(ctx, usecase.CaptureAuthorizationInput{
		UserID:          req.Params.XUserId.String(),
		AuthorizationID: req.Id.String(),
	})
	if err != nil {
		if err.Error() == "authorization not found" {
			return api.CaptureAuthorization404JSONResponse{Error: "authorization_not_found"}, nil
		}
		return api.CaptureAuthorization400JSONResponse{Error: err.Error()}, nil
	}
	return api.CaptureAuthorization200JSONResponse(toAuthorization(*auth)), nil
}

func toCard(c domain.Card) api.Card {
	id, _ := uuid.Parse(c.ID)
	userID, _ := uuid.Parse(c.UserID)
	walletID, _ := uuid.Parse(c.WalletID)
	return api.Card{
		Id:          openapi_types.UUID(id),
		UserId:      openapi_types.UUID(userID),
		WalletId:    openapi_types.UUID(walletID),
		LastFour:    c.LastFour,
		Status:      string(c.Status),
		ExpiryMonth: c.ExpiryMonth,
		ExpiryYear:  c.ExpiryYear,
	}
}

func toAuthorization(a domain.Authorization) api.Authorization {
	id, _ := uuid.Parse(a.ID)
	cardID, _ := uuid.Parse(a.CardID)
	userID, _ := uuid.Parse(a.UserID)
	out := api.Authorization{
		Id:       openapi_types.UUID(id),
		CardId:   openapi_types.UUID(cardID),
		UserId:   openapi_types.UUID(userID),
		Amount:   a.Amount,
		Currency: a.Currency,
		Status:   string(a.Status),
	}
	if a.MerchantName != "" {
		out.MerchantName = &a.MerchantName
	}
	if a.LedgerHoldID != "" {
		out.LedgerHoldId = &a.LedgerHoldID
	}
	if a.LedgerTransferID != "" {
		out.LedgerTransferId = &a.LedgerTransferID
	}
	if a.FailureReason != "" {
		out.FailureReason = &a.FailureReason
	}
	if !a.CreatedAt.IsZero() {
		createdAt := a.CreatedAt.UTC()
		out.CreatedAt = &createdAt
	}
	if a.CapturedAt != nil {
		capturedAt := a.CapturedAt.UTC()
		out.CapturedAt = &capturedAt
	}
	return out
}