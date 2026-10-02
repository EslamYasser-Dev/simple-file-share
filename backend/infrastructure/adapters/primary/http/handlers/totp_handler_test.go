package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

type totpFixture struct {
	handler *TotpHandler
	reset   *AdminTotpResetHandler
	svc     *services.TOTPService
	users   *fs.UserFileRepository
}

func newTotpFixture(t *testing.T) totpFixture {
	t.Helper()
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	roles := services.NewRoleCatalog(fs.NewRoleFileRepository(dir))
	svc := services.NewTOTPService(userRepo, hasher, roles)
	h := NewTOTPHandler(svc)
	h.SetAudit(nil)
	reset := NewAdminTotpResetHandler(svc)
	return totpFixture{handler: h, reset: reset, svc: svc, users: userRepo}
}

func (f *totpFixture) createUser(t *testing.T, username, role string) *models.User {
	t.Helper()
	hash, _ := auth.NewPBKDF2Hasher().Hash("secret")
	u := &models.User{Username: username, PasswordHash: hash, Role: role, Enabled: true, CreatedAt: time.Now().UTC()}
	if role == models.RoleAdmin {
		u.IsAdmin = true
	}
	if err := f.users.CreateUser(u); err != nil {
		t.Fatal(err)
	}
	return u
}

func asUser(u *models.User, r *http.Request) *http.Request {
	return r.WithContext(authctx.WithUser(r.Context(), u))
}

func totpPost(h http.Handler, u *models.User, path, body string) *httptest.ResponseRecorder {
	req := asUser(u, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
	// PathValue is populated by the ServeMux; set it manually for direct
	// handler invocation.
	segs := strings.Split(strings.Trim(path, "/"), "/")
	if len(segs) == 4 && segs[0] == "api" && segs[1] == "auth" && segs[2] == "totp" {
		req.SetPathValue("action", segs[3])
	}
	if len(segs) >= 4 && segs[0] == "api" && segs[1] == "admin" && segs[2] == "users" {
		req.SetPathValue("username", segs[3])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestTotpEnrollVerifyDisableFlow(t *testing.T) {
	f := newTotpFixture(t)
	alice := f.createUser(t, "alice", models.RoleMember)

	rec := totpPost(f.handler, alice, "/api/auth/totp/enroll", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("enroll status = %d body=%s", rec.Code, rec.Body)
	}
	var enroll struct {
		Secret     string `json:"secret"`
		OTPAuthURI string `json:"otpauthUri"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &enroll); err != nil || enroll.Secret == "" || enroll.OTPAuthURI == "" {
		t.Fatalf("enroll body = %s err=%v", rec.Body, err)
	}

	// Wrong code → 400, still pending.
	rec = totpPost(f.handler, alice, "/api/auth/totp/verify", `{"code":"000000"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("wrong code status = %d body=%s", rec.Code, rec.Body)
	}

	code, _ := auth.TOTPCode(enroll.Secret, time.Now())
	rec = totpPost(f.handler, alice, "/api/auth/totp/verify", `{"code":"`+code+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("verify status = %d body=%s", rec.Code, rec.Body)
	}
	var verify struct {
		BackupCodes []string `json:"backupCodes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &verify); err != nil || len(verify.BackupCodes) != 10 {
		t.Fatalf("verify body = %s err=%v", rec.Body, err)
	}

	// Enroll again → 409 once enabled.
	rec = totpPost(f.handler, alice, "/api/auth/totp/enroll", "")
	if rec.Code != http.StatusConflict {
		t.Fatalf("re-enroll status = %d, want 409", rec.Code)
	}

	// Disable with fresh code.
	code, _ = auth.TOTPCode(enroll.Secret, time.Now())
	rec = totpPost(f.handler, alice, "/api/auth/totp/disable", `{"code":"`+code+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("disable status = %d body=%s", rec.Code, rec.Body)
	}
	stored, _ := f.users.FindByUsername("alice")
	if stored.TOTPEnabled {
		t.Fatal("disable did not persist")
	}
}

func TestTotpHandlerRoutingAndMethod(t *testing.T) {
	f := newTotpFixture(t)
	alice := f.createUser(t, "alice", models.RoleMember)

	if rec := totpPost(f.handler, alice, "/api/auth/totp/nope", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown action status = %d, want 404", rec.Code)
	}
	req := asUser(alice, httptest.NewRequest(http.MethodGet, "/api/auth/totp/enroll", nil))
	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET status = %d, want 405", rec.Code)
	}
}

func TestAdminTotpReset(t *testing.T) {
	f := newTotpFixture(t)
	alice := f.createUser(t, "alice", models.RoleMember)
	alice.TOTPEnabled = true
	alice.TOTPSecret = "SECRET"
	if err := f.users.UpdateUser(alice, alice.Username); err != nil {
		t.Fatal(err)
	}
	admin := f.createUser(t, "root", models.RoleAdmin)
	member := f.createUser(t, "bob", models.RoleMember)

	if rec := totpPost(f.reset, member, "/api/admin/users/alice/totp/reset", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("member reset status = %d, want 403", rec.Code)
	}
	if rec := totpPost(f.reset, admin, "/api/admin/users/alice/totp/reset", ""); rec.Code != http.StatusOK {
		t.Fatalf("admin reset status = %d body=%s", rec.Code, rec.Body)
	}
	after, _ := f.users.FindByUsername("alice")
	if after.TOTPEnabled || after.TOTPSecret != "" {
		t.Fatalf("reset left state: %+v", after)
	}
}

// Token endpoint: enrolled account without otp gets the totp_required
// challenge; a valid code logs in.
func TestTokenHandlerTwoFactorChallenge(t *testing.T) {
	dir := t.TempDir()
	userRepo := fs.NewUserFileRepository(dir)
	hasher := auth.NewPBKDF2Hasher()
	provider := auth.NewUserAuthProvider(userRepo, hasher)
	jwt := auth.NewJWTManager("test-secret", time.Hour)
	tokens := services.NewTokenService(services.NewAuthenticateService(provider), jwt, userRepo)
	roles := services.NewRoleCatalog(nil)
	totp := services.NewTOTPService(userRepo, hasher, roles)
	tokens.SetTwoFactor(totp)
	tokenH := NewTokenHandler(tokens)
	tokenH.SetAudit(nil)

	hash, _ := hasher.Hash("secret")
	alice := &models.User{Username: "alice", PasswordHash: hash, Role: models.RoleMember, Enabled: true, CreatedAt: time.Now().UTC()}
	if err := userRepo.CreateUser(alice); err != nil {
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

	// No otp → 401 totp_required.
	rec := httptest.NewRecorder()
	tokenH.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/auth/token",
		strings.NewReader(`{"username":"alice","password":"secret"}`)))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body)
	}
	var errBody struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil || errBody.Error != "totp_required" {
		t.Fatalf("body = %s, want totp_required", rec.Body)
	}

	// With otp → 200.
	fresh, _ := auth.TOTPCode(secret, time.Now())
	rec = httptest.NewRecorder()
	tokenH.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/auth/token",
		strings.NewReader(`{"username":"alice","password":"secret","otp":"`+fresh+`"}`)))
	if rec.Code != http.StatusOK {
		t.Fatalf("otp login status = %d body=%s", rec.Code, rec.Body)
	}
}
