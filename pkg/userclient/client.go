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

package userclient

import (
	"context"
	"fmt"

	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/pkg/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	Addr string
}

type Client struct {
	conn *grpc.ClientConn
	rpc  neobankv1.UserInternalServiceClient
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:50052"
	}
	conn, err := grpcutil.Dial(ctx, cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("dial user service: %w", err)
	}
	return &Client{
		conn: conn,
		rpc:  neobankv1.NewUserInternalServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

type User struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

type Wallet struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	Currency        string `json:"currency"`
	LedgerAccountID string `json:"ledger_account_id"`
	Status          string `json:"status"`
}

type DeviceToken struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Platform string `json:"platform"`
	Token    string `json:"token"`
}

func (c *Client) GetByPhone(ctx context.Context, phone string) (User, error) {
	resp, err := c.rpc.GetUserByPhone(ctx, &neobankv1.GetUserByPhoneRequest{Phone: phone})
	if err != nil {
		return User{}, mapError(err)
	}
	return toUser(resp.GetUser()), nil
}

func (c *Client) GetByEmail(ctx context.Context, email string) (User, error) {
	resp, err := c.rpc.GetUserByEmail(ctx, &neobankv1.GetUserByEmailRequest{Email: email})
	if err != nil {
		return User{}, mapError(err)
	}
	return toUser(resp.GetUser()), nil
}

func (c *Client) GetByID(ctx context.Context, userID string) (User, error) {
	resp, err := c.rpc.GetUserByID(ctx, &neobankv1.GetUserByIDRequest{UserId: userID})
	if err != nil {
		return User{}, mapError(err)
	}
	return toUser(resp.GetUser()), nil
}

func (c *Client) GetWallet(ctx context.Context, userID, currency string) (Wallet, error) {
	resp, err := c.rpc.GetWallet(ctx, &neobankv1.GetWalletRequest{
		UserId:   userID,
		Currency: currency,
	})
	if err != nil {
		return Wallet{}, mapError(err)
	}
	w := resp.GetWallet()
	return Wallet{
		ID:              w.GetId(),
		UserID:          w.GetUserId(),
		Currency:        w.GetCurrency(),
		LedgerAccountID: w.GetLedgerAccountId(),
		Status:          w.GetStatus(),
	}, nil
}

func (c *Client) ListDeviceTokens(ctx context.Context, userID string) ([]DeviceToken, error) {
	resp, err := c.rpc.ListDeviceTokens(ctx, &neobankv1.ListDeviceTokensRequest{UserId: userID})
	if err != nil {
		return nil, mapError(err)
	}
	out := make([]DeviceToken, 0, len(resp.GetDeviceTokens()))
	for _, token := range resp.GetDeviceTokens() {
		out = append(out, DeviceToken{
			ID:       token.GetId(),
			UserID:   token.GetUserId(),
			Platform: token.GetPlatform(),
			Token:    token.GetToken(),
		})
	}
	return out, nil
}

func (c *Client) UpsertPayee(ctx context.Context, userID, payeeUserID, nickname, idempotencyKey string) error {
	ctx = grpcutil.WithIdempotencyKey(ctx, idempotencyKey)
	_, err := c.rpc.UpsertPayee(ctx, &neobankv1.UpsertPayeeRequest{
		UserId:         userID,
		PayeeUserId:    payeeUserID,
		Nickname:       nickname,
		IdempotencyKey: idempotencyKey,
	})
	return mapError(err)
}

func toUser(user *neobankv1.InternalUser) User {
	if user == nil {
		return User{}
	}
	return User{
		ID:     user.GetId(),
		Email:  user.GetEmail(),
		Phone:  user.GetPhone(),
		Status: user.GetStatus(),
	}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			return fmt.Errorf("user service status 404: %s", st.Message())
		case codes.InvalidArgument:
			return fmt.Errorf("user service status 400: %s", st.Message())
		default:
			return fmt.Errorf("user service status %s: %s", st.Code(), st.Message())
		}
	}
	return fmt.Errorf("user service request: %w", err)
}