package grpc

import (
	"time"

	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/services/card/internal/gen/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func timeRFC3339(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func cardToProto(c api.Card) *neobankv1.Card {
	out := &neobankv1.Card{
		Id:          c.Id.String(),
		UserId:      c.UserId.String(),
		WalletId:    c.WalletId.String(),
		LastFour:    c.LastFour,
		Status:      c.Status,
		ExpiryMonth: int32(c.ExpiryMonth),
		ExpiryYear:  int32(c.ExpiryYear),
		OnlineOnly:  c.OnlineOnly,
	}
	if c.DailyLimit != nil {
		out.DailyLimit = *c.DailyLimit
	}
	return out
}

func authorizationToProto(a api.Authorization) *neobankv1.Authorization {
	out := &neobankv1.Authorization{
		Id:       a.Id.String(),
		CardId:   a.CardId.String(),
		UserId:   a.UserId.String(),
		Amount:   a.Amount,
		Currency: a.Currency,
		Status:   a.Status,
	}
	if a.MerchantName != nil {
		out.MerchantName = *a.MerchantName
	}
	if a.MerchantCategoryCode != nil {
		out.MerchantCategoryCode = *a.MerchantCategoryCode
	}
	if a.LedgerHoldId != nil {
		out.LedgerHoldId = *a.LedgerHoldId
	}
	if a.LedgerTransferId != nil {
		out.LedgerTransferId = *a.LedgerTransferId
	}
	if a.FailureReason != nil {
		out.FailureReason = *a.FailureReason
	}
	if a.CreatedAt != nil {
		out.CreatedAt = timeRFC3339(a.CreatedAt)
	}
	if a.CapturedAt != nil {
		out.CapturedAt = timeRFC3339(a.CapturedAt)
	}
	return out
}

func mapCardResponse(resp api.IssueCardResponseObject) (*neobankv1.CardResponse, error) {
	switch r := resp.(type) {
	case api.IssueCard200JSONResponse:
		return &neobankv1.CardResponse{Card: cardToProto(api.Card(r)), HttpStatus: 200}, nil
	case api.IssueCard201JSONResponse:
		return &neobankv1.CardResponse{Card: cardToProto(api.Card(r)), HttpStatus: 201}, nil
	case api.IssueCard400JSONResponse:
		return &neobankv1.CardResponse{HttpStatus: 400, Error: r.Error}, nil
	case api.IssueCard422JSONResponse:
		return &neobankv1.CardResponse{HttpStatus: 422, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected issue card response")
	}
}

func mapGetCardResponse(resp api.GetCardResponseObject) (*neobankv1.CardResponse, error) {
	switch r := resp.(type) {
	case api.GetCard200JSONResponse:
		return &neobankv1.CardResponse{Card: cardToProto(api.Card(r)), HttpStatus: 200}, nil
	case api.GetCard404JSONResponse:
		return &neobankv1.CardResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected get card response")
	}
}

func mapUpdateCardControlsResponse(resp api.UpdateCardControlsResponseObject) (*neobankv1.CardResponse, error) {
	switch r := resp.(type) {
	case api.UpdateCardControls200JSONResponse:
		return &neobankv1.CardResponse{Card: cardToProto(api.Card(r)), HttpStatus: 200}, nil
	case api.UpdateCardControls400JSONResponse:
		return &neobankv1.CardResponse{HttpStatus: 400, Error: r.Error}, nil
	case api.UpdateCardControls404JSONResponse:
		return &neobankv1.CardResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected update card controls response")
	}
}

func mapFreezeCardResponse(resp api.FreezeCardResponseObject) (*neobankv1.CardResponse, error) {
	switch r := resp.(type) {
	case api.FreezeCard200JSONResponse:
		return &neobankv1.CardResponse{Card: cardToProto(api.Card(r)), HttpStatus: 200}, nil
	case api.FreezeCard404JSONResponse:
		return &neobankv1.CardResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected freeze card response")
	}
}

func mapUnfreezeCardResponse(resp api.UnfreezeCardResponseObject) (*neobankv1.CardResponse, error) {
	switch r := resp.(type) {
	case api.UnfreezeCard200JSONResponse:
		return &neobankv1.CardResponse{Card: cardToProto(api.Card(r)), HttpStatus: 200}, nil
	case api.UnfreezeCard404JSONResponse:
		return &neobankv1.CardResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected unfreeze card response")
	}
}

func mapListCardsResponse(resp api.ListCardsResponseObject) (*neobankv1.ListCardsResponse, error) {
	switch r := resp.(type) {
	case api.ListCards200JSONResponse:
		cards := make([]*neobankv1.Card, 0, len(r.Cards))
		for _, c := range r.Cards {
			cards = append(cards, cardToProto(c))
		}
		return &neobankv1.ListCardsResponse{Cards: cards, HttpStatus: 200}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected list cards response")
	}
}

func mapAuthorizeTransactionResponse(resp api.AuthorizeTransactionResponseObject) (*neobankv1.AuthorizationResponse, error) {
	switch r := resp.(type) {
	case api.AuthorizeTransaction200JSONResponse:
		return &neobankv1.AuthorizationResponse{
			Authorization: authorizationToProto(api.Authorization(r)),
			HttpStatus:    200,
		}, nil
	case api.AuthorizeTransaction201JSONResponse:
		return &neobankv1.AuthorizationResponse{
			Authorization: authorizationToProto(api.Authorization(r)),
			HttpStatus:    201,
		}, nil
	case api.AuthorizeTransaction400JSONResponse:
		return &neobankv1.AuthorizationResponse{HttpStatus: 400, Error: r.Error}, nil
	case api.AuthorizeTransaction404JSONResponse:
		return &neobankv1.AuthorizationResponse{HttpStatus: 404, Error: r.Error}, nil
	case api.AuthorizeTransaction422JSONResponse:
		return &neobankv1.AuthorizationResponse{
			Authorization: authorizationToProto(api.Authorization(r)),
			HttpStatus:    422,
		}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected authorize transaction response")
	}
}

func mapGetAuthorizationResponse(resp api.GetAuthorizationResponseObject) (*neobankv1.AuthorizationResponse, error) {
	switch r := resp.(type) {
	case api.GetAuthorization200JSONResponse:
		return &neobankv1.AuthorizationResponse{
			Authorization: authorizationToProto(api.Authorization(r)),
			HttpStatus:    200,
		}, nil
	case api.GetAuthorization404JSONResponse:
		return &neobankv1.AuthorizationResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected get authorization response")
	}
}

func mapCaptureAuthorizationResponse(resp api.CaptureAuthorizationResponseObject) (*neobankv1.AuthorizationResponse, error) {
	switch r := resp.(type) {
	case api.CaptureAuthorization200JSONResponse:
		return &neobankv1.AuthorizationResponse{
			Authorization: authorizationToProto(api.Authorization(r)),
			HttpStatus:    200,
		}, nil
	case api.CaptureAuthorization400JSONResponse:
		return &neobankv1.AuthorizationResponse{HttpStatus: 400, Error: r.Error}, nil
	case api.CaptureAuthorization404JSONResponse:
		return &neobankv1.AuthorizationResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected capture authorization response")
	}
}

func mapListAuthorizationsResponse(resp api.ListAuthorizationsResponseObject) (*neobankv1.ListAuthorizationsResponse, error) {
	switch r := resp.(type) {
	case api.ListAuthorizations200JSONResponse:
		auths := make([]*neobankv1.Authorization, 0, len(r.Authorizations))
		for _, a := range r.Authorizations {
			auths = append(auths, authorizationToProto(a))
		}
		return &neobankv1.ListAuthorizationsResponse{
			Authorizations: auths,
			HttpStatus:     200,
		}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected list authorizations response")
	}
}