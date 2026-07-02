package grpc

import (
	"context"
	"errors"

	"github.com/iho/neobank/pkg/audit"
	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	sqlcrepo "github.com/iho/neobank/services/user/internal/adapter/sqlc"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/usecase"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	neobankv1.UnimplementedUserInternalServiceServer

	users            *sqlcrepo.UserRepository
	wallets          *sqlcrepo.WalletRepository
	listDeviceTokens *usecase.ListDeviceTokensUseCase
	upsertPayee      *usecase.UpsertPayeeUseCase
	piiAccess        audit.PIIAccessRecorder
}

func NewServer(
	users *sqlcrepo.UserRepository,
	wallets *sqlcrepo.WalletRepository,
	listDeviceTokens *usecase.ListDeviceTokensUseCase,
	upsertPayee *usecase.UpsertPayeeUseCase,
	piiAccess audit.PIIAccessRecorder,
) *Server {
	return &Server{
		users:            users,
		wallets:          wallets,
		listDeviceTokens: listDeviceTokens,
		upsertPayee:      upsertPayee,
		piiAccess:        piiAccess,
	}
}

func (s *Server) GetUserByID(ctx context.Context, req *neobankv1.GetUserByIDRequest) (*neobankv1.GetUserResponse, error) {
	user, err := s.users.GetByID(ctx, req.GetUserId())
	if err != nil {
		return nil, mapUserLookupError(err)
	}
	if err := s.recordPIIAccess(ctx, user.ID, audit.PIIResourceInternalUserLookup, map[string]any{"lookup": "id"}); err != nil {
		return nil, status.Errorf(codes.Internal, "record pii access: %v", err)
	}
	return &neobankv1.GetUserResponse{User: toInternalUser(user)}, nil
}

func (s *Server) GetUserByPhone(ctx context.Context, req *neobankv1.GetUserByPhoneRequest) (*neobankv1.GetUserResponse, error) {
	user, err := s.users.GetByPhone(ctx, req.GetPhone())
	if err != nil {
		return nil, mapUserLookupError(err)
	}
	if err := s.recordPIIAccess(ctx, user.ID, audit.PIIResourceUserByPhone, nil); err != nil {
		return nil, status.Errorf(codes.Internal, "record pii access: %v", err)
	}
	return &neobankv1.GetUserResponse{User: toInternalUser(user)}, nil
}

func (s *Server) GetUserByEmail(ctx context.Context, req *neobankv1.GetUserByEmailRequest) (*neobankv1.GetUserResponse, error) {
	user, err := s.users.GetByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, mapUserLookupError(err)
	}
	if err := s.recordPIIAccess(ctx, user.ID, audit.PIIResourceInternalUserLookup, map[string]any{"lookup": "email"}); err != nil {
		return nil, status.Errorf(codes.Internal, "record pii access: %v", err)
	}
	return &neobankv1.GetUserResponse{User: toInternalUser(user)}, nil
}

func (s *Server) GetWallet(ctx context.Context, req *neobankv1.GetWalletRequest) (*neobankv1.GetWalletResponse, error) {
	currency := req.GetCurrency()
	if currency == "" {
		currency = "USD"
	}
	wallet, err := s.wallets.GetByUserAndCurrency(ctx, req.GetUserId(), currency)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "wallet_not_found")
		}
		return nil, status.Errorf(codes.Internal, "get wallet: %v", err)
	}
	if err := s.recordPIIAccess(ctx, wallet.UserID, audit.PIIResourceInternalWallet, map[string]any{
		"currency": currency,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "record pii access: %v", err)
	}
	return &neobankv1.GetWalletResponse{Wallet: toInternalWallet(wallet)}, nil
}

func (s *Server) ListDeviceTokens(ctx context.Context, req *neobankv1.ListDeviceTokensRequest) (*neobankv1.ListDeviceTokensResponse, error) {
	tokens, err := s.listDeviceTokens.Execute(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list device tokens: %v", err)
	}
	out := make([]*neobankv1.InternalDeviceToken, 0, len(tokens))
	for _, token := range tokens {
		out = append(out, &neobankv1.InternalDeviceToken{
			Id:       token.ID,
			UserId:   token.UserID,
			Platform: token.Platform,
			Token:    token.Token,
		})
	}
	return &neobankv1.ListDeviceTokensResponse{DeviceTokens: out}, nil
}

func (s *Server) UpsertPayee(ctx context.Context, req *neobankv1.UpsertPayeeRequest) (*neobankv1.UpsertPayeeResponse, error) {
	idempotencyKey := req.GetIdempotencyKey()
	if idempotencyKey == "" {
		idempotencyKey = grpcutil.IdempotencyKeyFromContext(ctx)
	}
	payee, err := s.upsertPayee.Execute(ctx, req.GetUserId(), req.GetPayeeUserId(), req.GetNickname())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	_ = idempotencyKey
	return &neobankv1.UpsertPayeeResponse{
		Id:          payee.ID,
		PayeeUserId: payee.PayeeUserID,
	}, nil
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

func mapUserLookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return status.Error(codes.NotFound, "user_not_found")
	}
	return status.Errorf(codes.Internal, "get user: %v", err)
}

func toInternalUser(user *domain.User) *neobankv1.InternalUser {
	return &neobankv1.InternalUser{
		Id:     user.ID,
		Email:  user.Email,
		Phone:  user.Phone,
		Status: string(user.Status),
	}
}

func toInternalWallet(wallet *domain.Wallet) *neobankv1.InternalWallet {
	return &neobankv1.InternalWallet{
		Id:              wallet.ID,
		UserId:          wallet.UserID,
		Currency:        wallet.Currency,
		LedgerAccountId: wallet.LedgerAccountID,
		Status:          wallet.Status,
	}
}