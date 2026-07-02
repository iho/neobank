package api

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/iho/neobank/pkg/auth"
	"github.com/iho/neobank/services/gateway/internal/client"
	gwmiddleware "github.com/iho/neobank/services/gateway/internal/middleware"
	"github.com/iho/neobank/services/gateway/internal/gen/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Server struct {
	jwt           *auth.JWT
	users         *client.UserClient
	payments      *client.PaymentClient
	cards         *client.CardClient
	notifications *client.NotificationClient
	allowDevAuth  bool
}

func NewServer(
	jwtAuth *auth.JWT,
	users *client.UserClient,
	payments *client.PaymentClient,
	cards *client.CardClient,
	notifications *client.NotificationClient,
	allowDevAuth bool,
) *Server {
	return &Server{
		jwt:           jwtAuth,
		users:         users,
		payments:      payments,
		cards:         cards,
		notifications: notifications,
		allowDevAuth:  allowDevAuth,
	}
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

func (s *Server) ChangePassword(ctx context.Context, req api.ChangePasswordRequestObject) (api.ChangePasswordResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ChangePassword401JSONResponse{Error: "unauthorized"}, nil
	}
	if req.Body == nil {
		return api.ChangePassword400JSONResponse{Error: "invalid_json"}, nil
	}
	status, err := s.users.ChangePassword(ctx, userID, req.Body.CurrentPassword, req.Body.NewPassword)
	if status == 401 {
		return api.ChangePassword401JSONResponse{Error: err.Error()}, nil
	}
	if err != nil {
		return api.ChangePassword502JSONResponse{Error: err.Error()}, nil
	}
	return api.ChangePassword204Response{}, nil
}

func (s *Server) Register(ctx context.Context, req api.RegisterRequestObject) (api.RegisterResponseObject, error) {
	if req.Body == nil {
		return api.Register400JSONResponse{Error: "invalid_json"}, nil
	}

	phone := ""
	if req.Body.Phone != nil {
		phone = *req.Body.Phone
	}

	inviteCode := ""
	if req.Body.InviteCode != nil {
		inviteCode = *req.Body.InviteCode
	}
	resp, err := s.users.Register(ctx, req.Params.IdempotencyKey, client.RegisterRequest{
		Email:      string(req.Body.Email),
		Phone:      phone,
		Password:   req.Body.Password,
		InviteCode: inviteCode,
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

	kycReq := client.SubmitKYCRequest{
		FullName:    req.Body.FullName,
		DateOfBirth: req.Body.DateOfBirth.String(),
		CountryCode: req.Body.CountryCode,
	}
	if req.Body.DocumentType != nil {
		kycReq.DocumentType = *req.Body.DocumentType
	}
	if req.Body.DocumentNumber != nil {
		kycReq.DocumentNumber = *req.Body.DocumentNumber
	}
	resp, err := s.users.SubmitKYC(ctx, userID, req.Params.IdempotencyKey, kycReq)
	if err != nil {
		return api.SubmitKYC502JSONResponse{Error: err.Error()}, nil
	}

	out := api.SubmitKYC200JSONResponse{
		KycCaseId: resp.KYCCaseID,
		Status:    resp.Status,
	}
	if resp.WalletID != "" {
		out.WalletId = &resp.WalletID
	}
	if resp.RejectionReason != "" {
		out.RejectionReason = &resp.RejectionReason
	}
	return out, nil
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

	cursor := ""
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	list, statusCode, err := s.users.ListWalletTransactions(ctx, userID, limit, cursor)
	if err != nil {
		if statusCode >= 400 && statusCode < 500 {
			return api.ListWalletTransactions401JSONResponse{Error: err.Error()}, nil
		}
		return api.ListWalletTransactions502JSONResponse{Error: err.Error()}, nil
	}

	transactions := make([]api.WalletTransaction, 0, len(list.Transactions))
	for _, t := range list.Transactions {
		tx := api.WalletTransaction{
			Id:        t.ID,
			Type:      t.Type,
			Amount:    t.Amount,
			Currency:  t.Currency,
			Direction: t.Direction,
			Status:    t.Status,
			CreatedAt: parseTimeOrNow(t.CreatedAt),
		}
		if t.Counterparty != "" {
			tx.Counterparty = &t.Counterparty
		}
		if t.Memo != "" {
			tx.Memo = &t.Memo
		}
		ref := t.ReferenceID
		if ref == "" {
			ref = t.ID
		}
		tx.ReferenceId = &ref
		transactions = append(transactions, tx)
	}
	resp := api.ListWalletTransactions200JSONResponse{Transactions: transactions}
	if list.NextCursor != "" {
		resp.NextCursor = &list.NextCursor
	}
	return resp, nil
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

	cursor := ""
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	list, _, err := s.payments.ListTransfers(ctx, userID, limit, cursor)
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
	resp := api.ListTransfers200JSONResponse{Transfers: transfers}
	if list.NextCursor != "" {
		resp.NextCursor = &list.NextCursor
	}
	return resp, nil
}

func (s *Server) GetLimits(ctx context.Context, req api.GetLimitsRequestObject) (api.GetLimitsResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.GetLimits401JSONResponse{Error: "unauthorized"}, nil
	}
	limits, _, err := s.payments.GetLimits(ctx, userID)
	if err != nil {
		return api.GetLimits502JSONResponse{Error: err.Error()}, nil
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

	transferReq := client.CreateTransferRequest{
		Amount:   req.Body.Amount,
		Currency: currency,
		Memo:     memo,
	}
	if req.Body.RecipientPhone != nil {
		transferReq.RecipientPhone = *req.Body.RecipientPhone
	}
	if req.Body.RecipientEmail != nil {
		transferReq.RecipientEmail = string(*req.Body.RecipientEmail)
	}
	if req.Body.RecipientUserId != nil {
		transferReq.RecipientUserID = *req.Body.RecipientUserId
	}

	transfer, statusCode, err := s.payments.CreateP2PTransfer(ctx, userID, req.Params.IdempotencyKey, transferReq)
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

	dailyLimit := ""
	if req.Body.DailyLimit != nil {
		dailyLimit = *req.Body.DailyLimit
	}
	onlineOnly := false
	if req.Body.OnlineOnly != nil {
		onlineOnly = *req.Body.OnlineOnly
	}
	card, statusCode, err := s.cards.IssueCard(ctx, userID, req.Params.IdempotencyKey, client.IssueCardRequest{
		WalletID:       walletID,
		CardholderName: req.Body.CardholderName,
		DailyLimit:     dailyLimit,
		OnlineOnly:     onlineOnly,
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

func (s *Server) UpdateCardControls(ctx context.Context, req api.UpdateCardControlsRequestObject) (api.UpdateCardControlsResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.UpdateCardControls401JSONResponse{Error: "unauthorized"}, nil
	}
	if req.Body == nil {
		return api.UpdateCardControls400JSONResponse{Error: "invalid_json"}, nil
	}
	card, statusCode, err := s.cards.UpdateCardControls(ctx, userID, req.Id, client.UpdateCardControlsRequest{
		DailyLimit: req.Body.DailyLimit,
		OnlineOnly: req.Body.OnlineOnly,
	})
	if statusCode == 404 {
		return api.UpdateCardControls404JSONResponse{Error: "card_not_found"}, nil
	}
	if statusCode == 400 {
		return api.UpdateCardControls400JSONResponse{Error: err.Error()}, nil
	}
	if err != nil {
		return api.UpdateCardControls502JSONResponse{Error: err.Error()}, nil
	}
	return api.UpdateCardControls200JSONResponse(toCardView(card)), nil
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

	cursor := ""
	if req.Params.Cursor != nil {
		cursor = *req.Params.Cursor
	}
	list, err := s.notifications.ListNotifications(ctx, userID, limit, cursor)
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
	resp := api.ListNotifications200JSONResponse{
		Notifications: notifications,
		UnreadCount:   list.UnreadCount,
	}
	if list.NextCursor != "" {
		resp.NextCursor = &list.NextCursor
	}
	return resp, nil
}

func (s *Server) ListPayees(ctx context.Context, req api.ListPayeesRequestObject) (api.ListPayeesResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ListPayees401JSONResponse{Error: "unauthorized"}, nil
	}
	limit := 20
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	list, _, err := s.users.ListPayees(ctx, userID, limit)
	if err != nil {
		return api.ListPayees502JSONResponse{Error: err.Error()}, nil
	}
	payees := make([]api.Payee, 0, len(list.Payees))
	for _, p := range list.Payees {
		payee := api.Payee{
			Id:          p.ID,
			PayeeUserId: p.PayeeUserID,
			LastUsedAt:  parseTimeOrNow(p.LastUsedAt),
			CreatedAt:   parseTimeOrNow(p.CreatedAt),
		}
		if p.Nickname != "" {
			payee.Nickname = &p.Nickname
		}
		if p.PayeeEmail != "" {
			payee.PayeeEmail = &p.PayeeEmail
		}
		if p.PayeePhone != "" {
			payee.PayeePhone = &p.PayeePhone
		}
		payees = append(payees, payee)
	}
	return api.ListPayees200JSONResponse{Payees: payees}, nil
}

func (s *Server) CreatePayee(ctx context.Context, req api.CreatePayeeRequestObject) (api.CreatePayeeResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.CreatePayee401JSONResponse{Error: "unauthorized"}, nil
	}
	if req.Body == nil {
		return api.CreatePayee400JSONResponse{Error: "invalid_json"}, nil
	}
	nickname := ""
	if req.Body.Nickname != nil {
		nickname = *req.Body.Nickname
	}
	payee, statusCode, err := s.users.CreatePayee(ctx, userID, req.Body.PayeeUserId, nickname)
	if statusCode == 400 {
		return api.CreatePayee400JSONResponse{Error: err.Error()}, nil
	}
	if err != nil {
		return api.CreatePayee502JSONResponse{Error: err.Error()}, nil
	}
	view := api.Payee{
		Id:          payee.ID,
		PayeeUserId: payee.PayeeUserID,
		LastUsedAt:  parseTimeOrNow(payee.LastUsedAt),
		CreatedAt:   parseTimeOrNow(payee.CreatedAt),
	}
	if payee.Nickname != "" {
		view.Nickname = &payee.Nickname
	}
	return api.CreatePayee201JSONResponse(view), nil
}

func (s *Server) DeletePayee(ctx context.Context, req api.DeletePayeeRequestObject) (api.DeletePayeeResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.DeletePayee401JSONResponse{Error: "unauthorized"}, nil
	}
	statusCode, err := s.users.DeletePayee(ctx, userID, req.Id)
	if statusCode == 404 {
		return api.DeletePayee404JSONResponse{Error: "payee_not_found"}, nil
	}
	if err != nil {
		return api.DeletePayee502JSONResponse{Error: err.Error()}, nil
	}
	return api.DeletePayee204Response{}, nil
}

func (s *Server) MarkNotificationRead(ctx context.Context, req api.MarkNotificationReadRequestObject) (api.MarkNotificationReadResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.MarkNotificationRead401JSONResponse{Error: "unauthorized"}, nil
	}
	n, statusCode, err := s.notifications.MarkNotificationRead(ctx, userID, req.Id.String())
	if statusCode == 404 {
		return api.MarkNotificationRead404JSONResponse{Error: "notification_not_found"}, nil
	}
	if err != nil {
		return api.MarkNotificationRead502JSONResponse{Error: err.Error()}, nil
	}
	createdAt, _ := time.Parse(time.RFC3339, n.CreatedAt)
	return api.MarkNotificationRead200JSONResponse{
		Id:        n.ID,
		UserId:    n.UserID,
		EventType: n.EventType,
		Title:     n.Title,
		Body:      n.Body,
		Read:      n.Read,
		CreatedAt: createdAt,
	}, nil
}

func (s *Server) MarkAllNotificationsRead(ctx context.Context, req api.MarkAllNotificationsReadRequestObject) (api.MarkAllNotificationsReadResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.MarkAllNotificationsRead401JSONResponse{Error: "unauthorized"}, nil
	}
	count, err := s.notifications.MarkAllNotificationsRead(ctx, userID)
	if err != nil {
		return api.MarkAllNotificationsRead502JSONResponse{Error: err.Error()}, nil
	}
	return api.MarkAllNotificationsRead200JSONResponse{MarkedCount: count}, nil
}

func (s *Server) DepositWallet(ctx context.Context, req api.DepositWalletRequestObject) (api.DepositWalletResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.DepositWallet401JSONResponse{Error: "unauthorized"}, nil
	}
	if req.Body == nil {
		return api.DepositWallet400JSONResponse{Error: "invalid_json"}, nil
	}
	currency := "USD"
	if req.Body.Currency != nil {
		currency = *req.Body.Currency
	}
	out, statusCode, err := s.users.DepositWallet(ctx, userID, req.Params.IdempotencyKey, client.DepositWalletRequest{
		Amount:   req.Body.Amount,
		Currency: currency,
	})
	if err != nil {
		if statusCode == 400 {
			return api.DepositWallet400JSONResponse{Error: err.Error()}, nil
		}
		return api.DepositWallet502JSONResponse{Error: err.Error()}, nil
	}
	view := api.DepositWalletResponse{
		Id:       out.ID,
		WalletId: out.WalletID,
		Amount:   out.Amount,
		Currency: out.Currency,
		Status:   out.Status,
	}
	if out.LedgerTransferID != "" {
		view.LedgerTransferId = &out.LedgerTransferID
	}
	if out.CreatedAt != "" {
		if createdAt, parseErr := time.Parse(time.RFC3339, out.CreatedAt); parseErr == nil {
			view.CreatedAt = &createdAt
		}
	}
	if statusCode == http.StatusCreated {
		return api.DepositWallet201JSONResponse(view), nil
	}
	return api.DepositWallet200JSONResponse(view), nil
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

	channel := "pos"
	if req.Body.Channel != nil {
		channel = string(*req.Body.Channel)
	}
	mcc := ""
	if req.Body.MerchantCategoryCode != nil {
		mcc = *req.Body.MerchantCategoryCode
	}
	auth, statusCode, err := s.cards.AuthorizeTransaction(ctx, userID, req.Id, req.Params.IdempotencyKey, client.AuthorizeRequest{
		Amount:               req.Body.Amount,
		Currency:             currency,
		MerchantName:         merchant,
		MerchantCategoryCode: mcc,
		Channel:              channel,
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

func toReferralInviteView(inv client.ReferralInviteView) api.ReferralInvite {
	out := api.ReferralInvite{
		Id:            inv.ID,
		InviterUserId: inv.InviterUserID,
		InviteCode:    inv.InviteCode,
		Status:        inv.Status,
		CreatedAt:     parseTimeOrNow(inv.CreatedAt),
	}
	if inv.InviteeUserID != "" {
		out.InviteeUserId = &inv.InviteeUserID
	}
	if inv.AcceptedAt != "" {
		out.AcceptedAt = parseTimePtr(inv.AcceptedAt)
	}
	return out
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
	if a.MerchantCategoryCode != "" {
		out.MerchantCategoryCode = &a.MerchantCategoryCode
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
	out := api.Card{
		Id:          c.ID,
		UserId:      c.UserID,
		WalletId:    c.WalletID,
		LastFour:    c.LastFour,
		Status:      c.Status,
		ExpiryMonth: c.ExpiryMonth,
		ExpiryYear:  c.ExpiryYear,
		OnlineOnly:  c.OnlineOnly,
	}
	if c.DailyLimit != "" {
		out.DailyLimit = &c.DailyLimit
	}
	return out
}

func (s *Server) RegisterDeviceToken(ctx context.Context, req api.RegisterDeviceTokenRequestObject) (api.RegisterDeviceTokenResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.RegisterDeviceToken401JSONResponse{Error: "unauthorized"}, nil
	}
	if req.Body == nil {
		return api.RegisterDeviceToken400JSONResponse{Error: "invalid_json"}, nil
	}
	token, status, err := s.users.RegisterDeviceToken(ctx, userID, req.Params.IdempotencyKey, client.RegisterDeviceTokenRequest{
		Platform: string(req.Body.Platform),
		Token:    req.Body.Token,
	})
	if status == 400 {
		return api.RegisterDeviceToken400JSONResponse{Error: err.Error()}, nil
	}
	if err != nil {
		return api.RegisterDeviceToken502JSONResponse{Error: err.Error()}, nil
	}
	createdAt, _ := time.Parse(time.RFC3339, token.CreatedAt)
	return api.RegisterDeviceToken201JSONResponse{
		Id:        token.ID,
		UserId:    token.UserID,
		Platform:  token.Platform,
		Token:     token.Token,
		CreatedAt: createdAt,
	}, nil
}

func (s *Server) DeleteDeviceToken(ctx context.Context, req api.DeleteDeviceTokenRequestObject) (api.DeleteDeviceTokenResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.DeleteDeviceToken401JSONResponse{Error: "unauthorized"}, nil
	}
	status, err := s.users.DeleteDeviceToken(ctx, userID, req.Id.String(), req.Params.IdempotencyKey)
	if status == 404 {
		return api.DeleteDeviceToken404JSONResponse{Error: "device_token_not_found"}, nil
	}
	if err != nil {
		return api.DeleteDeviceToken502JSONResponse{Error: err.Error()}, nil
	}
	return api.DeleteDeviceToken204Response{}, nil
}

func (s *Server) CloseAccount(ctx context.Context, req api.CloseAccountRequestObject) (api.CloseAccountResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.CloseAccount401JSONResponse{Error: "unauthorized"}, nil
	}

	cards, err := s.cards.ListCards(ctx, userID)
	if err != nil {
		return api.CloseAccount502JSONResponse{Error: err.Error()}, nil
	}
	for _, card := range cards.Cards {
		if card.Status == "active" {
			if _, _, err := s.cards.FreezeCard(ctx, userID, card.ID); err != nil {
				return api.CloseAccount502JSONResponse{Error: err.Error()}, nil
			}
		}
	}

	status, err := s.users.CloseAccount(ctx, userID, req.Params.IdempotencyKey)
	if status == 400 {
		return api.CloseAccount400JSONResponse{Error: err.Error()}, nil
	}
	if err != nil {
		return api.CloseAccount502JSONResponse{Error: err.Error()}, nil
	}
	return api.CloseAccount204Response{}, nil
}

func (s *Server) GetNotificationPreferences(ctx context.Context, req api.GetNotificationPreferencesRequestObject) (api.GetNotificationPreferencesResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.GetNotificationPreferences401JSONResponse{Error: "unauthorized"}, nil
	}
	prefs, err := s.notifications.GetNotificationPreferences(ctx, userID)
	if err != nil {
		return api.GetNotificationPreferences502JSONResponse{Error: err.Error()}, nil
	}
	return api.GetNotificationPreferences200JSONResponse{
		Transfers: prefs.Transfers,
		Cards:     prefs.Cards,
		Kyc:       prefs.KYC,
		Push:      prefs.Push,
		Email:     prefs.Email,
	}, nil
}

func (s *Server) UpdateNotificationPreferences(ctx context.Context, req api.UpdateNotificationPreferencesRequestObject) (api.UpdateNotificationPreferencesResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.UpdateNotificationPreferences401JSONResponse{Error: "unauthorized"}, nil
	}
	if req.Body == nil {
		return api.UpdateNotificationPreferences400JSONResponse{Error: "invalid_json"}, nil
	}
	prefs, err := s.notifications.UpdateNotificationPreferences(ctx, userID, client.UpdateNotificationPreferencesRequest{
		Transfers: req.Body.Transfers,
		Cards:     req.Body.Cards,
		KYC:       req.Body.Kyc,
		Push:      req.Body.Push,
		Email:     req.Body.Email,
	})
	if err != nil {
		return api.UpdateNotificationPreferences502JSONResponse{Error: err.Error()}, nil
	}
	return api.UpdateNotificationPreferences200JSONResponse{
		Transfers: prefs.Transfers,
		Cards:     prefs.Cards,
		Kyc:       prefs.KYC,
		Push:      prefs.Push,
		Email:     prefs.Email,
	}, nil
}

func (s *Server) ExportWalletTransactions(ctx context.Context, req api.ExportWalletTransactionsRequestObject) (api.ExportWalletTransactionsResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ExportWalletTransactions401JSONResponse{Error: "unauthorized"}, nil
	}
	format := "csv"
	if req.Params.Format != nil {
		format = *req.Params.Format
	}
	data, statusCode, err := s.users.ExportWalletTransactions(ctx, userID, format, req.Params.From.String(), req.Params.To.String())
	if statusCode == 400 {
		return api.ExportWalletTransactions400JSONResponse{Error: err.Error()}, nil
	}
	if err != nil {
		return api.ExportWalletTransactions502JSONResponse{Error: err.Error()}, nil
	}
	return api.ExportWalletTransactions200TextcsvResponse{
		Body:          bytes.NewReader(data),
		ContentLength: int64(len(data)),
	}, nil
}

func (s *Server) ListWallets(ctx context.Context, req api.ListWalletsRequestObject) (api.ListWalletsResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ListWallets401JSONResponse{Error: "unauthorized"}, nil
	}
	list, _, err := s.users.ListWallets(ctx, userID)
	if err != nil {
		return api.ListWallets502JSONResponse{Error: err.Error()}, nil
	}
	wallets := make([]api.WalletBalance, 0, len(list.Wallets))
	for _, w := range list.Wallets {
		view := api.WalletBalance{
			WalletId:         w.WalletID,
			Currency:         w.Currency,
			Balance:          w.Balance,
			AvailableBalance: w.AvailableBalance,
		}
		if w.LedgerAccountID != "" {
			view.LedgerAccountId = &w.LedgerAccountID
		}
		if w.EncumberedBalance != "" {
			view.EncumberedBalance = &w.EncumberedBalance
		}
		wallets = append(wallets, view)
	}
	return api.ListWallets200JSONResponse{Wallets: wallets}, nil
}

func (s *Server) CreateReferralInvite(ctx context.Context, req api.CreateReferralInviteRequestObject) (api.CreateReferralInviteResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.CreateReferralInvite401JSONResponse{Error: "unauthorized"}, nil
	}
	invite, statusCode, err := s.users.CreateReferralInvite(ctx, userID, req.Params.IdempotencyKey)
	if statusCode == 400 {
		return api.CreateReferralInvite502JSONResponse{Error: err.Error()}, nil
	}
	if err != nil {
		return api.CreateReferralInvite502JSONResponse{Error: err.Error()}, nil
	}
	return api.CreateReferralInvite201JSONResponse(toReferralInviteView(invite)), nil
}

func (s *Server) ListReferralInvites(ctx context.Context, req api.ListReferralInvitesRequestObject) (api.ListReferralInvitesResponseObject, error) {
	userID := s.resolveUserID(req.Params.Authorization, req.Params.XUserId)
	if userID == "" {
		return api.ListReferralInvites401JSONResponse{Error: "unauthorized"}, nil
	}
	limit := 20
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	list, _, err := s.users.ListReferralInvites(ctx, userID, limit)
	if err != nil {
		return api.ListReferralInvites502JSONResponse{Error: err.Error()}, nil
	}
	invites := make([]api.ReferralInvite, 0, len(list.Invites))
	for _, inv := range list.Invites {
		invites = append(invites, toReferralInviteView(inv))
	}
	return api.ListReferralInvites200JSONResponse{Invites: invites}, nil
}

func (s *Server) resolveUserID(authHeader, xUserID *string) string {
	var authVal, xUserVal string
	if authHeader != nil {
		authVal = *authHeader
	}
	if xUserID != nil {
		xUserVal = *xUserID
	}
	return gwmiddleware.ResolveUserID(authVal, xUserVal, s.jwt, s.allowDevAuth)
}

func ptr[T any](v T) *T {
	return &v
}
