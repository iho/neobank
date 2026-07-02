package ledgerclient

import (
	"context"

	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/iho/neobank/pkg/grpcutil"
)

type CreateAccountInput struct {
	Name                 string
	Currency             string
	AllowNegativeBalance bool
	AllowPositiveBalance bool
}

func (c *Client) CreateAccount(ctx context.Context, in CreateAccountInput) (*goledgerv1.Account, error) {
	ctx, cancel := grpcutil.DefaultTimeout(ctx)
	defer cancel()

	resp, err := c.accounts.CreateAccount(ctx, &goledgerv1.CreateAccountRequest{
		Name:                 in.Name,
		Currency:             in.Currency,
		AllowNegativeBalance: in.AllowNegativeBalance,
		AllowPositiveBalance: in.AllowPositiveBalance,
	})
	if err != nil {
		return nil, err
	}
	return resp.Account, nil
}

func (c *Client) GetAccount(ctx context.Context, id string) (*goledgerv1.Account, error) {
	ctx, cancel := grpcutil.DefaultTimeout(ctx)
	defer cancel()

	resp, err := c.accounts.GetAccount(ctx, &goledgerv1.GetAccountRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.Account, nil
}