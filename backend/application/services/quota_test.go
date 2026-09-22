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

func newQuotaFixture(t *testing.T) (*UpdateUserQuotaService, *UploadService, *fs.UserFileRepository) {
	t.Helper()
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	scoper := policy.NewPathScoper()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index)
	userRepo := fs.NewUserFileRepository(dir)

	if err := userRepo.CreateUser(&models.User{
		Username:   "quota-user",
		IsAdmin:    false,
		QuotaBytes: 0,
		CreatedAt:  time.Now().UTC(),
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	return NewUpdateUserQuotaService(userRepo, index, scoper),
		NewUploadService(fileRepo, scoper, index, userRepo, 0),
		userRepo
}

func TestUpdateUserQuotaServiceAdminOnly(t *testing.T) {
	quotaService, _, _ := newQuotaFixture(t)

	var forbidden *domainerrors.ForbiddenError
	if _, err := quotaService.Execute(&models.User{Username: "someone"}, "quota-user", "1GB"); !errors.As(err, &forbidden) {
		t.Fatalf("non-admin err = %v, want ForbiddenError", err)
	}

	admin := &models.User{Username: "quota-user", IsAdmin: true}
	if _, err := quotaService.Execute(admin, "quota-user", "unlimited"); err != nil {
		t.Fatalf("system-view clamp err = %v", err)
	}
}

func TestUpdateUserQuotaServicePersistsAndRejectsInvalid(t *testing.T) {
	quotaService, _, userRepo := newQuotaFixture(t)
	nilUser := &models.User{Username: "system", IsAdmin: true}

	if _, err := quotaService.Execute(nil, "quota-user", "gibberish"); err == nil {
		t.Fatal("expected validation error for invalid quota")
	}
	var validation *domainerrors.ValidationError
	if _, err := quotaService.Execute(nilUser, "quota-user", "-10GB"); !errors.As(err, &validation) {
		t.Fatalf("negative quota err = %v, want ValidationError", err)
	}

	if _, err := quotaService.Execute(nilUser, "quota-user", "10MB"); err != nil {
		t.Fatalf("set quota: %v", err)
	}
	q, err := userRepo.GetQuotaBytes("quota-user")
	if err != nil || q != 10<<20 {
		t.Fatalf("GetQuotaBytes = %d, %v; want %d", q, err, 10<<20)
	}

	if _, err := quotaService.Execute(nilUser, "missing-user", "10MB"); !errors.Is(err, domainerrors.ErrUserNotFound) {
		t.Fatalf("missing user err = %v, want ErrUserNotFound", err)
	}
}

func TestUploadEnforcesAccountQuota(t *testing.T) {
	_, uploadService, userRepo := newQuotaFixture(t)
	if err := userRepo.SetQuotaBytes("quota-user", 5); err != nil {
		t.Fatal(err)
	}

	user := &models.User{Username: "quota-user"}
	part := func(name, content string) models.UploadPart {
		return models.UploadPart{Name: name, Content: io.NopCloser(strings.NewReader(content))}
	}

	// Exactly at the remaining quota is fine.
	uploads, err := uploadService.Execute(user, []models.UploadPart{part("exact.txt", "12345")})
	if err != nil {
		t.Fatalf("exact-quota upload: %v", err)
	}
	if len(uploads) != 1 {
		t.Fatalf("uploads = %+v", uploads)
	}

	// The account is now full: any further write is rejected with a quota error.
	var quotaErr *domainerrors.QuotaExceededError
	if _, err := uploadService.Execute(user, []models.UploadPart{part("over.txt", "1")}); !errors.As(err, &quotaErr) {
		t.Fatalf("over-quota upload err = %v, want QuotaExceededError", err)
	}
}

func TestUploadUnlimitedWithoutQuota(t *testing.T) {
	_, uploadService, _ := newQuotaFixture(t)
	user := &models.User{Username: "quota-user"}

	uploads, err := uploadService.Execute(user, []models.UploadPart{{
		Name:    "big.txt",
		Content: io.NopCloser(strings.NewReader(strings.Repeat("x", 1024))),
	}})
	if err != nil || len(uploads) != 1 {
		t.Fatalf("unlimited upload = %+v, %v", uploads, err)
	}
}
