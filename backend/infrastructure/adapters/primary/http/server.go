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

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/EslamYasser-Dev/simple-file-share/api"
	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type RouteHandlers struct {
	Files           http.Handler
	Download        http.Handler
	View            http.Handler
	Upload          http.Handler
	ResumableUpload http.Handler
	Update          http.Handler
	Directory       http.Handler
	FileInfo        http.Handler
	Search          http.Handler
	Events          http.Handler
	P2P             http.Handler
	Register        http.Handler
	Me              http.Handler
	AuthInfo        http.Handler
	AdminUsers      http.Handler
	AdminUser       http.Handler
	AdminQuota      http.Handler
	AdminPass       http.Handler
	AdminRoles      http.Handler
	AdminRole       http.Handler
	AdminAnalytics  http.Handler
	SelfPass        http.Handler
	Shares          http.Handler
	Share           http.Handler
	Versions        http.Handler
	Version         http.Handler
	Restore         http.Handler
	Health          http.Handler
	Token           http.Handler
	Refresh         http.Handler
	Revoke          http.Handler
	OAuthStart      http.Handler
	OAuthCb         http.Handler
	GRPCCert        http.Handler
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
	// Chunked uploads issue one request per part; a dedicated higher budget
	// keeps resume traffic from tripping the general API limiter.
	uploadLimiter *IPLimiter
	grpcHandler   http.Handler
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
			ReadHeaderTimeout: DefaultReadHeaderTimeout,
			WriteTimeout:      DefaultWriteTimeout,
			IdleTimeout:       DefaultIdleTimeout,
			MaxHeaderBytes:    DefaultMaxHeaderBytes,
		},
		shareLimiter: NewIPLimiter(shareLimitRate, shareLimitBurst),
		apiLimiter:   NewIPLimiter(shareLimitRate, shareLimitBurst),
		// Stricter budget for credential endpoints (login/register/oauth).
		authLimiter: NewIPLimiter(0.5, 10),
		// Resumable chunk traffic: higher rate/burst than the general API.
		uploadLimiter: NewIPLimiter(50, 100),
	}
}

func (s *Server) SetStaticFileServer(dir string) {
	s.staticDir = dir
}

// SetGRPCHandler mounts a gRPC http.Handler alongside the REST mux on this
// server's listener. gRPC calls are detected by HTTP/2 + gRPC content type;
// everything else goes to the normal mux.
func (s *Server) SetGRPCHandler(h http.Handler) {
	s.grpcHandler = h
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

	// Wrap the whole mux so the static SPA, share links, and swagger routes
	// receive the same security headers as the API chains.
	base := securityHeaders(mux)
	if s.grpcHandler != nil {
		grpcHandler := s.grpcHandler
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor >= 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
				grpcHandler.ServeHTTP(w, r)
				return
			}
			base.ServeHTTP(w, r)
		})
		// h2c lets HTTP/2 cleartext (gRPC) through proxies or local clients
		// share the listener; TLS mode gets HTTP/2 from Go's ALPN instead.
		s.httpServer.Handler = h2c.NewHandler(inner, &http2.Server{})
	} else {
		s.httpServer.Handler = base
	}

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
	mux.Handle("/api/events", apiChain(s.handlers.Events))
	// WebRTC signaling: presence + short SDP/ICE bodies (rate-limited like API).
	mux.Handle("/api/p2p/stream", apiChain(s.handlers.P2P))
	mux.Handle("/api/p2p/peers", apiChain(s.handlers.P2P))
	mux.Handle("/api/p2p/signal", apiChain(s.handlers.P2P))
	mux.Handle("/api/upload", apiChain(s.handlers.Upload))
	// Resumable sessions need their own limiter (many small PATCH requests).
	mux.Handle("/api/uploads", chainMiddleware(
		s.handlers.ResumableUpload,
		corsMiddleware,
		securityHeaders,
		func(next http.Handler) http.Handler {
			return loggingMiddleware(next, s.logger)
		},
		s.authMiddlewareIfEnabled(),
		RateLimitMiddleware(s.uploadLimiter),
	))
	mux.Handle("/api/uploads/", chainMiddleware(
		s.handlers.ResumableUpload,
		corsMiddleware,
		securityHeaders,
		func(next http.Handler) http.Handler {
			return loggingMiddleware(next, s.logger)
		},
		s.authMiddlewareIfEnabled(),
		RateLimitMiddleware(s.uploadLimiter),
	))
	mux.Handle("/api/files/content", apiChain(s.handlers.Update))
	mux.Handle("/api/directories", apiChain(s.handlers.Directory))
	mux.Handle("/api/auth/me", apiChain(s.handlers.Me))
	mux.Handle("/api/auth/password", apiChain(s.handlers.SelfPass))
	mux.Handle("/api/admin/users", apiChain(s.handlers.AdminUser))
	mux.Handle("/api/admin/users/{username}", apiChain(s.handlers.AdminUser))
	mux.Handle("/api/admin/users/{username}/quota", apiChain(s.handlers.AdminQuota))
	mux.Handle("/api/admin/users/{username}/password", apiChain(s.handlers.AdminPass))
	mux.Handle("/api/admin/roles", apiChain(s.handlers.AdminRoles))
	mux.Handle("/api/admin/roles/{name}", apiChain(s.handlers.AdminRole))
	mux.Handle("/api/admin/analytics/{report}", apiChain(s.handlers.AdminAnalytics))
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

	// Public: the gRPC TLS certificate for client-side pinning, fetched over
	// the trusted HTTPS API before the mobile app opens the proxied channel.
	mux.Handle("/api/grpc/cert", publicMiddleware(s.handlers.GRPCCert))

	// Public share links: the token is the credential. The route is rate-limited
	// per client address so it cannot be swept for valid tokens.
	mux.Handle(
		"/api/share/",
		chainMiddleware(
			s.handlers.Share,
			corsMiddleware,
			securityHeaders,
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

// authMiddlewareIfEnabled returns the auth middleware when auth is on, or a
// no-op pass-through when auth is disabled (system view).
func (s *Server) authMiddlewareIfEnabled() func(http.Handler) http.Handler {
	if s.enableAuth && s.authService != nil {
		return AuthMiddleware(s.authService, s.tokenService)
	}
	return func(next http.Handler) http.Handler { return next }
}

// securityHeaders mitigates MIME sniffing, reduces cross-origin leakage of
// credentials-bearing URLs (including OAuth redirects), blocks framing, and
// restricts which origins may serve scripts/styles/fonts for responses served
// from this origin. The Content-Security-Policy is skipped for the swagger UI
// (which loads its bundle from a CDN) while every other header still applies.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		if requestIsHTTPS(r) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		}
		if !strings.HasPrefix(r.URL.Path, "/swagger") {
			w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		}
		next.ServeHTTP(w, r)
	})
}

// contentSecurityPolicy keeps same-origin as the only script/connect source,
// allows inline style attributes and the pre-paint theme script, and lets the
// app load its webfonts and blob/data previews. frame-ancestors duplicates
// X-Frame-Options for modern browsers.
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self' 'unsafe-inline'; " +
	"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
	"font-src 'self' data: https://fonts.gstatic.com; " +
	"img-src 'self' data: blob:; " +
	"media-src 'self' blob:; " +
	"connect-src 'self'; " +
	"frame-src 'self' blob:; " +
	"object-src 'none'; " +
	"base-uri 'self'; " +
	"form-action 'self'; " +
	"frame-ancestors 'none'"

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
