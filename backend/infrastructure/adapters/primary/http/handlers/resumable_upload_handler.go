package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

// ResumableUploadHandler exposes the chunked upload session API:
//
//	POST   /api/uploads              create (or resume by fingerprint)
//	GET    /api/uploads              list active sessions for the caller
//	GET    /api/uploads/{id}         status / resume probe
//	PATCH  /api/uploads/{id}         append one chunk (Upload-Offset header)
//	POST   /api/uploads/{id}/complete promote staged bytes to the final path
//	DELETE /api/uploads/{id}         abort and discard staged bytes
type ResumableUploadHandler struct {
	svc *services.ResumableUploadService
}

func NewResumableUploadHandler(svc *services.ResumableUploadService) *ResumableUploadHandler {
	return &ResumableUploadHandler{svc: svc}
}

func (h *ResumableUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/uploads":
		h.create(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/api/uploads":
		h.list(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/uploads/"):
		h.status(w, r)
	case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/uploads/"):
		h.append(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/complete"):
		h.complete(w, r)
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/uploads/"):
		h.abort(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ResumableUploadHandler) list(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.svc.List(currentUser(r))
	if err != nil {
		respondWithError(w, err)
		return
	}
	out := make([]dto.PendingUploadSession, 0, len(sessions))
	for _, session := range sessions {
		out = append(out, toPendingDTO(session))
	}
	respondJSON(w, http.StatusOK, out)
}

func (h *ResumableUploadHandler) create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUploadSessionRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	session, err := h.svc.Start(currentUser(r), req.Path, req.Filename, req.Fingerprint, req.Size)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, toSessionDTO(session, h.svc.ChunkSize()))
}

func (h *ResumableUploadHandler) status(w http.ResponseWriter, r *http.Request) {
	id, ok := sessionIDFromPath(r.URL.Path, "/api/uploads/")
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid upload id")
		return
	}
	session, err := h.svc.Status(currentUser(r), id)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, toSessionDTO(session, h.svc.ChunkSize()))
}

func (h *ResumableUploadHandler) append(w http.ResponseWriter, r *http.Request) {
	id, ok := sessionIDFromPath(r.URL.Path, "/api/uploads/")
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid upload id")
		return
	}
	offsetHeader := r.Header.Get("Upload-Offset")
	if offsetHeader == "" {
		respondError(w, http.StatusBadRequest, "Upload-Offset header is required")
		return
	}
	offset, err := strconv.ParseInt(offsetHeader, 10, 64)
	if err != nil || offset < 0 {
		respondError(w, http.StatusBadRequest, "invalid Upload-Offset")
		return
	}

	session, err := h.svc.Append(currentUser(r), id, offset, r.Body)
	if err != nil {
		var mismatch *services.ErrOffsetMismatch
		if errors.As(err, &mismatch) {
			respondJSON(w, http.StatusConflict, dto.OffsetMismatchError{
				Error:    "upload offset mismatch",
				Expected: mismatch.Expected,
			})
			return
		}
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, toSessionDTO(session, h.svc.ChunkSize()))
}

func (h *ResumableUploadHandler) complete(w http.ResponseWriter, r *http.Request) {
	const prefix = "/api/uploads/"
	const suffix = "/complete"
	trimmed := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), suffix)
	if trimmed == "" || strings.Contains(trimmed, "/") {
		respondError(w, http.StatusBadRequest, "invalid upload id")
		return
	}
	result, err := h.svc.Complete(currentUser(r), trimmed)
	if err != nil {
		var mismatch *services.ErrOffsetMismatch
		if errors.As(err, &mismatch) {
			respondJSON(w, http.StatusConflict, dto.OffsetMismatchError{
				Error:    "upload incomplete",
				Expected: mismatch.Expected,
			})
			return
		}
		respondWithError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, dto.UploadResult{Path: result.Filename, Size: result.Size})
}

func (h *ResumableUploadHandler) abort(w http.ResponseWriter, r *http.Request) {
	id, ok := sessionIDFromPath(r.URL.Path, "/api/uploads/")
	if !ok {
		respondError(w, http.StatusBadRequest, "invalid upload id")
		return
	}
	if err := h.svc.Abort(currentUser(r), id); err != nil {
		respondWithError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func sessionIDFromPath(path, prefix string) (string, bool) {
	trimmed := strings.TrimPrefix(path, prefix)
	if trimmed == "" || strings.Contains(trimmed, "/") {
		return "", false
	}
	return trimmed, true
}

func toSessionDTO(session *models.UploadSession, chunkSize int64) dto.UploadSession {
	return dto.UploadSession{
		ID:        session.ID,
		Offset:    session.Offset,
		Size:      session.Size,
		ChunkSize: chunkSize,
		ExpiresAt: session.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toPendingDTO(session *models.UploadSession) dto.PendingUploadSession {
	return dto.PendingUploadSession{
		ID:          session.ID,
		Fingerprint: session.Fingerprint,
		Destination: session.Destination,
		Filename:    session.Filename,
		Size:        session.Size,
		Offset:      session.Offset,
		UpdatedAt:   session.UpdatedAt.UTC().Format(time.RFC3339),
		ExpiresAt:   session.ExpiresAt.UTC().Format(time.RFC3339),
	}
}
