//
// Copyright (c) 2026 Sumicare
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
	"errors"
	"fmt"

	"google.golang.org/grpc"

	goledgerv1 "github.com/iho/neobank/pkg/gen/goledger/v1"
	"github.com/iho/neobank/pkg/grpcutil"
)

// Config holds connection settings for the goledger service.
type Config struct {
	Addr string
}

// Client wraps gRPC calls to goledger.
type Client struct {
	conn      *grpc.ClientConn
	accounts  goledgerv1.AccountServiceClient
	transfers goledgerv1.TransferServiceClient
	holds     goledgerv1.HoldServiceClient
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:50051"
	}

	conn, err := grpcutil.Dial(ctx, cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("dial ledger: %w", err)
	}

	return &Client{
		conn:      conn,
		accounts:  goledgerv1.NewAccountServiceClient(conn),
		transfers: goledgerv1.NewTransferServiceClient(conn),
		holds:     goledgerv1.NewHoldServiceClient(conn),
	}, nil
}

func (c *Client) Accounts() goledgerv1.AccountServiceClient {
	return c.accounts
}

func (c *Client) Transfers() goledgerv1.TransferServiceClient {
	return c.transfers
}

func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// HealthCheck verifies the gRPC channel is ready.
func (c *Client) HealthCheck(ctx context.Context) error {
	if c.conn.GetState().String() == "SHUTDOWN" {
		return errors.New("ledger connection shutdown")
	}

	return nil
}
