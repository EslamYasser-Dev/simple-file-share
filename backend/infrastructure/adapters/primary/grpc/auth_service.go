package grpcapi

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	filesharev1 "github.com/EslamYasser-Dev/simple-file-share/api/proto/fileshare/v1"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

// AuthService adapts account use cases to the gRPC transport.
type AuthService struct {
	filesharev1.UnimplementedAuthServiceServer

	register      *services.RegisterUserService
	users         *services.ListUsersService
	authProvider  ports.AuthProvider
	signupEnabled bool
}

func NewAuthService(
	register *services.RegisterUserService,
	users *services.ListUsersService,
	authProvider ports.AuthProvider,
	signupEnabled bool,
) *AuthService {
	return &AuthService{
		register:      register,
		users:         users,
		authProvider:  authProvider,
		signupEnabled: signupEnabled,
	}
}

func (s *AuthService) Register(_ context.Context, req *filesharev1.RegisterRequest) (*filesharev1.User, error) {
	user, err := s.register.Execute(req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, toStatus(err)
	}
	return toProtoUser(user), nil
}

func (s *AuthService) Authenticate(_ context.Context, req *filesharev1.AuthenticateRequest) (*filesharev1.User, error) {
	user, err := s.authProvider.Authenticate(req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, toStatus(err)
	}
	return toProtoUser(user), nil
}

func (s *AuthService) Me(ctx context.Context, _ *filesharev1.MeRequest) (*filesharev1.User, error) {
	user := authctx.UserFromContext(ctx)
	if user == nil {
		// Auth disabled — expose the system view so clients render as admin.
		return &filesharev1.User{IsAdmin: true}, nil
	}
	return toProtoUser(user), nil
}

func (s *AuthService) GetAuthInfo(_ context.Context, _ *filesharev1.GetAuthInfoRequest) (*filesharev1.AuthInfoResponse, error) {
	return &filesharev1.AuthInfoResponse{SignupEnabled: s.signupEnabled}, nil
}

func (s *AuthService) ListUsers(ctx context.Context, _ *filesharev1.ListUsersRequest) (*filesharev1.ListUsersResponse, error) {
	user := authctx.UserFromContext(ctx)
	if user == nil || !user.IsAdmin {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	stats, err := s.users.Execute()
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*filesharev1.UserStats, 0, len(stats))
	for _, st := range stats {
		out = append(out, &filesharev1.UserStats{
			Username:  st.Username,
			FileCount: int64(st.Files),
			TotalSize: st.Size,
			IsAdmin:   st.IsAdmin,
		})
	}
	return &filesharev1.ListUsersResponse{Users: out}, nil
}

func toProtoUser(u *models.User) *filesharev1.User {
	if u == nil {
		return nil
	}
	return &filesharev1.User{Username: u.Username, IsAdmin: u.IsAdmin}
}
