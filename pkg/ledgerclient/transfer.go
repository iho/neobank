package ledgerclient

import (
	"context"

	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateTransferInput struct {
	FromAccountID  string
	ToAccountID    string
	Amount         string
	IdempotencyKey string
	Metadata       map[string]string
}

func (c *Client) CreateTransfer(ctx context.Context, in CreateTransferInput) (*goledgerv1.Transfer, error) {
	ctx, cancel := grpcutil.DefaultTimeout(ctx)
	defer cancel()

	if in.IdempotencyKey != "" {
		ctx = grpcutil.WithIdempotencyKey(ctx, in.IdempotencyKey)
	}

	req := &goledgerv1.CreateTransferRequest{
		FromAccountId: in.FromAccountID,
		ToAccountId:   in.ToAccountID,
		Amount:        in.Amount,
		Metadata:      in.Metadata,
	}
	if in.IdempotencyKey != "" {
		req.IdempotencyKey = &in.IdempotencyKey
	}

	resp, err := c.transfers.CreateTransfer(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Transfer, nil
}

// GetTransfer looks up a ledger transfer by ID, returning (nil, nil) if it does not exist.
func (c *Client) GetTransfer(ctx context.Context, id string) (*goledgerv1.Transfer, error) {
	ctx, cancel := grpcutil.DefaultTimeout(ctx)
	defer cancel()

	resp, err := c.transfers.GetTransfer(ctx, &goledgerv1.GetTransferRequest{Id: id})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	return resp.Transfer, nil
}

func (c *Client) ReverseTransfer(ctx context.Context, transferID string, metadata map[string]string) (*goledgerv1.Transfer, error) {
	ctx, cancel := grpcutil.DefaultTimeout(ctx)
	defer cancel()

	resp, err := c.transfers.ReverseTransfer(ctx, &goledgerv1.ReverseTransferRequest{
		TransferId: transferID,
		Metadata:   metadata,
	})
	if err != nil {
		return nil, err
	}
	return resp.Transfer, nil
}
