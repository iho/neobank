package ledger

import (
	"context"

	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/iho/neobank/pkg/ledgerclient"
)

type Adapter struct {
	client *ledgerclient.Client
}

func New(client *ledgerclient.Client) *Adapter {
	return &Adapter{client: client}
}

func (a *Adapter) CreateAccount(ctx context.Context, in ledgerclient.CreateAccountInput) (*goledgerv1.Account, error) {
	if a == nil || a.client == nil {
		return nil, ledgerclient.ErrUnavailable
	}
	return a.client.CreateAccount(ctx, in)
}

func (a *Adapter) GetAccount(ctx context.Context, id string) (*goledgerv1.Account, error) {
	if a == nil || a.client == nil {
		return nil, ledgerclient.ErrUnavailable
	}
	return a.client.GetAccount(ctx, id)
}