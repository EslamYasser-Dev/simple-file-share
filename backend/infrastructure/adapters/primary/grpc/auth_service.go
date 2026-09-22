package grpcapi

import (
	"context"

	filesharev1 "github.com/EslamYasser-Dev/simple-file-share/api/proto/fileshare/v1"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

// AuthService adapts account use cases to the gRPC transport.
type AuthService struct {
	filesharev1.UnimplementedAuthServiceServer

	register      *services.RegisterUserService
	users         *services.ListUsersService
	authenticate  *services.AuthenticateService
	signupEnabled bool
}

func NewAuthService(
	register *services.RegisterUserService,
	users *services.ListUsersService,
	authenticate *services.AuthenticateService,
	signupEnabled bool,
) *AuthService {
	return &AuthService{
		register:      register,
		users:         users,
		authenticate:  authenticate,
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
	user, err := s.authenticate.Execute(req.GetUsername(), req.GetPassword())
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
	stats, err := s.users.Execute(authctx.UserFromContext(ctx))
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
