package services

import (
	"errors"
	"io"
	"strings"
	"testing"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

type socialFixture struct {
	follows    *fs.FollowFileRepository
	visibility *fs.VisibilityFileRepository
	timeline   *fs.TimelineFileRepository
	followSvc  *FollowService
	visSvc     *VisibilityService
	feedSvc    *TimelineService
	alice      *models.User
	bob        *models.User
	carol      *models.User
}

func newSocialFixture(t *testing.T) *socialFixture {
	t.Helper()
	dir := t.TempDir()
	fileRepo := fs.NewLocalFileRepository(dir)
	userRepo := fs.NewUserFileRepository(dir)
	scoper := policy.NewPathScoper()

	mkuser := func(name string) *models.User {
		u := &models.User{Username: name}
		if err := userRepo.CreateUser(u); err != nil {
			t.Fatalf("create user %s: %v", name, err)
		}
		stored, err := userRepo.FindByUsername(name)
		if err != nil {
			t.Fatalf("find user %s: %v", name, err)
		}
		return stored
	}
	alice, bob, carol := mkuser("alice"), mkuser("bob"), mkuser("carol")

	// One real file owned by alice for the visibility matrix.
	if _, err := fileRepo.WriteFile("users/alice/a.txt", io.NopCloser(strings.NewReader("hello"))); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	follows := fs.NewFollowFileRepository(dir)
	visRepo := fs.NewVisibilityFileRepository(dir)
	timeline := fs.NewTimelineFileRepository(dir)
	downloads := NewDownloadService(
		NewDownloadFileService(fileRepo, scoper),
		NewDownloadZipService(fileRepo, scoper),
	)
	return &socialFixture{
		follows:    follows,
		visibility: visRepo,
		timeline:   timeline,
		followSvc:  NewFollowService(follows, userRepo),
		visSvc:     NewVisibilityService(fileRepo, scoper, visRepo, follows, timeline, downloads, userRepo),
		feedSvc:    NewTimelineService(timeline, follows),
		alice:      alice,
		bob:        bob,
		carol:      carol,
	}
}

// TestFollowLifecycle covers self-follow rejection, duplicates, unknown
// accounts, and ordered follower/following lists.
func TestFollowLifecycle(t *testing.T) {
	f := newSocialFixture(t)

	if err := f.followSvc.Follow("bob", "bob"); !errors.Is(err, domainerrors.ErrCannotFollowSelf) {
		t.Fatalf("self-follow: got %v", err)
	}
	if err := f.followSvc.Follow("bob", "nobody"); !errors.Is(err, domainerrors.ErrUserNotFound) {
		t.Fatalf("unknown followee: got %v", err)
	}
	if err := f.followSvc.Follow("bob", "alice"); err != nil {
		t.Fatalf("follow: %v", err)
	}
	if err := f.followSvc.Follow("bob", "alice"); !errors.Is(err, domainerrors.ErrAlreadyFollowing) {
		t.Fatalf("duplicate: got %v", err)
	}
	if ok, err := f.followSvc.IsFollowing("bob", "alice"); err != nil || !ok {
		t.Fatalf("IsFollowing: ok=%v err=%v", ok, err)
	}
	if err := f.followSvc.Follow("carol", "alice"); err != nil {
		t.Fatalf("follow carol: %v", err)
	}
	followers, err := f.followSvc.Followers("alice")
	if err != nil || len(followers) != 2 || followers[0] != "bob" || followers[1] != "carol" {
		t.Fatalf("followers: %v err=%v", followers, err)
	}
	following, err := f.followSvc.Following("bob")
	if err != nil || len(following) != 1 || following[0] != "alice" {
		t.Fatalf("following: %v err=%v", following, err)
	}
	if err := f.followSvc.Unfollow("bob", "alice"); err != nil {
		t.Fatalf("unfollow: %v", err)
	}
	if err := f.followSvc.Unfollow("bob", "alice"); !errors.Is(err, domainerrors.ErrFollowNotFound) {
		t.Fatalf("double unfollow: got %v", err)
	}
}

// TestVisibilityAccessMatrix verifies the audience rules: private stays with
// the owner, link needs a follow edge, public needs authentication, and the
// streaming toggle only affects non-owners.
func TestVisibilityAccessMatrix(t *testing.T) {
	f := newSocialFixture(t)

	// Default is private.
	if ok, _ := f.visSvc.CanView(f.bob, "alice", "a.txt"); ok {
		t.Fatal("default must be private")
	}
	if ok, _ := f.visSvc.CanView(f.alice, "alice", "a.txt"); !ok {
		t.Fatal("owner must always view")
	}

	if _, err := f.visSvc.Set(f.alice, "a.txt", "everyone", true); err == nil {
		t.Fatal("bogus level must be rejected")
	}
	if _, err := f.visSvc.Set(f.alice, "missing.txt", models.VisibilityPublic, true); err == nil {
		t.Fatal("missing file must be rejected")
	}
	if _, err := f.visSvc.Set(f.bob, "a.txt", models.VisibilityPublic, true); err == nil {
		t.Fatal("non-owner must be forbidden")
	}

	if _, err := f.visSvc.Set(f.alice, "a.txt", models.VisibilityLink, true); err != nil {
		t.Fatalf("set link: %v", err)
	}
	if ok, _ := f.visSvc.CanView(f.bob, "alice", "a.txt"); ok {
		t.Fatal("stranger must not see link-scoped file")
	}
	if err := f.followSvc.Follow("bob", "alice"); err != nil {
		t.Fatalf("follow: %v", err)
	}
	if ok, _ := f.visSvc.CanView(f.bob, "alice", "a.txt"); !ok {
		t.Fatal("follower must see link-scoped file")
	}
	if ok, _ := f.visSvc.CanStream(f.bob, "alice", "a.txt"); !ok {
		t.Fatal("follower must stream when allowed")
	}

	if _, err := f.visSvc.Set(f.alice, "a.txt", models.VisibilityLink, false); err != nil {
		t.Fatalf("disable streaming: %v", err)
	}
	if ok, _ := f.visSvc.CanView(f.bob, "alice", "a.txt"); !ok {
		t.Fatal("metadata still visible without streaming")
	}
	if ok, _ := f.visSvc.CanStream(f.bob, "alice", "a.txt"); ok {
		t.Fatal("streaming toggle must gate playback")
	}
	if ok, _ := f.visSvc.CanStream(f.alice, "alice", "a.txt"); !ok {
		t.Fatal("owner always streams")
	}

	if _, err := f.visSvc.Set(f.alice, "a.txt", models.VisibilityPublic, true); err != nil {
		t.Fatalf("set public: %v", err)
	}
	if ok, _ := f.visSvc.CanView(f.carol, "alice", "a.txt"); !ok {
		t.Fatal("public visible to any authenticated user")
	}
	if ok, _ := f.visSvc.CanView(nil, "alice", "a.txt"); ok {
		t.Fatal("public still requires authentication")
	}
}

// TestVisibilityServe verifies cross-user byte serving: the owner always,
// a follower when the file is link-scoped with streaming on, and nobody
// else — anonymous, stranger, streaming-off, or unknown owner.
func TestVisibilityServe(t *testing.T) {
	f := newSocialFixture(t)

	readBody := func(t *testing.T, viewer *models.User, owner, path string) (string, error) {
		t.Helper()
		dl, err := f.visSvc.Serve(viewer, owner, path)
		if err != nil {
			return "", err
		}
		defer dl.Stream.Close()
		body, err := io.ReadAll(dl.Stream)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

	if body, err := readBody(t, f.alice, "alice", "a.txt"); err != nil || body != "hello" {
		t.Fatalf("owner serve: body=%q err=%v", body, err)
	}
	if _, err := readBody(t, nil, "alice", "a.txt"); !isForbidden(err) {
		t.Fatalf("anonymous serve: got %v", err)
	}
	if _, err := readBody(t, f.bob, "alice", "a.txt"); !isForbidden(err) {
		t.Fatalf("stranger serve of private: got %v", err)
	}

	if _, err := f.visSvc.Set(f.alice, "a.txt", models.VisibilityLink, true); err != nil {
		t.Fatalf("set link: %v", err)
	}
	if err := f.followSvc.Follow("bob", "alice"); err != nil {
		t.Fatalf("follow: %v", err)
	}
	if body, err := readBody(t, f.bob, "alice", "a.txt"); err != nil || body != "hello" {
		t.Fatalf("follower serve: body=%q err=%v", body, err)
	}
	if _, err := readBody(t, f.carol, "alice", "a.txt"); !isForbidden(err) {
		t.Fatalf("non-follower serve: got %v", err)
	}

	if _, err := f.visSvc.Set(f.alice, "a.txt", models.VisibilityLink, false); err != nil {
		t.Fatalf("disable streaming: %v", err)
	}
	if _, err := readBody(t, f.bob, "alice", "a.txt"); !isForbidden(err) {
		t.Fatalf("streaming-off serve: got %v", err)
	}

	// Unknown owner is indistinguishable from private: the gate runs before
	// any account lookup, so there is no existence oracle.
	if _, err := readBody(t, f.bob, "nobody", "a.txt"); !isForbidden(err) {
		t.Fatalf("unknown owner serve: got %v", err)
	}
}

func isForbidden(err error) bool {
	var forbidden *domainerrors.ForbiddenError
	return errors.As(err, &forbidden)
}

// TestTimelineFeedAuthZ verifies the feed shows owners their own entries,
// followers link-scoped entries, everyone public entries — and that a
// visibility downgrade hides older entries with cursor pagination intact.
func TestTimelineFeedAuthZ(t *testing.T) {
	f := newSocialFixture(t)

	if _, err := f.feedSvc.Record(models.TimelineUpload, f.alice, "a.txt", 5, models.VisibilityPrivate); err != nil {
		t.Fatalf("record: %v", err)
	}
	if _, err := f.feedSvc.Record("bogus", f.alice, "a.txt", 5, models.VisibilityPrivate); err == nil {
		t.Fatal("bogus kind must be rejected")
	}

	// Owner always sees their own private entry; others do not.
	own, err := f.feedSvc.Feed(f.alice, "", 10)
	if err != nil || len(own) != 1 {
		t.Fatalf("owner feed: %v err=%v", own, err)
	}
	stranger, err := f.feedSvc.Feed(f.bob, "", 10)
	if err != nil || len(stranger) != 0 {
		t.Fatalf("stranger feed: %v err=%v", stranger, err)
	}
	if _, err := f.feedSvc.Feed(nil, "", 10); err == nil {
		t.Fatal("anonymous feed must be rejected")
	}

	// Link-scope + follow: visible. Downgrade to private: hidden again,
	// including the older entry (snapshot refresh).
	if _, err := f.visSvc.Set(f.alice, "a.txt", models.VisibilityLink, true); err != nil {
		t.Fatalf("set link: %v", err)
	}
	if err := f.followSvc.Follow("bob", "alice"); err != nil {
		t.Fatalf("follow: %v", err)
	}
	feed, err := f.feedSvc.Feed(f.bob, "", 10)
	if err != nil || len(feed) != 2 {
		t.Fatalf("follower feed after link: %v err=%v", len(feed), err)
	}
	carolFeed, err := f.feedSvc.Feed(f.carol, "", 10)
	if err != nil || len(carolFeed) != 0 {
		t.Fatalf("non-follower feed: %v err=%v", len(carolFeed), err)
	}
	if _, err := f.visSvc.Set(f.alice, "a.txt", models.VisibilityPrivate, false); err != nil {
		t.Fatalf("downgrade: %v", err)
	}
	after, err := f.feedSvc.Feed(f.bob, "", 10)
	if err != nil {
		t.Fatalf("feed after downgrade: %v", err)
	}
	for _, e := range after {
		if e.Owner == "alice" && e.Path == "a.txt" && e.Visibility != models.VisibilityPrivate {
			t.Fatalf("downgraded entry still exposed: %+v", e)
		}
	}

	// Cursor pagination: page through public entries one at a time.
	if _, err := f.visSvc.Set(f.alice, "a.txt", models.VisibilityPublic, true); err != nil {
		t.Fatalf("set public: %v", err)
	}
	page1, err := f.feedSvc.Feed(f.carol, "", 1)
	if err != nil || len(page1) != 1 {
		t.Fatalf("page1: %v err=%v", page1, err)
	}
	page2, err := f.feedSvc.Feed(f.carol, page1[0].ID, 10)
	if err != nil || len(page2) == 0 {
		t.Fatalf("page2: %v err=%v", page2, err)
	}
	if page2[0].ID == page1[0].ID {
		t.Fatal("cursor must advance past the last seen entry")
	}
}
