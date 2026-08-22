package handlers

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
)

type UploadHandler struct {
	uploadService *services.UploadService
}

func NewUploadHandler(uploadService *services.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	destPrefix := strings.TrimPrefix(r.URL.Query().Get("path"), "/")

	reader, err := r.MultipartReader()
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			respondJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "upload exceeds size limit"})
			return
		}
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart request"})
		return
	}

	var parts []models.UploadPart
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if part.FileName() == "" {
			if part.FormName() == "path" {
				if b, readErr := io.ReadAll(part); readErr == nil {
					destPrefix = strings.TrimPrefix(string(b), "/")
				}
			}
			part.Close()
			continue
		}

		filename := part.FileName()
		if destPrefix != "" {
			filename = filepath.ToSlash(filepath.Join(destPrefix, filename))
		}
		parts = append(parts, &uploadPartWithName{name: filename, rc: part})
	}

	uploads, err := h.uploadService.Execute(parts)
	if err != nil {
		respondWithError(w, err)
		return
	}

	if len(uploads) == 0 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "no files uploaded"})
		return
	}

	if len(uploads) == 1 {
		respondJSON(w, http.StatusOK, dto.UploadResult{Path: uploads[0].Filename, Size: uploads[0].Size})
		return
	}

	results := make([]dto.UploadResult, len(uploads))
	for i, u := range uploads {
		results[i] = dto.UploadResult{Path: u.Filename, Size: u.Size}
	}
	respondJSON(w, http.StatusOK, results)
}

type uploadPartWithName struct {
	name string
	rc   models.ReadCloser
}

func (u *uploadPartWithName) Filename() string           { return u.name }
func (u *uploadPartWithName) Content() models.ReadCloser { return u.rc }
