package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
)

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, err error) {
	switch err.(type) {
	case *domainerrors.NotFoundError:
		respondJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case *domainerrors.ValidationError:
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	default:
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func serveDownload(w http.ResponseWriter, stream io.ReadCloser, filename, contentType string) {
	defer stream.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if _, err := io.Copy(w, stream); err != nil {
		http.Error(w, "stream copy failed", http.StatusInternalServerError)
	}
}

func pathFromQuery(r *http.Request) string {
	path := r.URL.Query().Get("path")
	if path == "" {
		return "/"
	}
	return path
}
