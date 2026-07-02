package client

import (
	"context"
	"fmt"

	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	"google.golang.org/grpc"
)

type UserClient struct {
	conn *grpc.ClientConn
	rpc  neobankv1.UserServiceClient
}

func NewUserClient(ctx context.Context, cfg Config) (*UserClient, error) {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:50052"
	}
	conn, err := grpcutil.Dial(ctx, cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("dial user service: %w", err)
	}
	return &UserClient{
		conn: conn,
		rpc:  neobankv1.NewUserServiceClient(conn),
	}, nil
}

func (c *UserClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

type RegisterRequest struct {
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	InviteCode string `json:"invite_code,omitempty"`
}

type RegisterResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SubmitKYCRequest struct {
	FullName       string `json:"full_name"`
	DateOfBirth    string `json:"date_of_birth"`
	CountryCode    string `json:"country_code"`
	DocumentType   string `json:"document_type,omitempty"`
	DocumentNumber string `json:"document_number,omitempty"`
}

type SubmitKYCResponse struct {
	KYCCaseID       string `json:"kyc_case_id"`
	Status          string `json:"status"`
	WalletID        string `json:"wallet_id,omitempty"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type KYCStatusResponse struct {
	Status          string `json:"status"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type WalletBalance struct {
	WalletID          string `json:"wallet_id"`
	LedgerAccountID   string `json:"ledger_account_id,omitempty"`
	Currency          string `json:"currency"`
	Balance           string `json:"balance"`
	EncumberedBalance string `json:"encumbered_balance,omitempty"`
	AvailableBalance  string `json:"available_balance"`
}

type ProvisionWalletResponse struct {
	WalletID        string `json:"wallet_id"`
	LedgerAccountID string `json:"ledger_account_id"`
}

type WalletTransactionView struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	Direction    string `json:"direction"`
	Status       string `json:"status"`
	Counterparty string `json:"counterparty,omitempty"`
	Memo         string `json:"memo,omitempty"`
	ReferenceID  string `json:"reference_id,omitempty"`
	CreatedAt    string `json:"created_at"`
}

type WalletTransactionList struct {
	Transactions []WalletTransactionView `json:"transactions"`
	NextCursor   string                  `json:"next_cursor,omitempty"`
}

type PayeeView struct {
	ID          string `json:"id"`
	PayeeUserID string `json:"payee_user_id"`
	Nickname    string `json:"nickname,omitempty"`
	PayeeEmail  string `json:"payee_email,omitempty"`
	PayeePhone  string `json:"payee_phone,omitempty"`
	LastUsedAt  string `json:"last_used_at"`
	CreatedAt   string `json:"created_at"`
}

type PayeeList struct {
	Payees []PayeeView `json:"payees"`
}

type ProfileView struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Status      string `json:"status"`
	FullName    string `json:"full_name,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	KYCStatus   string `json:"kyc_status"`
	CreatedAt   string `json:"created_at"`
}

func (c *UserClient) RefreshToken(ctx context.Context, refreshToken string) (LoginResponse, error) {
	resp, err := c.rpc.RefreshToken(ctx, &neobankv1.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return LoginResponse{}, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return LoginResponse{}, statusError("user", status, resp.GetError())
	}
	return toLoginResponse(resp), nil
}

func (c *UserClient) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	resp, err := c.rpc.Login(ctx, &neobankv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return LoginResponse{}, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return LoginResponse{}, statusError("user", status, resp.GetError())
	}
	return toLoginResponse(resp), nil
}

func (c *UserClient) Register(ctx context.Context, idempotencyKey string, req RegisterRequest) (RegisterResponse, error) {
	ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	resp, err := c.rpc.Register(ctx, &neobankv1.RegisterRequest{
		Email:          req.Email,
		Phone:          req.Phone,
		Password:       req.Password,
		InviteCode:     req.InviteCode,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return RegisterResponse{}, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 201 {
		return RegisterResponse{}, statusError("user", status, resp.GetError())
	}
	return RegisterResponse{
		UserID:       resp.GetUserId(),
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	}, nil
}

func (c *UserClient) SubmitKYC(ctx context.Context, userID, idempotencyKey string, req SubmitKYCRequest) (SubmitKYCResponse, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	resp, err := c.rpc.SubmitKYC(ctx, &neobankv1.SubmitKYCRequest{
		UserId:         userID,
		FullName:       req.FullName,
		DateOfBirth:    req.DateOfBirth,
		CountryCode:    req.CountryCode,
		DocumentType:   req.DocumentType,
		DocumentNumber: req.DocumentNumber,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return SubmitKYCResponse{}, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return SubmitKYCResponse{}, statusError("user", status, resp.GetError())
	}
	return SubmitKYCResponse{
		KYCCaseID:       resp.GetKycCaseId(),
		Status:          resp.GetStatus(),
		WalletID:        resp.GetWalletId(),
		RejectionReason: resp.GetRejectionReason(),
	}, nil
}

func (c *UserClient) GetProfile(ctx context.Context, userID string) (ProfileView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.GetProfile(ctx, &neobankv1.GetProfileRequest{UserId: userID})
	if err != nil {
		return ProfileView{}, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return ProfileView{}, status, statusError("user", status, resp.GetError())
	}
	return toProfileView(resp.GetProfile()), status, nil
}

func (c *UserClient) GetKYCStatus(ctx context.Context, userID string) (KYCStatusResponse, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.GetKYCStatus(ctx, &neobankv1.GetKYCStatusRequest{UserId: userID})
	if err != nil {
		return KYCStatusResponse{}, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return KYCStatusResponse{}, statusError("user", status, resp.GetError())
	}
	return KYCStatusResponse{
		Status:          resp.GetStatus(),
		RejectionReason: resp.GetRejectionReason(),
	}, nil
}

func (c *UserClient) GetWalletBalance(ctx context.Context, userID, currency string) (WalletBalance, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.GetWalletBalance(ctx, &neobankv1.GetWalletBalanceRequest{
		UserId:   userID,
		Currency: currency,
	})
	if err != nil {
		return WalletBalance{}, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return WalletBalance{}, status, statusError("user", status, resp.GetError())
	}
	return toWalletBalance(resp.GetBalance()), status, nil
}

func (c *UserClient) ListWalletTransactions(ctx context.Context, userID string, limit int, cursor string) (WalletTransactionList, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.ListWalletTransactions(ctx, &neobankv1.ListWalletTransactionsRequest{
		UserId: userID,
		Limit:  int32(limit),
		Cursor: cursor,
	})
	if err != nil {
		return WalletTransactionList{}, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return WalletTransactionList{}, status, statusError("user", status, resp.GetError())
	}
	out := WalletTransactionList{NextCursor: resp.GetNextCursor()}
	for _, tx := range resp.GetTransactions() {
		out.Transactions = append(out.Transactions, toWalletTransactionView(tx))
	}
	return out, status, nil
}

func (c *UserClient) ProvisionWallet(ctx context.Context, userID, idempotencyKey, currency string) (ProvisionWalletResponse, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	resp, err := c.rpc.ProvisionWallet(ctx, &neobankv1.ProvisionWalletRequest{
		UserId:         userID,
		Currency:       currency,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return ProvisionWalletResponse{}, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 201 && status != 200 {
		return ProvisionWalletResponse{}, statusError("user", status, resp.GetError())
	}
	return ProvisionWalletResponse{
		WalletID:        resp.GetWalletId(),
		LedgerAccountID: resp.GetLedgerAccountId(),
	}, nil
}

type DepositWalletRequest struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency,omitempty"`
}

type DepositWalletResponse struct {
	ID               string `json:"id"`
	WalletID         string `json:"wallet_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	LedgerTransferID string `json:"ledger_transfer_id,omitempty"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at,omitempty"`
}

func (c *UserClient) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) (int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.ChangePassword(ctx, &neobankv1.ChangePasswordRequest{
		UserId:          userID,
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
	})
	if err != nil {
		return 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 204 {
		return status, statusError("user", status, resp.GetError())
	}
	return status, nil
}

func (c *UserClient) ListPayees(ctx context.Context, userID string, limit int) (PayeeList, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.ListPayees(ctx, &neobankv1.ListPayeesRequest{
		UserId: userID,
		Limit:  int32(limit),
	})
	if err != nil {
		return PayeeList{}, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return PayeeList{}, status, statusError("user", status, resp.GetError())
	}
	out := PayeeList{}
	for _, p := range resp.GetPayees() {
		out.Payees = append(out.Payees, toPayeeView(p))
	}
	return out, status, nil
}

func (c *UserClient) CreatePayee(ctx context.Context, userID, payeeUserID, nickname string) (PayeeView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.CreatePayee(ctx, &neobankv1.CreatePayeeRequest{
		UserId:      userID,
		PayeeUserId: payeeUserID,
		Nickname:    nickname,
	})
	if err != nil {
		return PayeeView{}, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 201 {
		return PayeeView{}, status, statusError("user", status, resp.GetError())
	}
	return toPayeeView(resp.GetPayee()), status, nil
}

func (c *UserClient) DeletePayee(ctx context.Context, userID, payeeID string) (int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.DeletePayee(ctx, &neobankv1.DeletePayeeRequest{
		UserId:  userID,
		PayeeId: payeeID,
	})
	if err != nil {
		return 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 204 {
		return status, statusError("user", status, resp.GetError())
	}
	return status, nil
}

type DeviceTokenView struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Platform  string `json:"platform"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

type RegisterDeviceTokenRequest struct {
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

func (c *UserClient) RegisterDeviceToken(ctx context.Context, userID, idempotencyKey string, req RegisterDeviceTokenRequest) (DeviceTokenView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	if idempotencyKey != "" {
		ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	}
	resp, err := c.rpc.RegisterDeviceToken(ctx, &neobankv1.RegisterDeviceTokenRequest{
		UserId:         userID,
		Platform:       req.Platform,
		Token:          req.Token,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return DeviceTokenView{}, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 201 {
		return DeviceTokenView{}, status, statusError("user", status, resp.GetError())
	}
	return toDeviceTokenView(resp.GetDeviceToken()), status, nil
}

func (c *UserClient) DeleteDeviceToken(ctx context.Context, userID, tokenID, idempotencyKey string) (int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	if idempotencyKey != "" {
		ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	}
	resp, err := c.rpc.DeleteDeviceToken(ctx, &neobankv1.DeleteDeviceTokenRequest{
		UserId:         userID,
		TokenId:        tokenID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 204 {
		return status, statusError("user", status, resp.GetError())
	}
	return status, nil
}

func (c *UserClient) CloseAccount(ctx context.Context, userID, idempotencyKey string) (int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	resp, err := c.rpc.CloseAccount(ctx, &neobankv1.CloseAccountRequest{
		UserId:         userID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 204 {
		return status, statusError("user", status, resp.GetError())
	}
	return status, nil
}

type WalletBalanceList struct {
	Wallets []WalletBalance `json:"wallets"`
}

type ReferralInviteView struct {
	ID            string `json:"id"`
	InviterUserID string `json:"inviter_user_id"`
	InviteCode    string `json:"invite_code"`
	InviteeUserID string `json:"invitee_user_id,omitempty"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	AcceptedAt    string `json:"accepted_at,omitempty"`
}

type ReferralInviteList struct {
	Invites []ReferralInviteView `json:"invites"`
}

func (c *UserClient) ExportWalletTransactions(ctx context.Context, userID, format string, from, to string) ([]byte, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.ExportWalletTransactions(ctx, &neobankv1.ExportWalletTransactionsRequest{
		UserId: userID,
		Format: format,
		From:   from,
		To:     to,
	})
	if err != nil {
		return nil, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return nil, status, statusError("user", status, resp.GetError())
	}
	return resp.GetCsvData(), status, nil
}

func (c *UserClient) ListWallets(ctx context.Context, userID string) (WalletBalanceList, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.ListWallets(ctx, &neobankv1.ListWalletsRequest{UserId: userID})
	if err != nil {
		return WalletBalanceList{}, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return WalletBalanceList{}, status, statusError("user", status, resp.GetError())
	}
	out := WalletBalanceList{}
	for _, w := range resp.GetWallets() {
		out.Wallets = append(out.Wallets, toWalletBalance(w))
	}
	return out, status, nil
}

func (c *UserClient) CreateReferralInvite(ctx context.Context, userID, idempotencyKey string) (ReferralInviteView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	if idempotencyKey != "" {
		ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	}
	resp, err := c.rpc.CreateReferralInvite(ctx, &neobankv1.CreateReferralInviteRequest{
		UserId:         userID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return ReferralInviteView{}, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 201 {
		return ReferralInviteView{}, status, statusError("user", status, resp.GetError())
	}
	return toReferralInviteView(resp.GetInvite()), status, nil
}

func (c *UserClient) ListReferralInvites(ctx context.Context, userID string, limit int) (ReferralInviteList, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.ListReferralInvites(ctx, &neobankv1.ListReferralInvitesRequest{
		UserId: userID,
		Limit:  int32(limit),
	})
	if err != nil {
		return ReferralInviteList{}, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return ReferralInviteList{}, status, statusError("user", status, resp.GetError())
	}
	out := ReferralInviteList{}
	for _, inv := range resp.GetInvites() {
		out.Invites = append(out.Invites, toReferralInviteView(inv))
	}
	return out, status, nil
}

func (c *UserClient) DepositWallet(ctx context.Context, userID, idempotencyKey string, req DepositWalletRequest) (DepositWalletResponse, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	resp, err := c.rpc.DepositWallet(ctx, &neobankv1.DepositWalletRequest{
		UserId:         userID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return DepositWalletResponse{}, 0, dialError("user", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 201 && status != 200 {
		return DepositWalletResponse{}, status, statusError("user", status, resp.GetError())
	}
	d := resp.GetDeposit()
	return DepositWalletResponse{
		ID:               d.GetId(),
		WalletID:         d.GetWalletId(),
		Amount:           d.GetAmount(),
		Currency:         d.GetCurrency(),
		LedgerTransferID: d.GetLedgerTransferId(),
		Status:           d.GetStatus(),
		CreatedAt:        d.GetCreatedAt(),
	}, status, nil
}

func toLoginResponse(resp *neobankv1.AuthResponse) LoginResponse {
	if resp == nil {
		return LoginResponse{}
	}
	return LoginResponse{
		UserID:       resp.GetUserId(),
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	}
}

func toProfileView(p *neobankv1.Profile) ProfileView {
	if p == nil {
		return ProfileView{}
	}
	return ProfileView{
		UserID:      p.GetUserId(),
		Email:       p.GetEmail(),
		Phone:       p.GetPhone(),
		Status:      p.GetStatus(),
		FullName:    p.GetFullName(),
		DateOfBirth: p.GetDateOfBirth(),
		CountryCode: p.GetCountryCode(),
		KYCStatus:   p.GetKycStatus(),
		CreatedAt:   p.GetCreatedAt(),
	}
}

func toWalletBalance(w *neobankv1.WalletBalance) WalletBalance {
	if w == nil {
		return WalletBalance{}
	}
	return WalletBalance{
		WalletID:          w.GetWalletId(),
		LedgerAccountID:   w.GetLedgerAccountId(),
		Currency:          w.GetCurrency(),
		Balance:           w.GetBalance(),
		EncumberedBalance: w.GetEncumberedBalance(),
		AvailableBalance:  w.GetAvailableBalance(),
	}
}

func toWalletTransactionView(tx *neobankv1.WalletTransaction) WalletTransactionView {
	if tx == nil {
		return WalletTransactionView{}
	}
	return WalletTransactionView{
		ID:           tx.GetId(),
		Type:         tx.GetType(),
		Amount:       tx.GetAmount(),
		Currency:     tx.GetCurrency(),
		Direction:    tx.GetDirection(),
		Status:       tx.GetStatus(),
		Counterparty: tx.GetCounterparty(),
		Memo:         tx.GetMemo(),
		ReferenceID:  tx.GetReferenceId(),
		CreatedAt:    tx.GetCreatedAt(),
	}
}

func toPayeeView(p *neobankv1.Payee) PayeeView {
	if p == nil {
		return PayeeView{}
	}
	return PayeeView{
		ID:          p.GetId(),
		PayeeUserID: p.GetPayeeUserId(),
		Nickname:    p.GetNickname(),
		PayeeEmail:  p.GetPayeeEmail(),
		PayeePhone:  p.GetPayeePhone(),
		LastUsedAt:  p.GetLastUsedAt(),
		CreatedAt:   p.GetCreatedAt(),
	}
}

func toDeviceTokenView(t *neobankv1.DeviceToken) DeviceTokenView {
	if t == nil {
		return DeviceTokenView{}
	}
	return DeviceTokenView{
		ID:        t.GetId(),
		UserID:    t.GetUserId(),
		Platform:  t.GetPlatform(),
		Token:     t.GetToken(),
		CreatedAt: t.GetCreatedAt(),
	}
}

func toReferralInviteView(inv *neobankv1.ReferralInvite) ReferralInviteView {
	if inv == nil {
		return ReferralInviteView{}
	}
	return ReferralInviteView{
		ID:            inv.GetId(),
		InviterUserID: inv.GetInviterUserId(),
		InviteCode:    inv.GetInviteCode(),
		InviteeUserID: inv.GetInviteeUserId(),
		Status:        inv.GetStatus(),
		CreatedAt:     inv.GetCreatedAt(),
		AcceptedAt:    inv.GetAcceptedAt(),
	}
}