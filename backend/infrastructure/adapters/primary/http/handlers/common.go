package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

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
	var sharePassword *domainerrors.SharePasswordError
	var quotaExceeded *domainerrors.QuotaExceededError

	status := http.StatusInternalServerError
	message := "internal server error"

	switch {
	case errors.Is(err, domainerrors.ErrTwoFactorRequired):
		// Distinct from invalid credentials: clients use this to prompt for
		// a code instead of retrying the password.
		status, message = http.StatusUnauthorized, "totp_required"
	case errors.Is(err, domainerrors.ErrUserAlreadyExists):
		status, message = http.StatusConflict, err.Error()
	case errors.Is(err, domainerrors.ErrInvalidCredentials):
		// Never distinguish "bad password" from "no such user" — one generic
		// 401 prevents account enumeration through login responses.
		status, message = http.StatusUnauthorized, "authentication failed"
	case errors.Is(err, domainerrors.ErrUserNotFound):
		// User-lookup misses on public auth surfaces must not leak existence.
		status, message = http.StatusUnauthorized, "authentication failed"
	case errors.Is(err, domainerrors.ErrNotFound):
		status, message = http.StatusNotFound, err.Error()
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
	case errors.Is(err, domainerrors.ErrShareLimitReached):
		// Machine-readable like totp_required: clients retry nothing, they
		// surface "link exhausted".
		status, message = http.StatusGone, "share_limit_reached"
	case errors.As(err, &sharePassword):
		status = http.StatusUnauthorized
		if sharePassword.Missing {
			message = "share_password_required"
		} else {
			message = "share_password_invalid"
		}
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

// actorName is the audit-trail identity for this request: the authenticated
// username, or "-" when auth is disabled.
func actorName(r *http.Request) string {
	if u := currentUser(r); u != nil {
		return u.Username
	}
	return "-"
}

// clientIP strips the port from the peer address for audit records.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// serveDownload streams the resolved download as an attachment with
// conditional-request revalidation and byte-range support when the underlying
// stream is seekable (local files); non-seekable streams (S3 bodies, zip
// archives) degrade to a plain full-body response.
func serveDownload(w http.ResponseWriter, r *http.Request, dl *models.Download) {
	serveBody(w, r, dl, false)
}

// serveInline streams content for in-browser rendering (e.g. PDF previews).
// Callers must only pass content types that are safe to render on our origin.
func serveInline(w http.ResponseWriter, r *http.Request, dl *models.Download) {
	serveBody(w, r, dl, true)
}

func serveBody(w http.ResponseWriter, r *http.Request, dl *models.Download, inline bool) {
	defer dl.Stream.Close()

	// Revalidation first: a 304 carries no Content-Disposition.
	if etag := downloadETag(dl); etag != "" {
		w.Header().Set("ETag", etag)
		if !dl.ModTime.IsZero() {
			w.Header().Set("Last-Modified", dl.ModTime.UTC().Format(http.TimeFormat))
		}
		if etagMatches(r.Header.Get("If-None-Match"), etag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	disposition := fmt.Sprintf("attachment; filename=%q", dl.Filename)
	if inline {
		disposition = fmt.Sprintf("inline; filename=%q", dl.Filename)
	}
	w.Header().Set("Content-Type", dl.ContentType)
	w.Header().Set("Content-Disposition", disposition)
	if inline {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "private, max-age=60")
	}

	// Byte ranges need a seekable stream; ServeContent handles If-Range,
	// suffix/multipart ranges, and 416 itself.
	if r.Header.Get("Range") != "" {
		if seeker, ok := dl.Stream.(io.ReadSeeker); ok {
			http.ServeContent(w, r, dl.Filename, dl.ModTime, seeker)
			return
		}
	}
	if _, err := io.Copy(w, dl.Stream); err != nil {
		// Client likely disconnected; headers already sent, cannot write error status.
		return
	}
}

// downloadETag derives a strong validator from identity metadata (name, size,
// mtime). It never hashes file contents, so it stays O(1); files with identical
// name+size+mtime are the same representation, which is exactly what a
// filesystem validator can promise. Unknown metadata (zip archives, version
// blobs) yields "" — no ETag, no false revalidation.
func downloadETag(dl *models.Download) string {
	if dl.ModTime.IsZero() {
		return ""
	}
	sum := sha256.New()
	fmt.Fprintf(sum, "%s\x00%d\x00%d", dl.Filename, dl.Size, dl.ModTime.UnixNano())
	return `"` + hex.EncodeToString(sum.Sum(nil)) + `"`
}

// etagMatches implements If-None-Match comparison (weak tags and lists).
func etagMatches(header, etag string) bool {
	if header == "" {
		return false
	}
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" {
			return true
		}
		if strings.TrimPrefix(candidate, "W/") == etag {
			return true
		}
	}
	return false
}
