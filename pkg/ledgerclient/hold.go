package ledgerclient

import (
	"context"

	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/iho/neobank/pkg/grpcutil"
)

type HoldFundsInput struct {
	AccountID string
	Amount    string
}

func (c *Client) HoldFunds(ctx context.Context, in HoldFundsInput) (*goledgerv1.Hold, error) {
	ctx, cancel := grpcutil.DefaultTimeout(ctx)
	defer cancel()

	resp, err := c.holds.HoldFunds(ctx, &goledgerv1.HoldFundsRequest{
		AccountId: in.AccountID,
		Amount:    in.Amount,
	})
	if err != nil {
		return nil, err
	}
	return resp.Hold, nil
}

func (c *Client) VoidHold(ctx context.Context, holdID string) error {
	ctx, cancel := grpcutil.DefaultTimeout(ctx)
	defer cancel()

	_, err := c.holds.VoidHold(ctx, &goledgerv1.VoidHoldRequest{HoldId: holdID})
	return err
}

// ListHoldsByAccount returns holds for an account, used by reconciliation to
// cross-check card.authorizations against goledger's view of open/settled holds.
func (c *Client) ListHoldsByAccount(ctx context.Context, accountID string, limit int) ([]*goledgerv1.Hold, error) {
	ctx, cancel := grpcutil.DefaultTimeout(ctx)
	defer cancel()

	resp, err := c.holds.ListHoldsByAccount(ctx, &goledgerv1.ListHoldsByAccountRequest{
		AccountId: accountID,
		Limit:     int32(limit),
	})
	if err != nil {
		return nil, err
	}
	return resp.Holds, nil
}

type CaptureHoldInput struct {
	HoldID      string
	ToAccountID string
}

func (c *Client) CaptureHold(ctx context.Context, in CaptureHoldInput) (*goledgerv1.Transfer, error) {
	ctx, cancel := grpcutil.DefaultTimeout(ctx)
	defer cancel()

	resp, err := c.holds.CaptureHold(ctx, &goledgerv1.CaptureHoldRequest{
		HoldId:      in.HoldID,
		ToAccountId: in.ToAccountID,
	})
	if err != nil {
		return nil, err
	}
	return resp.Transfer, nil
}
