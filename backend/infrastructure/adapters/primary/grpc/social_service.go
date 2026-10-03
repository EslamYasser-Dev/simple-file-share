package grpcapi

import (
	"context"
	"errors"
	"io"
	"time"

	filesharev1 "github.com/EslamYasser-Dev/simple-file-share/api/proto/fileshare/v1"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SocialService adapts follow, visibility, and timeline use cases to the gRPC
// transport. Every method requires authentication (no publicMethods entry).
// Account-existence errors map to NotFound here: unlike the public auth
// surface, callers are already authenticated, so there is nothing to
// enumerate that follower lists do not already reveal.
type SocialService struct {
	filesharev1.UnimplementedSocialServiceServer

	follows    *services.FollowService
	visibility *services.VisibilityService
	timeline   *services.TimelineService
}

func NewSocialService(
	follows *services.FollowService,
	visibility *services.VisibilityService,
	timeline *services.TimelineService,
) *SocialService {
	return &SocialService{follows: follows, visibility: visibility, timeline: timeline}
}

func (s *SocialService) Follow(ctx context.Context, req *filesharev1.FollowRequest) (*filesharev1.FollowResponse, error) {
	me := authctx.UserFromContext(ctx)
	username := req.GetUsername()
	if err := s.follows.Follow(callerName(me), username); err != nil {
		return nil, toSocialStatus(err)
	}
	return &filesharev1.FollowResponse{Username: username, Following: true}, nil
}

func (s *SocialService) Unfollow(ctx context.Context, req *filesharev1.UnfollowRequest) (*filesharev1.UnfollowResponse, error) {
	me := authctx.UserFromContext(ctx)
	username := req.GetUsername()
	if err := s.follows.Unfollow(callerName(me), username); err != nil {
		return nil, toSocialStatus(err)
	}
	return &filesharev1.UnfollowResponse{Username: username, Following: false}, nil
}

func (s *SocialService) IsFollowing(ctx context.Context, req *filesharev1.IsFollowingRequest) (*filesharev1.IsFollowingResponse, error) {
	me := authctx.UserFromContext(ctx)
	ok, err := s.follows.IsFollowing(callerName(me), req.GetUsername())
	if err != nil {
		return nil, toSocialStatus(err)
	}
	return &filesharev1.IsFollowingResponse{Username: req.GetUsername(), Following: ok}, nil
}

func (s *SocialService) ListFollowers(_ context.Context, req *filesharev1.ListFollowersRequest) (*filesharev1.ListFollowersResponse, error) {
	usernames, err := s.follows.Followers(req.GetUsername())
	if err != nil {
		return nil, toSocialStatus(err)
	}
	return &filesharev1.ListFollowersResponse{Usernames: usernames}, nil
}

func (s *SocialService) ListFollowing(_ context.Context, req *filesharev1.ListFollowingRequest) (*filesharev1.ListFollowingResponse, error) {
	usernames, err := s.follows.Following(req.GetUsername())
	if err != nil {
		return nil, toSocialStatus(err)
	}
	return &filesharev1.ListFollowingResponse{Usernames: usernames}, nil
}

func (s *SocialService) SetVisibility(ctx context.Context, req *filesharev1.SetVisibilityRequest) (*filesharev1.FileVisibility, error) {
	vis, err := s.visibility.Set(
		authctx.UserFromContext(ctx),
		req.GetPath(),
		models.VisibilityLevel(req.GetLevel()),
		req.GetAllowStream(),
	)
	if err != nil {
		return nil, toSocialStatus(err)
	}
	return toProtoVisibility(vis), nil
}

func (s *SocialService) GetVisibility(_ context.Context, req *filesharev1.GetVisibilityRequest) (*filesharev1.FileVisibility, error) {
	vis, err := s.visibility.Get(req.GetOwner(), req.GetPath())
	if err != nil {
		return nil, toSocialStatus(err)
	}
	return toProtoVisibility(vis), nil
}

func (s *SocialService) ListFeed(ctx context.Context, req *filesharev1.ListFeedRequest) (*filesharev1.ListFeedResponse, error) {
	events, err := s.timeline.Feed(authctx.UserFromContext(ctx), req.GetCursor(), int(req.GetLimit()))
	if err != nil {
		return nil, toSocialStatus(err)
	}
	out := make([]*filesharev1.TimelineEvent, 0, len(events))
	for _, e := range events {
		out = append(out, toProtoTimelineEvent(e))
	}
	resp := &filesharev1.ListFeedResponse{Events: out}
	if len(events) > 0 {
		resp.NextCursor = events[len(events)-1].ID
	}
	return resp, nil
}

// DownloadSharedFile streams another owner's file after the visibility
// service re-checks the streaming gate. Chunking mirrors FileService so
// mobile can reuse its download-then-play path.
func (s *SocialService) DownloadSharedFile(req *filesharev1.DownloadSharedRequest, stream filesharev1.SocialService_DownloadSharedFileServer) error {
	download, err := s.visibility.Serve(
		authctx.UserFromContext(stream.Context()),
		req.GetOwner(),
		req.GetPath(),
	)
	if err != nil {
		return toSocialStatus(err)
	}
	defer download.Stream.Close()

	buf := make([]byte, streamChunkSize)
	first := true
	for {
		n, readErr := download.Stream.Read(buf)
		if n > 0 {
			chunk := &filesharev1.DownloadChunk{Data: buf[:n]}
			if first {
				chunk.Filename = download.Filename
				chunk.ContentType = download.ContentType
				first = false
			}
			if sendErr := stream.Send(chunk); sendErr != nil {
				return sendErr
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return status.Error(codes.Internal, "stream read failed")
		}
	}
}

func callerName(user *models.User) string {
	if user == nil {
		return ""
	}
	return user.Username
}

func toProtoVisibility(v *models.FileVisibility) *filesharev1.FileVisibility {
	if v == nil {
		return nil
	}
	out := &filesharev1.FileVisibility{
		Owner:       v.Owner,
		Path:        v.Path,
		Level:       string(v.Level),
		AllowStream: v.AllowStream,
	}
	if !v.UpdatedAt.IsZero() {
		out.UpdatedAt = v.UpdatedAt.UTC().Format(time.RFC3339)
	}
	return out
}

func toProtoTimelineEvent(e *models.TimelineEvent) *filesharev1.TimelineEvent {
	if e == nil {
		return nil
	}
	out := &filesharev1.TimelineEvent{
		Id:         e.ID,
		Owner:      e.Owner,
		Kind:       e.Kind,
		Path:       e.Path,
		Name:       e.Name,
		Size:       e.Size,
		Visibility: string(e.Visibility),
	}
	if !e.CreatedAt.IsZero() {
		out.CreatedAt = e.CreatedAt.UTC().Format(time.RFC3339)
	}
	return out
}

// toSocialStatus maps service errors to codes, overriding the global
// unauthenticated collapse for account lookup: social callers are already
// authenticated.
func toSocialStatus(err error) error {
	switch {
	case errors.Is(err, domainerrors.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, domainerrors.ErrAlreadyFollowing):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domainerrors.ErrCannotFollowSelf):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domainerrors.ErrFollowNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return toStatus(err)
	}
}
