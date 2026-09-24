package services

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

type shareFixture struct {
	fileRepo  *fs.IndexedFileRepository
	shareRepo *fs.ShareFileRepository
	scoper    *policy.PathScoper
	create    *CreateShareService
	list      *ListSharesService
	revoke    *RevokeShareService
	resolve   *ResolveShareService
}

func newShareFixture(t *testing.T) *shareFixture {
	t.Helper()
	dir := t.TempDir()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memoryIndex(), nil)
	shareRepo := fs.NewShareFileRepository(dir)
	scoper := policy.NewPathScoper()
	downloadService := NewDownloadService(
		NewDownloadFileService(fileRepo, scoper),
		NewDownloadZipService(fileRepo, scoper),
	)
	return &shareFixture{
		fileRepo:  fileRepo,
		shareRepo: shareRepo,
		scoper:    scoper,
		create:    NewCreateShareService(fileRepo, shareRepo, scoper),
		list:      NewListSharesService(shareRepo, scoper, NewRoleCatalog(nil)),
		revoke:    NewRevokeShareService(shareRepo, scoper, NewRoleCatalog(nil)),
		resolve:   NewResolveShareService(shareRepo, scoper, downloadService),
	}
}

func memoryIndex() *memory.FileIndexRepository {
	return memory.NewFileIndexRepository()
}

// TestShareCreateRejectsMissingPath verifies that shares only point at paths
// that actually exist, and that validity is bounded.
func TestCreateShareRejectsMissingPathAndBadValidity(t *testing.T) {
	f := newShareFixture(t)
	alice := &models.User{Username: "alice"}

	if _, err := f.create.Execute(alice, "/missing.txt", 60); err == nil {
		t.Fatal("expected NotFound for missing path")
	}
	if _, err := f.create.Execute(alice, "notes.txt", -1); err == nil {
		t.Fatal("expected validation error for negative validity")
	}
	if _, err := f.create.Execute(alice, "notes.txt", 400*24*3600); err == nil {
		t.Fatal("expected validation error for excessive validity")
	}
}

// TestShareLifecycleForUser covers create → list → resolve → revoke for a
// private file, including the public download path through the same services
// the HTTP handler uses.
func TestShareLifecycleForUser(t *testing.T) {
	f := newShareFixture(t)
	alice := &models.User{Username: "alice"}

	if _, err := f.fileRepo.WriteFile("users/alice/notes.txt", io.NopCloser(strings.NewReader("hi"))); err != nil {
		t.Fatalf("write file: %v", err)
	}

	share, err := f.create.Execute(alice, "/notes.txt", 3600)
	if err != nil {
		t.Fatalf("create share: %v", err)
	}
	if share.Token == "" {
		t.Fatal("empty token")
	}
	if share.ExpiresAt.IsZero() {
		t.Fatal("expected an expiration for a finite validity")
	}
	if share.Path != "notes.txt" {
		t.Errorf("stored path = %q, want notes.txt (virtual)", share.Path)
	}

	listed, err := f.list.Execute(alice)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 1 || listed[0].Token != share.Token {
		t.Errorf("alice shares = %v, want exactly the created one", listed)
	}

	// Another user cannot see alice's shares.
	bobShares, err := f.list.Execute(&models.User{Username: "bob"})
	if err != nil {
		t.Fatalf("bob list: %v", err)
	}
	if len(bobShares) != 0 {
		t.Errorf("bob sees %d shares, want 0", len(bobShares))
	}

	download, err := f.resolve.Execute(share.Token)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	body, _ := io.ReadAll(download.Stream)
	download.Stream.Close()
	if string(body) != "hi" {
		t.Errorf("served body = %q, want hi", body)
	}

	if err := f.revoke.Execute(alice, share.Token); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := f.resolve.Execute(share.Token); err == nil {
		t.Fatal("expected error resolving a revoked share")
	}
}

// TestCreateShareMakesNeverExpiringLinksWhenAsked verifies the "0 = never"
// contract.
func TestCreateShareNeverExpires(t *testing.T) {
	f := newShareFixture(t)
	if _, err := f.fileRepo.WriteFile("notes.txt", io.NopCloser(strings.NewReader("x"))); err != nil {
		t.Fatalf("write file: %v", err)
	}
	share, err := f.create.Execute(nil, "/notes.txt", 0)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !share.ExpiresAt.IsZero() {
		t.Errorf("expected no expiration, got %v", share.ExpiresAt)
	}
}

// TestResolveShareRejectsExpiredLink verifies links stop working after their
// validity window.
func TestResolveShareRejectsExpiredLink(t *testing.T) {
	f := newShareFixture(t)
	alice := &models.User{Username: "alice"}
	if _, err := f.fileRepo.WriteFile("users/alice/notes.txt", io.NopCloser(strings.NewReader("x"))); err != nil {
		t.Fatalf("write file: %v", err)
	}
	f.create.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	share, err := f.create.Execute(alice, "/notes.txt", 3600)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	f.create.now = time.Now

	var expired *domainerrors.ShareExpiredError
	if _, err := f.resolve.Execute(share.Token); !errors.As(err, &expired) {
		t.Fatalf("resolve = %v, want ShareExpiredError", err)
	}
}

// TestRevokeShareEnforcesOwnership verifies a regular user cannot revoke
// someone else's link, while an admin can.
func TestRevokeShareEnforcesOwnership(t *testing.T) {
	f := newShareFixture(t)
	alice := &models.User{Username: "alice"}
	if _, err := f.fileRepo.WriteFile("users/alice/notes.txt", io.NopCloser(strings.NewReader("x"))); err != nil {
		t.Fatalf("write file: %v", err)
	}
	share, err := f.create.Execute(alice, "/notes.txt", 60)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	var forbidden *domainerrors.ForbiddenError
	if err := f.revoke.Execute(&models.User{Username: "bob"}, share.Token); !errors.As(err, &forbidden) {
		t.Fatalf("bob revoke = %v, want ForbiddenError", err)
	}
	if err := f.revoke.Execute(&models.User{Username: "alice", IsAdmin: true}, share.Token); err != nil {
		t.Fatalf("admin revoke = %v, want nil", err)
	}
}

// TestResolveShareServesDeletedFileGone verifies a share whose backing file was
// deleted yields a not-found error rather than a corrupted stream.
func TestResolveShareServesDeletedFileGone(t *testing.T) {
	f := newShareFixture(t)
	if _, err := f.fileRepo.WriteFile("notes.txt", io.NopCloser(strings.NewReader("x"))); err != nil {
		t.Fatalf("write file: %v", err)
	}
	share, err := f.create.Execute(nil, "/notes.txt", 0)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := f.fileRepo.DeletePath("notes.txt"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	var notFound *domainerrors.NotFoundError
	if _, err := f.resolve.Execute(share.Token); !errors.As(err, &notFound) {
		t.Fatalf("resolve = %v, want NotFoundError", err)
	}
}

// TestShareCreatedByAdminOfAnotherUsersHome verifies an admin can share a file
// living under another account's home and that the link re-scopes correctly at
// serve time (previously a 403: the reconstructed owner looked like a regular
// user).
func TestShareCreatedByAdminOfAnotherUsersHome(t *testing.T) {
	f := newShareFixture(t)
	admin := &models.User{Username: "root", IsAdmin: true}
	if _, err := f.fileRepo.WriteFile("users/alice/notes.txt", io.NopCloser(strings.NewReader("secret"))); err != nil {
		t.Fatalf("write file: %v", err)
	}

	share, err := f.create.Execute(admin, "/users/alice/notes.txt", 0)
	if err != nil {
		t.Fatalf("admin create: %v", err)
	}
	if !share.IsAdmin {
		t.Fatal("expected the admin flag to be persisted on the share")
	}

	download, err := f.resolve.Execute(share.Token)
	if err != nil {
		t.Fatalf("admin share resolve = %v, want nil", err)
	}
	body, _ := io.ReadAll(download.Stream)
	download.Stream.Close()
	if string(body) != "secret" {
		t.Errorf("served body = %q, want secret", body)
	}

	// A regular user's share never marks IsAdmin, so a forged flag on the
	// stored share is the only path to admin scope and is enforced by the
	// repository being server-owned.
	alice := &models.User{Username: "alice"}
	if _, err := f.fileRepo.WriteFile("users/alice/own.txt", io.NopCloser(strings.NewReader("x"))); err != nil {
		t.Fatalf("write file: %v", err)
	}
	regular, err := f.create.Execute(alice, "/own.txt", 0)
	if err != nil {
		t.Fatalf("regular create: %v", err)
	}
	if regular.IsAdmin {
		t.Fatal("regular-user share must not carry the admin flag")
	}
}
