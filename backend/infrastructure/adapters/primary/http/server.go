package xhttp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "embed"

	"github.com/EslamYasser-Dev/simple-file-share/api"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type RouteHandlers struct {
	Files     http.Handler
	Download  http.Handler
	Upload    http.Handler
	Directory http.Handler
	FileInfo  http.Handler
	Search    http.Handler
	Health    http.Handler
}

type Server struct {
	port           string
	tlsGenerator   ports.TLSCertGenerator
	logger         ports.Logger
	handlers       RouteHandlers
	authProvider   ports.AuthProvider
	enableAuth     bool
	maxUploadBytes int64
	httpServer     *http.Server
	staticDir      string
	useTLS         bool
}

func NewServer(
	port string,
	tlsGen ports.TLSCertGenerator,
	logger ports.Logger,
	handlers RouteHandlers,
	authProvider ports.AuthProvider,
	enableAuth bool,
	maxUploadBytes int64,
) *Server {
	return &Server{
		port:           port,
		tlsGenerator:   tlsGen,
		logger:         logger,
		handlers:       handlers,
		authProvider:   authProvider,
		enableAuth:     enableAuth,
		maxUploadBytes: maxUploadBytes,
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
	}
}

func (s *Server) SetStaticFileServer(dir string) {
	s.staticDir = dir
}

func (s *Server) ConfigureTLS(enableTLS bool) {
	s.useTLS = enableTLS
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
	mux.Handle("/api/files/info", apiChain(s.handlers.FileInfo))
	mux.Handle("/api/files/search", apiChain(s.handlers.Search))
	mux.Handle("/api/upload", chainMiddleware(s.handlers.Upload, append(s.apiMiddlewareFuncs(), MaxBytesMiddleware(s.maxUploadBytes))...))
	mux.Handle("/api/directories", apiChain(s.handlers.Directory))

	mux.Handle("/health", chainMiddleware(s.handlers.Health, corsMiddleware, func(next http.Handler) http.Handler {
		return loggingMiddleware(next, s.logger)
	}))

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

	if s.staticDir == "" {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				http.Redirect(w, r, "/swagger", http.StatusFound)
				return
			}
			http.NotFound(w, r)
		})
	}

	return mux
}

func (s *Server) apiMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return chainMiddleware(h, s.apiMiddlewareFuncs()...)
	}
}

func (s *Server) apiMiddlewareFuncs() []func(http.Handler) http.Handler {
	middlewares := []func(http.Handler) http.Handler{
		corsMiddleware,
		func(next http.Handler) http.Handler {
			return loggingMiddleware(next, s.logger)
		},
	}
	if s.enableAuth && s.authProvider != nil {
		middlewares = append(middlewares, AuthMiddleware(s.authProvider))
	}
	return middlewares
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
