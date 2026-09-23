package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// RegisterHandler creates a new account (public endpoint).
type RegisterHandler struct {
	registerService *services.RegisterUserService
}

func NewRegisterHandler(registerService *services.RegisterUserService) *RegisterHandler {
	return &RegisterHandler{registerService: registerService}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.registerService.Execute(req.Username, req.Password)
	if err != nil {
		respondWithError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, dto.FromUser(user))
}

// MeHandler reports the authenticated user's identity, role, and storage use.
type MeHandler struct {
	infoService *services.UserInfoService
}

func NewMeHandler(infoService *services.UserInfoService) *MeHandler {
	return &MeHandler{infoService: infoService}
}

func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.infoService.Execute(currentUser(r))
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dto.FromUserStatsSingle(*stats))
}

// AuthInfoHandler reports runtime auth capabilities (public endpoint).
type AuthInfoHandler struct {
	signupEnabled bool
	oauth         []string
}

func NewAuthInfoHandler(signupEnabled bool, oauthProviders []string) *AuthInfoHandler {
	return &AuthInfoHandler{signupEnabled: signupEnabled, oauth: oauthProviders}
}

func (h *AuthInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	respondJSON(w, http.StatusOK, dto.AuthInfoResponse{SignupEnabled: h.signupEnabled, OAuth: h.oauth})
}

// AdminUsersHandler lists accounts with storage usage. The authorization gate
// (system view only) lives in ListUsersService.
type AdminUsersHandler struct {
	usersService *services.ListUsersService
}

func NewAdminUsersHandler(usersService *services.ListUsersService) *AdminUsersHandler {
	return &AdminUsersHandler{usersService: usersService}
}

func (h *AdminUsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.usersService.Execute(currentUser(r))
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dto.FromUserStats(stats))
}

// AdminQuotaHandler sets an account's storage quota. The authorization gate
// (system view only) lives in UpdateUserQuotaService.
type AdminQuotaHandler struct {
	quotaService *services.UpdateUserQuotaService
}

func NewAdminQuotaHandler(quotaService *services.UpdateUserQuotaService) *AdminQuotaHandler {
	return &AdminQuotaHandler{quotaService: quotaService}
}

func (h *AdminQuotaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.PathValue("username")
	if username == "" {
		respondError(w, http.StatusBadRequest, "username is required")
		return
	}

	var req dto.SetQuotaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	stats, err := h.quotaService.Execute(currentUser(r), username, req.Quota)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dto.FromUserStatsSingle(*stats))
}
