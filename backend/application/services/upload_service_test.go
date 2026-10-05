package services

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

func newUploadFixture(t *testing.T) (*UploadService, *fs.UserFileRepository, string, *policy.PathScoper, *models.User) {
	t.Helper()
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	scoper := policy.NewPathScoper()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index, nil)
	userRepo := fs.NewUserFileRepository(dir)

	if err := userRepo.CreateUser(&models.User{
		Username:   "quota-user",
		IsAdmin:    false,
		QuotaBytes: 0,
		CreatedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return NewUploadService(fileRepo, scoper, index, userRepo, 0), userRepo, dir, scoper, &models.User{Username: "quota-user"}
}

func uploadPart(name, content string) models.UploadPart {
	return models.UploadPart{Name: name, Content: io.NopCloser(strings.NewReader(content))}
}

// A mid-write quota overflow must fail without committing the truncated
// bytes: previously the budget reader returned io.EOF, io.Copy reported
// success, and a partial (or empty) file replaced the original.
func TestUploadQuotaOverflowDoesNotCommitPartialFile(t *testing.T) {
	uploadService, userRepo, dir, scoper, user := newUploadFixture(t)
	if err := userRepo.SetQuotaBytes("quota-user", 10); err != nil {
		t.Fatal(err)
	}

	// Fill 6 of the 10 bytes so 4 remain.
	if _, err := uploadService.Execute(user, []models.UploadPart{uploadPart("keep.txt", "123456")}); err != nil {
		t.Fatalf("seed upload: %v", err)
	}

	// An 8-byte file does not fit the 4 remaining bytes: rejected mid-copy.
	var quotaErr *domainerrors.QuotaExceededError
	if _, err := uploadService.Execute(user, []models.UploadPart{uploadPart("over.txt", "12345678")}); !errors.As(err, &quotaErr) {
		t.Fatalf("over-quota upload err = %v, want QuotaExceededError", err)
	}

	// The truncated 4-byte file must not exist.
	overPath := filepath.Join(dir, mustWritePath(t, scoper, user, "over.txt"))
	if _, err := os.Stat(overPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial over.txt left behind (stat err = %v)", err)
	}

	// Overwriting a good file must leave the original intact on failure.
	keepPath := filepath.Join(dir, mustWritePath(t, scoper, user, "keep.txt"))
	if _, err := uploadService.Execute(user, []models.UploadPart{uploadPart("keep.txt", "12345678")}); !errors.As(err, &quotaErr) {
		t.Fatalf("overwrite err = %v, want QuotaExceededError", err)
	}
	got, err := os.ReadFile(keepPath)
	if err != nil {
		t.Fatalf("read keep.txt: %v", err)
	}
	if string(got) != "123456" {
		t.Fatalf("keep.txt = %q, want original %q", got, "123456")
	}
}

func mustWritePath(t *testing.T, scoper *policy.PathScoper, user *models.User, rel string) string {
	t.Helper()
	physical, err := scoper.WritePath(user, rel)
	if err != nil {
		t.Fatalf("WritePath(%q): %v", rel, err)
	}
	return physical
}

func newResumableFixture(t *testing.T) (*ResumableUploadService, string, *models.User) {
	t.Helper()
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	scoper := policy.NewPathScoper()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index, nil)
	userRepo := fs.NewUserFileRepository(dir)
	if err := userRepo.CreateUser(&models.User{
		Username:  "resumable-user",
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	sessions := fs.NewUploadSessionRepository(dir)
	return NewResumableUploadService(sessions, fileRepo, scoper, index, userRepo, 0), dir,
		&models.User{Username: "resumable-user"}
}

// A chunk larger than the declared session size must be rejected after at most
// one extra byte is staged, not streamed to disk without bound.
func TestResumableAppendBoundsChunkToDeclaredSize(t *testing.T) {
	svc, _, user := newResumableFixture(t)

	session, err := svc.Start(user, "", "big.bin", "fp-big", 3)
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	var validation *domainerrors.ValidationError
	if _, err := svc.Append(user, session.ID, 0, strings.NewReader("0123456789")); !errors.As(err, &validation) {
		t.Fatalf("oversized chunk err = %v, want ValidationError", err)
	}
	if _, err := svc.Status(user, session.ID); err == nil {
		t.Fatal("lying session should have been discarded")
	}
}

// A staging file that drifted from the recorded offset (for example after a
// crashed writer) must surface as ErrOffsetMismatch so the handler answers
// 409 with the true offset instead of 500.
func TestResumableAppendMapsStagingConflictToOffsetMismatch(t *testing.T) {
	svc, _, user := newResumableFixture(t)

	session, err := svc.Start(user, "", "drift.bin", "fp-drift", 10)
	if err != nil {
		t.Fatalf("start: %v", err)
	}

	// Simulate a crashed writer: staged bytes exist but metadata still says 0.
	staged, err := svc.sessions.OpenStaging(session.ID, true)
	if err != nil {
		t.Fatalf("open staging: %v", err)
	}
	if _, err := staged.Write([]byte("ab")); err != nil {
		t.Fatalf("stage drift: %v", err)
	}
	if err := staged.Close(); err != nil {
		t.Fatalf("close staging: %v", err)
	}

	var mismatch *ErrOffsetMismatch
	if _, err := svc.Append(user, session.ID, 0, strings.NewReader("x")); !errors.As(err, &mismatch) {
		t.Fatalf("drifted staging err = %v, want ErrOffsetMismatch", err)
	}
	if mismatch.Expected != 2 {
		t.Fatalf("Expected = %d, want staged size 2", mismatch.Expected)
	}
}
