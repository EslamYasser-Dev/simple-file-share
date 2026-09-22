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

func writeIsolationFile(t *testing.T, repo *fs.LocalFileRepository, path, content string) {
	t.Helper()
	if _, err := repo.WriteFile(path, io.NopCloser(strings.NewReader(content))); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

type isolationFixture struct {
	list     *ListFilesService
	download *DownloadFileService
	search   *SearchFilesService
	upload   *UploadService
	mkdir    *CreateDirectoryService
	remove   *DeletePathService
	update   *UpdateFileContentService
}

func newIsolationFixture(t *testing.T) *isolationFixture {
	t.Helper()

	dir := t.TempDir()
	repo := fs.NewLocalFileRepository(dir)
	scoper := policy.NewPathScoper()

	for _, home := range []string{"users/alice", "users/bob", "shared"} {
		if err := repo.CreateDirectory(home); err != nil {
			t.Fatalf("create %s: %v", home, err)
		}
	}
	writeIsolationFile(t, repo, "users/alice/alice-report.txt", "alice")
	writeIsolationFile(t, repo, "users/bob/bob-report.txt", "bob")
	writeIsolationFile(t, repo, "shared/shared-report.txt", "public")

	index := memory.NewFileIndexRepository()
	for _, entry := range []*models.FileInfo{
		{Name: "alice-report.txt", Path: "users/alice/alice-report.txt", Size: 5, Modified: time.Now().UTC()},
		{Name: "bob-report.txt", Path: "users/bob/bob-report.txt", Size: 3, Modified: time.Now().UTC()},
		{Name: "shared-report.txt", Path: "shared/shared-report.txt", Size: 6, Modified: time.Now().UTC()},
	} {
		if err := index.Upsert(entry); err != nil {
			t.Fatalf("index %s: %v", entry.Path, err)
		}
	}

	return &isolationFixture{
		list:     NewListFilesService(repo, scoper),
		download: NewDownloadFileService(repo, scoper),
		search:   NewSearchFilesService(index, scoper),
		upload:   NewUploadService(repo, scoper, 0),
		mkdir:    NewCreateDirectoryService(repo, scoper),
		remove:   NewDeletePathService(repo, scoper),
		update:   NewUpdateFileContentService(repo, scoper),
	}
}

func TestUsersSeeOnlyTheirOwnRootListing(t *testing.T) {
	fixture := newIsolationFixture(t)
	alice := &models.User{Username: "alice"}
	bob := &models.User{Username: "bob"}

	aliceRoot, err := fixture.list.Execute(alice, "/")
	if err != nil {
		t.Fatalf("list alice root: %v", err)
	}
	if len(aliceRoot.Files) != 1 || aliceRoot.Files[0].Path != "alice-report.txt" {
		t.Fatalf("alice root = %+v, want only alice-report.txt", aliceRoot.Files)
	}

	bobRoot, err := fixture.list.Execute(bob, "/")
	if err != nil {
		t.Fatalf("list bob root: %v", err)
	}
	if len(bobRoot.Files) != 1 || bobRoot.Files[0].Path != "bob-report.txt" {
		t.Fatalf("bob root = %+v, want only bob-report.txt", bobRoot.Files)
	}
}

func TestUsersCannotReachAnotherUsersTree(t *testing.T) {
	fixture := newIsolationFixture(t)
	alice := &models.User{Username: "alice"}

	for _, path := range []string{"users", "users/bob", "users/bob/bob-report.txt"} {
		_, err := fixture.list.Execute(alice, path)
		var forbidden *domainerrors.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Errorf("list %q error = %v, want ForbiddenError", path, err)
		}
	}

	_, err := fixture.download.Execute(alice, "users/bob/bob-report.txt")
	var forbidden *domainerrors.ForbiddenError
	if !errors.As(err, &forbidden) {
		t.Errorf("download error = %v, want ForbiddenError", err)
	}
}

func TestUsersCanReadOwnAndSharedFiles(t *testing.T) {
	fixture := newIsolationFixture(t)
	alice := &models.User{Username: "alice"}

	own, err := fixture.download.Execute(alice, "alice-report.txt")
	if err != nil {
		t.Fatalf("download own file: %v", err)
	}
	body, err := io.ReadAll(own.Stream)
	_ = own.Stream.Close()
	if err != nil {
		t.Fatalf("read own file: %v", err)
	}
	if string(body) != "alice" {
		t.Errorf("own body = %q, want %q", body, "alice")
	}

	shared, err := fixture.download.Execute(alice, "shared/shared-report.txt")
	if err != nil {
		t.Fatalf("download shared file: %v", err)
	}
	_ = shared.Stream.Close()

	results, err := fixture.search.Execute(alice, "report", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	got := map[string]bool{}
	for _, info := range results {
		got[info.Path] = true
	}
	if !got["alice-report.txt"] || !got["shared/shared-report.txt"] || got["users/bob/bob-report.txt"] || got["bob-report.txt"] {
		t.Errorf("search paths = %v, want only alice-report.txt and shared/shared-report.txt", got)
	}
}

func TestUsersCannotWriteOutsideTheirNamespace(t *testing.T) {
	fixture := newIsolationFixture(t)
	alice := &models.User{Username: "alice"}

	operations := map[string]func() error{
		"upload": func() error {
			_, err := fixture.upload.Execute(alice, []models.UploadPart{{
				Name:    "users/bob/evil.txt",
				Content: io.NopCloser(strings.NewReader("evil")),
			}})
			return err
		},
		"mkdir": func() error {
			return fixture.mkdir.Execute(alice, "users/bob/evil")
		},
		"update": func() error {
			_, err := fixture.update.Execute(alice, "users/bob/bob-report.txt", "changed")
			return err
		},
		"delete": func() error {
			return fixture.remove.Execute(alice, "users/bob/bob-report.txt")
		},
		"shared upload": func() error {
			_, err := fixture.upload.Execute(alice, []models.UploadPart{{
				Name:    "shared/evil.txt",
				Content: io.NopCloser(strings.NewReader("evil")),
			}})
			return err
		},
	}

	for name, operation := range operations {
		err := operation()
		var forbidden *domainerrors.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Errorf("%s error = %v, want ForbiddenError", name, err)
		}
	}
}

func TestAdminKeepsFullVisibility(t *testing.T) {
	fixture := newIsolationFixture(t)
	admin := &models.User{Username: "root", IsAdmin: true}

	other, err := fixture.list.Execute(admin, "users/bob")
	if err != nil {
		t.Fatalf("admin list bob tree: %v", err)
	}
	if len(other.Files) != 1 || other.Files[0].Path != "users/bob/bob-report.txt" {
		t.Fatalf("admin bob tree = %+v, want physical users/bob/bob-report.txt", other.Files)
	}
}

func TestUploadRespectsSizeLimit(t *testing.T) {
	dir := t.TempDir()
	scoper := policy.NewPathScoper()
	repo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memory.NewFileIndexRepository())
	service := NewUploadService(repo, scoper, 5)

	// Exactly on the limit is allowed.
	uploads, err := service.Execute(nil, []models.UploadPart{{
		Name:    "exact.txt",
		Content: io.NopCloser(strings.NewReader("12345")),
	}})
	if err != nil {
		t.Fatalf("exact-limit upload = %v", err)
	}
	if len(uploads) != 1 || uploads[0].Size != 5 {
		t.Fatalf("uploads = %+v, want 1 file of 5 bytes", uploads)
	}

	// Past the limit is rejected with a validation error.
	var validation *domainerrors.ValidationError
	if _, err = service.Execute(nil, []models.UploadPart{{
		Name:    "over.txt",
		Content: io.NopCloser(strings.NewReader("123456")),
	}}); !errors.As(err, &validation) {
		t.Fatalf("over-limit upload = %v, want ValidationError", err)
	}
}

func TestUploadRejectsEmptyPayloadAndWritesToDestination(t *testing.T) {
	fixture := newIsolationFixture(t)

	var validation *domainerrors.ValidationError
	if _, err := fixture.upload.Execute(nil, nil); !errors.As(err, &validation) {
		t.Fatalf("no-part upload = %v, want ValidationError", err)
	}

	// The destination prefix joins with the filename.
	uploads, err := fixture.upload.Execute(nil, []models.UploadPart{{
		Name:        "a.txt",
		Destination: "docs",
		Content:     io.NopCloser(strings.NewReader("hi")),
	}})
	if err != nil {
		t.Fatalf("upload = %v", err)
	}
	if len(uploads) != 1 || uploads[0].Filename != "docs/a.txt" {
		t.Fatalf("uploads = %+v, want docs/a.txt", uploads)
	}
}
