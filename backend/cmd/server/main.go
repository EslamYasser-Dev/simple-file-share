package main

import (
	"log"
	"os"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	xhttp "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/handlers"
	config "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/config"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/logging"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/tls"
)

func main() {
	// === CONFIG ===
	var cfg ports.ConfigProvider
	var err error

	if os.Getenv("APP_ENV") == "production" {
		cfg, err = config.NewEnvConfigProvider()
		if err != nil {
			log.Fatal("Failed to load config: ", err)
		}
	} else {
		// Use development configuration
		cfg, err = config.NewDevConfigProvider()
		if err != nil {
			log.Fatal("Failed to load development config: ", err)
		}
		log.Println("⚠️  Running in DEVELOPMENT mode with TLS disabled")
	}

	// === LOGGING ===
	logger := logging.NewStdLogger()

	// === SECONDARY ADAPTERS ===
	fileRepo := fs.NewLocalFileRepository(cfg.GetRootDir())
	// authProvider := auth.NewStaticAuthProvider(cfg.GetUsername(), cfg.GetPassword())
	tlsGenerator := &tls.InMemoryTLSCertGenerator{}

	// === APPLICATION SERVICES ===
	listService := services.NewListFilesService(fileRepo)
	downloadService := services.NewDownloadFileService(fileRepo)
	zipService := services.NewDownloadZipService(fileRepo)
	uploadService := services.NewUploadService(fileRepo)

	// === PRIMARY ADAPTERS (HTTP HANDLERS) ===
	rootHandler := handlers.NewRootHandler(listService, downloadService, zipService, cfg.GetPort())
	uploadHandler := handlers.NewUploadHandler(uploadService)

	// === HTTP SERVER ===
	server := xhttp.NewServer(
		cfg.GetPort(),
		tlsGenerator,
		logger,
		rootHandler,
		uploadHandler,
	)

	// In development, we'll serve the frontend files directly
	if os.Getenv("APP_ENV") != "production" {
		// For development, serve the built frontend files if present.
		// The binary is executed from the project root (via `make run`),
		// so the correct relative path is "frontend/dist".
		frontendDir := "frontend/dist"
		if _, err := os.Stat(frontendDir); !os.IsNotExist(err) {
			server.SetStaticFileServer(frontendDir)
		}
	}

	// === START ===
	if err := server.Start(); err != nil {
		logger.Fatal("Server failed", "error", err)
	}
}
