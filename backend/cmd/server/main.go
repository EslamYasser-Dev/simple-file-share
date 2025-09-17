package main

import (
	"log"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	xhttp "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/handlers"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/config"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/logging"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/tls"
)

func main() {
	// === CONFIG ===
	cfg, err := config.NewEnvConfigProvider()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
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

	// === START ===
	if err := server.Start(); err != nil {
		logger.Fatal("Server failed", "error", err)
	}
}
