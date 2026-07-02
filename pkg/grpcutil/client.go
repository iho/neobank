package grpcutil

import (
	"context"
	"time"

	"github.com/iho/neobank/pkg/reqctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const IdempotencyKeyHeader = "x-idempotency-key"
const CorrelationIDHeader = "x-correlation-id"

// Dial opens an insecure gRPC connection (mTLS in production).
func Dial(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	base := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(correlationInterceptor),
	}
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

// DefaultTimeout wraps a context with a sensible RPC timeout.
func DefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 10*time.Second)
}
