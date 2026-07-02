package grpcutil

import (
	"context"

	"github.com/iho/neobank/pkg/otel"
	"github.com/iho/neobank/pkg/reqctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewServer builds a gRPC server with correlation propagation, optional mTLS, and OTel stats.
func NewServer() (*grpc.Server, error) {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(serverCorrelationInterceptor),
	}
	tlsCfg := LoadTLSConfigFromEnv()
	if tlsCfg.Enabled {
		creds, err := tlsCfg.ServerCredentials()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}
	opts = append(opts, otel.GRPCServerOptions()...)
	return grpc.NewServer(opts...), nil
}

func serverCorrelationInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(CorrelationIDHeader); len(vals) > 0 {
			ctx = reqctx.WithCorrelationID(ctx, vals[0])
		}
	}
	return handler(ctx, req)
}