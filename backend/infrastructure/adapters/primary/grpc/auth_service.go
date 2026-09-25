package grpcapi

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

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
	userinfo      *services.UserInfoService
	tokens        *services.TokenService
	signupEnabled bool
}

func NewAuthService(
	register *services.RegisterUserService,
	users *services.ListUsersService,
	authenticate *services.AuthenticateService,
	userinfo *services.UserInfoService,
	signupEnabled bool,
	tokens *services.TokenService,
) *AuthService {
	return &AuthService{
		register:      register,
		users:         users,
		authenticate:  authenticate,
		userinfo:      userinfo,
		signupEnabled: signupEnabled,
		tokens:        tokens,
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
	if s.userinfo == nil {
		user := authctx.UserFromContext(ctx)
		if user == nil {
			return &filesharev1.User{IsAdmin: true, Role: models.RoleAdmin, Enabled: true}, nil
		}
		return toProtoUser(user), nil
	}
	stats, err := s.userinfo.Execute(authctx.UserFromContext(ctx))
	if err != nil {
		return nil, toStatus(err)
	}
	out := &filesharev1.User{
		Username:    stats.Username,
		IsAdmin:     stats.IsAdmin,
		Role:        stats.Role,
		Enabled:     stats.Enabled,
		QuotaBytes:  stats.QuotaBytes,
		Size:        stats.Size,
		Files:       int64(stats.Files),
		Permissions: models.AllPermissionsIfAdmin(stats.IsAdmin),
	}
	if !stats.CreatedAt.IsZero() {
		out.CreatedAt = stats.CreatedAt.UTC().Format(time.RFC3339)
	}
	return out, nil
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
			Username:   st.Username,
			FileCount:  int64(st.Files),
			TotalSize:  st.Size,
			IsAdmin:    st.IsAdmin,
			Role:       st.Role,
			Enabled:    st.Enabled,
			QuotaBytes: st.QuotaBytes,
		})
	}
	return &filesharev1.ListUsersResponse{Users: out}, nil
}

func (s *AuthService) Login(_ context.Context, req *filesharev1.LoginRequest) (*filesharev1.LoginResponse, error) {
	if s.tokens == nil {
		return nil, status.Error(codes.Unimplemented, "token service unavailable")
	}
	pair, err := s.tokens.Login(req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, toStatus(err)
	}
	expiresIn := int64(0)
	if !pair.ExpiresAt.IsZero() {
		expiresIn = int64(time.Until(pair.ExpiresAt).Seconds())
		if expiresIn < 0 {
			expiresIn = 0
		}
	}
	return &filesharev1.LoginResponse{
		AccessToken: pair.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, _ *filesharev1.LogoutRequest) (*filesharev1.LogoutResponse, error) {
	if s.tokens != nil {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("authorization"); len(vals) > 0 && strings.HasPrefix(vals[0], bearerPrefix) {
				_ = s.tokens.Revoke(strings.TrimSpace(vals[0][len(bearerPrefix):]))
			}
		}
	}
	return &filesharev1.LogoutResponse{Status: "revoked"}, nil
}

func toProtoUser(u *models.User) *filesharev1.User {
	if u == nil {
		return nil
	}
	return &filesharev1.User{
		Username: u.Username,
		IsAdmin:  u.IsAdmin,
		Role:     u.Role,
		Enabled:  u.Enabled,
	}
}
