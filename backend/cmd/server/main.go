package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/application/events"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	grpcapi "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/grpc"
	xhttp "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/handlers"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/analytics"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	config "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/config"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/logging"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/s3"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/tls"
)

const productionEnv = "production"

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
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

	// Ensure the global shared folder exists so listing /shared never 404s
	// on a fresh install (list_file_service requires the physical directory).
	sharedDir := filepath.Join(rootDir, policy.SharedRoot)
	if err := os.MkdirAll(sharedDir, 0o700); err != nil {
		logger.Fatal("Failed to create shared folder", "error", err)
		return
	}

	indexRepo := memory.NewFileIndexRepository()
	textIndex := memory.NewTextIndex()

	// Content storage: local filesystem (default) or S3-compatible object store.
	// Metadata (users/roles/shares/revocations) always stays under rootDir.
	storageBackend := cfg.GetStorageBackend()
	switch storageBackend {
	case "local":
	case "s3":
	default:
		logger.Fatal("Unsupported STORAGE_BACKEND (want local or s3)", "value", storageBackend)
		return
	}

	var contentRepo ports.FileRepository
	var versionRepo ports.VersionRepository
	var mover ports.StorageMover
	var walkRoot func(string) ([]*models.FileInfo, error)
	var walkText func(string) ([]ports.TextDocument, error)
	var setVersionKeep func(int)

	localRepo := fs.NewLocalFileRepository(rootDir)
	if storageBackend == "s3" {
		s3Settings := cfg.GetS3Settings()
		if s3Settings.Bucket == "" {
			logger.Fatal("STORAGE_BACKEND=s3 requires S3_BUCKET")
			return
		}
		if s3Settings.AccessKey == "" || s3Settings.SecretKey == "" {
			logger.Fatal("STORAGE_BACKEND=s3 requires S3_ACCESS_KEY and S3_SECRET_KEY")
			return
		}
		s3Repo := s3.NewFileRepository(s3Settings)
		contentRepo = s3Repo
		versionRepo = s3Repo
		mover = s3Repo
		walkRoot = s3Repo.Walk
		walkText = s3Repo.WalkTextDocuments
		setVersionKeep = s3Repo.SetVersionKeep
		logger.Info("Storage backend: s3", "bucket", s3Settings.Bucket, "region", s3Settings.Region, "endpoint", s3Settings.Endpoint)
	} else {
		contentRepo = localRepo
		versionRepo = localRepo
		mover = localRepo
		walkRoot = fs.WalkRoot
		walkText = fs.WalkRootTextDocuments
		setVersionKeep = localRepo.SetVersionKeep
		logger.Info("Storage backend: local", "path", rootDir)
	}

	fileRepo := fs.NewIndexedFileRepository(contentRepo, indexRepo, textIndex)

	rebuildService := services.NewRebuildIndexService(indexRepo, textIndex)
	if err := rebuildService.Execute(rootDir, walkRoot, walkText); err != nil {
		logger.Warn("File index rebuild failed", "error", err)
	} else {
		logger.Info("File search index ready")
	}

	scoper := policy.NewPathScoper()
	userRepo := fs.NewUserFileRepository(rootDir)
	shareRepo := fs.NewShareFileRepository(rootDir)
	roleRepo := fs.NewRoleFileRepository(rootDir)
	roleCatalog := services.NewRoleCatalog(roleRepo)
	hasher := auth.NewPBKDF2Hasher()

	seedService := services.NewSeedAdminService(userRepo, hasher, fileRepo, scoper)
	if !cfg.EnableAuth() {
		logger.Info("Auth disabled — running as system admin view")
	} else if err := refuseWeakAdminSeed(cfg, userRepo); err != nil {
		logger.Fatal(err.Error())
		return
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

	analyticsStore := analytics.NewStore(rootDir)
	analyticsService := services.NewAnalyticsService(analyticsStore, roleCatalog)

	listService := services.NewListFilesService(fileRepo, scoper)
	fileDownloadService := services.NewDownloadFileService(fileRepo, scoper)
	zipService := services.NewDownloadZipService(fileRepo, scoper)
	downloadService := services.NewDownloadService(fileDownloadService, zipService)
	eventBus := events.NewBus()
	p2pService := services.NewP2PService()
	uploadService := services.NewUploadService(fileRepo, scoper, indexRepo, userRepo, cfg.GetMaxUploadBytes())
	uploadSessions := fs.NewUploadSessionRepository(rootDir)
	resumableUploadService := services.NewResumableUploadService(uploadSessions, fileRepo, scoper, indexRepo, userRepo, cfg.GetMaxUploadBytes())
	updateService := services.NewUpdateFileContentService(fileRepo, scoper)
	createDirService := services.NewCreateDirectoryService(fileRepo, scoper)
	deleteService := services.NewDeletePathService(fileRepo, scoper)
	infoService := services.NewGetFileInfoService(fileRepo, scoper)
	searchService := services.NewSearchFilesService(indexRepo, scoper, textIndex)
	registerService := services.NewRegisterUserService(userRepo, hasher, fileRepo, scoper, cfg.EnableSignup(), cfg.GetDefaultQuotaBytes())
	usersService := services.NewListUsersService(userRepo, indexRepo, scoper, roleCatalog)
	userInfoService := services.NewUserInfoService(userRepo, indexRepo, scoper)
	quotaService := services.NewUpdateUserQuotaService(userRepo, indexRepo, scoper, roleCatalog)
	createUserService := services.NewCreateUserService(userRepo, hasher, fileRepo, scoper, roleCatalog, cfg.GetDefaultQuotaBytes())
	updateUserService := services.NewUpdateUserService(userRepo, roleCatalog, scoper, mover, shareRepo, indexRepo)
	deleteUserService := services.NewDeleteUserService(userRepo, roleCatalog, scoper, fileRepo, shareRepo)
	resetPasswordService := services.NewResetPasswordService(userRepo, hasher, roleCatalog)
	changePasswordService := services.NewChangePasswordService(userRepo, hasher, jwtManager)
	listRolesService := services.NewListRolesService(roleCatalog)
	upsertRoleService := services.NewCreateOrUpdateRoleService(userRepo, roleRepo, roleCatalog)
	deleteRoleService := services.NewDeleteRoleService(userRepo, roleRepo, roleCatalog)

	createUserService.SetTokenManager(jwtManager)
	updateUserService.SetTokenManager(jwtManager)
	deleteUserService.SetTokenManager(jwtManager)
	resetPasswordService.SetTokenManager(jwtManager)

	uploadService.SetEventBus(eventBus)
	resumableUploadService.SetEventBus(eventBus)
	updateService.SetEventBus(eventBus)
	createDirService.SetEventBus(eventBus)
	deleteService.SetEventBus(eventBus)
	quotaService.SetEventBus(eventBus)
	downloadService.SetEventBus(eventBus)
	tokenService.SetEventBus(eventBus)

	// Persist every bus event to the analytics log (PII-free rollups for admins).
	go func() {
		ch, stop := eventBus.Subscribe(context.Background())
		defer stop()
		for e := range ch {
			_ = analyticsStore.Record(ports.AnalyticsEvent{
				Type:  e.Type,
				Path:  e.Path,
				User:  e.User,
				Bytes: e.Bytes,
				At:    e.At,
			})
		}
	}()

	listVersionsService := services.NewListVersionsService(fileRepo, versionRepo, scoper)
	downloadVersionService := services.NewDownloadVersionService(fileRepo, versionRepo, scoper)
	restoreVersionService := services.NewRestoreVersionService(fileRepo, versionRepo, scoper)
	restoreVersionService.SetEventBus(eventBus)
	if raw := os.Getenv("VERSION_KEEP"); raw != "" {
		if n, convErr := strconv.Atoi(raw); convErr == nil && n >= 0 {
			setVersionKeep(n)
			logger.Info("Version retention set", "keep", n)
		}
	}

	// Public share links: management handlers require auth, resolution does not.
	createShareService := services.NewCreateShareService(fileRepo, shareRepo, scoper)
	listSharesService := services.NewListSharesService(shareRepo, scoper, roleCatalog)
	revokeShareService := services.NewRevokeShareService(shareRepo, scoper, roleCatalog)
	resolveShareService := services.NewResolveShareService(shareRepo, scoper, downloadService)
	purgeSharesService := services.NewPurgeExpiredSharesService(shareRepo)
	createShareService.SetEventBus(eventBus)
	revokeShareService.SetEventBus(eventBus)
	if purged, err := purgeSharesService.Execute(); err != nil {
		logger.Warn("Share cleanup failed", "error", err)
	} else if purged > 0 {
		logger.Info("Purged expired share links", "count", purged)
	}

	// Drop upload sessions that outlived their TTL (crash/restart leftovers).
	if purged, err := resumableUploadService.PurgeExpired(); err != nil {
		logger.Warn("Upload session cleanup failed", "error", err)
	} else if purged > 0 {
		logger.Info("Purged expired upload sessions", "count", purged)
	}
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if n, err := resumableUploadService.PurgeExpired(); err == nil && n > 0 {
				logger.Info("Purged expired upload sessions", "count", n)
			}
		}
	}()

	listHandler := handlers.NewListHandler(listService)
	deleteHandler := handlers.NewDeleteHandler(deleteService)
	filesHandler := handlers.NewFilesHandler(listHandler, deleteHandler)

	routeHandlers := xhttp.RouteHandlers{
		Files:           filesHandler,
		Download:        handlers.NewDownloadHandler(downloadService),
		View:            handlers.NewViewHandler(fileDownloadService),
		Upload:          handlers.NewUploadHandler(uploadService),
		ResumableUpload: handlers.NewResumableUploadHandler(resumableUploadService),
		Update:          handlers.NewUpdateFileHandler(updateService),
		Directory:       handlers.NewDirectoryHandler(createDirService),
		FileInfo:        handlers.NewFileInfoHandler(infoService),
		Search:          handlers.NewSearchHandler(searchService),
		Events:          handlers.NewEventsHandler(eventBus),
		P2P:             handlers.NewP2PHandler(p2pService),
		Register:        handlers.NewRegisterHandler(registerService),
		Me:              handlers.NewMeHandler(userInfoService),
		AuthInfo:        handlers.NewAuthInfoHandler(cfg.EnableSignup(), oauthService.ProviderNames()),
		AdminUsers:      handlers.NewAdminUsersHandler(usersService),
		AdminUser:       handlers.NewAdminUserItemHandler(usersService, createUserService, updateUserService, deleteUserService, resetPasswordService),
		AdminQuota:      handlers.NewAdminQuotaHandler(quotaService),
		AdminPass:       handlers.NewAdminUserPasswordHandler(resetPasswordService),
		AdminRoles:      handlers.NewAdminRolesHandler(listRolesService, upsertRoleService),
		AdminRole:       handlers.NewAdminRoleItemHandler(deleteRoleService),
		AdminAnalytics:  handlers.NewAdminAnalyticsHandler(analyticsService),
		SelfPass:        handlers.NewSelfPasswordHandler(changePasswordService),
		Shares:          handlers.NewSharesHandler(createShareService, listSharesService, revokeShareService),
		Share:           handlers.NewShareDownloadHandler(resolveShareService),
		Versions:        handlers.NewVersionsHandler(listVersionsService),
		Version:         handlers.NewVersionDownloadHandler(downloadVersionService),
		Restore:         handlers.NewVersionRestoreHandler(restoreVersionService),
		Health:          handlers.NewHealthHandler(),
		Token:           handlers.NewTokenHandler(tokenService),
		Refresh:         handlers.NewRefreshHandler(tokenService),
		Revoke:          handlers.NewRevokeHandler(tokenService),
		OAuthStart:      handlers.NewOAuthStartHandler(oauthService),
		OAuthCb:         handlers.NewOAuthCallbackHandler(oauthService),
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

// refuseWeakAdminSeed blocks a production first boot that would seed the
// bootstrap admin with a well-known default password. Existing installs
// (any account already present) are never affected.
func refuseWeakAdminSeed(cfg ports.ConfigProvider, users ports.UserRepository) error {
	if os.Getenv("APP_ENV") != productionEnv {
		return nil
	}
	count, err := users.CountUsers()
	if err != nil || count > 0 {
		return nil
	}
	password := strings.ToLower(strings.TrimSpace(cfg.GetPassword()))
	if knownWeakAdminPasswords[password] {
		return fmt.Errorf("refusing to start: ADMIN_PASSWORD is a known default value; choose a strong password for the bootstrap admin")
	}
	return nil
}

var knownWeakAdminPasswords = map[string]bool{
	"admin": true, "changeme": true, "password": true, "password123": true,
	"changeme123": true, "admin123": true, "123456": true, "12345678": true,
	"root": true, "test": true, "qwerty": true, "letmein": true,
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
