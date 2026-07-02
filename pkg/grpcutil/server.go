package grpcutil

import (
	"context"

	"github.com/iho/neobank/pkg/otel"
	"github.com/iho/neobank/pkg/reqctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewServer builds a gRPC server with correlation propagation and optional OTel stats.
func NewServer() *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(serverCorrelationInterceptor),
	}
	opts = append(opts, otel.GRPCServerOptions()...)
	return grpc.NewServer(opts...)
}

func serverCorrelationInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(CorrelationIDHeader); len(vals) > 0 {
			ctx = reqctx.WithCorrelationID(ctx, vals[0])
		}
	}
	return handler(ctx, req)
}