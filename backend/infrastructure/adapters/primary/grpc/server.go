// Package grpcapi is the gRPC primary adapter. It translates gRPC requests
// into the same application use cases used by the HTTP adapter.
package grpcapi

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	filesharev1 "github.com/EslamYasser-Dev/simple-file-share/api/proto/fileshare/v1"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// Server hosts the gRPC services. It is started alongside the HTTP server and
// shut down when the process receives a termination signal.
type Server struct {
	port         string
	logger       ports.Logger
	grpcServer   *grpc.Server
	shutdownWait time.Duration
}

// NewServer builds the gRPC server, registering the health and reflection
// services plus the application services.
func NewServer(
	port string,
	logger ports.Logger,
	tlsGenerator ports.TLSCertGenerator,
	enableTLS bool,
	authService *services.AuthenticateService,
	enableAuth bool,
	authStore *AuthService,
	fileService *FileService,
) (*Server, error) {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryAuthInterceptor(authService, enableAuth)),
		grpc.ChainStreamInterceptor(streamAuthInterceptor(authService, enableAuth)),
	}

	if enableTLS {
		certPEM, keyPEM, err := tlsGenerator.GenerateCert()
		if err != nil {
			return nil, fmt.Errorf("generate gRPC TLS cert: %w", err)
		}
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("parse gRPC TLS key pair: %w", err)
		}
		creds := credentials.NewTLS(&tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{cert},
		})
		opts = append(opts, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(opts...)
	filesharev1.RegisterAuthServiceServer(grpcServer, authStore)
	filesharev1.RegisterFileServiceServer(grpcServer, fileService)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	reflection.Register(grpcServer)

	return &Server{
		port:         port,
		logger:       logger,
		grpcServer:   grpcServer,
		shutdownWait: 5 * time.Second,
	}, nil
}

// Start listens on the configured port and blocks until the server stops.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("listen gRPC on :%s: %w", s.port, err)
	}
	s.logger.Info("gRPC server starting", "address", "0.0.0.0:"+s.port)
	return s.Serve(lis)
}

// Serve blocks serving on the provided listener. Tests use it with an
// in-memory listener.
func (s *Server) Serve(lis net.Listener) error {
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("serve gRPC: %w", err)
	}
	return nil
}

// Stop gracefully drains in-flight RPCs, forcing a hard stop after a timeout.
func (s *Server) Stop() {
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(s.shutdownWait):
		s.grpcServer.Stop()
	}
	s.logger.Info("gRPC server stopped")
}
