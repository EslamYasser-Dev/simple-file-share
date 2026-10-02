package services

import (
	"time"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// Audit action vocabulary. Handlers share these strings so the admin query
// API can filter on them; keep them dot-separated and stable.
const (
	AuditLoginOK        = "login.ok"
	AuditLoginFail      = "login.fail"
	AuditTokenRevoke    = "token.revoke"
	AuditUserRegister   = "user.register"
	AuditUserCreate     = "user.create"
	AuditUserUpdate     = "user.update"
	AuditUserDelete     = "user.delete"
	AuditUserPassword   = "user.password"
	AuditUserQuota      = "user.quota"
	AuditRoleUpsert     = "role.upsert"
	AuditRoleDelete     = "role.delete"
	AuditShareCreate    = "share.create"
	AuditShareRevoke    = "share.revoke"
	AuditPasswordChange = "password.change"
	AuditTotpEnroll     = "totp.enroll"
	AuditTotpConfirm    = "totp.confirm"
	AuditTotpDisable    = "totp.disable"
	AuditTotpAdminReset = "totp.admin_reset"
	AuditAPIKeyCreate   = "apikey.create"
	AuditAPIKeyRevoke   = "apikey.revoke"
)

// AuditService records security events. The store may be nil (audit
// disabled), in which case every method is a no-op. Record never fails the
// caller: an audit write error must not break the business operation.
type AuditService struct {
	store ports.AuditLog
	roles *RoleCatalog
}

// NewAuditService wraps a store (nil disables auditing).
func NewAuditService(store ports.AuditLog, roles *RoleCatalog) *AuditService {
	return &AuditService{store: store, roles: roles}
}

// Enabled reports whether a durable trail exists.
func (s *AuditService) Enabled() bool {
	return s != nil && s.store != nil
}

// Record appends one entry. Free-form fields are clipped so a hostile
// client cannot bloat the trail with megabyte "usernames".
func (s *AuditService) Record(action, actor, remote, target, detail string) {
	if !s.Enabled() {
		return
	}
	if actor == "" {
		actor = "-"
	}
	_ = s.store.Record(ports.AuditEntry{
		At:     time.Now().UTC(),
		Action: action,
		Actor:  clipAudit(actor, 64),
		Target: clipAudit(target, 128),
		Remote: clipAudit(remote, 64),
		Detail: clipAudit(detail, 256),
	})
}

// Query returns one page of the trail for the admin console. Authorization
// mirrors ListUsersService: system view (auth disabled) or users.read.
func (s *AuditService) Query(user *models.User, q ports.AuditQuery) ([]ports.AuditEntry, string, error) {
	if !s.Enabled() {
		return []ports.AuditEntry{}, "", nil
	}
	if user == nil {
		// auth disabled — system view
	} else if !user.IsSystemView() && !user.HasPermission(models.PermUsersRead, s.roles) {
		return nil, "", &domainerrors.ForbiddenError{Action: "read audit log"}
	}
	entries, next, err := s.store.Query(q)
	if err != nil {
		return nil, "", err
	}
	if entries == nil {
		entries = []ports.AuditEntry{}
	}
	return entries, next, nil
}

func clipAudit(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
