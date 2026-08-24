package main

import (
	"os"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	xhttp "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/handlers"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	config "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/config"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/logging"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/tls"
)

const productionEnv = "production"

func main() {
	logger := logging.NewStdLogger()

	cfg, err := loadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to load config", "error", err)
		return
	}

	indexRepo := memory.NewFileIndexRepository()
	localRepo := fs.NewLocalFileRepository(cfg.GetRootDir())
	fileRepo := fs.NewIndexedFileRepository(localRepo, indexRepo)

	rebuildService := services.NewRebuildIndexService(indexRepo)
	if err := rebuildService.Execute(cfg.GetRootDir(), fs.WalkRoot); err != nil {
		logger.Warn("File index rebuild failed", "error", err)
	} else {
		logger.Info("File search index ready")
	}

	authProvider := auth.NewStaticAuthProvider(cfg.GetUsername(), cfg.GetPassword())
	tlsGenerator := &tls.InMemoryTLSCertGenerator{}

	listService := services.NewListFilesService(fileRepo)
	downloadService := services.NewDownloadFileService(fileRepo)
	zipService := services.NewDownloadZipService(fileRepo)
	uploadService := services.NewUploadService(fileRepo)
	createDirService := services.NewCreateDirectoryService(fileRepo)
	deleteService := services.NewDeletePathService(fileRepo)
	infoService := services.NewGetFileInfoService(fileRepo)
	searchService := services.NewSearchFilesService(indexRepo)

	listHandler := handlers.NewListHandler(listService)
	deleteHandler := handlers.NewDeleteHandler(deleteService)
	filesHandler := handlers.NewFilesHandler(listHandler, deleteHandler)

	routeHandlers := xhttp.RouteHandlers{
		Files:     filesHandler,
		Download:  handlers.NewDownloadHandler(downloadService, zipService),
		Upload:    handlers.NewUploadHandler(uploadService),
		Directory: handlers.NewDirectoryHandler(createDirService),
		FileInfo:  handlers.NewFileInfoHandler(infoService),
		Search:    handlers.NewSearchHandler(searchService),
		Health:    handlers.NewHealthHandler(),
	}

	server := xhttp.NewServer(
		cfg.GetPort(),
		tlsGenerator,
		logger,
		routeHandlers,
		authProvider,
		cfg.EnableAuth(),
		cfg.GetMaxUploadBytes(),
	)
	server.ConfigureTLS(cfg.EnableTLS())

	if os.Getenv("APP_ENV") != productionEnv {
		server.SetStaticFileServer("frontend/dist")
	}

	if err := server.Start(); err != nil {
		logger.Fatal("Server failed", "error", err)
	}
}

// loadConfig selects the config provider based on APP_ENV.
func loadConfig(logger ports.Logger) (ports.ConfigProvider, error) {
	if os.Getenv("APP_ENV") == productionEnv {
		logger.Info("Running in PRODUCTION mode")
		return config.NewEnvConfigProvider()
	}
	logger.Info("Running in DEVELOPMENT mode (auth disabled, TLS disabled)")
	return config.NewDevConfigProvider()
}
