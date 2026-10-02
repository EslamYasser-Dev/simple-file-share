package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// TotpHandler serves the self-service two-factor endpoints:
//
//	POST /api/auth/totp/enroll   -> {secret, otpauthUri}  (start enrollment)
//	POST /api/auth/totp/verify   -> {backupCodes:[...]}   (confirm + issue codes)
//	POST /api/auth/totp/disable  -> {status:"disabled"}   (turn 2FA off)
type TotpHandler struct {
	svc   *services.TOTPService
	audit *services.AuditService
}

// NewTOTPHandler constructs the enrollment routes (nil svc disables them).
func NewTOTPHandler(svc *services.TOTPService) *TotpHandler {
	return &TotpHandler{svc: svc}
}

// SetAudit attaches the security audit trail (nil disables recording).
func (h *TotpHandler) SetAudit(a *services.AuditService) { h.audit = a }

type totpCodeRequest struct {
	Code string `json:"code"`
}

func (h *TotpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	switch r.PathValue("action") {
	case "enroll":
		h.handleEnroll(w, r)
	case "verify":
		h.handleVerify(w, r)
	case "disable":
		h.handleDisable(w, r)
	default:
		respondError(w, http.StatusNotFound, "unknown two-factor action")
	}
}

func (h *TotpHandler) handleEnroll(w http.ResponseWriter, r *http.Request) {
	actor := currentUser(r)
	if actor != nil && actor.TOTPEnabled {
		respondError(w, http.StatusConflict, "two-factor already enabled")
		return
	}
	secret, uri, err := h.svc.Enroll(actor)
	if err != nil {
		respondWithError(w, err)
		return
	}
	h.audit.Record(services.AuditTotpEnroll, actorName(r), clientIP(r), actorName(r), "")
	respondJSON(w, http.StatusOK, map[string]string{"secret": secret, "otpauthUri": uri})
}

func (h *TotpHandler) handleVerify(w http.ResponseWriter, r *http.Request) {
	var req totpCodeRequest
	if err := decodeSmallJSON(w, r, &req); err != nil {
		return
	}
	codes, err := h.svc.Confirm(currentUser(r), req.Code)
	if err != nil {
		respondWithError(w, err)
		return
	}
	h.audit.Record(services.AuditTotpConfirm, actorName(r), clientIP(r), "", "backup codes issued")
	respondJSON(w, http.StatusOK, map[string]any{"backupCodes": codes})
}

func (h *TotpHandler) handleDisable(w http.ResponseWriter, r *http.Request) {
	var req totpCodeRequest
	if err := decodeSmallJSON(w, r, &req); err != nil {
		return
	}
	if err := h.svc.Disable(currentUser(r), req.Code); err != nil {
		respondWithError(w, err)
		return
	}
	h.audit.Record(services.AuditTotpDisable, actorName(r), clientIP(r), "", "")
	respondJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
}

// AdminTotpResetHandler serves POST /api/admin/users/{username}/totp/reset —
// an admin clears two-factor state for a locked-out account (users.update).
type AdminTotpResetHandler struct {
	svc   *services.TOTPService
	audit *services.AuditService
}

// NewAdminTotpResetHandler constructs the admin reset route.
func NewAdminTotpResetHandler(svc *services.TOTPService) *AdminTotpResetHandler {
	return &AdminTotpResetHandler{svc: svc}
}

// SetAudit attaches the security audit trail (nil disables recording).
func (h *AdminTotpResetHandler) SetAudit(a *services.AuditService) { h.audit = a }

func (h *AdminTotpResetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.PathValue("username")
	if username == "" {
		respondError(w, http.StatusBadRequest, "username is required")
		return
	}
	if err := h.svc.AdminReset(currentUser(r), username); err != nil {
		respondWithError(w, err)
		return
	}
	h.audit.Record(services.AuditTotpAdminReset, actorName(r), clientIP(r), username, "")
	respondJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}

// decodeSmallJSON reads a ≤1MiB JSON body, writing the error response on
// failure and returning a sentinel the caller should abort on.
func decodeSmallJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return err
	}
	return nil
}
