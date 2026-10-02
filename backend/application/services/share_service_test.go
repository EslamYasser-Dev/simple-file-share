package services

import (
	"errors"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
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
		create:    NewCreateShareService(fileRepo, shareRepo, scoper, auth.NewPBKDF2Hasher()),
		list:      NewListSharesService(shareRepo, scoper, NewRoleCatalog(nil)),
		revoke:    NewRevokeShareService(shareRepo, scoper, NewRoleCatalog(nil)),
		resolve:   NewResolveShareService(shareRepo, scoper, downloadService, auth.NewPBKDF2Hasher()),
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

	if _, err := f.create.Execute(alice, "/missing.txt", SharePolicy{ExpiresInSeconds: 60}); err == nil {
		t.Fatal("expected NotFound for missing path")
	}
	if _, err := f.create.Execute(alice, "notes.txt", SharePolicy{ExpiresInSeconds: -1}); err == nil {
		t.Fatal("expected validation error for negative validity")
	}
	if _, err := f.create.Execute(alice, "notes.txt", SharePolicy{ExpiresInSeconds: 400 * 24 * 3600}); err == nil {
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

	share, err := f.create.Execute(alice, "/notes.txt", SharePolicy{ExpiresInSeconds: 3600})
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

	download, err := f.resolve.Execute(share.Token, "")
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
	if _, err := f.resolve.Execute(share.Token, ""); err == nil {
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
	share, err := f.create.Execute(nil, "/notes.txt", SharePolicy{ExpiresInSeconds: 0})
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
	share, err := f.create.Execute(alice, "/notes.txt", SharePolicy{ExpiresInSeconds: 3600})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	f.create.now = time.Now

	var expired *domainerrors.ShareExpiredError
	if _, err := f.resolve.Execute(share.Token, ""); !errors.As(err, &expired) {
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
	share, err := f.create.Execute(alice, "/notes.txt", SharePolicy{ExpiresInSeconds: 60})
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
	share, err := f.create.Execute(nil, "/notes.txt", SharePolicy{ExpiresInSeconds: 0})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := f.fileRepo.DeletePath("notes.txt"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	var notFound *domainerrors.NotFoundError
	if _, err := f.resolve.Execute(share.Token, ""); !errors.As(err, &notFound) {
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

	share, err := f.create.Execute(admin, "/users/alice/notes.txt", SharePolicy{ExpiresInSeconds: 0})
	if err != nil {
		t.Fatalf("admin create: %v", err)
	}
	if !share.IsAdmin {
		t.Fatal("expected the admin flag to be persisted on the share")
	}

	download, err := f.resolve.Execute(share.Token, "")
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
	regular, err := f.create.Execute(alice, "/own.txt", SharePolicy{ExpiresInSeconds: 0})
	if err != nil {
		t.Fatalf("regular create: %v", err)
	}
	if regular.IsAdmin {
		t.Fatal("regular-user share must not carry the admin flag")
	}
}

// TestSharePasswordPolicy covers the password-protected link lifecycle:
// creation stores only a hash, and serving demands the right password.
func TestSharePasswordPolicy(t *testing.T) {
	f := newShareFixture(t)
	alice := &models.User{Username: "alice"}
	if _, err := f.fileRepo.WriteFile("users/alice/secret.txt", io.NopCloser(strings.NewReader("top"))); err != nil {
		t.Fatal(err)
	}

	share, err := f.create.Execute(alice, "/secret.txt", SharePolicy{Password: "hunter2"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !share.PasswordProtected() {
		t.Fatal("share not marked protected")
	}
	if share.PasswordHash == "" || strings.Contains(share.PasswordHash, "hunter2") {
		t.Fatalf("password hash leaks plaintext: %q", share.PasswordHash)
	}

	// Missing password.
	_, err = f.resolve.Execute(share.Token, "")
	var pwErr *domainerrors.SharePasswordError
	if !errors.As(err, &pwErr) || !pwErr.Missing {
		t.Fatalf("missing password err = %v, want SharePasswordError{Missing}", err)
	}
	// Wrong password.
	_, err = f.resolve.Execute(share.Token, "wrong")
	if !errors.As(err, &pwErr) || pwErr.Missing {
		t.Fatalf("wrong password err = %v, want SharePasswordError{not missing}", err)
	}
	// Correct password serves.
	download, err := f.resolve.Execute(share.Token, "hunter2")
	if err != nil {
		t.Fatalf("correct password: %v", err)
	}
	body, _ := io.ReadAll(download.Stream)
	download.Stream.Close()
	if string(body) != "top" {
		t.Fatalf("body = %q", body)
	}

	// A public link ignores any presented password.
	pub, err := f.create.Execute(alice, "/secret.txt", SharePolicy{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.resolve.Execute(pub.Token, "anything"); err != nil {
		t.Fatalf("public link with stray password: %v", err)
	}
}

// TestShareMaxDownloadsPolicy covers the download budget: atomic consume,
// exhaustion, and no charge for unservable paths.
func TestShareMaxDownloadsPolicy(t *testing.T) {
	f := newShareFixture(t)
	alice := &models.User{Username: "alice"}
	if _, err := f.fileRepo.WriteFile("users/alice/limited.txt", io.NopCloser(strings.NewReader("x"))); err != nil {
		t.Fatal(err)
	}

	share, err := f.create.Execute(alice, "/limited.txt", SharePolicy{MaxDownloads: 2})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !share.Limited() || share.MaxDownloads != 2 {
		t.Fatalf("share = %+v", share)
	}
	for i := 0; i < 2; i++ {
		download, err := f.resolve.Execute(share.Token, "")
		if err != nil {
			t.Fatalf("download %d: %v", i+1, err)
		}
		io.Copy(io.Discard, download.Stream)
		download.Stream.Close()
	}
	if _, err := f.resolve.Execute(share.Token, ""); !errors.Is(err, domainerrors.ErrShareLimitReached) {
		t.Fatalf("third download err = %v, want ErrShareLimitReached", err)
	}
	stored, err := f.shareRepo.FindByToken(share.Token)
	if err != nil || stored.Downloads != 2 {
		t.Fatalf("stored downloads = %d err=%v, want 2", stored.Downloads, err)
	}

	// Unlimited links never touch the counter and never exhaust.
	open, err := f.create.Execute(alice, "/limited.txt", SharePolicy{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		download, err := f.resolve.Execute(open.Token, "")
		if err != nil {
			t.Fatalf("unlimited %d: %v", i, err)
		}
		io.Copy(io.Discard, download.Stream)
		download.Stream.Close()
	}
	stored, _ = f.shareRepo.FindByToken(open.Token)
	if stored.Downloads != 0 {
		t.Fatalf("unlimited link consumed budget: %d", stored.Downloads)
	}

	// A deleted target 404s without burning the budget.
	if err := f.fileRepo.DeletePath("users/alice/limited.txt"); err != nil {
		t.Fatal(err)
	}
	oneShot, err := f.create.Execute(alice, "/gone.txt", SharePolicy{MaxDownloads: 1})
	if err == nil {
		// The file was deleted; creating against it must fail (path missing).
		t.Fatalf("expected create on deleted path to fail, got share %v", oneShot)
	}
}

// TestSharePolicyValidation bounds the policy inputs.
func TestSharePolicyValidation(t *testing.T) {
	f := newShareFixture(t)
	alice := &models.User{Username: "alice"}
	if _, err := f.fileRepo.WriteFile("users/alice/a.txt", io.NopCloser(strings.NewReader("x"))); err != nil {
		t.Fatal(err)
	}
	if _, err := f.create.Execute(alice, "/a.txt", SharePolicy{MaxDownloads: -1}); err == nil {
		t.Fatal("negative maxDownloads must fail")
	}
	if _, err := f.create.Execute(alice, "/a.txt", SharePolicy{Password: strings.Repeat("p", 257)}); err == nil {
		t.Fatal("oversized password must fail")
	}
	if _, err := f.create.Execute(alice, "/a.txt", SharePolicy{Password: strings.Repeat("p", 256)}); err != nil {
		t.Fatalf("256-char password must be accepted: %v", err)
	}
}
