package api

import (
	"context"
	"strings"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/iho/neobank/pkg/auth"
	"github.com/iho/neobank/services/gateway/internal/client"
	"github.com/iho/neobank/services/gateway/internal/gen/api"
)

type Server struct {
	jwt           *auth.JWT
	users         *client.UserClient
	payments      *client.PaymentClient
	cards         *client.CardClient
	notifications *client.NotificationClient
}

func NewServer(
	jwtAuth *auth.JWT,
	users *client.UserClient,
	payments *client.PaymentClient,
	cards *client.CardClient,
	notifications *client.NotificationClient,
) *Server {
	return &Server{jwt: jwtAuth, users: users, payments: payments, cards: cards, notifications: notifications}
}

func (s *Server) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: "ok", Service: "gateway"}, nil
}

func (s *Server) Login(ctx context.Context, req api.LoginRequestObject) (api.LoginResponseObject, error) {
	if req.Body == nil {
		return api.Login401JSONResponse{Error: "invalid_json"}, nil
	}

	resp, err := s.users.Login(ctx, client.LoginRequest{
		Email:    string(req.Body.Email),
		Password: req.Body.Password,
	})
	if err != nil {
		return api.Login502JSONResponse{Error: err.Error()}, nil
	}

	return api.Login200JSONResponse{
		UserId:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (s *Server) RefreshToken(ctx context.Context, req api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	if req.Body == nil {
		return api.RefreshToken401JSONResponse{Error: "invalid_json"}, nil
	}

	resp, err := s.users.RefreshToken(ctx, req.Body.RefreshToken)
	if err != nil {
		return api.RefreshToken502JSONResponse{Error: err.Error()}, nil
	}

	return api.RefreshToken200JSONResponse{
		UserId:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (s *Server) Register(ctx context.Context, req api.RegisterRequestObject) (api.RegisterResponseObject, error) {
	if req.Body == nil {
		return api.Register400JSONResponse{Error: "invalid_json"}, nil
	}

	phone := ""
	if req.Body.Phone != nil {
		phone = *req.Body.Phone
	}

	resp, err := s.users.Register(ctx, req.Params.IdempotencyKey, client.RegisterRequest{
		Email:    string(req.Body.Email),
		Phone:    phone,
		Password: req.Body.Password,
	})
	if err != nil {
		return api.Register502JSONResponse{Error: err.Error()}, nil
	}

	return api.Register201JSONResponse{
		UserId:       resp.UserID,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (s *Server) SubmitKYC(ctx context.Context, req api.SubmitKYCRequestObject) (api.SubmitKYCResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.SubmitKYC401JSONResponse{Error: "unauthorized"}, nil
	}
	if req.Body == nil {
		return api.SubmitKYC400JSONResponse{Error: "invalid_json"}, nil
	}

	resp, err := s.users.SubmitKYC(ctx, userID, req.Params.IdempotencyKey, client.SubmitKYCRequest{
		FullName:    req.Body.FullName,
		DateOfBirth: req.Body.DateOfBirth.String(),
		CountryCode: req.Body.CountryCode,
	})
	if err != nil {
		return api.SubmitKYC502JSONResponse{Error: err.Error()}, nil
	}

	return api.SubmitKYC200JSONResponse{
		KycCaseId: resp.KYCCaseID,
		Status:    resp.Status,
		WalletId:  resp.WalletID,
	}, nil
}

func (s *Server) GetProfile(ctx context.Context, req api.GetProfileRequestObject) (api.GetProfileResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.GetProfile401JSONResponse{Error: "unauthorized"}, nil
	}

	profile, statusCode, err := s.users.GetProfile(ctx, userID)
	if statusCode == 404 {
		return api.GetProfile404JSONResponse{Error: "user_not_found"}, nil
	}
	if err != nil {
		return api.GetProfile502JSONResponse{Error: err.Error()}, nil
	}

	createdAt, _ := time.Parse(time.RFC3339, profile.CreatedAt)
	resp := api.Profile{
		UserId:    profile.UserID,
		Email:     profile.Email,
		Phone:     profile.Phone,
		Status:    profile.Status,
		KycStatus: profile.KYCStatus,
		CreatedAt: createdAt,
	}
	if profile.FullName != "" {
		resp.FullName = &profile.FullName
	}
	if profile.CountryCode != "" {
		resp.CountryCode = &profile.CountryCode
	}
	if profile.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", profile.DateOfBirth)
		if err == nil {
			resp.DateOfBirth = &openapi_types.Date{Time: dob}
		}
	}
	return api.GetProfile200JSONResponse(resp), nil
}

func (s *Server) GetKYCStatus(ctx context.Context, req api.GetKYCStatusRequestObject) (api.GetKYCStatusResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.GetKYCStatus401JSONResponse{Error: "unauthorized"}, nil
	}

	resp, err := s.users.GetKYCStatus(ctx, userID)
	if err != nil {
		return api.GetKYCStatus502JSONResponse{Error: err.Error()}, nil
	}

	out := api.KYCStatusResponse{Status: resp.Status}
	if resp.RejectionReason != "" {
		out.RejectionReason = &resp.RejectionReason
	}
	return api.GetKYCStatus200JSONResponse(out), nil
}

func (s *Server) GetWalletBalance(ctx context.Context, req api.GetWalletBalanceRequestObject) (api.GetWalletBalanceResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.GetWalletBalance401JSONResponse{Error: "unauthorized"}, nil
	}

	currency := "USD"
	if req.Params.Currency != nil {
		currency = *req.Params.Currency
	}

	balance, statusCode, err := s.users.GetWalletBalance(ctx, userID, currency)
	if statusCode == 404 {
		return api.GetWalletBalance404JSONResponse{Error: "wallet_not_found"}, nil
	}
	if err != nil {
		return api.GetWalletBalance502JSONResponse{Error: err.Error()}, nil
	}

	resp := api.WalletBalance{
		WalletId:         balance.WalletID,
		Currency:         balance.Currency,
		Balance:          balance.Balance,
		AvailableBalance: balance.AvailableBalance,
	}
	if balance.LedgerAccountID != "" {
		resp.LedgerAccountId = &balance.LedgerAccountID
	}
	if balance.EncumberedBalance != "" {
		resp.EncumberedBalance = &balance.EncumberedBalance
	}
	return api.GetWalletBalance200JSONResponse(resp), nil
}

func (s *Server) ListWalletTransactions(ctx context.Context, req api.ListWalletTransactionsRequestObject) (api.ListWalletTransactionsResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ListWalletTransactions401JSONResponse{Error: "unauthorized"}, nil
	}

	limit := 20
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	transfers, _, err := s.payments.ListTransfers(ctx, userID, limit)
	if err != nil {
		return api.ListWalletTransactions502JSONResponse{Error: err.Error()}, nil
	}
	auths, err := s.cards.ListAuthorizations(ctx, userID, limit)
	if err != nil {
		return api.ListWalletTransactions502JSONResponse{Error: err.Error()}, nil
	}

	transactions := buildWalletTransactions(userID, transfers, auths, limit)
	return api.ListWalletTransactions200JSONResponse{Transactions: transactions}, nil
}

func (s *Server) ListTransfers(ctx context.Context, req api.ListTransfersRequestObject) (api.ListTransfersResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ListTransfers401JSONResponse{Error: "unauthorized"}, nil
	}

	limit := 20
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	list, _, err := s.payments.ListTransfers(ctx, userID, limit)
	if err != nil {
		return api.ListTransfers502JSONResponse{Error: err.Error()}, nil
	}

	transfers := make([]api.Transfer, 0, len(list.Transfers))
	for _, t := range list.Transfers {
		view := api.Transfer{
			Id:              ptr(t.ID),
			Status:          ptr(t.Status),
			SenderUserId:    ptr(t.SenderUserID),
			RecipientUserId: ptr(t.RecipientUserID),
			Amount:          ptr(t.Amount),
			Currency:        ptr(t.Currency),
		}
		if t.LedgerTransferID != "" {
			view.LedgerTransferId = ptr(t.LedgerTransferID)
		}
		if t.FailureReason != "" {
			view.FailureReason = ptr(t.FailureReason)
		}
		if t.Memo != "" {
			view.Memo = ptr(t.Memo)
		}
		view.CreatedAt = parseTimePtr(t.CreatedAt)
		view.CompletedAt = parseTimePtr(t.CompletedAt)
		transfers = append(transfers, view)
	}
	return api.ListTransfers200JSONResponse{Transfers: transfers}, nil
}

func (s *Server) GetTransfer(ctx context.Context, req api.GetTransferRequestObject) (api.GetTransferResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.GetTransfer401JSONResponse{Error: "unauthorized"}, nil
	}

	transfer, statusCode, err := s.payments.GetTransfer(ctx, userID, req.Id)
	if statusCode == 404 {
		return api.GetTransfer404JSONResponse{Error: "transfer_not_found"}, nil
	}
	if err != nil {
		return api.GetTransfer502JSONResponse{Error: err.Error()}, nil
	}
	return api.GetTransfer200JSONResponse(toTransferView(transfer)), nil
}

func (s *Server) ProvisionWallet(ctx context.Context, req api.ProvisionWalletRequestObject) (api.ProvisionWalletResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ProvisionWallet401JSONResponse{Error: "unauthorized"}, nil
	}

	currency := "USD"
	if req.Body != nil && req.Body.Currency != nil {
		currency = *req.Body.Currency
	}

	resp, err := s.users.ProvisionWallet(ctx, userID, req.Params.IdempotencyKey, currency)
	if err != nil {
		return api.ProvisionWallet502JSONResponse{Error: err.Error()}, nil
	}

	return api.ProvisionWallet201JSONResponse{
		WalletId:        resp.WalletID,
		LedgerAccountId: resp.LedgerAccountID,
	}, nil
}

func (s *Server) CreateTransfer(ctx context.Context, req api.CreateTransferRequestObject) (api.CreateTransferResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.CreateTransfer401JSONResponse{Error: "unauthorized"}, nil
	}
	if req.Body == nil {
		return api.CreateTransfer502JSONResponse{Error: "invalid_json"}, nil
	}

	currency := "USD"
	if req.Body.Currency != nil {
		currency = *req.Body.Currency
	}
	memo := ""
	if req.Body.Memo != nil {
		memo = *req.Body.Memo
	}

	transfer, statusCode, err := s.payments.CreateP2PTransfer(ctx, userID, req.Params.IdempotencyKey, client.CreateTransferRequest{
		RecipientPhone: req.Body.RecipientPhone,
		Amount:         req.Body.Amount,
		Currency:       currency,
		Memo:           memo,
	})
	if statusCode == 0 {
		return api.CreateTransfer502JSONResponse{Error: err.Error()}, nil
	}

	view := toTransferView(transfer)

	switch statusCode {
	case 422:
		return api.CreateTransfer422JSONResponse(view), nil
	case 201:
		return api.CreateTransfer201JSONResponse(view), nil
	default:
		return api.CreateTransfer200JSONResponse(view), nil
	}
}

func (s *Server) IssueCard(ctx context.Context, req api.IssueCardRequestObject) (api.IssueCardResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.IssueCard401JSONResponse{Error: "unauthorized"}, nil
	}
	if req.Body == nil {
		return api.IssueCard502JSONResponse{Error: "invalid_json"}, nil
	}

	walletID := ""
	if req.Body.WalletId != nil {
		walletID = *req.Body.WalletId
	}

	card, statusCode, err := s.cards.IssueCard(ctx, userID, req.Params.IdempotencyKey, client.IssueCardRequest{
		WalletID:       walletID,
		CardholderName: req.Body.CardholderName,
	})
	if statusCode == 422 {
		return api.IssueCard422JSONResponse{Error: err.Error()}, nil
	}
	if err != nil {
		return api.IssueCard502JSONResponse{Error: err.Error()}, nil
	}

	view := toCardView(card)
	if statusCode == 200 {
		return api.IssueCard200JSONResponse(view), nil
	}
	return api.IssueCard201JSONResponse(view), nil
}

func (s *Server) ListCards(ctx context.Context, req api.ListCardsRequestObject) (api.ListCardsResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ListCards401JSONResponse{Error: "unauthorized"}, nil
	}

	list, err := s.cards.ListCards(ctx, userID)
	if err != nil {
		return api.ListCards502JSONResponse{Error: err.Error()}, nil
	}

	cards := make([]api.Card, 0, len(list.Cards))
	for _, c := range list.Cards {
		cards = append(cards, toCardView(c))
	}
	return api.ListCards200JSONResponse{Cards: cards}, nil
}

func (s *Server) GetCard(ctx context.Context, req api.GetCardRequestObject) (api.GetCardResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.GetCard401JSONResponse{Error: "unauthorized"}, nil
	}

	card, statusCode, err := s.cards.GetCard(ctx, userID, req.Id)
	if statusCode == 404 {
		return api.GetCard404JSONResponse{Error: "card_not_found"}, nil
	}
	if err != nil {
		return api.GetCard502JSONResponse{Error: err.Error()}, nil
	}
	return api.GetCard200JSONResponse(toCardView(card)), nil
}

func (s *Server) FreezeCard(ctx context.Context, req api.FreezeCardRequestObject) (api.FreezeCardResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.FreezeCard401JSONResponse{Error: "unauthorized"}, nil
	}

	card, statusCode, err := s.cards.FreezeCard(ctx, userID, req.Id)
	if statusCode == 404 {
		return api.FreezeCard404JSONResponse{Error: "card_not_found"}, nil
	}
	if err != nil {
		return api.FreezeCard502JSONResponse{Error: err.Error()}, nil
	}
	return api.FreezeCard200JSONResponse(toCardView(card)), nil
}

func (s *Server) UnfreezeCard(ctx context.Context, req api.UnfreezeCardRequestObject) (api.UnfreezeCardResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.UnfreezeCard401JSONResponse{Error: "unauthorized"}, nil
	}

	card, statusCode, err := s.cards.UnfreezeCard(ctx, userID, req.Id)
	if statusCode == 404 {
		return api.UnfreezeCard404JSONResponse{Error: "card_not_found"}, nil
	}
	if err != nil {
		return api.UnfreezeCard502JSONResponse{Error: err.Error()}, nil
	}
	return api.UnfreezeCard200JSONResponse(toCardView(card)), nil
}

func (s *Server) ListNotifications(ctx context.Context, req api.ListNotificationsRequestObject) (api.ListNotificationsResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ListNotifications401JSONResponse{Error: "unauthorized"}, nil
	}

	limit := 20
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	list, err := s.notifications.ListNotifications(ctx, userID, limit)
	if err != nil {
		return api.ListNotifications502JSONResponse{Error: err.Error()}, nil
	}

	notifications := make([]api.Notification, 0, len(list.Notifications))
	for _, n := range list.Notifications {
		createdAt, _ := time.Parse(time.RFC3339, n.CreatedAt)
		notifications = append(notifications, api.Notification{
			Id:        n.ID,
			UserId:    n.UserID,
			EventType: n.EventType,
			Title:     n.Title,
			Body:      n.Body,
			Read:      n.Read,
			CreatedAt: createdAt,
		})
	}
	return api.ListNotifications200JSONResponse{Notifications: notifications}, nil
}

func (s *Server) AuthorizeTransaction(ctx context.Context, req api.AuthorizeTransactionRequestObject) (api.AuthorizeTransactionResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.AuthorizeTransaction401JSONResponse{Error: "unauthorized"}, nil
	}
	if req.Body == nil {
		return api.AuthorizeTransaction502JSONResponse{Error: "invalid_json"}, nil
	}

	currency := "USD"
	if req.Body.Currency != nil {
		currency = *req.Body.Currency
	}
	merchant := ""
	if req.Body.MerchantName != nil {
		merchant = *req.Body.MerchantName
	}

	auth, statusCode, err := s.cards.AuthorizeTransaction(ctx, userID, req.Id, req.Params.IdempotencyKey, client.AuthorizeRequest{
		Amount:       req.Body.Amount,
		Currency:     currency,
		MerchantName: merchant,
	})
	if statusCode == 422 {
		return api.AuthorizeTransaction422JSONResponse(toAuthorizationView(auth)), nil
	}
	if err != nil {
		return api.AuthorizeTransaction502JSONResponse{Error: err.Error()}, nil
	}
	if statusCode == 200 {
		return api.AuthorizeTransaction200JSONResponse(toAuthorizationView(auth)), nil
	}
	return api.AuthorizeTransaction201JSONResponse(toAuthorizationView(auth)), nil
}

func (s *Server) ListAuthorizations(ctx context.Context, req api.ListAuthorizationsRequestObject) (api.ListAuthorizationsResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ListAuthorizations401JSONResponse{Error: "unauthorized"}, nil
	}

	limit := 20
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}

	list, err := s.cards.ListAuthorizations(ctx, userID, limit)
	if err != nil {
		return api.ListAuthorizations502JSONResponse{Error: err.Error()}, nil
	}

	auths := make([]api.CardAuthorization, 0, len(list.Authorizations))
	for _, a := range list.Authorizations {
		auths = append(auths, toAuthorizationView(a))
	}
	return api.ListAuthorizations200JSONResponse{Authorizations: auths}, nil
}

func (s *Server) GetAuthorization(ctx context.Context, req api.GetAuthorizationRequestObject) (api.GetAuthorizationResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.GetAuthorization401JSONResponse{Error: "unauthorized"}, nil
	}

	auth, statusCode, err := s.cards.GetAuthorization(ctx, userID, req.Id)
	if statusCode == 404 {
		return api.GetAuthorization404JSONResponse{Error: "authorization_not_found"}, nil
	}
	if err != nil {
		return api.GetAuthorization502JSONResponse{Error: err.Error()}, nil
	}
	return api.GetAuthorization200JSONResponse(toAuthorizationView(auth)), nil
}

func (s *Server) CaptureAuthorization(ctx context.Context, req api.CaptureAuthorizationRequestObject) (api.CaptureAuthorizationResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.CaptureAuthorization401JSONResponse{Error: "unauthorized"}, nil
	}

	auth, statusCode, err := s.cards.CaptureAuthorization(ctx, userID, req.Id)
	if statusCode == 404 {
		return api.CaptureAuthorization404JSONResponse{Error: "authorization_not_found"}, nil
	}
	if err != nil {
		return api.CaptureAuthorization502JSONResponse{Error: err.Error()}, nil
	}
	return api.CaptureAuthorization200JSONResponse(toAuthorizationView(auth)), nil
}

func toAuthorizationView(a client.AuthorizationView) api.CardAuthorization {
	out := api.CardAuthorization{
		Id:       a.ID,
		CardId:   a.CardID,
		UserId:   a.UserID,
		Amount:   a.Amount,
		Currency: a.Currency,
		Status:   a.Status,
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
	out.CreatedAt = parseTimePtr(a.CreatedAt)
	out.CapturedAt = parseTimePtr(a.CapturedAt)
	return out
}

func toTransferView(t client.TransferView) api.Transfer {
	view := api.Transfer{
		Id:              ptr(t.ID),
		Status:          ptr(t.Status),
		SenderUserId:    ptr(t.SenderUserID),
		RecipientUserId: ptr(t.RecipientUserID),
		Amount:          ptr(t.Amount),
		Currency:        ptr(t.Currency),
	}
	if t.LedgerTransferID != "" {
		view.LedgerTransferId = ptr(t.LedgerTransferID)
	}
	if t.FailureReason != "" {
		view.FailureReason = ptr(t.FailureReason)
	}
	if t.Memo != "" {
		view.Memo = ptr(t.Memo)
	}
	view.CreatedAt = parseTimePtr(t.CreatedAt)
	view.CompletedAt = parseTimePtr(t.CompletedAt)
	return view
}

func toCardView(c client.CardView) api.Card {
	return api.Card{
		Id:          c.ID,
		UserId:      c.UserID,
		WalletId:    c.WalletID,
		LastFour:    c.LastFour,
		Status:      c.Status,
		ExpiryMonth: c.ExpiryMonth,
		ExpiryYear:  c.ExpiryYear,
	}
}

func (s *Server) resolveUserID(authHeader, xUserID *string) string {
	if xUserID != nil && *xUserID != "" {
		return *xUserID
	}
	if authHeader == nil {
		return ""
	}
	raw := strings.TrimSpace(*authHeader)
	if !strings.HasPrefix(raw, "Bearer ") {
		return ""
	}
	token := strings.TrimSpace(strings.TrimPrefix(raw, "Bearer "))
	if token == "" {
		return ""
	}

	if strings.HasPrefix(token, "access.") {
		parts := strings.Split(token, ".")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	if s.jwt != nil {
		userID, err := s.jwt.ValidateAccessToken(token)
		if err == nil {
			return userID
		}
	}
	return ""
}

func ptr[T any](v T) *T {
	return &v
}