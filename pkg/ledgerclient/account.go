//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	return resp.GetAccount(), nil
}

func (c *Client) GetAccount(ctx context.Context, id string) (*goledgerv1.Account, error) {
	ctx, cancel := grpcutil.DefaultTimeout(ctx)
	defer cancel()

	resp, err := c.accounts.GetAccount(ctx, &goledgerv1.GetAccountRequest{Id: id})
	if err != nil {
		return nil, err
	}

	return resp.GetAccount(), nil
}
