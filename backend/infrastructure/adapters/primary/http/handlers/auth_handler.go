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

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
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

// MeHandler reports the authenticated user's identity and role.
type MeHandler struct{}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := currentUser(r)
	if user == nil {
		// Auth disabled — expose the system view so the UI renders as admin.
		respondJSON(w, http.StatusOK, dto.UserResponse{IsAdmin: true})
		return
	}
	respondJSON(w, http.StatusOK, dto.FromUser(user))
}

// AuthInfoHandler reports runtime auth capabilities (public endpoint).
type AuthInfoHandler struct {
	signupEnabled bool
}

func NewAuthInfoHandler(signupEnabled bool) *AuthInfoHandler {
	return &AuthInfoHandler{signupEnabled: signupEnabled}
}

func (h *AuthInfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	respondJSON(w, http.StatusOK, dto.AuthInfoResponse{SignupEnabled: h.signupEnabled})
}

// AdminUsersHandler lists accounts with storage usage (admin only).
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

	user := currentUser(r)
	if user == nil || !user.IsAdmin {
		respondError(w, http.StatusForbidden, "admin access required")
		return
	}

	stats, err := h.usersService.Execute()
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dto.FromUserStats(stats))
}
