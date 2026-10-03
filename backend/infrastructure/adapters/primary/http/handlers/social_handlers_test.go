package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

type socialHTTPFixture struct {
	follows *FollowsHandler
	vis     *VisibilityHandler
	feed    *FeedHandler
	alice   *models.User
	bob     *models.User
}

func newSocialHTTPFixture(t *testing.T) *socialHTTPFixture {
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
	return &socialHTTPFixture{
		follows: NewFollowsHandler(services.NewFollowService(follows, userRepo)),
		vis:     NewVisibilityHandler(services.NewVisibilityService(fileRepo, scoper, visRepo, follows, timeline, downloads, userRepo)),
		feed:    NewFeedHandler(services.NewTimelineService(timeline, follows)),
		alice:   alice,
		bob:     bob,
	}
}

func socialRequest(t *testing.T, h http.Handler, method, target, body string, user *models.User) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	req = req.WithContext(authctx.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// TestSocialRoutesAuthZ drives the new endpoints end to end: follow,
// visibility set/get, and feed gating for a stranger vs a follower, plus the
// anonymous feed rejection and the 404/409/400 mappings.
func TestSocialRoutesAuthZ(t *testing.T) {
	f := newSocialHTTPFixture(t)

	// Follow unknown user: 404, not 401.
	rec := socialRequest(t, f.follows, http.MethodPost, "/api/follows", `{"username":"nobody"}`, f.bob)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("follow unknown: got %d %s", rec.Code, rec.Body.String())
	}
	// Follow self: 400.
	rec = socialRequest(t, f.follows, http.MethodPost, "/api/follows", `{"username":"bob"}`, f.bob)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("follow self: got %d", rec.Code)
	}
	// Follow alice: 201. Again: 409.
	rec = socialRequest(t, f.follows, http.MethodPost, "/api/follows", `{"username":"alice"}`, f.bob)
	if rec.Code != http.StatusCreated {
		t.Fatalf("follow: got %d %s", rec.Code, rec.Body.String())
	}
	rec = socialRequest(t, f.follows, http.MethodPost, "/api/follows", `{"username":"alice"}`, f.bob)
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate follow: got %d", rec.Code)
	}
	// Check + lists.
	rec = socialRequest(t, f.follows, http.MethodGet, "/api/follows?check=alice", "", f.bob)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"following":true`) {
		t.Fatalf("check: got %d %s", rec.Code, rec.Body.String())
	}
	rec = socialRequest(t, f.follows, http.MethodGet, "/api/follows?user=alice&direction=followers", "", f.bob)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "bob") {
		t.Fatalf("followers: got %d %s", rec.Code, rec.Body.String())
	}

	// Visibility: stranger cannot set (404 in their own namespace — no
	// cross-user existence oracle); owner sets link (200).
	rec = socialRequest(t, f.vis, http.MethodPut, "/api/visibility", `{"path":"a.txt","level":"link","allowStream":true}`, f.bob)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("stranger set: got %d %s", rec.Code, rec.Body.String())
	}
	rec = socialRequest(t, f.vis, http.MethodPut, "/api/visibility", `{"path":"a.txt","level":"link","allowStream":true}`, f.alice)
	if rec.Code != http.StatusOK {
		t.Fatalf("owner set: got %d %s", rec.Code, rec.Body.String())
	}
	rec = socialRequest(t, f.vis, http.MethodGet, "/api/visibility?owner=alice&path=a.txt", "", f.bob)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"level":"link"`) {
		t.Fatalf("get: got %d %s", rec.Code, rec.Body.String())
	}
	// Bad level: 400.
	rec = socialRequest(t, f.vis, http.MethodPut, "/api/visibility", `{"path":"a.txt","level":"everyone"}`, f.alice)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad level: got %d", rec.Code)
	}

	// Feed records an upload for alice via the service-facing path: upload a
	// file through the timeline recorder indirectly by recording here would
	// bypass authZ, so only visibility+follow state matters; the feed must
	// show the link-scoped visibility event to the follower.
	rec = socialRequest(t, f.feed, http.MethodGet, "/api/feed", "", f.bob)
	if rec.Code != http.StatusOK {
		t.Fatalf("feed: got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"owner":"alice"`) {
		t.Fatalf("follower must see link-scoped entry: %s", rec.Body.String())
	}
	// Anonymous feed: 403.
	rec = socialRequest(t, f.feed, http.MethodGet, "/api/feed", "", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("anonymous feed: got %d", rec.Code)
	}
	// Unfollow: entry disappears.
	rec = socialRequest(t, f.follows, http.MethodDelete, "/api/follows", `{"username":"alice"}`, f.bob)
	if rec.Code != http.StatusOK {
		t.Fatalf("unfollow: got %d", rec.Code)
	}
	rec = socialRequest(t, f.feed, http.MethodGet, "/api/feed", "", f.bob)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), `"owner":"alice"`) {
		t.Fatalf("feed after unfollow: got %d %s", rec.Code, rec.Body.String())
	}
	// Double unfollow: 404.
	rec = socialRequest(t, f.follows, http.MethodDelete, "/api/follows", `{"username":"alice"}`, f.bob)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("double unfollow: got %d", rec.Code)
	}
}
