package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// AdminUserItemHandler handles /api/admin/users for list/create and
// /api/admin/users/{username} for update/delete.
type AdminUserItemHandler struct {
	list   *services.ListUsersService
	create *services.CreateUserService
	update *services.UpdateUserService
	del    *services.DeleteUserService
	reset  *services.ResetPasswordService
}

func NewAdminUserItemHandler(
	list *services.ListUsersService,
	create *services.CreateUserService,
	update *services.UpdateUserService,
	del *services.DeleteUserService,
	reset *services.ResetPasswordService,
) *AdminUserItemHandler {
	return &AdminUserItemHandler{list: list, create: create, update: update, del: del, reset: reset}
}

func (h *AdminUserItemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.PathValue("username"))
	if username == "" {
		h.handleCollection(w, r)
		return
	}
	h.handleItem(w, r, username)
}

func (h *AdminUserItemHandler) handleCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		stats, err := h.list.Execute(currentUser(r))
		if err != nil {
			respondWithError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, dto.FromUserStats(stats))
	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req dto.CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		user, err := h.create.Execute(currentUser(r), services.CreateUserInput{
			Username:        req.Username,
			Password:        req.Password,
			Role:            req.Role,
			QuotaBytes:      req.QuotaBytes,
			Enabled:         req.Enabled,
			UseDefaultQuota: req.UseDefaultQuota,
		})
		if err != nil {
			respondWithError(w, err)
			return
		}
		respondJSON(w, http.StatusCreated, dto.FromUser(user))
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AdminUserItemHandler) handleItem(w http.ResponseWriter, r *http.Request, username string) {
	switch r.Method {
	case http.MethodPatch, http.MethodPut:
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req dto.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		in := services.UpdateUserInput{
			Username: req.Username,
			Role:     req.Role,
			Enabled:  req.Enabled,
		}
		user, err := h.update.Execute(currentUser(r), username, in)
		if err != nil {
			respondWithError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, dto.FromUser(user))
	case http.MethodDelete:
		if err := h.del.Execute(currentUser(r), username); err != nil {
			respondWithError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminUserPasswordHandler handles /api/admin/users/{username}/password.
type AdminUserPasswordHandler struct {
	reset *services.ResetPasswordService
}

func NewAdminUserPasswordHandler(reset *services.ResetPasswordService) *AdminUserPasswordHandler {
	return &AdminUserPasswordHandler{reset: reset}
}

func (h *AdminUserPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := strings.TrimSpace(r.PathValue("username"))
	if username == "" {
		respondError(w, http.StatusBadRequest, "username is required")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.reset.Execute(currentUser(r), username, req.Password); err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}

// SelfPasswordHandler handles POST /api/auth/password (change own password).
type SelfPasswordHandler struct {
	change *services.ChangePasswordService
}

func NewSelfPasswordHandler(change *services.ChangePasswordService) *SelfPasswordHandler {
	return &SelfPasswordHandler{change: change}
}

func (h *SelfPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.change.Execute(currentUser(r), req.CurrentPassword, req.NewPassword); err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "password changed"})
}

// AdminRolesHandler handles /api/admin/roles list + create/upsert.
type AdminRolesHandler struct {
	list   *services.ListRolesService
	upsert *services.CreateOrUpdateRoleService
}

func NewAdminRolesHandler(list *services.ListRolesService, upsert *services.CreateOrUpdateRoleService) *AdminRolesHandler {
	return &AdminRolesHandler{list: list, upsert: upsert}
}

func (h *AdminRolesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		roles, err := h.list.Execute(currentUser(r))
		if err != nil {
			respondWithError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, dto.FromRoles(roles))
	case http.MethodPost, http.MethodPut:
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req dto.UpsertRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		role, err := h.upsert.Execute(currentUser(r), req.Name, req.Description, req.Permissions)
		if err != nil {
			respondWithError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, dto.FromRole(role))
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminRoleItemHandler handles DELETE /api/admin/roles/{name}.
type AdminRoleItemHandler struct {
	del *services.DeleteRoleService
}

func NewAdminRoleItemHandler(del *services.DeleteRoleService) *AdminRoleItemHandler {
	return &AdminRoleItemHandler{del: del}
}

func (h *AdminRoleItemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := strings.TrimSpace(r.PathValue("name"))
	if name == "" {
		respondError(w, http.StatusBadRequest, "role name is required")
		return
	}
	if err := h.del.Execute(currentUser(r), name); err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ensure models import used when building responses without stats.
var _ = models.RoleAdmin
