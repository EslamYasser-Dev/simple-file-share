package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	// The storage root is owned exclusively by this server: validate it, create
	// it when missing, and lock it down to owner-only permissions.
	rootDir, err := fs.PrepareStorageRoot(cfg.GetRootDir())
	if err != nil {
		logger.Fatal("Invalid storage root", "error", err)
		return
	}
	logger.Info("Storage root ready", "path", rootDir)

	indexRepo := memory.NewFileIndexRepository()
	textIndex := memory.NewTextIndex()
	localRepo := fs.NewLocalFileRepository(rootDir)
	fileRepo := fs.NewIndexedFileRepository(localRepo, indexRepo, textIndex)

	rebuildService := services.NewRebuildIndexService(indexRepo, textIndex)
	if err := rebuildService.Execute(rootDir, fs.WalkRoot, fs.WalkRootTextDocuments); err != nil {
		logger.Warn("File index rebuild failed", "error", err)
	} else {
		logger.Info("File search index ready")
	}

	scoper := policy.NewPathScoper()
	userRepo := fs.NewUserFileRepository(rootDir)
	shareRepo := fs.NewShareFileRepository(rootDir)
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
	authenticateService := services.NewAuthenticateService(authProvider)
	tlsGenerator := &tls.InMemoryTLSCertGenerator{}

	jwtSecret := cfg.GetJWTSecret()
	enforceJWTSecretPolicy(logger, cfg, jwtSecret)
	revocationPath := filepath.Join(rootDir, ".file-share", "token_revocations.json")
	jwtManager := auth.NewJWTManagerWithStore(jwtSecret, time.Duration(cfg.GetJWTTTLSeconds())*time.Second, revocationPath)
	tokenService := services.NewTokenService(authenticateService, jwtManager, userRepo)
	oauthProviders := auth.NewOAuthProviders()
	if err := enforceOAuthConfig(cfg, oauthProviders); err != nil {
		logger.Fatal("Invalid OAuth configuration", "error", err)
		return
	}
	oauthService := services.NewOAuthLoginService(oauthProviders, userRepo, fileRepo, scoper, tokenService, cfg.EnableSignup(), cfg.GetDefaultQuotaBytes())
	if names := oauthService.ProviderNames(); len(names) > 0 {
		logger.Info("OAuth providers configured", "providers", names)
	}

	listService := services.NewListFilesService(fileRepo, scoper)
	fileDownloadService := services.NewDownloadFileService(fileRepo, scoper)
	zipService := services.NewDownloadZipService(fileRepo, scoper)
	downloadService := services.NewDownloadService(fileDownloadService, zipService)
	uploadService := services.NewUploadService(fileRepo, scoper, indexRepo, userRepo, cfg.GetMaxUploadBytes())
	updateService := services.NewUpdateFileContentService(fileRepo, scoper)
	createDirService := services.NewCreateDirectoryService(fileRepo, scoper)
	deleteService := services.NewDeletePathService(fileRepo, scoper)
	infoService := services.NewGetFileInfoService(fileRepo, scoper)
	searchService := services.NewSearchFilesService(indexRepo, scoper, textIndex)
	registerService := services.NewRegisterUserService(userRepo, hasher, fileRepo, scoper, cfg.EnableSignup(), cfg.GetDefaultQuotaBytes())
	usersService := services.NewListUsersService(userRepo, indexRepo, scoper)
	userInfoService := services.NewUserInfoService(userRepo, indexRepo, scoper)
	quotaService := services.NewUpdateUserQuotaService(userRepo, indexRepo, scoper)

	listVersionsService := services.NewListVersionsService(fileRepo, localRepo, scoper)
	downloadVersionService := services.NewDownloadVersionService(fileRepo, localRepo, scoper)
	restoreVersionService := services.NewRestoreVersionService(fileRepo, localRepo, scoper)
	if raw := os.Getenv("VERSION_KEEP"); raw != "" {
		if n, convErr := strconv.Atoi(raw); convErr == nil && n >= 0 {
			localRepo.SetVersionKeep(n)
			logger.Info("Version retention set", "keep", n)
		}
	}

	// Public share links: management handlers require auth, resolution does not.
	createShareService := services.NewCreateShareService(fileRepo, shareRepo, scoper)
	listSharesService := services.NewListSharesService(shareRepo, scoper)
	revokeShareService := services.NewRevokeShareService(shareRepo, scoper)
	resolveShareService := services.NewResolveShareService(shareRepo, scoper, downloadService)
	purgeSharesService := services.NewPurgeExpiredSharesService(shareRepo)
	if purged, err := purgeSharesService.Execute(); err != nil {
		logger.Warn("Share cleanup failed", "error", err)
	} else if purged > 0 {
		logger.Info("Purged expired share links", "count", purged)
	}

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
		Me:         handlers.NewMeHandler(userInfoService),
		AuthInfo:   handlers.NewAuthInfoHandler(cfg.EnableSignup(), oauthService.ProviderNames()),
		AdminUsers: handlers.NewAdminUsersHandler(usersService),
		AdminQuota: handlers.NewAdminQuotaHandler(quotaService),
		Shares:     handlers.NewSharesHandler(createShareService, listSharesService, revokeShareService),
		Share:      handlers.NewShareDownloadHandler(resolveShareService),
		Versions:   handlers.NewVersionsHandler(listVersionsService),
		Version:    handlers.NewVersionDownloadHandler(downloadVersionService),
		Restore:    handlers.NewVersionRestoreHandler(restoreVersionService),
		Health:     handlers.NewHealthHandler(),
		Token:      handlers.NewTokenHandler(tokenService),
		Refresh:    handlers.NewRefreshHandler(tokenService),
		Revoke:     handlers.NewRevokeHandler(tokenService),
		OAuthStart: handlers.NewOAuthStartHandler(oauthService),
		OAuthCb:    handlers.NewOAuthCallbackHandler(oauthService),
	}

	server := xhttp.NewServer(
		cfg.GetPort(),
		tlsGenerator,
		logger,
		routeHandlers,
		authenticateService,
		tokenService,
		cfg.EnableAuth(),
	)
	server.ConfigureTLS(cfg.EnableTLS())

	if cfg.EnableGRPC() {
		authService := grpcapi.NewAuthService(registerService, usersService, authenticateService, cfg.EnableSignup())
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
			authenticateService,
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
	// frontend from the Go binary lets one container host the whole app (the API
	// and the UI) behind a single origin.
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

// enforceJWTSecretPolicy refuses forgeable secrets when auth is on in
// production (default placeholder or too-short keys).
func enforceJWTSecretPolicy(logger ports.Logger, cfg ports.ConfigProvider, secret string) {
	if !cfg.EnableAuth() || os.Getenv("APP_ENV") != productionEnv {
		return
	}
	insecure := secret == "" ||
		secret == "change-me-in-production" ||
		len(secret) < 32
	if insecure {
		logger.Fatal("Refusing to start: JWT_SECRET must be set to a random value of at least 32 characters in production")
	}
}

// enforceOAuthConfig requires an explicit public origin when OAuth is enabled
// in production so Host/X-Forwarded-* cannot steal the access-token redirect.
func enforceOAuthConfig(cfg ports.ConfigProvider, providers map[string]*auth.OAuthProvider) error {
	if len(providers) == 0 || os.Getenv("APP_ENV") != productionEnv {
		return nil
	}
	base := strings.TrimSpace(os.Getenv("OAUTH_REDIRECT_BASE"))
	if err := auth.ValidateRedirectBase(base); err != nil {
		return fmt.Errorf("OAUTH_REDIRECT_BASE (required with OAuth in production): %w", err)
	}
	if spa := strings.TrimSpace(os.Getenv("OAUTH_SPA_BASE")); spa != "" {
		if err := auth.ValidateRedirectBase(spa); err != nil {
			return fmt.Errorf("OAUTH_SPA_BASE: %w", err)
		}
	}
	_ = cfg
	return nil
}
