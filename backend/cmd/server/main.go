package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
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
	auditstore "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/audit"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	config "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/config"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/indexstore"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/logging"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/metrics"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/s3"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/sessions"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/tls"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/webhook"
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

	// Observability: one recorder shared by the HTTP listener, the gRPC
	// interceptors, and the SSE handler. Disabled entirely when
	// ENABLE_METRICS=false (request ids still get stamped).
	var recorder ports.MetricsRecorder
	var metricsExporter ports.MetricsExporter
	var registry *metrics.Registry
	if cfg.EnableMetrics() {
		registry = metrics.NewRegistry()
		rec := metrics.NewRecorder(registry)
		recorder = rec
		metricsExporter = rec
		registry.NewGauge(
			"fileshare_process_start_time_seconds",
			"Unix time when this process started.",
		).WithLabelValues().Set(float64(time.Now().Unix()))
		indexFilesGauge := registry.NewGauge(
			"fileshare_index_files",
			"Files currently in the metadata search index.",
		).WithLabelValues()
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if files, _, err := indexRepo.PrefixStats(""); err == nil {
					indexFilesGauge.Set(float64(files))
				}
			}
		}()
		logger.Info("Metrics enabled", "endpoint", "/metrics")
	}

	rebuildService := services.NewRebuildIndexService(indexRepo, textIndex)
	if cfg.EnableIndexSnapshot() {
		// Incremental boot: unchanged text documents are restored from
		// ROOT_DIR/.file-share/index.snapshot; only new/changed files are
		// re-read through the active content backend (local or S3).
		rebuildService.EnableSnapshotPersist(indexstore.NewStore(rootDir), fileRepo.ServeFile)
		logger.Info("Index snapshot persistence enabled")
	}
	var indexReady atomic.Bool
	build, err := rebuildService.Execute(rootDir, walkRoot, walkText)
	if err != nil {
		logger.Warn("File index rebuild failed", "error", err)
	} else {
		indexReady.Store(true)
		logger.Info("File search index ready",
			"source", build.Source,
			"files", build.Files,
			"text_docs", build.TextDocs,
			"duration_ms", build.Duration.Milliseconds(),
		)
		if build.SnapshotErr != nil {
			logger.Warn("Index snapshot save failed (next boot re-extracts)", "error", build.SnapshotErr)
		}
		if registry != nil {
			registry.NewGauge(
				"fileshare_index_build_seconds",
				"Duration of the last startup index build in seconds.",
			).WithLabelValues().Set(build.Duration.Seconds())
			registry.NewGauge(
				"fileshare_index_source",
				"1 for the source the last index build used (snapshot|full).",
				"source",
			).WithLabelValues(build.Source).Set(1)
			registry.NewGauge(
				"fileshare_index_text_docs",
				"Documents in the full-text index after the last build.",
			).WithLabelValues().Set(float64(build.TextDocs))
		}
	}

	scoper := policy.NewPathScoper()
	userRepo := fs.NewUserFileRepository(rootDir)
	shareRepo := fs.NewShareFileRepository(rootDir)
	roleRepo := fs.NewRoleFileRepository(rootDir)
	roleCatalog := services.NewRoleCatalog(roleRepo)
	hasher := auth.NewPBKDF2Hasher()

	// Security audit trail (logins, admin changes). The store is a plain
	// interface so a disabled audit is a true nil, never a typed-nil.
	var auditStore ports.AuditLog
	if cfg.EnableAuditLog() {
		jsonlStore := auditstore.NewJSONLLog(rootDir, cfg.GetAuditMaxBytes(), cfg.GetAuditKeep())
		auditStore = jsonlStore
		logger.Info("Security audit log enabled", "path", jsonlStore.Path())
	}
	auditService := services.NewAuditService(auditStore, roleCatalog)

	// Two-factor authentication: state lives on the user account, so this
	// service only needs the user repo, hasher (backup codes) and roles.
	totpService := services.NewTOTPService(userRepo, hasher, roleCatalog)

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
	tlsGenerator := tls.NewPersistedTLSCertGenerator(
		&tls.InMemoryTLSCertGenerator{},
		filepath.Join(rootDir, ".file-share", "tls"),
	)

	jwtSecret := cfg.GetJWTSecret()
	enforceJWTSecretPolicy(logger, cfg, jwtSecret)
	revocationPath := filepath.Join(rootDir, ".file-share", "token_revocations.json")
	jwtManager := auth.NewJWTManagerWithStore(jwtSecret, time.Duration(cfg.GetJWTTTLSeconds())*time.Second, revocationPath)
	tokenService := services.NewTokenService(authenticateService, jwtManager, userRepo)
	tokenService.SetTwoFactor(totpService)
	sessionService := services.NewSessionService(sessions.NewRegistry(rootDir), jwtManager)
	tokenService.SetSessions(sessionService)
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

	// Optional webhooks: when WEBHOOK_URLS is configured, every bus event is
	// also POSTed to those endpoints (HMAC-signed when WEBHOOK_SECRET is set).
	// Delivery is best-effort and never fails the request that published it.
	if webhookURLs := cfg.GetWebhookURLs(); len(webhookURLs) > 0 {
		dispatcher := webhook.NewDispatcher(webhookURLs, cfg.GetWebhookSecret(), logger)
		go dispatcher.Run(context.Background(), eventBus)
		logger.Info("Webhooks enabled",
			"endpoints", len(webhookURLs),
			"signed", cfg.GetWebhookSecret() != "",
		)
	}

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
	createShareService := services.NewCreateShareService(fileRepo, shareRepo, scoper, hasher)
	listSharesService := services.NewListSharesService(shareRepo, scoper, roleCatalog)
	revokeShareService := services.NewRevokeShareService(shareRepo, scoper, roleCatalog)
	resolveShareService := services.NewResolveShareService(shareRepo, scoper, downloadService, hasher)
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

	eventsHandler := handlers.NewEventsHandler(eventBus)
	eventsHandler.SetMetricsRecorder(recorder)

	// Handlers that record security events get the audit trail attached on
	// the concrete type before they are stored as http.Handler.
	registerHandler := handlers.NewRegisterHandler(registerService)
	registerHandler.SetAudit(auditService)
	adminUserHandler := handlers.NewAdminUserItemHandler(usersService, createUserService, updateUserService, deleteUserService, resetPasswordService)
	adminUserHandler.SetAudit(auditService)
	adminQuotaHandler := handlers.NewAdminQuotaHandler(quotaService)
	adminQuotaHandler.SetAudit(auditService)
	adminPassHandler := handlers.NewAdminUserPasswordHandler(resetPasswordService)
	adminPassHandler.SetAudit(auditService)
	adminRolesHandler := handlers.NewAdminRolesHandler(listRolesService, upsertRoleService)
	adminRolesHandler.SetAudit(auditService)
	adminRoleHandler := handlers.NewAdminRoleItemHandler(deleteRoleService)
	adminRoleHandler.SetAudit(auditService)
	selfPassHandler := handlers.NewSelfPasswordHandler(changePasswordService)
	selfPassHandler.SetAudit(auditService)
	sharesHandler := handlers.NewSharesHandler(createShareService, listSharesService, revokeShareService)
	sharesHandler.SetAudit(auditService)
	tokenHandler := handlers.NewTokenHandler(tokenService)
	tokenHandler.SetAudit(auditService)
	tokenHandler.SetSessions(sessionService)
	revokeHandler := handlers.NewRevokeHandler(tokenService)
	revokeHandler.SetAudit(auditService)
	refreshHandler := handlers.NewRefreshHandler(tokenService)
	refreshHandler.SetSessions(sessionService)
	authInfoHandler := handlers.NewAuthInfoHandler(cfg.EnableSignup(), oauthService.ProviderNames())
	authInfoHandler.SetTwoFactor(cfg.EnableTwoFactor())
	totpHandler := handlers.NewTOTPHandler(totpService)
	totpHandler.SetAudit(auditService)
	adminTotpResetHandler := handlers.NewAdminTotpResetHandler(totpService)
	adminTotpResetHandler.SetAudit(auditService)

	// API keys: nil service disables both the routes and "sfs_…" bearer
	// credentials (existing keys stop authenticating until re-enabled).
	var apiKeyService *services.APIKeyService
	var apiKeyHandler *handlers.APIKeysHandler
	if cfg.EnableAPIKeys() {
		apiKeyService = services.NewAPIKeyService(fs.NewAPIKeyFileRepository(rootDir), userRepo, hasher)
		apiKeyHandler = handlers.NewAPIKeysHandler(apiKeyService)
		apiKeyHandler.SetAudit(auditService)
	}

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
		Events:          eventsHandler,
		P2P:             handlers.NewP2PHandler(p2pService),
		Register:        registerHandler,
		Me:              handlers.NewMeHandler(userInfoService),
		AuthInfo:        authInfoHandler,
		AdminUsers:      handlers.NewAdminUsersHandler(usersService),
		AdminUser:       adminUserHandler,
		AdminQuota:      adminQuotaHandler,
		AdminPass:       adminPassHandler,
		AdminRoles:      adminRolesHandler,
		AdminRole:       adminRoleHandler,
		AdminAnalytics:  handlers.NewAdminAnalyticsHandler(analyticsService),
		SelfPass:        selfPassHandler,
		Shares:          sharesHandler,
		Share:           handlers.NewShareDownloadHandler(resolveShareService),
		Versions:        handlers.NewVersionsHandler(listVersionsService),
		Version:         handlers.NewVersionDownloadHandler(downloadVersionService),
		Restore:         handlers.NewVersionRestoreHandler(restoreVersionService),
		Health:          handlers.NewHealthHandler(),
		Token:           tokenHandler,
		Refresh:         refreshHandler,
		Revoke:          revokeHandler,
		Sessions:        handlers.NewSessionsHandler(sessionService),
		OAuthStart:      handlers.NewOAuthStartHandler(oauthService),
		OAuthCb:         handlers.NewOAuthCallbackHandler(oauthService),
		GRPCCert:        handlers.NewGRPCCertHandler(tlsGenerator),
		Ready: handlers.NewReadyHandler(
			handlers.ReadinessCheck{Name: "storage", Check: func() error {
				return checkStorageWritable(rootDir)
			}},
			handlers.ReadinessCheck{Name: "index", Check: func() error {
				if !indexReady.Load() {
					return fmt.Errorf("metadata index not rebuilt")
				}
				return nil
			}},
		),
	}
	if metricsExporter != nil {
		routeHandlers.Metrics = handlers.NewMetricsHandler(
			metricsExporter,
			cfg.GetMetricsToken(),
			cfg.EnableAuth(),
			func(r *http.Request) (*models.User, error) {
				return xhttp.ResolveAuthenticatedUser(r, authenticateService, tokenService)
			},
		)
	}
	if auditService.Enabled() {
		routeHandlers.AdminAudit = handlers.NewAdminAuditHandler(auditService)
	}
	if cfg.EnableTwoFactor() {
		routeHandlers.Totp = totpHandler
		routeHandlers.AdminTotpReset = adminTotpResetHandler
	}
	if apiKeyHandler != nil {
		routeHandlers.APIKeys = apiKeyHandler
	}

	server := xhttp.NewServer(
		cfg.GetPort(),
		tlsGenerator,
		logger,
		routeHandlers,
		authenticateService,
		tokenService,
		apiKeyService,
		cfg.EnableAuth(),
	)
	server.ConfigureTLS(cfg.EnableTLS())
	server.SetMetricsRecorder(recorder)

	if cfg.EnablePprof() {
		go servePprof(logger)
	}

	if cfg.EnableGRPC() {
		authService := grpcapi.NewAuthService(registerService, usersService, authenticateService, userInfoService, cfg.EnableSignup(), tokenService)
		shareService := grpcapi.NewShareService(createShareService, listSharesService, revokeShareService)
		eventsService := grpcapi.NewEventsService(eventBus)
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
			cfg.EnableGRPCTLS(),
			authenticateService,
			tokenService,
			cfg.EnableAuth(),
			authService,
			shareService,
			eventsService,
			fileService,
			grpcapi.WithMetrics(recorder),
		)
		if err != nil {
			logger.Fatal("Failed to create gRPC server", "error", err)
			return
		}
		server.SetGRPCHandler(grpcServer.Handler())
		go func() {
			if err := grpcServer.Start(); err != nil {
				logger.Error("gRPC server failed", "error", err)
			}
		}()
		defer grpcServer.Stop()
	}

	// Serve a prebuilt SPA from the same origin only when STATIC_DIR points at
	// a directory containing index.html + assets. Unset by default: the UI is
	// hosted separately (Vercel) and the server mounts the API alone.
	server.SetStaticFileServer(os.Getenv("STATIC_DIR"))

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

// checkStorageWritable is the readiness probe for the storage root: create,
// fsync, and remove a tiny temp file. It catches read-only mounts and full
// disks — the failures that would break every write after boot.
func checkStorageWritable(rootDir string) error {
	f, err := os.CreateTemp(rootDir, ".ready-*")
	if err != nil {
		return err
	}
	name := f.Name()
	if _, err := f.Write([]byte("ok")); err != nil {
		_ = f.Close()
		_ = os.Remove(name)
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(name)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return err
	}
	return os.Remove(name)
}

// servePprof exposes net/http/pprof on a loopback-only side listener. It is
// intentionally never mounted on the public mux: profiling endpoints leak
// internals and must not be reachable from the network.
func servePprof(logger ports.Logger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))

	addr := "127.0.0.1:6060"
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	logger.Info("pprof listening (loopback only)", "address", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("pprof server failed", "error", err)
	}
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
