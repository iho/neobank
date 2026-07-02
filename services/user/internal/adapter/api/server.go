package api

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/audit"
	"github.com/iho/neobank/pkg/events"
	openapi_types "github.com/oapi-codegen/runtime/types"
	sqlcrepo "github.com/iho/neobank/services/user/internal/adapter/sqlc"
	"github.com/iho/neobank/services/user/internal/domain"
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
	exportGDPR          *usecase.ExportGDPRUseCase
	maskGDPR            *usecase.MaskGDPRUseCase
	depositWallet       *usecase.DepositWalletUseCase
	changePassword      *usecase.ChangePasswordUseCase
	users               *sqlcrepo.UserRepository
	wallets             *sqlcrepo.WalletRepository
	piiAccess           audit.PIIAccessRecorder
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
	exportGDPR *usecase.ExportGDPRUseCase,
	maskGDPR *usecase.MaskGDPRUseCase,
	depositWallet *usecase.DepositWalletUseCase,
	changePassword *usecase.ChangePasswordUseCase,
	users *sqlcrepo.UserRepository,
	wallets *sqlcrepo.WalletRepository,
	piiAccess audit.PIIAccessRecorder,
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
		exportGDPR:         exportGDPR,
		maskGDPR:           maskGDPR,
		depositWallet:      depositWallet,
		changePassword:     changePassword,
		users:              users,
		wallets:            wallets,
		piiAccess:          piiAccess,
	}
}

func (s *Server) recordPIIAccess(ctx context.Context, subjectUserID, resource string, metadata map[string]any) error {
	if s.piiAccess == nil {
		return nil
	}
	return s.piiAccess.RecordPIIAccess(ctx, audit.PIIAccessEntry{
		SubjectUserID: subjectUserID,
		Resource:      resource,
		Metadata:      metadata,
	})
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
	if err := s.recordPIIAccess(ctx, profile.UserID, audit.PIIResourceProfile, nil); err != nil {
		return nil, err
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
	if err := s.recordPIIAccess(ctx, req.Params.XUserId.String(), audit.PIIResourceKYCStatus, nil); err != nil {
		return nil, err
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
	if err := s.recordPIIAccess(ctx, req.Params.XUserId.String(), audit.PIIResourceWalletTransactions, map[string]any{
		"count": len(views),
	}); err != nil {
		return nil, err
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
	if err := s.recordPIIAccess(ctx, req.Params.XUserId.String(), audit.PIIResourceWalletBalance, map[string]any{
		"currency": currency,
	}); err != nil {
		return nil, err
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

func (s *Server) ChangePassword(ctx context.Context, req api.ChangePasswordRequestObject) (api.ChangePasswordResponseObject, error) {
	if req.Body == nil {
		return api.ChangePassword400JSONResponse{Error: "invalid_json"}, nil
	}
	if err := s.changePassword.Execute(ctx, usecase.ChangePasswordInput{
		UserID:          req.Params.XUserId.String(),
		CurrentPassword: req.Body.CurrentPassword,
		NewPassword:     req.Body.NewPassword,
	}); err != nil {
		if err.Error() == "invalid current password" {
			return api.ChangePassword401JSONResponse{Error: err.Error()}, nil
		}
		return api.ChangePassword400JSONResponse{Error: err.Error()}, nil
	}
	return api.ChangePassword204Response{}, nil
}

func (s *Server) DepositWallet(ctx context.Context, req api.DepositWalletRequestObject) (api.DepositWalletResponseObject, error) {
	currency := "USD"
	if req.Body.Currency != nil {
		currency = *req.Body.Currency
	}
	out, err := s.depositWallet.Execute(ctx, usecase.DepositWalletInput{
		UserID:         req.Params.XUserId.String(),
		Amount:         req.Body.Amount,
		Currency:       currency,
		IdempotencyKey: req.Params.IdempotencyKey,
	})
	if err != nil {
		return api.DepositWallet400JSONResponse{Error: err.Error()}, nil
	}
	view, err := toDepositView(out.Deposit)
	if err != nil {
		return nil, err
	}
	if out.Created {
		return api.DepositWallet201JSONResponse(view), nil
	}
	return api.DepositWallet200JSONResponse(view), nil
}

func toDepositView(out domain.Deposit) (api.DepositWalletResponse, error) {
	id, err := uuid.Parse(out.ID)
	if err != nil {
		return api.DepositWalletResponse{}, err
	}
	walletID, err := uuid.Parse(out.WalletID)
	if err != nil {
		return api.DepositWalletResponse{}, err
	}
	view := api.DepositWalletResponse{
		Id:       openapi_types.UUID(id),
		WalletId: openapi_types.UUID(walletID),
		Amount:   out.Amount,
		Currency: out.Currency,
		Status:   string(out.Status),
	}
	if out.LedgerTransferID != "" {
		view.LedgerTransferId = &out.LedgerTransferID
	}
	if !out.CreatedAt.IsZero() {
		view.CreatedAt = &out.CreatedAt
	}
	return view, nil
}

func (s *Server) GetUserByEmail(ctx context.Context, req api.GetUserByEmailRequestObject) (api.GetUserByEmailResponseObject, error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetUserByEmail404JSONResponse{Error: "user_not_found"}, nil
		}
		return nil, err
	}
	if err := s.recordPIIAccess(ctx, user.ID, audit.PIIResourceInternalUserLookup, map[string]any{"lookup": "email"}); err != nil {
		return nil, err
	}
	view, err := toUserView(user)
	if err != nil {
		return nil, err
	}
	return api.GetUserByEmail200JSONResponse(view), nil
}

func (s *Server) GetInternalUser(ctx context.Context, req api.GetInternalUserRequestObject) (api.GetInternalUserResponseObject, error) {
	user, err := s.users.GetByID(ctx, req.UserId.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetInternalUser404JSONResponse{Error: "user_not_found"}, nil
		}
		return nil, err
	}
	if err := s.recordPIIAccess(ctx, user.ID, audit.PIIResourceInternalUserLookup, map[string]any{"lookup": "id"}); err != nil {
		return nil, err
	}
	view, err := toUserView(user)
	if err != nil {
		return nil, err
	}
	return api.GetInternalUser200JSONResponse(view), nil
}

func toUserView(user *domain.User) (api.User, error) {
	id, err := uuid.Parse(user.ID)
	if err != nil {
		return api.User{}, err
	}
	return api.User{
		Id:     openapi_types.UUID(id),
		Email:  user.Email,
		Phone:  user.Phone,
		Status: string(user.Status),
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
	if err := s.recordPIIAccess(ctx, user.ID, audit.PIIResourceUserByPhone, nil); err != nil {
		return nil, err
	}
	view, err := toUserView(user)
	if err != nil {
		return nil, err
	}
	return api.GetUserByPhone200JSONResponse(view), nil
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
	if err := s.recordPIIAccess(ctx, wallet.UserID, audit.PIIResourceInternalWallet, map[string]any{
		"currency": currency,
	}); err != nil {
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

func (s *Server) ExportUserGDPR(ctx context.Context, req api.ExportUserGDPRRequestObject) (api.ExportUserGDPRResponseObject, error) {
	userID := req.UserId.String()
	out, err := s.exportGDPR.Execute(ctx, userID)
	if err != nil {
		if err.Error() == "user not found" {
			return api.ExportUserGDPR404JSONResponse{Error: "user_not_found"}, nil
		}
		return nil, err
	}
	if err := s.recordPIIAccess(ctx, userID, audit.PIIResourceGDPRExport, nil); err != nil {
		return nil, err
	}

	profile := out.Profile
	userUUID, _ := uuid.Parse(profile.UserID)
	resp := api.GDPRExportResponse{
		UserId:                 openapi_types.UUID(userUUID),
		ExportedAt:             out.ExportedAt,
		WalletTransactionCount: out.WalletTransactionCount,
		Profile: api.Profile{
			UserId:    openapi_types.UUID(userUUID),
			Email:     profile.Email,
			Phone:     profile.Phone,
			Status:    profile.Status,
			KycStatus: profile.KYCStatus,
			CreatedAt: profile.CreatedAt,
		},
	}
	if profile.FullName != "" {
		resp.Profile.FullName = &profile.FullName
	}
	if profile.CountryCode != "" {
		resp.Profile.CountryCode = &profile.CountryCode
	}
	if profile.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", profile.DateOfBirth)
		if err == nil {
			resp.Profile.DateOfBirth = &openapi_types.Date{Time: dob}
		}
	}

	resp.KycSubmissions = make([]api.GDPRExportKYCSubmission, 0, len(out.KYCSubmissions))
	for _, sub := range out.KYCSubmissions {
		subID, _ := uuid.Parse(sub.ID)
		caseID, _ := uuid.Parse(sub.KYCCaseID)
		view := api.GDPRExportKYCSubmission{
			Id:                openapi_types.UUID(subID),
			KycCaseId:         openapi_types.UUID(caseID),
			Provider:          sub.Provider,
			ScreeningDecision: sub.ScreeningDecision,
			CreatedAt:         sub.CreatedAt.UTC(),
		}
		if sub.DocumentType != "" {
			view.DocumentType = &sub.DocumentType
		}
		if sub.DocumentNumber != "" {
			view.DocumentNumber = &sub.DocumentNumber
		}
		if sub.ScreeningReason != "" {
			view.ScreeningReason = &sub.ScreeningReason
		}
		resp.KycSubmissions = append(resp.KycSubmissions, view)
	}

	resp.Wallets = make([]api.Wallet, 0, len(out.Wallets))
	for _, wallet := range out.Wallets {
		walletID, _ := uuid.Parse(wallet.ID)
		walletUserID, _ := uuid.Parse(wallet.UserID)
		resp.Wallets = append(resp.Wallets, api.Wallet{
			Id:              openapi_types.UUID(walletID),
			UserId:          openapi_types.UUID(walletUserID),
			Currency:        wallet.Currency,
			LedgerAccountId: wallet.LedgerAccountID,
			Status:          wallet.Status,
		})
	}

	return api.ExportUserGDPR200JSONResponse(resp), nil
}

func (s *Server) MaskUserGDPR(ctx context.Context, req api.MaskUserGDPRRequestObject) (api.MaskUserGDPRResponseObject, error) {
	userID := req.UserId.String()
	if err := s.maskGDPR.Execute(ctx, userID); err != nil {
		if err.Error() == "user not found" {
			return api.MaskUserGDPR404JSONResponse{Error: "user_not_found"}, nil
		}
		return nil, err
	}
	userUUID, _ := uuid.Parse(userID)
	return api.MaskUserGDPR200JSONResponse{
		UserId: openapi_types.UUID(userUUID),
		Status: "masked",
	}, nil
}