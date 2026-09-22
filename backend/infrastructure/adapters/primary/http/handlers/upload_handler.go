package handlers

import (
	"io"
	"net/http"
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

	// An optional path chooses the virtual destination directory. It can also
	// be carried as a "path" multipart form field, which wins for parity with
	// the web form the frontend submits.
	destPrefix := strings.TrimPrefix(r.URL.Query().Get("path"), "/")

	reader, err := r.MultipartReader()
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid multipart request")
		return
	}

	var uploads []models.FileUpload
	var execErrors []error

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			// A non-EOF error means the stream is broken; retrying would
			// loop forever, so fail the request instead.
			respondError(w, http.StatusBadRequest, "malformed multipart request")
			return
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

		// mime/multipart shares one buffered reader across parts, so the next
		// part can only be parsed once this one has been fully consumed. The
		// upload therefore happens inline: it drains the part and releases the
		// descriptor before NextPart() advances the stream. Reading a part
		// lazily and then calling NextPart() again silently discards its data.
		written, execErr := h.uploadService.Execute(currentUser(r), []models.UploadPart{{
			Name:        part.FileName(),
			Destination: destPrefix,
			Content:     part,
		}})
		if execErr != nil {
			execErrors = append(execErrors, execErr)
			continue
		}
		uploads = append(uploads, written...)
	}

	if len(uploads) == 0 && len(execErrors) > 0 {
		respondWithError(w, execErrors[0])
		return
	}
	if len(uploads) == 0 {
		respondError(w, http.StatusBadRequest, "no files uploaded")
		return
	}

	results := make([]dto.UploadResult, len(uploads))
	for i, u := range uploads {
		results[i] = dto.UploadResult{Path: u.Filename, Size: u.Size}
	}
	respondJSON(w, http.StatusOK, results)
}
