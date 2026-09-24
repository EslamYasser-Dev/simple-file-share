package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/authctx"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// respondError writes a standard JSON error body with the given status.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, dto.ErrorResponse{Error: message})
}

func respondWithError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var notFound *domainerrors.NotFoundError
	var validation *domainerrors.ValidationError
	var forbidden *domainerrors.ForbiddenError
	var isDir *domainerrors.IsDirectoryError
	var notDir *domainerrors.NotDirectoryError
	var shareNotFound *domainerrors.ShareNotFoundError
	var shareExpired *domainerrors.ShareExpiredError
	var quotaExceeded *domainerrors.QuotaExceededError

	status := http.StatusInternalServerError
	message := "internal server error"

	switch {
	case errors.Is(err, domainerrors.ErrUserAlreadyExists):
		status, message = http.StatusConflict, err.Error()
	case errors.Is(err, domainerrors.ErrInvalidCredentials):
		// Never distinguish "bad password" from "no such user" — one generic
		// 401 prevents account enumeration through login responses.
		status, message = http.StatusUnauthorized, "authentication failed"
	case errors.Is(err, domainerrors.ErrUserNotFound):
		// User-lookup misses on public auth surfaces must not leak existence.
		status, message = http.StatusUnauthorized, "authentication failed"
	case errors.As(err, &notFound):
		status, message = http.StatusNotFound, err.Error()
	case errors.As(err, &validation):
		status, message = http.StatusBadRequest, err.Error()
	case errors.As(err, &isDir):
		status, message = http.StatusConflict, err.Error()
	case errors.As(err, &notDir):
		status, message = http.StatusBadRequest, err.Error()
	case errors.As(err, &forbidden):
		status, message = http.StatusForbidden, err.Error()
	case errors.As(err, &quotaExceeded):
		status, message = http.StatusRequestEntityTooLarge, err.Error()
	case errors.As(err, &shareExpired):
		status, message = http.StatusGone, err.Error()
	case errors.As(err, &shareNotFound):
		status, message = http.StatusNotFound, err.Error()
	}

	respondJSON(w, status, dto.ErrorResponse{Error: message})
}

// currentUser returns the authenticated user for this request. It is nil when
// auth is disabled, in which case handlers treat the request as a system view.
func currentUser(r *http.Request) *models.User {
	return authctx.UserFromContext(r.Context())
}

func serveDownload(w http.ResponseWriter, stream io.ReadCloser, filename, contentType string) {
	defer stream.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if _, err := io.Copy(w, stream); err != nil {
		// Client likely disconnected; headers already sent, cannot write error status.
		return
	}
}

// serveInline streams content for in-browser rendering (e.g. PDF previews).
// Callers must only pass content types that are safe to render on our origin.
func serveInline(w http.ResponseWriter, stream io.ReadCloser, filename, contentType string) {
	defer stream.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, max-age=60")
	if _, err := io.Copy(w, stream); err != nil {
		// Headers already flushed; client disconnected.
		return
	}
}
