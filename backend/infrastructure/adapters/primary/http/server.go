package xhttp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "embed"

	"github.com/EslamYasser-Dev/simple-file-share/api"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type RouteHandlers struct {
	Files      http.Handler
	Download   http.Handler
	View       http.Handler
	Upload     http.Handler
	Update     http.Handler
	Directory  http.Handler
	FileInfo   http.Handler
	Search     http.Handler
	Register   http.Handler
	Me         http.Handler
	AuthInfo   http.Handler
	AdminUsers http.Handler
	AdminQuota http.Handler
	Shares     http.Handler
	Share      http.Handler
	Versions   http.Handler
	Version    http.Handler
	Restore    http.Handler
	Health     http.Handler
	Token      http.Handler
	Refresh    http.Handler
	Revoke     http.Handler
	OAuthStart http.Handler
	OAuthCb    http.Handler
}

type Server struct {
	port         string
	tlsGenerator ports.TLSCertGenerator
	logger       ports.Logger
	handlers     RouteHandlers
	authService  *services.AuthenticateService
	tokenService *services.TokenService
	enableAuth   bool
	httpServer   *http.Server
	staticDir    string
	useTLS       bool
	shareLimiter *IPLimiter
	apiLimiter   *IPLimiter
	authLimiter  *IPLimiter
}

func NewServer(
	port string,
	tlsGen ports.TLSCertGenerator,
	logger ports.Logger,
	handlers RouteHandlers,
	authService *services.AuthenticateService,
	tokenService *services.TokenService,
	enableAuth bool,
) *Server {
	return &Server{
		port:         port,
		tlsGenerator: tlsGen,
		logger:       logger,
		handlers:     handlers,
		authService:  authService,
		tokenService: tokenService,
		enableAuth:   enableAuth,
		httpServer: &http.Server{
			Addr: ":" + port,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
			ReadTimeout:       DefaultReadTimeout,
			ReadHeaderTimeout: DefaultReadTimeout,
			WriteTimeout:      DefaultWriteTimeout,
			IdleTimeout:       DefaultIdleTimeout,
			MaxHeaderBytes:    DefaultMaxHeaderBytes,
		},
		shareLimiter: NewIPLimiter(shareLimitRate, shareLimitBurst),
		apiLimiter:   NewIPLimiter(shareLimitRate, shareLimitBurst),
		// Stricter budget for credential endpoints (login/register/oauth).
		authLimiter: NewIPLimiter(0.5, 10),
	}
}

func (s *Server) SetStaticFileServer(dir string) {
	s.staticDir = dir
}

func (s *Server) ConfigureTLS(enableTLS bool) {
	s.useTLS = enableTLS
}

// RegisterRoutesForTest exposes the internal mux for integration tests.
func (s *Server) RegisterRoutesForTest() *http.ServeMux {
	return s.registerRoutes()
}

func (s *Server) Start() error {
	mux := s.registerRoutes()

	if s.staticDir != "" {
		if _, err := os.Stat(s.staticDir); !os.IsNotExist(err) {
			fs := http.FileServer(http.Dir(s.staticDir))
			mux.Handle("/", fs)
			s.logger.Info("Serving static files", "directory", s.staticDir)
		} else {
			s.logger.Warn("Static directory missing", "directory", s.staticDir)
		}
	}

	s.httpServer.Handler = mux

	if s.useTLS {
		certPEM, keyPEM, err := s.tlsGenerator.GenerateCert()
		if err != nil {
			return fmt.Errorf("generate TLS cert: %w", err)
		}
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return fmt.Errorf("parse TLS key pair: %w", err)
		}
		s.httpServer.TLSConfig.Certificates = []tls.Certificate{cert}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		protocol := "http"
		if s.useTLS {
			protocol = "https"
		}
		s.logger.Info("Server starting", "protocol", protocol, "address", "0.0.0.0:"+s.port)

		var err error
		if s.useTLS {
			err = s.httpServer.ListenAndServeTLS("", "")
		} else {
			err = s.httpServer.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Server failed", "error", err)
		}
	}()

	<-ctx.Done()
	s.logger.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Forced shutdown", "error", err)
		return err
	}

	s.logger.Info("Server exited gracefully")
	return nil
}

func (s *Server) registerRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	apiChain := s.apiMiddleware()
	mux.Handle("/api/files", apiChain(s.handlers.Files))
	mux.Handle("/api/files/download", apiChain(s.handlers.Download))
	mux.Handle("/api/files/view", apiChain(s.handlers.View))
	mux.Handle("/api/files/info", apiChain(s.handlers.FileInfo))
	mux.Handle("/api/files/search", apiChain(s.handlers.Search))
	mux.Handle("/api/upload", apiChain(s.handlers.Upload))
	mux.Handle("/api/files/content", apiChain(s.handlers.Update))
	mux.Handle("/api/directories", apiChain(s.handlers.Directory))
	mux.Handle("/api/auth/me", apiChain(s.handlers.Me))
	mux.Handle("/api/admin/users", apiChain(s.handlers.AdminUsers))
	mux.Handle("/api/admin/users/{username}/quota", apiChain(s.handlers.AdminQuota))
	mux.Handle("/api/shares", apiChain(s.handlers.Shares))
	mux.Handle("/api/files/versions", apiChain(s.handlers.Versions))
	mux.Handle("/api/files/version", apiChain(s.handlers.Version))
	mux.Handle("/api/files/version/restore", apiChain(s.handlers.Restore))

	// Public auth endpoints (no credentials required). Rate-limited harder
	// than the general API so password guessing and state flooding are blunt.
	publicMiddleware := func(h http.Handler) http.Handler {
		return chainMiddleware(
			h,
			corsMiddleware,
			securityHeaders,
			RateLimitMiddleware(s.authLimiter),
			func(next http.Handler) http.Handler {
				return loggingMiddleware(next, s.logger)
			},
		)
	}
	mux.Handle("/api/auth/register", publicMiddleware(s.handlers.Register))
	mux.Handle("/api/auth/info", publicMiddleware(s.handlers.AuthInfo))
	mux.Handle("/api/auth/token", publicMiddleware(s.handlers.Token))
	mux.Handle("/api/auth/refresh", publicMiddleware(s.handlers.Refresh))
	mux.Handle("/api/auth/revoke", publicMiddleware(s.handlers.Revoke))
	mux.Handle("/api/auth/oauth/{provider}/start", publicMiddleware(s.handlers.OAuthStart))
	mux.Handle("/api/auth/oauth/{provider}/callback", publicMiddleware(s.handlers.OAuthCb))

	// Public share links: the token is the credential. The route is rate-limited
	// per client address so it cannot be swept for valid tokens.
	mux.Handle(
		"/api/share/",
		chainMiddleware(
			s.handlers.Share,
			corsMiddleware,
			RateLimitMiddleware(s.shareLimiter),
			func(next http.Handler) http.Handler {
				return loggingMiddleware(next, s.logger)
			},
		),
	)

	mux.Handle("/health", publicMiddleware(s.handlers.Health))

	if swaggerEnabled() {
		mux.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write(api.SwaggerSpec)
		})

		mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Swagger UI</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@4/swagger-ui.css"/></head>
<body><div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@4/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({url:'/swagger.yaml',dom_id:'#swagger-ui'})</script>
</body></html>`))
		})
	}

	if s.staticDir == "" {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" && swaggerEnabled() {
				http.Redirect(w, r, "/swagger", http.StatusFound)
				return
			}
			http.NotFound(w, r)
		})
	}

	return mux
}

// swaggerEnabled honors ENABLE_SWAGGER. Unset defaults to on outside
// production so local docs stay available without exposing the UI publicly.
func swaggerEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ENABLE_SWAGGER"))) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return os.Getenv("APP_ENV") != "production"
	}
}

func (s *Server) apiMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return chainMiddleware(h, s.apiMiddlewareFuncs()...)
	}
}

func (s *Server) apiMiddlewareFuncs() []func(http.Handler) http.Handler {
	middlewares := []func(http.Handler) http.Handler{
		corsMiddleware,
		securityHeaders,
		func(next http.Handler) http.Handler {
			return loggingMiddleware(next, s.logger)
		},
	}
	if s.enableAuth && s.authService != nil {
		middlewares = append(middlewares, AuthMiddleware(s.authService, s.tokenService))
	}
	middlewares = append(middlewares, RateLimitMiddleware(s.apiLimiter))
	return middlewares
}

// securityHeaders mitigates MIME sniffing and reduces cross-origin leakage of
// credentials-bearing URLs (including OAuth redirects).
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Flush passes through streaming flushes so large downloads stream properly.
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
