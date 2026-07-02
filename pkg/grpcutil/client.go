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

package grpcutil

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/iho/neobank/pkg/otel"
	"github.com/iho/neobank/pkg/reqctx"
)

const (
	IdempotencyKeyHeader = "x-idempotency-key"
	CorrelationIDHeader  = "x-correlation-id"
	UserIDHeader         = "x-user-id"
)

// Dial opens a gRPC connection. Uses mTLS when GRPC_MTLS_ENABLED=true.
func Dial(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	tlsCfg := LoadTLSConfigFromEnv()
	var creds credentials.TransportCredentials
	if tlsCfg.Enabled {
		var err error
		creds, err = tlsCfg.ClientCredentials()
		if err != nil {
			return nil, err
		}
	} else {
		creds = insecure.NewCredentials()
	}

	base := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithChainUnaryInterceptor(correlationInterceptor),
	}
	base = append(base, otel.GRPCDialOptions()...)

	return grpc.NewClient(addr, append(base, opts...)...)
}

// DialInsecure opens a plaintext gRPC connection regardless of GRPC_MTLS_ENABLED.
// Use for third-party services (e.g. goledger) that do not yet terminate mTLS.
func DialInsecure(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	base := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(correlationInterceptor),
	}
	base = append(base, otel.GRPCDialOptions()...)
	return grpc.NewClient(addr, append(base, opts...)...)
}

// correlationInterceptor forwards the request's correlation ID (see pkg/reqctx)
// onto outgoing gRPC metadata so downstream calls (e.g. into goledger) can be
// tied back to the originating API request in logs.
func correlationInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	if id := reqctx.CorrelationID(ctx); id != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, CorrelationIDHeader, id)
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}

// WithIdempotencyKey attaches an idempotency key to outgoing gRPC metadata.
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, IdempotencyKeyHeader, key)
}

// WithUserID attaches the authenticated user id to outgoing gRPC metadata.
func WithUserID(ctx context.Context, userID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, UserIDHeader, userID)
}

// DefaultTimeout wraps a context with a sensible RPC timeout.
func DefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 10*time.Second)
}

// IdempotencyKeyFromContext reads an idempotency key from incoming gRPC metadata.
func IdempotencyKeyFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(IdempotencyKeyHeader); len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}
