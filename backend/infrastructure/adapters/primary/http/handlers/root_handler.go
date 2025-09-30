package handlers

import (
    "encoding/json"
    "net/http"
    "path/filepath"
    "strings"
    "time"

    "github.com/EslamYasser-Dev/simple-file-share/application/services"
)

// RootHandler decides between directory listing and file/zip download for GET "/".
type RootHandler struct {
	listService     *services.ListFilesService
	fileService     *services.DownloadFileService
	zipService      *services.DownloadZipService
	port            string
}

func NewRootHandler(list *services.ListFilesService, file *services.DownloadFileService, zip *services.DownloadZipService, port string) *RootHandler {
	return &RootHandler{listService: list, fileService: file, zipService: zip, port: port}
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // API endpoints
    if r.Method == http.MethodGet && r.URL.Path == "/api/files" {
        // JSON directory listing using ?path=
        reqPath := r.URL.Query().Get("path")
        if reqPath == "" {
            reqPath = "/"
        }
        pageData, err := h.listService.Execute(reqPath)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }
        // Map to frontend shape
        type item struct {
            Name     string `json:"name"`
            Path     string `json:"path"`
            Size     int64  `json:"size"`
            IsDir    bool   `json:"isDir"`
            Modified string `json:"modified"`
        }
        var items []item
        if pageData != nil {
            for _, f := range pageData.Files {
                p := strings.TrimPrefix(f.URL, "/")
                items = append(items, item{
                    Name:     f.Name,
                    Path:     p,
                    Size:     0, // repository doesn’t expose raw bytes; fill later if needed
                    IsDir:    f.IsDir,
                    Modified: time.Now().Format(time.RFC3339),
                })
            }
        }
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(items)
        return
    }

    if r.Method == http.MethodGet && r.URL.Path == "/api/files/download" {
        reqPath := r.URL.Query().Get("path")
        if reqPath == "" {
            http.Error(w, "Missing path", http.StatusBadRequest)
            return
        }
        // Normalize
        safePath := filepath.ToSlash(strings.TrimPrefix(reqPath, "/"))
        stream, filename, err := h.fileService.Execute(safePath)
        if err != nil {
            respondWithError(w, err)
            return
        }
        if stream == nil {
            http.Error(w, "Is a directory", http.StatusConflict)
            return
        }
        serveDownload(w, stream, filename, "application/octet-stream")
        return
    }

    path := cleanPath(r.URL.Path)
	if containsPathTraversal(path) {
		http.Error(w, "Path traversal detected", http.StatusForbidden)
		return
	}

	// ZIP request
	if strings.HasSuffix(path, ".zip") {
		path = strings.TrimSuffix(path, ".zip")
		stream, filename, err := h.zipService.Execute(path)
		if err != nil {
			respondWithError(w, err)
			return
		}
		if stream == nil {
			http.Error(w, "Not a directory", http.StatusBadRequest)
			return
		}
		serveDownload(w, stream, filename, "application/zip")
		return
	}

	// Try as directory first
	pageData, err := h.listService.Execute(path)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if pageData != nil {
		pageData.Port = h.port
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Directory listing available via API. Integrate template to render UI."))
		return
	}

	// Fallback: serve as file
	stream, filename, err := h.fileService.Execute(path)
	if err != nil {
		respondWithError(w, err)
		return
	}
	if stream == nil {
		http.Error(w, "Is a directory", http.StatusConflict)
		return
	}
	serveDownload(w, stream, filename, "application/octet-stream")
}
