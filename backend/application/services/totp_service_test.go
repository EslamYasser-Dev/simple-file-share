package services

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

func totpFixture(t *testing.T) (*TOTPService, *fs.UserFileRepository, *auth.PBKDF2Hasher) {
	t.Helper()
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	roles := NewRoleCatalog(fs.NewRoleFileRepository(dir))
	return NewTOTPService(userRepo, hasher, roles), userRepo, hasher
}

func createAlice(t *testing.T, users *fs.UserFileRepository, hasher *auth.PBKDF2Hasher) *models.User {
	t.Helper()
	hash, err := hasher.Hash("secret")
	if err != nil {
		t.Fatal(err)
	}
	alice := &models.User{Username: "alice", PasswordHash: hash, Role: models.RoleMember, Enabled: true, CreatedAt: time.Now().UTC()}
	if err := users.CreateUser(alice); err != nil {
		t.Fatal(err)
	}
	return alice
}

func TestTOTPEnrollConfirmLoginDisableLifecycle(t *testing.T) {
	svc, users, hasher := totpFixture(t)
	alice := createAlice(t, users, hasher)

	// Enroll starts pending: secret set, not yet enforced.
	secret, uri, err := svc.Enroll(alice)
	if err != nil {
		t.Fatalf("enroll: %v", err)
	}
	if secret == "" || uri == "" || alice.TOTPEnabled {
		t.Fatalf("enroll result: secret=%q uri=%q enabled=%v", secret, uri, alice.TOTPEnabled)
	}
	// Pending state persists across reload.
	reloaded, err := users.FindByUsername("alice")
	if err != nil || reloaded.TOTPSecret != secret || reloaded.TOTPEnabled {
		t.Fatalf("pending not persisted: err=%v user=%+v", err, reloaded)
	}

	// Wrong code rejected, still pending.
	if _, err := svc.Confirm(alice, "000000"); err == nil {
		t.Fatal("wrong code must not confirm")
	}
	if alice.TOTPEnabled {
		t.Fatal("failed confirm must leave account pending")
	}

	// Correct code confirms and issues backup codes.
	code, err := auth.TOTPCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	codes, err := svc.Confirm(alice, code)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if !alice.TOTPEnabled || len(codes) != 10 {
		t.Fatalf("confirmed: enabled=%v codes=%d", alice.TOTPEnabled, len(codes))
	}
	// Backup codes stored HASHED, not in the clear.
	stored, err := users.FindByUsername("alice")
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range codes {
		if raw == stored.BackupCodes[0] {
			t.Fatal("backup code stored in the clear")
		}
	}
	if !hasher.Verify(codes[0], stored.BackupCodes[0]) {
		t.Fatal("stored backup hash does not verify")
	}

	// Login gate: challenge without otp, reject wrong, accept right.
	if err := svc.VerifyLogin(alice, ""); !errors.Is(err, domainerrors.ErrTwoFactorRequired) {
		t.Fatalf("empty otp err = %v, want totp_required", err)
	}
	if err := svc.VerifyLogin(alice, "999999"); !errors.Is(err, domainerrors.ErrInvalidCredentials) {
		t.Fatalf("wrong otp err = %v, want invalid credentials", err)
	}
	if err := svc.VerifyLogin(alice, code); err != nil {
		t.Fatalf("valid otp: %v", err)
	}
	// Backup code works once and is consumed.
	backup := codes[3]
	if err := svc.VerifyLogin(alice, backup); err != nil {
		t.Fatalf("backup login: %v", err)
	}
	after, _ := users.FindByUsername("alice")
	if len(after.BackupCodes) != 9 {
		t.Fatalf("backup code not consumed: %d left", len(after.BackupCodes))
	}
	if err := svc.VerifyLogin(alice, backup); !errors.Is(err, domainerrors.ErrInvalidCredentials) {
		t.Fatalf("reused backup err = %v, want invalid credentials", err)
	}

	// Disable with a valid code clears everything.
	fresh, _ := auth.TOTPCode(alice.TOTPSecret, time.Now())
	if err := svc.Disable(alice, fresh); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if alice.TOTPEnabled || alice.TOTPSecret != "" || len(alice.BackupCodes) != 0 {
		t.Fatalf("disable left state behind: %+v", alice)
	}
	// Login no longer requires a second factor.
	if err := svc.VerifyLogin(alice, ""); err != nil {
		t.Fatalf("disabled account must not challenge: %v", err)
	}
}

func TestTOTPEnrollRestartAndGuards(t *testing.T) {
	svc, users, hasher := totpFixture(t)
	alice := createAlice(t, users, hasher)

	if _, err := svc.Confirm(alice, "123456"); err == nil {
		t.Fatal("confirm without enrollment must fail")
	}
	first, _, err := svc.Enroll(alice)
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := svc.Enroll(alice) // restart replaces the secret
	if err != nil || second == first {
		t.Fatalf("restart must mint a fresh secret: err=%v same=%v", err, second == first)
	}
	if alice.TOTPSecret != second {
		t.Fatalf("pending secret = %q, want %q", alice.TOTPSecret, second)
	}

	// Pending disable needs no code.
	if err := svc.Disable(alice, ""); err != nil {
		t.Fatalf("pending disable: %v", err)
	}
	if alice.TOTPSecret != "" {
		t.Fatal("pending disable must clear the secret")
	}
	// Disable when nothing is enabled fails cleanly.
	if err := svc.Disable(alice, "123456"); err == nil {
		t.Fatal("disable without enrollment must fail")
	}
	// Enroll when enabled fails.
	alice.TOTPEnabled = true
	alice.TOTPSecret = first
	if _, _, err := svc.Enroll(alice); err == nil {
		t.Fatal("enroll on enabled account must fail")
	}
	if _, err := svc.Confirm(alice, "123456"); err == nil {
		t.Fatal("confirm on enabled account must fail")
	}
}

func TestTOTPAdminReset(t *testing.T) {
	svc, users, hasher := totpFixture(t)
	alice := createAlice(t, users, hasher)
	alice.TOTPEnabled = true
	alice.TOTPSecret = "SECRET"
	alice.BackupCodes = []string{"hash"}
	if err := users.UpdateUser(alice, alice.Username); err != nil {
		t.Fatal(err)
	}

	admin := &models.User{Username: "root", Role: models.RoleAdmin, IsAdmin: true, Enabled: true}
	member := &models.User{Username: "bob", Role: models.RoleMember, Enabled: true}
	if err := svc.AdminReset(member, "alice"); err == nil {
		t.Fatal("member must not reset two-factor")
	}
	if err := svc.AdminReset(admin, "alice"); err != nil {
		t.Fatalf("admin reset: %v", err)
	}
	after, err := users.FindByUsername("alice")
	if err != nil {
		t.Fatal(err)
	}
	if after.TOTPEnabled || after.TOTPSecret != "" || len(after.BackupCodes) != 0 {
		t.Fatalf("reset left state: %+v", after)
	}
	if err := svc.AdminReset(admin, "ghost"); err == nil {
		t.Fatal("reset of unknown account must fail")
	}
}

func TestTokenServiceLoginEnforcesTwoFactor(t *testing.T) {
	tokens, users, hasher := newTokenServiceFixture(t)
	roles := NewRoleCatalog(nil)
	totp := NewTOTPService(users, hasher, roles)
	tokens.SetTwoFactor(totp)

	hash, _ := hasher.Hash("secret")
	alice := &models.User{Username: "alice", PasswordHash: hash, Role: models.RoleMember, Enabled: true, CreatedAt: time.Now().UTC()}
	if err := users.CreateUser(alice); err != nil {
		t.Fatal(err)
	}
	secret, _, err := totp.Enroll(alice)
	if err != nil {
		t.Fatal(err)
	}
	code, _ := auth.TOTPCode(secret, time.Now())
	if _, err := totp.Confirm(alice, code); err != nil {
		t.Fatal(err)
	}

	// Enrolled: no otp → challenge; wrong → invalid; right → token.
	if _, err := tokens.Login("alice", "secret"); !errors.Is(err, domainerrors.ErrTwoFactorRequired) {
		t.Fatalf("login without otp err = %v, want totp_required", err)
	}
	if _, err := tokens.LoginWithOTP("alice", "secret", "000000"); !errors.Is(err, domainerrors.ErrInvalidCredentials) {
		t.Fatalf("wrong otp err = %v, want invalid credentials", err)
	}
	fresh, _ := auth.TOTPCode(secret, time.Now())
	pair, err := tokens.LoginWithOTP("alice", "secret", fresh)
	if err != nil || pair.AccessToken == "" {
		t.Fatalf("valid otp login: err=%v pair=%+v", err, pair)
	}
	// Password still checked before the second factor.
	if _, err := tokens.LoginWithOTP("alice", "wrong", fresh); !errors.Is(err, domainerrors.ErrInvalidCredentials) {
		t.Fatalf("bad password err = %v", err)
	}
}
