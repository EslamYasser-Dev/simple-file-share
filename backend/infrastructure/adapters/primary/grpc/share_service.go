package grpcapi

import (
	"context"
	"path/filepath"
	"time"

	filesharev1 "github.com/EslamYasser-Dev/simple-file-share/api/proto/fileshare/v1"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
)

// ShareService adapts share-link use cases to the gRPC transport.
type ShareService struct {
	filesharev1.UnimplementedShareServiceServer

	create *services.CreateShareService
	list   *services.ListSharesService
	revoke *services.RevokeShareService
}

func NewShareService(
	create *services.CreateShareService,
	list *services.ListSharesService,
	revoke *services.RevokeShareService,
) *ShareService {
	return &ShareService{create: create, list: list, revoke: revoke}
}

func (s *ShareService) CreateShare(ctx context.Context, req *filesharev1.CreateShareRequest) (*filesharev1.Share, error) {
	share, err := s.create.Execute(authctx.UserFromContext(ctx), req.GetPath(), req.GetExpiresInSeconds())
	if err != nil {
		return nil, toStatus(err)
	}
	return toProtoShare(share), nil
}

func (s *ShareService) ListShares(ctx context.Context, _ *filesharev1.ListSharesRequest) (*filesharev1.ListSharesResponse, error) {
	shares, err := s.list.Execute(authctx.UserFromContext(ctx))
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*filesharev1.Share, 0, len(shares))
	for _, share := range shares {
		out = append(out, toProtoShare(share))
	}
	return &filesharev1.ListSharesResponse{Shares: out}, nil
}

func (s *ShareService) RevokeShare(ctx context.Context, req *filesharev1.RevokeShareRequest) (*filesharev1.RevokeShareResponse, error) {
	if err := s.revoke.Execute(authctx.UserFromContext(ctx), req.GetToken()); err != nil {
		return nil, toStatus(err)
	}
	return &filesharev1.RevokeShareResponse{Message: "revoked"}, nil
}

func toProtoShare(s *models.Share) *filesharev1.Share {
	if s == nil {
		return nil
	}
	out := &filesharev1.Share{
		Token: s.Token,
		Path:  s.Path,
		Name:  filepath.Base(s.Path),
		Owner: s.Owner,
	}
	if !s.CreatedAt.IsZero() {
		out.CreatedAt = s.CreatedAt.UTC().Format(time.RFC3339)
	}
	if !s.ExpiresAt.IsZero() {
		out.ExpiresAt = s.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return out
}
