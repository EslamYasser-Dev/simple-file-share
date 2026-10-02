package grpcapi

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// unaryMetricsInterceptor records every unary call, including ones rejected
// by the auth interceptor: it must be chained first (outermost).
func unaryMetricsInterceptor(rec ports.MetricsRecorder) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		rec.ObserveRPC(info.FullMethod, status.Code(err).String(), time.Since(start))
		return resp, err
	}
}

// streamMetricsInterceptor records stream lifetime; streams are long-lived by
// design (event subscription), so duration lands in the top bucket/+Inf.
func streamMetricsInterceptor(rec ports.MetricsRecorder) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		err := handler(srv, ss)
		rec.ObserveRPC(info.FullMethod, status.Code(err).String(), time.Since(start))
		return err
	}
}
