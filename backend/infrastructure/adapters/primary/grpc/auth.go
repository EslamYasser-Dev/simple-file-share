package grpcapi

import (
	"context"
	"encoding/base64"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

// publicMethods are callable without credentials. Authenticate validates its
// own credentials, and Register/GetAuthInfo are account-bootstrap endpoints.
var publicMethods = map[string]bool{
	"/fileshare.v1.AuthService/Register":                             true,
	"/fileshare.v1.AuthService/Authenticate":                         true,
	"/fileshare.v1.AuthService/GetAuthInfo":                          true,
	"/grpc.health.v1.Health/Check":                                   true,
	"/grpc.health.v1.Health/Watch":                                   true,
	"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo":      true,
	"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo": true,
}

const basicPrefix = "Basic "

func unaryAuthInterceptor(provider ports.AuthProvider, enabled bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !enabled || provider == nil || publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}
		user, err := authenticate(ctx, provider)
		if err != nil {
			return nil, err
		}
		return handler(authctx.WithUser(ctx, user), req)
	}
}

func streamAuthInterceptor(provider ports.AuthProvider, enabled bool) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !enabled || provider == nil || publicMethods[info.FullMethod] {
			return handler(srv, ss)
		}
		user, err := authenticate(ss.Context(), provider)
		if err != nil {
			return err
		}
		return handler(srv, &streamWithUser{ServerStream: ss, ctx: authctx.WithUser(ss.Context(), user)})
	}
}

// authenticate reads HTTP Basic credentials from incoming metadata and resolves
// them to an account via the AuthProvider port.
func authenticate(ctx context.Context, provider ports.AuthProvider) (*models.User, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing credentials")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing credentials")
	}
	username, password, ok := parseBasic(values[0])
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
	}
	user, err := provider.Authenticate(username, password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	return user, nil
}

func parseBasic(header string) (username, password string, ok bool) {
	if !strings.HasPrefix(header, basicPrefix) {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(header[len(basicPrefix):])
	if err != nil {
		return "", "", false
	}
	creds := string(decoded)
	i := strings.IndexByte(creds, ':')
	if i < 0 {
		return "", "", false
	}
	return creds[:i], creds[i+1:], true
}

// streamWithUser overrides the context of a server stream with the
// authenticated user so downstream handlers can read it.
type streamWithUser struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *streamWithUser) Context() context.Context { return s.ctx }
