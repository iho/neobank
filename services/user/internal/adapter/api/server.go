package api

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/events"
	openapi_types "github.com/oapi-codegen/runtime/types"
	sqlcrepo "github.com/iho/neobank/services/user/internal/adapter/sqlc"
	"github.com/iho/neobank/services/user/internal/gen/api"
	"github.com/iho/neobank/services/user/internal/usecase"
	"github.com/jackc/pgx/v5"
)

type Server struct {
	register        *usecase.RegisterUseCase
	login           *usecase.LoginUseCase
	refresh         *usecase.RefreshTokenUseCase
	submitKYC         *usecase.SubmitKYCUseCase
	getKYCStatus      *usecase.GetKYCStatusUseCase
	getProfile        *usecase.GetProfileUseCase
	walletBalance       *usecase.GetWalletBalanceUseCase
	listWalletTx        *usecase.ListWalletTransactionsUseCase
	projectWalletEvent  *usecase.ProjectWalletEventUseCase
	provisionWallet     *usecase.ProvisionWalletUseCase
	users               *sqlcrepo.UserRepository
	wallets             *sqlcrepo.WalletRepository
}

func NewServer(
	register *usecase.RegisterUseCase,
	login *usecase.LoginUseCase,
	refresh *usecase.RefreshTokenUseCase,
	submitKYC *usecase.SubmitKYCUseCase,
	getKYCStatus *usecase.GetKYCStatusUseCase,
	getProfile *usecase.GetProfileUseCase,
	walletBalance *usecase.GetWalletBalanceUseCase,
	listWalletTx *usecase.ListWalletTransactionsUseCase,
	projectWalletEvent *usecase.ProjectWalletEventUseCase,
	provisionWallet *usecase.ProvisionWalletUseCase,
	users *sqlcrepo.UserRepository,
	wallets *sqlcrepo.WalletRepository,
) *Server {
	return &Server{
		register:           register,
		login:              login,
		refresh:            refresh,
		submitKYC:          submitKYC,
		getKYCStatus:       getKYCStatus,
		getProfile:         getProfile,
		walletBalance:      walletBalance,
		listWalletTx:       listWalletTx,
		projectWalletEvent: projectWalletEvent,
		provisionWallet:    provisionWallet,
		users:              users,
		wallets:            wallets,
	}
}

func (s *Server) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: "ok"}, nil
}

func (s *Server) Login(ctx context.Context, req api.LoginRequestObject) (api.LoginResponseObject, error) {
	if req.Body == nil {
		return api.Login401JSONResponse{Error: "invalid_json"}, nil
	}
	out, err := s.login.Execute(ctx, usecase.LoginInput{
		Email:    string(req.Body.Email),
		Password: req.Body.Password,
	})
	if err != nil {
		return api.Login401JSONResponse{Error: err.Error()}, nil
	}
	userID, err := uuid.Parse(out.UserID)
	if err != nil {
		return nil, err
	}
	return api.Login200JSONResponse{
		UserId:       openapi_types.UUID(userID),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}, nil
}

func (s *Server) RefreshToken(ctx context.Context, req api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	if req.Body == nil {
		return api.RefreshToken401JSONResponse{Error: "invalid_json"}, nil
	}
	out, err := s.refresh.Execute(ctx, req.Body.RefreshToken)
	if err != nil {
		return api.RefreshToken401JSONResponse{Error: err.Error()}, nil
	}
	userID, err := uuid.Parse(out.UserID)
	if err != nil {
		return nil, err
	}
	return api.RefreshToken200JSONResponse{
		UserId:       openapi_types.UUID(userID),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
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
	out, err := s.register.Execute(ctx, usecase.RegisterInput{
		Email:          string(req.Body.Email),
		Phone:          phone,
		Password:       req.Body.Password,
		IdempotencyKey: req.Params.IdempotencyKey,
	})
	if err != nil {
		return api.Register400JSONResponse{Error: err.Error()}, nil
	}
	userID, err := uuid.Parse(out.UserID)
	if err != nil {
		return nil, err
	}
	return api.Register201JSONResponse{
		UserId:       openapi_types.UUID(userID),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}, nil
}

func (s *Server) GetProfile(ctx context.Context, req api.GetProfileRequestObject) (api.GetProfileResponseObject, error) {
	profile, err := s.getProfile.Execute(ctx, req.Params.XUserId.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetProfile404JSONResponse{Error: "user_not_found"}, nil
		}
		return nil, err
	}
	userID, err := uuid.Parse(profile.UserID)
	if err != nil {
		return nil, err
	}
	resp := api.Profile{
		UserId:    openapi_types.UUID(userID),
		Email:     profile.Email,
		Phone:     profile.Phone,
		Status:    profile.Status,
		KycStatus: profile.KYCStatus,
		CreatedAt: profile.CreatedAt,
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

func (s *Server) SubmitKYC(ctx context.Context, req api.SubmitKYCRequestObject) (api.SubmitKYCResponseObject, error) {
	if req.Body == nil {
		return api.SubmitKYC400JSONResponse{Error: "invalid_json"}, nil
	}
	in := usecase.SubmitKYCInput{
		UserID:         req.Params.XUserId.String(),
		FullName:       req.Body.FullName,
		DateOfBirth:    req.Body.DateOfBirth.String(),
		CountryCode:    req.Body.CountryCode,
		IdempotencyKey: req.Params.IdempotencyKey,
	}
	if req.Body.DocumentType != nil {
		in.DocumentType = *req.Body.DocumentType
	}
	if req.Body.DocumentNumber != nil {
		in.DocumentNumber = *req.Body.DocumentNumber
	}
	out, err := s.submitKYC.Execute(ctx, in)
	if err != nil {
		return api.SubmitKYC400JSONResponse{Error: err.Error()}, nil
	}
	caseID, _ := uuid.Parse(out.KYCCaseID)
	walletID, _ := uuid.Parse(out.WalletID)
	return api.SubmitKYC200JSONResponse{
		KycCaseId: openapi_types.UUID(caseID),
		Status:    string(out.Status),
		WalletId:  openapi_types.UUID(walletID),
	}, nil
}

func (s *Server) GetKYCStatus(ctx context.Context, req api.GetKYCStatusRequestObject) (api.GetKYCStatusResponseObject, error) {
	kycCase, err := s.getKYCStatus.Execute(ctx, req.Params.XUserId.String())
	if err != nil {
		return nil, err
	}
	resp := api.KYCStatusResponse{Status: string(kycCase.Status)}
	if kycCase.RejectionReason != "" {
		resp.RejectionReason = &kycCase.RejectionReason
	}
	return api.GetKYCStatus200JSONResponse(resp), nil
}

func (s *Server) ListWalletTransactions(ctx context.Context, req api.ListWalletTransactionsRequestObject) (api.ListWalletTransactionsResponseObject, error) {
	limit := 20
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	rows, err := s.listWalletTx.Execute(ctx, req.Params.XUserId.String(), limit)
	if err != nil {
		return nil, err
	}
	views := make([]api.WalletTransaction, 0, len(rows))
	for _, row := range rows {
		view := api.WalletTransaction{
			Id:        row.ID,
			Type:      row.Type,
			Amount:    row.Amount,
			Currency:  row.Currency,
			Direction: row.Direction,
			Status:    row.Status,
			CreatedAt: row.CreatedAt.UTC(),
		}
		ref := row.ID
		view.ReferenceId = &ref
		if row.Counterparty != "" {
			view.Counterparty = &row.Counterparty
		}
		if row.Memo != "" {
			view.Memo = &row.Memo
		}
		views = append(views, view)
	}
	return api.ListWalletTransactions200JSONResponse{Transactions: views}, nil
}

func (s *Server) IngestEvent(ctx context.Context, req api.IngestEventRequestObject) (api.IngestEventResponseObject, error) {
	if req.Body == nil {
		return api.IngestEvent400JSONResponse{Error: "invalid_json"}, nil
	}
	payload, err := json.Marshal(req.Body.Payload)
	if err != nil {
		return api.IngestEvent400JSONResponse{Error: "invalid_payload"}, nil
	}
	envelope := events.Envelope{
		EventID:       req.Body.EventId,
		EventType:     req.Body.EventType,
		AggregateType: req.Body.AggregateType,
		AggregateID:   req.Body.AggregateId,
		Payload:       payload,
	}
	if req.Body.EventVersion != nil {
		envelope.EventVersion = *req.Body.EventVersion
	}
	if req.Body.OccurredAt != nil {
		envelope.OccurredAt = *req.Body.OccurredAt
	}
	if err := s.projectWalletEvent.Execute(ctx, envelope); err != nil {
		return api.IngestEvent400JSONResponse{Error: err.Error()}, nil
	}
	return api.IngestEvent202Response{}, nil
}

func (s *Server) GetWalletBalance(ctx context.Context, req api.GetWalletBalanceRequestObject) (api.GetWalletBalanceResponseObject, error) {
	currency := "USD"
	if req.Params.Currency != nil {
		currency = *req.Params.Currency
	}
	balance, err := s.walletBalance.Execute(ctx, usecase.GetWalletBalanceInput{
		UserID:   req.Params.XUserId.String(),
		Currency: currency,
	})
	if err != nil {
		if err.Error() == "wallet not found" {
			return api.GetWalletBalance404JSONResponse{Error: "wallet_not_found"}, nil
		}
		return api.GetWalletBalance404JSONResponse{Error: err.Error()}, nil
	}
	walletID, _ := uuid.Parse(balance.WalletID)
	return api.GetWalletBalance200JSONResponse{
		WalletId:          openapi_types.UUID(walletID),
		LedgerAccountId:   &balance.LedgerAccountID,
		Currency:          balance.Currency,
		Balance:           balance.Balance,
		EncumberedBalance: &balance.EncumberedBalance,
		AvailableBalance:  balance.AvailableBalance,
	}, nil
}

func (s *Server) ProvisionWallet(ctx context.Context, req api.ProvisionWalletRequestObject) (api.ProvisionWalletResponseObject, error) {
	userID := req.Params.XUserId.String()
	if req.Body != nil && req.Body.UserId != nil {
		userID = req.Body.UserId.String()
	}
	currency := "USD"
	if req.Body != nil && req.Body.Currency != nil {
		currency = *req.Body.Currency
	}
	out, err := s.provisionWallet.Execute(ctx, usecase.ProvisionWalletInput{
		UserID:         userID,
		Currency:       currency,
		IdempotencyKey: req.Params.IdempotencyKey,
	})
	if err != nil {
		return api.ProvisionWallet400JSONResponse{Error: err.Error()}, nil
	}
	walletID, err := uuid.Parse(out.WalletID)
	if err != nil {
		return nil, err
	}
	return api.ProvisionWallet201JSONResponse{
		WalletId:        openapi_types.UUID(walletID),
		LedgerAccountId: out.LedgerAccountID,
	}, nil
}

func (s *Server) GetUserByPhone(ctx context.Context, req api.GetUserByPhoneRequestObject) (api.GetUserByPhoneResponseObject, error) {
	user, err := s.users.GetByPhone(ctx, req.Phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetUserByPhone404JSONResponse{Error: "user_not_found"}, nil
		}
		return nil, err
	}
	id, err := uuid.Parse(user.ID)
	if err != nil {
		return nil, err
	}
	return api.GetUserByPhone200JSONResponse{
		Id:     openapi_types.UUID(id),
		Email:  user.Email,
		Phone:  user.Phone,
		Status: string(user.Status),
	}, nil
}

func (s *Server) GetWallet(ctx context.Context, req api.GetWalletRequestObject) (api.GetWalletResponseObject, error) {
	currency := "USD"
	if req.Params.Currency != nil {
		currency = *req.Params.Currency
	}
	wallet, err := s.wallets.GetByUserAndCurrency(ctx, req.Params.UserId.String(), currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetWallet404JSONResponse{Error: "wallet_not_found"}, nil
		}
		return nil, err
	}
	id, err := uuid.Parse(wallet.ID)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(wallet.UserID)
	if err != nil {
		return nil, err
	}
	return api.GetWallet200JSONResponse{
		Id:              openapi_types.UUID(id),
		UserId:          openapi_types.UUID(userID),
		Currency:        wallet.Currency,
		LedgerAccountId: wallet.LedgerAccountID,
		Status:          wallet.Status,
	}, nil
}