package services

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

type versionFixture struct {
	dir      string
	repo     *fs.LocalFileRepository
	list     *ListVersionsService
	download *DownloadVersionService
	restore  *RestoreVersionService
}

func newVersionFixture(t *testing.T) *versionFixture {
	t.Helper()
	dir := t.TempDir()
	repo := fs.NewLocalFileRepository(dir)
	scoper := policy.NewPathScoper()

	for _, home := range []string{"users/alice", "users/bob"} {
		if err := repo.CreateDirectory(home); err != nil {
			t.Fatalf("create %s: %v", home, err)
		}
	}
	for _, content := range []string{"alice-v1", "alice-v2"} {
		if _, err := repo.WriteFile("users/alice/report.txt", io.NopCloser(strings.NewReader(content))); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := repo.WriteFile("users/bob/notes.txt", io.NopCloser(strings.NewReader("bob"))); err != nil {
		t.Fatal(err)
	}

	return &versionFixture{
		dir:      dir,
		repo:     repo,
		list:     NewListVersionsService(repo, repo, scoper),
		download: NewDownloadVersionService(repo, repo, scoper),
		restore:  NewRestoreVersionService(repo, repo, scoper),
	}
}

func TestListVersionsScoping(t *testing.T) {
	f := newVersionFixture(t)
	alice := &models.User{Username: "alice"}
	bob := &models.User{Username: "bob"}

	versions, err := f.list.Execute(alice, "report.txt")
	if err != nil {
		t.Fatalf("alice lists own versions: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("alice versions = %d, want 1", len(versions))
	}
	if versions[0].Path != "report.txt" {
		t.Fatalf("virtual path = %q, want report.txt", versions[0].Path)
	}

	if _, err := f.list.Execute(bob, "report.txt"); err == nil {
		t.Fatal("bob must not see alice's version history")
	}

	var forbidden *domainerrors.ForbiddenError
	if _, err := f.list.Execute(bob, "users/alice/report.txt"); !errors.As(err, &forbidden) {
		t.Fatalf("bob direct path error = %v, want forbidden", err)
	}

	if _, err := f.list.Execute(alice, "missing.txt"); err == nil {
		t.Fatal("expected missing file to fail")
	}
}

func TestDownloadVersionScoping(t *testing.T) {
	f := newVersionFixture(t)
	alice := &models.User{Username: "alice"}
	bob := &models.User{Username: "bob"}

	download, err := f.download.Execute(alice, "report.txt", 1)
	if err != nil {
		t.Fatalf("alice downloads version: %v", err)
	}
	data, _ := io.ReadAll(download.Stream)
	download.Stream.Close()
	if string(data) != "alice-v1" {
		t.Fatalf("version content = %q, want alice-v1", data)
	}
	if download.Filename != "report.txt" {
		t.Fatalf("filename = %q, want report.txt", download.Filename)
	}

	if _, err := f.download.Execute(alice, "report.txt", 0); err == nil {
		t.Fatal("expected version 0 to fail validation")
	}
	if _, err := f.download.Execute(bob, "report.txt", 1); err == nil {
		t.Fatal("bob must not download alice's version")
	}
	if _, err := f.download.Execute(alice, "report.txt", 42); err == nil {
		t.Fatal("expected missing version to fail")
	}
}

func TestRestoreVersion(t *testing.T) {
	f := newVersionFixture(t)
	alice := &models.User{Username: "alice"}

	if err := f.restore.Execute(alice, "report.txt", 1); err != nil {
		t.Fatalf("restore: %v", err)
	}
	got, err := readFileString(f.dir + "/users/alice/report.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got != "alice-v1" {
		t.Fatalf("restored = %q, want alice-v1", got)
	}

	versions, err := f.list.Execute(alice, "report.txt")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 {
		t.Fatalf("versions after restore = %d, want 2", len(versions))
	}

	bob := &models.User{Username: "bob"}
	if err := f.restore.Execute(bob, "report.txt", 1); err == nil {
		t.Fatal("bob must not restore alice's file")
	}
}

func readFileString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
