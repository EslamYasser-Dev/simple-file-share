package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http"
)

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	if errors.Is(err, domainerrors.ErrUserAlreadyExists) {
		respondJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	if errors.Is(err, domainerrors.ErrInvalidCredentials) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}
	if errors.Is(err, domainerrors.ErrUserNotFound) {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	var notFound *domainerrors.NotFoundError
	var validation *domainerrors.ValidationError
	var forbidden *domainerrors.ForbiddenError
	switch {
	case errors.As(err, &notFound):
		respondJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.As(err, &validation):
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.As(err, &forbidden):
		respondJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	default:
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

// currentUser returns the authenticated user for this request. It is nil when
// auth is disabled, in which case handlers treat the request as a system view.
func currentUser(r *http.Request) *models.User {
	return xhttp.UserFromContext(r.Context())
}

func serveDownload(w http.ResponseWriter, stream io.ReadCloser, filename, contentType string) {
	defer stream.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if _, err := io.Copy(w, stream); err != nil {
		http.Error(w, "stream copy failed", http.StatusInternalServerError)
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
		// Headers are already flushed; the client sees a truncated body.
		return
	}
}

func pathFromQuery(r *http.Request) string {
	path := r.URL.Query().Get("path")
	if path == "" {
		return "/"
	}
	return path
}
