package grpc

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	apiadapter "github.com/iho/neobank/services/user/internal/adapter/api"
	genapi "github.com/iho/neobank/services/user/internal/gen/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GatewayServer struct {
	neobankv1.UnimplementedUserServiceServer
	api *apiadapter.Server
}

func NewGatewayServer(api *apiadapter.Server) *GatewayServer {
	return &GatewayServer{api: api}
}

func (s *GatewayServer) Register(ctx context.Context, req *neobankv1.RegisterRequest) (*neobankv1.AuthResponse, error) {
	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		idempotencyKey = grpcutil.IdempotencyKeyFromContext(ctx)
	}
	body := &genapi.RegisterJSONRequestBody{
		Email:    openapi_types.Email(req.GetEmail()),
		Password: req.GetPassword(),
	}
	if phone := req.GetPhone(); phone != "" {
		body.Phone = &phone
	}
	if invite := req.GetInviteCode(); invite != "" {
		body.InviteCode = &invite
	}
	resp, err := s.api.Register(ctx, genapi.RegisterRequestObject{
		Params: genapi.RegisterParams{IdempotencyKey: idempotencyKey},
		Body:   body,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "register: %v", err)
	}
	return mapRegisterResponse(resp)
}

func (s *GatewayServer) Login(ctx context.Context, req *neobankv1.LoginRequest) (*neobankv1.AuthResponse, error) {
	resp, err := s.api.Login(ctx, genapi.LoginRequestObject{
		Body: &genapi.LoginJSONRequestBody{
			Email:    openapi_types.Email(req.GetEmail()),
			Password: req.GetPassword(),
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "login: %v", err)
	}
	return mapLoginResponse(resp)
}

func (s *GatewayServer) RefreshToken(ctx context.Context, req *neobankv1.RefreshTokenRequest) (*neobankv1.AuthResponse, error) {
	resp, err := s.api.RefreshToken(ctx, genapi.RefreshTokenRequestObject{
		Body: &genapi.RefreshTokenJSONRequestBody{RefreshToken: req.GetRefreshToken()},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "refresh token: %v", err)
	}
	return mapRefreshTokenResponse(resp)
}

func (s *GatewayServer) ChangePassword(ctx context.Context, req *neobankv1.ChangePasswordRequest) (*neobankv1.EmptyResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	resp, err := s.api.ChangePassword(ctx, genapi.ChangePasswordRequestObject{
		Params: genapi.ChangePasswordParams{XUserId: userID},
		Body: &genapi.ChangePasswordJSONRequestBody{
			CurrentPassword: req.GetCurrentPassword(),
			NewPassword:     req.GetNewPassword(),
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "change password: %v", err)
	}
	return mapEmptyResponse(resp)
}

func (s *GatewayServer) GetProfile(ctx context.Context, req *neobankv1.GetProfileRequest) (*neobankv1.ProfileResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	resp, err := s.api.GetProfile(ctx, genapi.GetProfileRequestObject{
		Params: genapi.GetProfileParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get profile: %v", err)
	}
	return mapProfileResponse(resp)
}

func (s *GatewayServer) SubmitKYC(ctx context.Context, req *neobankv1.SubmitKYCRequest) (*neobankv1.SubmitKYCResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		idempotencyKey = grpcutil.IdempotencyKeyFromContext(ctx)
	}
	dob, err := time.Parse("2006-01-02", req.GetDateOfBirth())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid date_of_birth: %v", err)
	}
	body := &genapi.SubmitKYCJSONRequestBody{
		FullName:    req.GetFullName(),
		DateOfBirth: openapi_types.Date{Time: dob},
		CountryCode: req.GetCountryCode(),
	}
	if docType := req.GetDocumentType(); docType != "" {
		body.DocumentType = &docType
	}
	if docNum := req.GetDocumentNumber(); docNum != "" {
		body.DocumentNumber = &docNum
	}
	resp, err := s.api.SubmitKYC(ctx, genapi.SubmitKYCRequestObject{
		Params: genapi.SubmitKYCParams{
			IdempotencyKey: idempotencyKey,
			XUserId:        userID,
		},
		Body: body,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "submit kyc: %v", err)
	}
	return mapSubmitKYCResponse(resp)
}

func (s *GatewayServer) GetKYCStatus(ctx context.Context, req *neobankv1.GetKYCStatusRequest) (*neobankv1.KYCStatusResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	resp, err := s.api.GetKYCStatus(ctx, genapi.GetKYCStatusRequestObject{
		Params: genapi.GetKYCStatusParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get kyc status: %v", err)
	}
	return mapKYCStatusResponse(resp)
}

func (s *GatewayServer) GetWalletBalance(ctx context.Context, req *neobankv1.GetWalletBalanceRequest) (*neobankv1.WalletBalanceResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	params := genapi.GetWalletBalanceParams{XUserId: userID}
	if currency := req.GetCurrency(); currency != "" {
		params.Currency = &currency
	}
	resp, err := s.api.GetWalletBalance(ctx, genapi.GetWalletBalanceRequestObject{Params: params})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get wallet balance: %v", err)
	}
	return mapWalletBalanceResponse(resp)
}

func (s *GatewayServer) ListWallets(ctx context.Context, req *neobankv1.ListWalletsRequest) (*neobankv1.ListWalletsResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	resp, err := s.api.ListWallets(ctx, genapi.ListWalletsRequestObject{
		Params: genapi.ListWalletsParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list wallets: %v", err)
	}
	return mapListWalletsResponse(resp)
}

func (s *GatewayServer) ProvisionWallet(ctx context.Context, req *neobankv1.ProvisionWalletRequest) (*neobankv1.ProvisionWalletResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		idempotencyKey = grpcutil.IdempotencyKeyFromContext(ctx)
	}
	body := &genapi.ProvisionWalletJSONRequestBody{}
	if currency := req.GetCurrency(); currency != "" {
		body.Currency = &currency
	}
	resp, err := s.api.ProvisionWallet(ctx, genapi.ProvisionWalletRequestObject{
		Params: genapi.ProvisionWalletParams{
			IdempotencyKey: idempotencyKey,
			XUserId:        userID,
		},
		Body: body,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "provision wallet: %v", err)
	}
	return mapProvisionWalletResponse(resp)
}

func (s *GatewayServer) DepositWallet(ctx context.Context, req *neobankv1.DepositWalletRequest) (*neobankv1.DepositWalletResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		idempotencyKey = grpcutil.IdempotencyKeyFromContext(ctx)
	}
	body := &genapi.DepositWalletJSONRequestBody{Amount: req.GetAmount()}
	if currency := req.GetCurrency(); currency != "" {
		body.Currency = &currency
	}
	resp, err := s.api.DepositWallet(ctx, genapi.DepositWalletRequestObject{
		Params: genapi.DepositWalletParams{
			IdempotencyKey: idempotencyKey,
			XUserId:        userID,
		},
		Body: body,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deposit wallet: %v", err)
	}
	return mapDepositWalletResponse(resp)
}

func (s *GatewayServer) ListWalletTransactions(ctx context.Context, req *neobankv1.ListWalletTransactionsRequest) (*neobankv1.ListWalletTransactionsResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	params := genapi.ListWalletTransactionsParams{XUserId: userID}
	if limit := req.GetLimit(); limit > 0 {
		l := int(limit)
		params.Limit = &l
	}
	if cursor := req.GetCursor(); cursor != "" {
		params.Cursor = &cursor
	}
	resp, err := s.api.ListWalletTransactions(ctx, genapi.ListWalletTransactionsRequestObject{Params: params})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list wallet transactions: %v", err)
	}
	return mapListWalletTransactionsResponse(resp)
}

func (s *GatewayServer) ExportWalletTransactions(ctx context.Context, req *neobankv1.ExportWalletTransactionsRequest) (*neobankv1.ExportWalletTransactionsResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	from, err := time.Parse("2006-01-02", req.GetFrom())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid from date: %v", err)
	}
	to, err := time.Parse("2006-01-02", req.GetTo())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid to date: %v", err)
	}
	params := genapi.ExportWalletTransactionsParams{
		XUserId: userID,
		From:    openapi_types.Date{Time: from},
		To:      openapi_types.Date{Time: to},
	}
	if format := req.GetFormat(); format != "" {
		params.Format = &format
	}
	resp, err := s.api.ExportWalletTransactions(ctx, genapi.ExportWalletTransactionsRequestObject{Params: params})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "export wallet transactions: %v", err)
	}
	return mapExportWalletTransactionsResponse(resp)
}

func (s *GatewayServer) ListPayees(ctx context.Context, req *neobankv1.ListPayeesRequest) (*neobankv1.ListPayeesResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	params := genapi.ListPayeesParams{XUserId: userID}
	if limit := req.GetLimit(); limit > 0 {
		l := int(limit)
		params.Limit = &l
	}
	resp, err := s.api.ListPayees(ctx, genapi.ListPayeesRequestObject{Params: params})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list payees: %v", err)
	}
	return mapListPayeesResponse(resp)
}

func (s *GatewayServer) CreatePayee(ctx context.Context, req *neobankv1.CreatePayeeRequest) (*neobankv1.PayeeResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	payeeUserID, err := parseXUserID(req.GetPayeeUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payee_user_id: %v", err)
	}
	body := &genapi.CreatePayeeJSONRequestBody{PayeeUserId: payeeUserID}
	if nickname := req.GetNickname(); nickname != "" {
		body.Nickname = &nickname
	}
	resp, err := s.api.CreatePayee(ctx, genapi.CreatePayeeRequestObject{
		Params: genapi.CreatePayeeParams{XUserId: userID},
		Body:   body,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create payee: %v", err)
	}
	return mapPayeeResponse(resp)
}

func (s *GatewayServer) DeletePayee(ctx context.Context, req *neobankv1.DeletePayeeRequest) (*neobankv1.EmptyResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	payeeID, err := parseUUID(req.GetPayeeId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payee_id: %v", err)
	}
	resp, err := s.api.DeletePayee(ctx, genapi.DeletePayeeRequestObject{
		Id:     payeeID,
		Params: genapi.DeletePayeeParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "delete payee: %v", err)
	}
	return mapDeletePayeeResponse(resp)
}

func (s *GatewayServer) RegisterDeviceToken(ctx context.Context, req *neobankv1.RegisterDeviceTokenRequest) (*neobankv1.DeviceTokenResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	resp, err := s.api.RegisterDeviceToken(ctx, genapi.RegisterDeviceTokenRequestObject{
		Params: genapi.RegisterDeviceTokenParams{XUserId: userID},
		Body: &genapi.RegisterDeviceTokenJSONRequestBody{
			Platform: genapi.RegisterDeviceTokenRequestPlatform(req.GetPlatform()),
			Token:    req.GetToken(),
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "register device token: %v", err)
	}
	return mapDeviceTokenResponse(resp)
}

func (s *GatewayServer) DeleteDeviceToken(ctx context.Context, req *neobankv1.DeleteDeviceTokenRequest) (*neobankv1.EmptyResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	tokenID, err := parseUUID(req.GetTokenId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token_id: %v", err)
	}
	resp, err := s.api.DeleteDeviceToken(ctx, genapi.DeleteDeviceTokenRequestObject{
		Id:     tokenID,
		Params: genapi.DeleteDeviceTokenParams{XUserId: userID},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "delete device token: %v", err)
	}
	return mapDeleteDeviceTokenResponse(resp)
}

func (s *GatewayServer) CloseAccount(ctx context.Context, req *neobankv1.CloseAccountRequest) (*neobankv1.EmptyResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		idempotencyKey = grpcutil.IdempotencyKeyFromContext(ctx)
	}
	resp, err := s.api.CloseAccount(ctx, genapi.CloseAccountRequestObject{
		Params: genapi.CloseAccountParams{
			IdempotencyKey: idempotencyKey,
			XUserId:        userID,
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "close account: %v", err)
	}
	return mapCloseAccountResponse(resp)
}

func (s *GatewayServer) CreateReferralInvite(ctx context.Context, req *neobankv1.CreateReferralInviteRequest) (*neobankv1.ReferralInviteResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		idempotencyKey = grpcutil.IdempotencyKeyFromContext(ctx)
	}
	resp, err := s.api.CreateReferralInvite(ctx, genapi.CreateReferralInviteRequestObject{
		Params: genapi.CreateReferralInviteParams{
			IdempotencyKey: idempotencyKey,
			XUserId:        userID,
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create referral invite: %v", err)
	}
	return mapReferralInviteResponse(resp)
}

func (s *GatewayServer) ListReferralInvites(ctx context.Context, req *neobankv1.ListReferralInvitesRequest) (*neobankv1.ListReferralInvitesResponse, error) {
	userID, err := parseXUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	params := genapi.ListReferralInvitesParams{XUserId: userID}
	if limit := req.GetLimit(); limit > 0 {
		l := int(limit)
		params.Limit = &l
	}
	resp, err := s.api.ListReferralInvites(ctx, genapi.ListReferralInvitesRequestObject{Params: params})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list referral invites: %v", err)
	}
	return mapListReferralInvitesResponse(resp)
}

func parseXUserID(userID string) (genapi.XUserId, error) {
	id, err := parseUUID(userID)
	if err != nil {
		return genapi.XUserId{}, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}
	return id, nil
}

func parseUUID(value string) (openapi_types.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return openapi_types.UUID{}, err
	}
	return openapi_types.UUID(id), nil
}

func mapRegisterResponse(resp genapi.RegisterResponseObject) (*neobankv1.AuthResponse, error) {
	switch r := resp.(type) {
	case genapi.Register201JSONResponse:
		return &neobankv1.AuthResponse{
			UserId:       r.UserId.String(),
			AccessToken:  r.AccessToken,
			RefreshToken: r.RefreshToken,
			HttpStatus:   201,
		}, nil
	case genapi.Register400JSONResponse:
		return &neobankv1.AuthResponse{HttpStatus: 400, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected register response")
	}
}

func mapLoginResponse(resp genapi.LoginResponseObject) (*neobankv1.AuthResponse, error) {
	switch r := resp.(type) {
	case genapi.Login200JSONResponse:
		return &neobankv1.AuthResponse{
			UserId:       r.UserId.String(),
			AccessToken:  r.AccessToken,
			RefreshToken: r.RefreshToken,
			HttpStatus:   200,
		}, nil
	case genapi.Login401JSONResponse:
		return &neobankv1.AuthResponse{HttpStatus: 401, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected login response")
	}
}

func mapRefreshTokenResponse(resp genapi.RefreshTokenResponseObject) (*neobankv1.AuthResponse, error) {
	switch r := resp.(type) {
	case genapi.RefreshToken200JSONResponse:
		return &neobankv1.AuthResponse{
			UserId:       r.UserId.String(),
			AccessToken:  r.AccessToken,
			RefreshToken: r.RefreshToken,
			HttpStatus:   200,
		}, nil
	case genapi.RefreshToken401JSONResponse:
		return &neobankv1.AuthResponse{HttpStatus: 401, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected refresh token response")
	}
}

func mapEmptyResponse(resp genapi.ChangePasswordResponseObject) (*neobankv1.EmptyResponse, error) {
	switch r := resp.(type) {
	case genapi.ChangePassword204Response:
		return &neobankv1.EmptyResponse{HttpStatus: 204}, nil
	case genapi.ChangePassword400JSONResponse:
		return &neobankv1.EmptyResponse{HttpStatus: 400, Error: r.Error}, nil
	case genapi.ChangePassword401JSONResponse:
		return &neobankv1.EmptyResponse{HttpStatus: 401, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected empty response")
	}
}

func mapDeletePayeeResponse(resp genapi.DeletePayeeResponseObject) (*neobankv1.EmptyResponse, error) {
	switch r := resp.(type) {
	case genapi.DeletePayee204Response:
		return &neobankv1.EmptyResponse{HttpStatus: 204}, nil
	case genapi.DeletePayee404JSONResponse:
		return &neobankv1.EmptyResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected delete payee response")
	}
}

func mapDeleteDeviceTokenResponse(resp genapi.DeleteDeviceTokenResponseObject) (*neobankv1.EmptyResponse, error) {
	switch r := resp.(type) {
	case genapi.DeleteDeviceToken204Response:
		return &neobankv1.EmptyResponse{HttpStatus: 204}, nil
	case genapi.DeleteDeviceToken404JSONResponse:
		return &neobankv1.EmptyResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected delete device token response")
	}
}

func mapCloseAccountResponse(resp genapi.CloseAccountResponseObject) (*neobankv1.EmptyResponse, error) {
	switch r := resp.(type) {
	case genapi.CloseAccount204Response:
		return &neobankv1.EmptyResponse{HttpStatus: 204}, nil
	case genapi.CloseAccount400JSONResponse:
		return &neobankv1.EmptyResponse{HttpStatus: 400, Error: r.Error}, nil
	case genapi.CloseAccount401JSONResponse:
		return &neobankv1.EmptyResponse{HttpStatus: 401, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected close account response")
	}
}

func mapProfileResponse(resp genapi.GetProfileResponseObject) (*neobankv1.ProfileResponse, error) {
	switch r := resp.(type) {
	case genapi.GetProfile200JSONResponse:
		return &neobankv1.ProfileResponse{Profile: toProtoProfile(genapi.Profile(r)), HttpStatus: 200}, nil
	case genapi.GetProfile404JSONResponse:
		return &neobankv1.ProfileResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected profile response")
	}
}

func mapSubmitKYCResponse(resp genapi.SubmitKYCResponseObject) (*neobankv1.SubmitKYCResponse, error) {
	switch r := resp.(type) {
	case genapi.SubmitKYC200JSONResponse:
		out := &neobankv1.SubmitKYCResponse{
			KycCaseId: r.KycCaseId.String(),
			Status:    r.Status,
			HttpStatus: 200,
		}
		if r.WalletId != nil {
			out.WalletId = r.WalletId.String()
		}
		if r.RejectionReason != nil {
			out.RejectionReason = *r.RejectionReason
		}
		return out, nil
	case genapi.SubmitKYC400JSONResponse:
		return &neobankv1.SubmitKYCResponse{HttpStatus: 400, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected submit kyc response")
	}
}

func mapKYCStatusResponse(resp genapi.GetKYCStatusResponseObject) (*neobankv1.KYCStatusResponse, error) {
	switch r := resp.(type) {
	case genapi.GetKYCStatus200JSONResponse:
		out := &neobankv1.KYCStatusResponse{Status: r.Status, HttpStatus: 200}
		if r.RejectionReason != nil {
			out.RejectionReason = *r.RejectionReason
		}
		return out, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected kyc status response")
	}
}

func mapWalletBalanceResponse(resp genapi.GetWalletBalanceResponseObject) (*neobankv1.WalletBalanceResponse, error) {
	switch r := resp.(type) {
	case genapi.GetWalletBalance200JSONResponse:
		balance := toProtoWalletBalance(genapi.WalletBalance(r))
		return &neobankv1.WalletBalanceResponse{
			Balance:    &balance,
			HttpStatus: 200,
		}, nil
	case genapi.GetWalletBalance404JSONResponse:
		return &neobankv1.WalletBalanceResponse{HttpStatus: 404, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected wallet balance response")
	}
}

func mapListWalletsResponse(resp genapi.ListWalletsResponseObject) (*neobankv1.ListWalletsResponse, error) {
	switch r := resp.(type) {
	case genapi.ListWallets200JSONResponse:
		wallets := make([]*neobankv1.WalletBalance, 0, len(r.Wallets))
		for _, w := range r.Wallets {
			balance := toProtoWalletBalance(w)
			wallets = append(wallets, &balance)
		}
		return &neobankv1.ListWalletsResponse{Wallets: wallets, HttpStatus: 200}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected list wallets response")
	}
}

func mapProvisionWalletResponse(resp genapi.ProvisionWalletResponseObject) (*neobankv1.ProvisionWalletResponse, error) {
	switch r := resp.(type) {
	case genapi.ProvisionWallet201JSONResponse:
		return &neobankv1.ProvisionWalletResponse{
			WalletId:        r.WalletId.String(),
			LedgerAccountId: r.LedgerAccountId,
			HttpStatus:      201,
		}, nil
	case genapi.ProvisionWallet400JSONResponse:
		return &neobankv1.ProvisionWalletResponse{HttpStatus: 400, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected provision wallet response")
	}
}

func mapDepositWalletResponse(resp genapi.DepositWalletResponseObject) (*neobankv1.DepositWalletResponse, error) {
	switch r := resp.(type) {
	case genapi.DepositWallet200JSONResponse:
		return &neobankv1.DepositWalletResponse{Deposit: toProtoDeposit(genapi.DepositWalletResponse(r)), HttpStatus: 200}, nil
	case genapi.DepositWallet201JSONResponse:
		return &neobankv1.DepositWalletResponse{Deposit: toProtoDeposit(genapi.DepositWalletResponse(r)), HttpStatus: 201}, nil
	case genapi.DepositWallet400JSONResponse:
		return &neobankv1.DepositWalletResponse{HttpStatus: 400, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected deposit wallet response")
	}
}

func mapListWalletTransactionsResponse(resp genapi.ListWalletTransactionsResponseObject) (*neobankv1.ListWalletTransactionsResponse, error) {
	switch r := resp.(type) {
	case genapi.ListWalletTransactions200JSONResponse:
		txs := make([]*neobankv1.WalletTransaction, 0, len(r.Transactions))
		for _, tx := range r.Transactions {
			item := toProtoWalletTransaction(tx)
			txs = append(txs, &item)
		}
		out := &neobankv1.ListWalletTransactionsResponse{Transactions: txs, HttpStatus: 200}
		if r.NextCursor != nil {
			out.NextCursor = *r.NextCursor
		}
		return out, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected list wallet transactions response")
	}
}

func mapExportWalletTransactionsResponse(resp genapi.ExportWalletTransactionsResponseObject) (*neobankv1.ExportWalletTransactionsResponse, error) {
	switch r := resp.(type) {
	case genapi.ExportWalletTransactions200TextcsvResponse:
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "read csv export: %v", err)
		}
		return &neobankv1.ExportWalletTransactionsResponse{CsvData: data, HttpStatus: 200}, nil
	case genapi.ExportWalletTransactions400JSONResponse:
		return &neobankv1.ExportWalletTransactionsResponse{HttpStatus: 400, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected export wallet transactions response")
	}
}

func mapListPayeesResponse(resp genapi.ListPayeesResponseObject) (*neobankv1.ListPayeesResponse, error) {
	switch r := resp.(type) {
	case genapi.ListPayees200JSONResponse:
		payees := make([]*neobankv1.Payee, 0, len(r.Payees))
		for _, p := range r.Payees {
			item := toProtoPayee(p)
			payees = append(payees, &item)
		}
		return &neobankv1.ListPayeesResponse{Payees: payees, HttpStatus: 200}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected list payees response")
	}
}

func mapPayeeResponse(resp genapi.CreatePayeeResponseObject) (*neobankv1.PayeeResponse, error) {
	switch r := resp.(type) {
	case genapi.CreatePayee201JSONResponse:
		payee := toProtoPayee(genapi.Payee(r))
		return &neobankv1.PayeeResponse{Payee: &payee, HttpStatus: 201}, nil
	case genapi.CreatePayee400JSONResponse:
		return &neobankv1.PayeeResponse{HttpStatus: 400, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected payee response")
	}
}

func mapDeviceTokenResponse(resp genapi.RegisterDeviceTokenResponseObject) (*neobankv1.DeviceTokenResponse, error) {
	switch r := resp.(type) {
	case genapi.RegisterDeviceToken201JSONResponse:
		token := toProtoDeviceToken(genapi.DeviceToken(r))
		return &neobankv1.DeviceTokenResponse{DeviceToken: &token, HttpStatus: 201}, nil
	case genapi.RegisterDeviceToken400JSONResponse:
		return &neobankv1.DeviceTokenResponse{HttpStatus: 400, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected device token response")
	}
}

func mapReferralInviteResponse(resp genapi.CreateReferralInviteResponseObject) (*neobankv1.ReferralInviteResponse, error) {
	switch r := resp.(type) {
	case genapi.CreateReferralInvite201JSONResponse:
		invite := toProtoReferralInvite(genapi.ReferralInvite(r))
		return &neobankv1.ReferralInviteResponse{Invite: &invite, HttpStatus: 201}, nil
	case genapi.CreateReferralInvite400JSONResponse:
		return &neobankv1.ReferralInviteResponse{HttpStatus: 400, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected referral invite response")
	}
}

func mapListReferralInvitesResponse(resp genapi.ListReferralInvitesResponseObject) (*neobankv1.ListReferralInvitesResponse, error) {
	switch r := resp.(type) {
	case genapi.ListReferralInvites200JSONResponse:
		invites := make([]*neobankv1.ReferralInvite, 0, len(r.Invites))
		for _, inv := range r.Invites {
			item := toProtoReferralInvite(inv)
			invites = append(invites, &item)
		}
		return &neobankv1.ListReferralInvitesResponse{Invites: invites, HttpStatus: 200}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected list referral invites response")
	}
}

func toProtoProfile(p genapi.Profile) *neobankv1.Profile {
	out := &neobankv1.Profile{
		UserId:    p.UserId.String(),
		Email:     p.Email,
		Phone:     p.Phone,
		Status:    p.Status,
		KycStatus: p.KycStatus,
		CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339),
	}
	if p.FullName != nil {
		out.FullName = *p.FullName
	}
	if p.CountryCode != nil {
		out.CountryCode = *p.CountryCode
	}
	if p.DateOfBirth != nil {
		out.DateOfBirth = p.DateOfBirth.String()
	}
	return out
}

func toProtoWalletBalance(w genapi.WalletBalance) neobankv1.WalletBalance {
	out := neobankv1.WalletBalance{
		WalletId:         w.WalletId.String(),
		Currency:         w.Currency,
		Balance:          w.Balance,
		AvailableBalance: w.AvailableBalance,
	}
	if w.LedgerAccountId != nil {
		out.LedgerAccountId = *w.LedgerAccountId
	}
	if w.EncumberedBalance != nil {
		out.EncumberedBalance = *w.EncumberedBalance
	}
	return out
}

func toProtoWalletTransaction(tx genapi.WalletTransaction) neobankv1.WalletTransaction {
	out := neobankv1.WalletTransaction{
		Id:        tx.Id,
		Type:      tx.Type,
		Amount:    tx.Amount,
		Currency:  tx.Currency,
		Direction: tx.Direction,
		Status:    tx.Status,
		CreatedAt: tx.CreatedAt.UTC().Format(time.RFC3339),
	}
	if tx.Counterparty != nil {
		out.Counterparty = *tx.Counterparty
	}
	if tx.Memo != nil {
		out.Memo = *tx.Memo
	}
	if tx.ReferenceId != nil {
		out.ReferenceId = *tx.ReferenceId
	}
	return out
}

func toProtoDeposit(d genapi.DepositWalletResponse) *neobankv1.Deposit {
	out := &neobankv1.Deposit{
		Id:       d.Id.String(),
		WalletId: d.WalletId.String(),
		Amount:   d.Amount,
		Currency: d.Currency,
		Status:   d.Status,
	}
	if d.LedgerTransferId != nil {
		out.LedgerTransferId = *d.LedgerTransferId
	}
	if d.CreatedAt != nil {
		out.CreatedAt = d.CreatedAt.UTC().Format(time.RFC3339)
	}
	return out
}

func toProtoPayee(p genapi.Payee) neobankv1.Payee {
	out := neobankv1.Payee{
		Id:          p.Id.String(),
		PayeeUserId: p.PayeeUserId.String(),
		LastUsedAt:  p.LastUsedAt.UTC().Format(time.RFC3339),
		CreatedAt:   p.CreatedAt.UTC().Format(time.RFC3339),
	}
	if p.Nickname != nil {
		out.Nickname = *p.Nickname
	}
	if p.PayeeEmail != nil {
		out.PayeeEmail = *p.PayeeEmail
	}
	if p.PayeePhone != nil {
		out.PayeePhone = *p.PayeePhone
	}
	return out
}

func toProtoDeviceToken(t genapi.DeviceToken) neobankv1.DeviceToken {
	return neobankv1.DeviceToken{
		Id:        t.Id.String(),
		UserId:    t.UserId.String(),
		Platform:  t.Platform,
		Token:     t.Token,
		CreatedAt: t.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toProtoReferralInvite(inv genapi.ReferralInvite) neobankv1.ReferralInvite {
	out := neobankv1.ReferralInvite{
		Id:            inv.Id.String(),
		InviterUserId: inv.InviterUserId.String(),
		InviteCode:    inv.InviteCode,
		Status:        inv.Status,
		CreatedAt:     inv.CreatedAt.UTC().Format(time.RFC3339),
	}
	if inv.InviteeUserId != nil {
		out.InviteeUserId = inv.InviteeUserId.String()
	}
	if inv.AcceptedAt != nil {
		out.AcceptedAt = inv.AcceptedAt.UTC().Format(time.RFC3339)
	}
	return out
}