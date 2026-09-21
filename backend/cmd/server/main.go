package main

import (
	"os"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	grpcapi "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/grpc"
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

	scoper := policy.NewPathScoper()
	userRepo := fs.NewUserFileRepository(cfg.GetRootDir())
	hasher := auth.NewPBKDF2Hasher()

	seedService := services.NewSeedAdminService(userRepo, hasher, fileRepo, scoper)
	if !cfg.EnableAuth() {
		logger.Info("Auth disabled — running as system admin view")
	} else if seeded, err := seedService.Execute(cfg.GetUsername(), cfg.GetPassword()); err != nil {
		logger.Warn("Admin seed failed", "error", err)
	} else if seeded {
		logger.Info("Seeded admin account", "username", cfg.GetUsername())
	}

	authProvider := auth.NewUserAuthProvider(userRepo, hasher)
	tlsGenerator := &tls.InMemoryTLSCertGenerator{}

	listService := services.NewListFilesService(fileRepo, scoper)
	fileDownloadService := services.NewDownloadFileService(fileRepo, scoper)
	zipService := services.NewDownloadZipService(fileRepo, scoper)
	downloadService := services.NewDownloadService(fileDownloadService, zipService)
	uploadService := services.NewUploadService(fileRepo, scoper)
	updateService := services.NewUpdateFileContentService(fileRepo, scoper)
	createDirService := services.NewCreateDirectoryService(fileRepo, scoper)
	deleteService := services.NewDeletePathService(fileRepo, scoper)
	infoService := services.NewGetFileInfoService(fileRepo, scoper)
	searchService := services.NewSearchFilesService(indexRepo, scoper)
	registerService := services.NewRegisterUserService(userRepo, hasher, fileRepo, scoper, cfg.EnableSignup())
	usersService := services.NewListUsersService(userRepo, indexRepo, scoper)

	listHandler := handlers.NewListHandler(listService)
	deleteHandler := handlers.NewDeleteHandler(deleteService)
	filesHandler := handlers.NewFilesHandler(listHandler, deleteHandler)

	routeHandlers := xhttp.RouteHandlers{
		Files:      filesHandler,
		Download:   handlers.NewDownloadHandler(downloadService),
		View:       handlers.NewViewHandler(fileDownloadService),
		Upload:     handlers.NewUploadHandler(uploadService),
		Update:     handlers.NewUpdateFileHandler(updateService),
		Directory:  handlers.NewDirectoryHandler(createDirService),
		FileInfo:   handlers.NewFileInfoHandler(infoService),
		Search:     handlers.NewSearchHandler(searchService),
		Register:   handlers.NewRegisterHandler(registerService),
		Me:         handlers.NewMeHandler(),
		AuthInfo:   handlers.NewAuthInfoHandler(cfg.EnableSignup()),
		AdminUsers: handlers.NewAdminUsersHandler(usersService),
		Health:     handlers.NewHealthHandler(),
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

	if cfg.EnableGRPC() {
		authService := grpcapi.NewAuthService(registerService, usersService, authProvider, cfg.EnableSignup())
		fileService := grpcapi.NewFileService(
			listService,
			infoService,
			searchService,
			createDirService,
			deleteService,
			updateService,
			uploadService,
			downloadService,
		)
		grpcServer, err := grpcapi.NewServer(
			cfg.GetGRPCPort(),
			logger,
			tlsGenerator,
			cfg.EnableTLS(),
			authProvider,
			cfg.EnableAuth(),
			authService,
			fileService,
		)
		if err != nil {
			logger.Fatal("Failed to create gRPC server", "error", err)
			return
		}
		go func() {
			if err := grpcServer.Start(); err != nil {
				logger.Error("gRPC server failed", "error", err)
			}
		}()
		defer grpcServer.Stop()
	}

	// Serve the built React frontend whenever present (both dev and production).
	// Point STATIC_DIR at the directory containing index.html + assets. Serving the
	// frontend from the Go binary lets Render host the whole app as a single
	// web service (one URL) for the API and the UI.
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "frontend/dist"
	}
	server.SetStaticFileServer(staticDir)

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
	logger.Info("Running in DEVELOPMENT mode (auth/TLS disabled unless explicitly enabled)")
	return config.NewDevConfigProvider()
}
