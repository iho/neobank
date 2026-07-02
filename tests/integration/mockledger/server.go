// Package mockledger provides an in-memory gRPC implementation of goledger
// for integration tests without an external ledger process.
package mockledger

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/grpcutil"
	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type accountState struct {
	id         string
	name       string
	currency   string
	balance    decimal.Decimal
	encumbered decimal.Decimal
}

type Server struct {
	goledgerv1.UnimplementedAccountServiceServer
	goledgerv1.UnimplementedTransferServiceServer
	goledgerv1.UnimplementedHoldServiceServer

	mu sync.Mutex

	accounts             map[string]*accountState
	transfers            map[string]*goledgerv1.Transfer
	transferByIdempotency map[string]string
	holds                map[string]*goledgerv1.Hold
	holdsByAccount       map[string][]string

	createTransferCalls int

	grpcServer *grpc.Server
	listener   net.Listener
}

func New() *Server {
	return &Server{
		accounts:              make(map[string]*accountState),
		transfers:             make(map[string]*goledgerv1.Transfer),
		transferByIdempotency: make(map[string]string),
		holds:                 make(map[string]*goledgerv1.Hold),
		holdsByAccount:        make(map[string][]string),
	}
}

func (s *Server) Start() (addr string, err error) {
	s.listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	s.grpcServer = grpc.NewServer()
	goledgerv1.RegisterAccountServiceServer(s.grpcServer, s)
	goledgerv1.RegisterTransferServiceServer(s.grpcServer, s)
	goledgerv1.RegisterHoldServiceServer(s.grpcServer, s)
	go func() {
		_ = s.grpcServer.Serve(s.listener)
	}()
	return s.listener.Addr().String(), nil
}

func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
}

func (s *Server) CreateTransferCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.createTransferCalls
}

// CreditAccount sets an account balance for test setup.
func (s *Server) CreditAccount(accountID, amount string) error {
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	acct, ok := s.accounts[accountID]
	if !ok {
		return fmt.Errorf("account %s not found", accountID)
	}
	acct.balance = amt
	return nil
}

// CreateSettlementAccount provisions the merchant settlement account used by card capture.
func (s *Server) CreateSettlementAccount() (*goledgerv1.Account, error) {
	resp, err := s.CreateAccount(context.Background(), &goledgerv1.CreateAccountRequest{
		Name:                 "SETTLEMENT:USD",
		Currency:             "USD",
		AllowNegativeBalance: true,
		AllowPositiveBalance: true,
	})
	if err != nil {
		return nil, err
	}
	return resp.Account, nil
}

func (s *Server) CreateAccount(_ context.Context, req *goledgerv1.CreateAccountRequest) (*goledgerv1.CreateAccountResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.NewString()
	now := timestamppb.Now()
	acct := &goledgerv1.Account{
		Id:                   id,
		Name:                 req.Name,
		Currency:             req.Currency,
		Balance:              "0",
		EncumberedBalance:    "0",
		Version:              1,
		AllowNegativeBalance: req.AllowNegativeBalance,
		AllowPositiveBalance: req.AllowPositiveBalance,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	s.accounts[id] = &accountState{
		id:       id,
		name:     req.Name,
		currency: req.Currency,
	}
	return &goledgerv1.CreateAccountResponse{Account: acct}, nil
}

func (s *Server) GetAccount(_ context.Context, req *goledgerv1.GetAccountRequest) (*goledgerv1.GetAccountResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.accounts[req.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "account not found")
	}
	return &goledgerv1.GetAccountResponse{Account: s.accountProto(st)}, nil
}

func (s *Server) ListAccounts(context.Context, *goledgerv1.ListAccountsRequest) (*goledgerv1.ListAccountsResponse, error) {
	return &goledgerv1.ListAccountsResponse{}, nil
}

func (s *Server) CreateTransfer(ctx context.Context, req *goledgerv1.CreateTransferRequest) (*goledgerv1.CreateTransferResponse, error) {
	key := idempotencyKey(ctx, req.IdempotencyKey)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.createTransferCalls++

	if key != "" {
		if existingID, ok := s.transferByIdempotency[key]; ok {
			if tr, ok := s.transfers[existingID]; ok {
				return &goledgerv1.CreateTransferResponse{Transfer: tr}, nil
			}
		}
	}

	transfer, err := s.applyTransferLocked(req.FromAccountId, req.ToAccountId, req.Amount, req.Metadata)
	if err != nil {
		return nil, err
	}
	if key != "" {
		s.transferByIdempotency[key] = transfer.Id
	}
	return &goledgerv1.CreateTransferResponse{Transfer: transfer}, nil
}

func (s *Server) CreateBatchTransfer(context.Context, *goledgerv1.CreateBatchTransferRequest) (*goledgerv1.CreateBatchTransferResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *Server) GetTransfer(_ context.Context, req *goledgerv1.GetTransferRequest) (*goledgerv1.GetTransferResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tr, ok := s.transfers[req.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "transfer not found")
	}
	return &goledgerv1.GetTransferResponse{Transfer: tr}, nil
}

func (s *Server) ListTransfersByAccount(context.Context, *goledgerv1.ListTransfersByAccountRequest) (*goledgerv1.ListTransfersByAccountResponse, error) {
	return &goledgerv1.ListTransfersByAccountResponse{}, nil
}

func (s *Server) ReverseTransfer(_ context.Context, req *goledgerv1.ReverseTransferRequest) (*goledgerv1.ReverseTransferResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	orig, ok := s.transfers[req.TransferId]
	if !ok {
		return nil, status.Error(codes.NotFound, "transfer not found")
	}
	reversal, err := s.applyTransferLocked(orig.ToAccountId, orig.FromAccountId, orig.Amount, req.Metadata)
	if err != nil {
		return nil, err
	}
	return &goledgerv1.ReverseTransferResponse{Transfer: reversal}, nil
}

func (s *Server) HoldFunds(_ context.Context, req *goledgerv1.HoldFundsRequest) (*goledgerv1.HoldFundsResponse, error) {
	amt, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	st, ok := s.accounts[req.AccountId]
	if !ok {
		return nil, status.Error(codes.NotFound, "account not found")
	}
	available := st.balance.Sub(st.encumbered)
	if available.LessThan(amt) {
		return nil, status.Error(codes.FailedPrecondition, "insufficient funds")
	}
	st.encumbered = st.encumbered.Add(amt)

	holdID := uuid.NewString()
	now := timestamppb.Now()
	hold := &goledgerv1.Hold{
		Id:        holdID,
		AccountId: req.AccountId,
		Amount:    req.Amount,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.holds[holdID] = hold
	s.holdsByAccount[req.AccountId] = append(s.holdsByAccount[req.AccountId], holdID)
	return &goledgerv1.HoldFundsResponse{Hold: hold}, nil
}

func (s *Server) VoidHold(_ context.Context, req *goledgerv1.VoidHoldRequest) (*goledgerv1.VoidHoldResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hold, ok := s.holds[req.HoldId]
	if !ok {
		return nil, status.Error(codes.NotFound, "hold not found")
	}
	amt, _ := decimal.NewFromString(hold.Amount)
	if st, ok := s.accounts[hold.AccountId]; ok {
		st.encumbered = st.encumbered.Sub(amt)
	}
	hold.Status = "voided"
	hold.UpdatedAt = timestamppb.Now()
	return &goledgerv1.VoidHoldResponse{}, nil
}

func (s *Server) CaptureHold(_ context.Context, req *goledgerv1.CaptureHoldRequest) (*goledgerv1.CaptureHoldResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hold, ok := s.holds[req.HoldId]
	if !ok {
		return nil, status.Error(codes.NotFound, "hold not found")
	}
	if hold.Status != "active" {
		return nil, status.Error(codes.FailedPrecondition, "hold not active")
	}

	amt, err := decimal.NewFromString(hold.Amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid hold amount")
	}
	st, ok := s.accounts[hold.AccountId]
	if !ok {
		return nil, status.Error(codes.NotFound, "account not found")
	}
	st.encumbered = st.encumbered.Sub(amt)
	st.balance = st.balance.Sub(amt)

	if to, ok := s.accounts[req.ToAccountId]; ok {
		to.balance = to.balance.Add(amt)
	}

	hold.Status = "captured"
	hold.UpdatedAt = timestamppb.Now()

	transfer := &goledgerv1.Transfer{
		Id:            uuid.NewString(),
		FromAccountId: hold.AccountId,
		ToAccountId:   req.ToAccountId,
		Amount:        hold.Amount,
		CreatedAt:     timestamppb.Now(),
		EventAt:       timestamppb.Now(),
	}
	s.transfers[transfer.Id] = transfer
	return &goledgerv1.CaptureHoldResponse{Transfer: transfer}, nil
}

func (s *Server) ListHoldsByAccount(_ context.Context, req *goledgerv1.ListHoldsByAccountRequest) (*goledgerv1.ListHoldsByAccountResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := s.holdsByAccount[req.AccountId]
	holds := make([]*goledgerv1.Hold, 0, len(ids))
	for _, id := range ids {
		if h, ok := s.holds[id]; ok {
			holds = append(holds, h)
		}
	}
	limit := int(req.Limit)
	if limit <= 0 || limit > len(holds) {
		limit = len(holds)
	}
	return &goledgerv1.ListHoldsByAccountResponse{Holds: holds[:limit]}, nil
}

func (s *Server) applyTransferLocked(fromID, toID, amount string, metadata map[string]string) (*goledgerv1.Transfer, error) {
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid amount")
	}
	from, ok := s.accounts[fromID]
	if !ok {
		return nil, status.Error(codes.NotFound, "from account not found")
	}
	to, ok := s.accounts[toID]
	if !ok {
		return nil, status.Error(codes.NotFound, "to account not found")
	}
	available := from.balance.Sub(from.encumbered)
	if available.LessThan(amt) && !from.allowNegative() {
		return nil, status.Error(codes.FailedPrecondition, "insufficient funds")
	}
	from.balance = from.balance.Sub(amt)
	to.balance = to.balance.Add(amt)

	now := time.Now().UTC()
	transfer := &goledgerv1.Transfer{
		Id:            uuid.NewString(),
		FromAccountId: fromID,
		ToAccountId:   toID,
		Amount:        amount,
		CreatedAt:     timestamppb.New(now),
		EventAt:       timestamppb.New(now),
		Metadata:      metadata,
	}
	s.transfers[transfer.Id] = transfer
	return transfer, nil
}

func (s *Server) accountProto(st *accountState) *goledgerv1.Account {
	return &goledgerv1.Account{
		Id:                st.id,
		Name:              st.name,
		Currency:          st.currency,
		Balance:           st.balance.StringFixed(2),
		EncumberedBalance: st.encumbered.StringFixed(2),
		Version:           1,
		UpdatedAt:         timestamppb.Now(),
		CreatedAt:         timestamppb.Now(),
	}
}

func (st *accountState) allowNegative() bool {
	return false
}

func idempotencyKey(ctx context.Context, reqKey *string) string {
	if reqKey != nil && *reqKey != "" {
		return *reqKey
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get(grpcutil.IdempotencyKeyHeader)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}