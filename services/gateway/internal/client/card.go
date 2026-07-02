package client

import (
	"context"
	"fmt"

	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	"google.golang.org/grpc"
)

type CardClient struct {
	conn *grpc.ClientConn
	rpc  neobankv1.CardServiceClient
}

func NewCardClient(ctx context.Context, cfg Config) (*CardClient, error) {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:50054"
	}
	conn, err := grpcutil.Dial(ctx, cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("dial card service: %w", err)
	}
	return &CardClient{
		conn: conn,
		rpc:  neobankv1.NewCardServiceClient(conn),
	}, nil
}

func (c *CardClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

type IssueCardRequest struct {
	WalletID       string `json:"wallet_id,omitempty"`
	CardholderName string `json:"cardholder_name"`
	DailyLimit     string `json:"daily_limit,omitempty"`
	OnlineOnly     bool   `json:"online_only,omitempty"`
}

type UpdateCardControlsRequest struct {
	DailyLimit *string `json:"daily_limit,omitempty"`
	OnlineOnly *bool   `json:"online_only,omitempty"`
}

type CardView struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	WalletID    string `json:"wallet_id"`
	LastFour    string `json:"last_four"`
	Status      string `json:"status"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	DailyLimit  string `json:"daily_limit,omitempty"`
	OnlineOnly  bool   `json:"online_only"`
}

type CardList struct {
	Cards []CardView `json:"cards"`
}

func (c *CardClient) IssueCard(ctx context.Context, userID, idempotencyKey string, req IssueCardRequest) (CardView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	resp, err := c.rpc.IssueCard(ctx, &neobankv1.IssueCardRequest{
		UserId:         userID,
		WalletId:       req.WalletID,
		CardholderName: req.CardholderName,
		DailyLimit:     req.DailyLimit,
		OnlineOnly:     req.OnlineOnly,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return CardView{}, 0, dialError("card", err)
	}
	status := int(resp.GetHttpStatus())
	if status >= 400 {
		return CardView{}, status, statusError("card", status, resp.GetError())
	}
	return toCardView(resp.GetCard()), status, nil
}

func (c *CardClient) ListCards(ctx context.Context, userID string) (CardList, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.ListCards(ctx, &neobankv1.ListCardsRequest{UserId: userID})
	if err != nil {
		return CardList{}, dialError("card", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return CardList{}, statusError("card", status, resp.GetError())
	}
	out := CardList{}
	for _, card := range resp.GetCards() {
		out.Cards = append(out.Cards, toCardView(card))
	}
	return out, nil
}

func (c *CardClient) GetCard(ctx context.Context, userID, cardID string) (CardView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.GetCard(ctx, &neobankv1.GetCardRequest{
		UserId: userID,
		CardId: cardID,
	})
	if err != nil {
		return CardView{}, 0, dialError("card", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return CardView{}, status, statusError("card", status, resp.GetError())
	}
	return toCardView(resp.GetCard()), status, nil
}

func (c *CardClient) FreezeCard(ctx context.Context, userID, cardID string) (CardView, int, error) {
	return c.cardAction(ctx, userID, cardID, c.rpc.FreezeCard)
}

func (c *CardClient) UnfreezeCard(ctx context.Context, userID, cardID string) (CardView, int, error) {
	return c.cardAction(ctx, userID, cardID, c.rpc.UnfreezeCard)
}

func (c *CardClient) UpdateCardControls(ctx context.Context, userID, cardID string, req UpdateCardControlsRequest) (CardView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	protoReq := &neobankv1.UpdateCardControlsRequest{
		UserId: userID,
		CardId: cardID,
	}
	if req.DailyLimit != nil {
		protoReq.DailyLimit = req.DailyLimit
	}
	if req.OnlineOnly != nil {
		protoReq.OnlineOnly = req.OnlineOnly
	}
	resp, err := c.rpc.UpdateCardControls(ctx, protoReq)
	if err != nil {
		return CardView{}, 0, dialError("card", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return CardView{}, status, statusError("card", status, resp.GetError())
	}
	return toCardView(resp.GetCard()), status, nil
}

type AuthorizeRequest struct {
	Amount               string `json:"amount"`
	Currency             string `json:"currency,omitempty"`
	MerchantName         string `json:"merchant_name,omitempty"`
	MerchantCategoryCode string `json:"merchant_category_code,omitempty"`
	Channel              string `json:"channel,omitempty"`
}

type AuthorizationView struct {
	ID                   string `json:"id"`
	CardID               string `json:"card_id"`
	UserID               string `json:"user_id"`
	MerchantName         string `json:"merchant_name,omitempty"`
	MerchantCategoryCode string `json:"merchant_category_code,omitempty"`
	Amount               string `json:"amount"`
	Currency             string `json:"currency"`
	Status               string `json:"status"`
	LedgerHoldID         string `json:"ledger_hold_id,omitempty"`
	LedgerTransferID     string `json:"ledger_transfer_id,omitempty"`
	FailureReason        string `json:"failure_reason,omitempty"`
	CreatedAt            string `json:"created_at,omitempty"`
	CapturedAt           string `json:"captured_at,omitempty"`
}

type AuthorizationList struct {
	Authorizations []AuthorizationView `json:"authorizations"`
}

func (c *CardClient) AuthorizeTransaction(ctx context.Context, userID, cardID, idempotencyKey string, req AuthorizeRequest) (AuthorizationView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	resp, err := c.rpc.AuthorizeTransaction(ctx, &neobankv1.AuthorizeTransactionRequest{
		UserId:               userID,
		CardId:               cardID,
		Amount:               req.Amount,
		Currency:             req.Currency,
		MerchantName:         req.MerchantName,
		MerchantCategoryCode: req.MerchantCategoryCode,
		Channel:              req.Channel,
		IdempotencyKey:       idempotencyKey,
	})
	if err != nil {
		return AuthorizationView{}, 0, dialError("card", err)
	}
	status := int(resp.GetHttpStatus())
	out := toAuthorizationView(resp.GetAuthorization())
	if status >= 400 {
		return out, status, statusError("card", status, "")
	}
	return out, status, nil
}

func (c *CardClient) ListAuthorizations(ctx context.Context, userID string, limit int) (AuthorizationList, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.ListAuthorizations(ctx, &neobankv1.ListAuthorizationsRequest{
		UserId: userID,
		Limit:  int32(limit),
	})
	if err != nil {
		return AuthorizationList{}, dialError("card", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return AuthorizationList{}, statusError("card", status, resp.GetError())
	}
	out := AuthorizationList{}
	for _, auth := range resp.GetAuthorizations() {
		out.Authorizations = append(out.Authorizations, toAuthorizationView(auth))
	}
	return out, nil
}

func (c *CardClient) GetAuthorization(ctx context.Context, userID, authID string) (AuthorizationView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.GetAuthorization(ctx, &neobankv1.GetAuthorizationRequest{
		UserId:          userID,
		AuthorizationId: authID,
	})
	if err != nil {
		return AuthorizationView{}, 0, dialError("card", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return AuthorizationView{}, status, statusError("card", status, resp.GetError())
	}
	return toAuthorizationView(resp.GetAuthorization()), status, nil
}

func (c *CardClient) CaptureAuthorization(ctx context.Context, userID, authID string) (AuthorizationView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.CaptureAuthorization(ctx, &neobankv1.CaptureAuthorizationRequest{
		UserId:          userID,
		AuthorizationId: authID,
	})
	if err != nil {
		return AuthorizationView{}, 0, dialError("card", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return AuthorizationView{}, status, statusError("card", status, resp.GetError())
	}
	return toAuthorizationView(resp.GetAuthorization()), status, nil
}

type cardActionFunc func(context.Context, *neobankv1.CardActionRequest, ...grpc.CallOption) (*neobankv1.CardResponse, error)

func (c *CardClient) cardAction(ctx context.Context, userID, cardID string, fn cardActionFunc) (CardView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := fn(ctx, &neobankv1.CardActionRequest{
		UserId: userID,
		CardId: cardID,
	})
	if err != nil {
		return CardView{}, 0, dialError("card", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return CardView{}, status, statusError("card", status, resp.GetError())
	}
	return toCardView(resp.GetCard()), status, nil
}

func toCardView(card *neobankv1.Card) CardView {
	if card == nil {
		return CardView{}
	}
	return CardView{
		ID:          card.GetId(),
		UserID:      card.GetUserId(),
		WalletID:    card.GetWalletId(),
		LastFour:    card.GetLastFour(),
		Status:      card.GetStatus(),
		ExpiryMonth: int(card.GetExpiryMonth()),
		ExpiryYear:  int(card.GetExpiryYear()),
		DailyLimit:  card.GetDailyLimit(),
		OnlineOnly:  card.GetOnlineOnly(),
	}
}

func toAuthorizationView(auth *neobankv1.Authorization) AuthorizationView {
	if auth == nil {
		return AuthorizationView{}
	}
	return AuthorizationView{
		ID:                   auth.GetId(),
		CardID:               auth.GetCardId(),
		UserID:               auth.GetUserId(),
		MerchantName:         auth.GetMerchantName(),
		MerchantCategoryCode: auth.GetMerchantCategoryCode(),
		Amount:               auth.GetAmount(),
		Currency:             auth.GetCurrency(),
		Status:               auth.GetStatus(),
		LedgerHoldID:         auth.GetLedgerHoldId(),
		LedgerTransferID:     auth.GetLedgerTransferId(),
		FailureReason:        auth.GetFailureReason(),
		CreatedAt:            auth.GetCreatedAt(),
		CapturedAt:           auth.GetCapturedAt(),
	}
}