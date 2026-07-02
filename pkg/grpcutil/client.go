package grpcutil

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const IdempotencyKeyHeader = "x-idempotency-key"

// Dial opens an insecure gRPC connection (mTLS in production).
func Dial(ctx context.Context, addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	base := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	return grpc.NewClient(addr, append(base, opts...)...)
}

// WithIdempotencyKey attaches an idempotency key to outgoing gRPC metadata.
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, IdempotencyKeyHeader, key)
}

// DefaultTimeout wraps a context with a sensible RPC timeout.
func DefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 10*time.Second)
}