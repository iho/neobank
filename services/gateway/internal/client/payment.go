package client

import (
	"context"
	"fmt"

	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	"google.golang.org/grpc"
)

type PaymentClient struct {
	conn *grpc.ClientConn
	rpc  neobankv1.PaymentServiceClient
}

func NewPaymentClient(ctx context.Context, cfg Config) (*PaymentClient, error) {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:50053"
	}
	conn, err := grpcutil.Dial(ctx, cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("dial payment service: %w", err)
	}
	return &PaymentClient{
		conn: conn,
		rpc:  neobankv1.NewPaymentServiceClient(conn),
	}, nil
}

func (c *PaymentClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

type CreateTransferRequest struct {
	RecipientPhone  string `json:"recipient_phone,omitempty"`
	RecipientEmail  string `json:"recipient_email,omitempty"`
	RecipientUserID string `json:"recipient_user_id,omitempty"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	Memo            string `json:"memo"`
}

type TransferView struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	SenderUserID     string `json:"sender_user_id"`
	RecipientUserID  string `json:"recipient_user_id"`
	Amount           string `json:"amount"`
	Currency         string `json:"currency"`
	LedgerTransferID string `json:"ledger_transfer_id,omitempty"`
	FailureReason    string `json:"failure_reason,omitempty"`
	Memo             string `json:"memo,omitempty"`
	CreatedAt        string `json:"created_at,omitempty"`
	CompletedAt      string `json:"completed_at,omitempty"`
}

type TransferList struct {
	Transfers  []TransferView `json:"transfers"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

type LimitGaugeView struct {
	Limit     string `json:"limit"`
	Used      string `json:"used"`
	Remaining string `json:"remaining"`
}

type TransferLimitsView struct {
	HourlyTransferCount LimitGaugeView `json:"hourly_transfer_count"`
	DailyTransferAmount LimitGaugeView `json:"daily_transfer_amount"`
	SingleTransferMax   string         `json:"single_transfer_max"`
}

type LimitsResponse struct {
	P2P TransferLimitsView `json:"p2p"`
}

func (c *PaymentClient) ListTransfers(ctx context.Context, userID string, limit int, cursor string) (TransferList, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.ListTransfers(ctx, &neobankv1.ListTransfersRequest{
		UserId: userID,
		Limit:  int32(limit),
		Cursor: cursor,
	})
	if err != nil {
		return TransferList{}, 0, dialError("payment", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return TransferList{}, status, statusError("payment", status, resp.GetError())
	}
	out := TransferList{NextCursor: resp.GetNextCursor()}
	for _, t := range resp.GetTransfers() {
		out.Transfers = append(out.Transfers, toTransferView(t))
	}
	return out, status, nil
}

func (c *PaymentClient) GetLimits(ctx context.Context, userID string) (LimitsResponse, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.GetLimits(ctx, &neobankv1.GetLimitsRequest{UserId: userID})
	if err != nil {
		return LimitsResponse{}, 0, dialError("payment", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return LimitsResponse{}, status, statusError("payment", status, resp.GetError())
	}
	return LimitsResponse{P2P: toTransferLimitsView(resp.GetP2P())}, status, nil
}

func (c *PaymentClient) GetTransfer(ctx context.Context, userID, transferID string) (TransferView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	resp, err := c.rpc.GetTransfer(ctx, &neobankv1.GetTransferRequest{
		UserId:     userID,
		TransferId: transferID,
	})
	if err != nil {
		return TransferView{}, 0, dialError("payment", err)
	}
	status := int(resp.GetHttpStatus())
	if status != 200 {
		return TransferView{}, status, statusError("payment", status, resp.GetError())
	}
	return toTransferView(resp.GetTransfer()), status, nil
}

func (c *PaymentClient) CreateP2PTransfer(ctx context.Context, userID, idempotencyKey string, req CreateTransferRequest) (TransferView, int, error) {
	ctx = grpcutil.WithUserID(ctx, userID)
	ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	resp, err := c.rpc.CreateP2PTransfer(ctx, &neobankv1.CreateP2PTransferRequest{
		UserId:          userID,
		RecipientPhone:  req.RecipientPhone,
		RecipientEmail:  req.RecipientEmail,
		RecipientUserId: req.RecipientUserID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Memo:            req.Memo,
		IdempotencyKey:  idempotencyKey,
	})
	if err != nil {
		return TransferView{}, 0, dialError("payment", err)
	}
	status := int(resp.GetHttpStatus())
	out := toTransferView(resp.GetTransfer())
	if status >= 400 {
		return out, status, statusError("payment", status, "")
	}
	return out, status, nil
}

func toTransferView(t *neobankv1.Transfer) TransferView {
	if t == nil {
		return TransferView{}
	}
	return TransferView{
		ID:               t.GetId(),
		Status:           t.GetStatus(),
		SenderUserID:     t.GetSenderUserId(),
		RecipientUserID:  t.GetRecipientUserId(),
		Amount:           t.GetAmount(),
		Currency:         t.GetCurrency(),
		LedgerTransferID: t.GetLedgerTransferId(),
		FailureReason:    t.GetFailureReason(),
		Memo:             t.GetMemo(),
		CreatedAt:        t.GetCreatedAt(),
		CompletedAt:      t.GetCompletedAt(),
	}
}

func toLimitGaugeView(g *neobankv1.LimitGauge) LimitGaugeView {
	if g == nil {
		return LimitGaugeView{}
	}
	return LimitGaugeView{
		Limit:     g.GetLimit(),
		Used:      g.GetUsed(),
		Remaining: g.GetRemaining(),
	}
}

func toTransferLimitsView(l *neobankv1.TransferLimits) TransferLimitsView {
	if l == nil {
		return TransferLimitsView{}
	}
	return TransferLimitsView{
		HourlyTransferCount: toLimitGaugeView(l.GetHourlyTransferCount()),
		DailyTransferAmount: toLimitGaugeView(l.GetDailyTransferAmount()),
		SingleTransferMax:   l.GetSingleTransferMax(),
	}
}