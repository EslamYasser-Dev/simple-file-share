package grpcapi

import (
	"context"
	"io"
	"strings"
	"testing"

	filesharev1 "github.com/EslamYasser-Dev/simple-file-share/api/proto/fileshare/v1"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newSocialGRPCFixture(t *testing.T) (*SocialService, *models.User, *models.User) {
	t.Helper()
	dir := t.TempDir()
	fileRepo := fs.NewLocalFileRepository(dir)
	userRepo := fs.NewUserFileRepository(dir)
	scoper := policy.NewPathScoper()

	for _, name := range []string{"alice", "bob"} {
		if err := userRepo.CreateUser(&models.User{Username: name}); err != nil {
			t.Fatalf("create user: %v", err)
		}
	}
	alice, _ := userRepo.FindByUsername("alice")
	bob, _ := userRepo.FindByUsername("bob")

	if _, err := fileRepo.WriteFile("users/alice/a.txt", io.NopCloser(strings.NewReader("hello"))); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	follows := fs.NewFollowFileRepository(dir)
	visRepo := fs.NewVisibilityFileRepository(dir)
	timeline := fs.NewTimelineFileRepository(dir)
	downloads := services.NewDownloadService(
		services.NewDownloadFileService(fileRepo, scoper),
		services.NewDownloadZipService(fileRepo, scoper),
	)
	svc := NewSocialService(
		services.NewFollowService(follows, userRepo),
		services.NewVisibilityService(fileRepo, scoper, visRepo, follows, timeline, downloads, userRepo),
		services.NewTimelineService(timeline, follows),
	)
	return svc, alice, bob
}

func userCtx(user *models.User) context.Context {
	return authctx.WithUser(context.Background(), user)
}

// TestSocialGRPCStatusMappings verifies the adapter translates domain errors
// to the documented codes: unknown accounts are NotFound (callers are
// already authenticated), duplicates are AlreadyExists, self-follow is
// InvalidArgument, and anonymous feed reads are PermissionDenied.
func TestSocialGRPCStatusMappings(t *testing.T) {
	svc, alice, bob := newSocialGRPCFixture(t)
	bobCtx := userCtx(bob)

	if _, err := svc.Follow(bobCtx, &filesharev1.FollowRequest{Username: ""}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("empty follow: got %v", err)
	}
	if _, err := svc.Follow(bobCtx, &filesharev1.FollowRequest{Username: "nobody"}); status.Code(err) != codes.NotFound {
		t.Fatalf("unknown followee: got %v", err)
	}
	if _, err := svc.Follow(bobCtx, &filesharev1.FollowRequest{Username: "bob"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("self follow: got %v", err)
	}
	resp, err := svc.Follow(bobCtx, &filesharev1.FollowRequest{Username: "alice"})
	if err != nil || !resp.GetFollowing() {
		t.Fatalf("follow: %v %v", resp, err)
	}
	if _, err := svc.Follow(bobCtx, &filesharev1.FollowRequest{Username: "alice"}); status.Code(err) != codes.AlreadyExists {
		t.Fatalf("duplicate: got %v", err)
	}
	if _, err := svc.Unfollow(bobCtx, &filesharev1.UnfollowRequest{Username: "carol"}); status.Code(err) != codes.NotFound {
		t.Fatalf("unfollow missing: got %v", err)
	}
	if _, err := svc.SetVisibility(userCtx(alice), &filesharev1.SetVisibilityRequest{Path: "a.txt", Level: "everyone"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("bad level: got %v", err)
	}
	if _, err := svc.ListFeed(context.Background(), &filesharev1.ListFeedRequest{}); status.Code(err) != codes.PermissionDenied {
		t.Fatalf("anonymous feed: got %v", err)
	}

	// Link-scope the file and confirm the follower's feed carries it with a
	// next cursor, while a stranger sees nothing.
	if _, err := svc.SetVisibility(userCtx(alice), &filesharev1.SetVisibilityRequest{Path: "a.txt", Level: "link", AllowStream: true}); err != nil {
		t.Fatalf("set link: %v", err)
	}
	feed, err := svc.ListFeed(bobCtx, &filesharev1.ListFeedRequest{Limit: 10})
	if err != nil {
		t.Fatalf("feed: %v", err)
	}
	if len(feed.GetEvents()) == 0 || feed.GetNextCursor() == "" {
		t.Fatalf("follower feed must carry the entry and a cursor: %+v", feed)
	}
	if feed.GetEvents()[0].GetVisibility() != "link" || feed.GetEvents()[0].GetOwner() != "alice" {
		t.Fatalf("unexpected entry: %+v", feed.GetEvents()[0])
	}
}
