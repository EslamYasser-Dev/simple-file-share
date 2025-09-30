package xhttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "embed"

	"github.com/EslamYasser-Dev/simple-file-share/api"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

type Server struct {
	port            string
	tlsGenerator    ports.TLSCertGenerator
	logger          ports.Logger
	rootHandler     http.Handler
	uploadHandler   http.Handler
	httpServer      *http.Server
	staticDir       string // Directory to serve static files from
	useTLS          bool   // Whether to use TLS/HTTPS
}

func NewServer(
	port string,
	tlsGen ports.TLSCertGenerator,
	logger ports.Logger,
	rootHandler, uploadHandler http.Handler,
) *Server {
	server := &http.Server{
		Addr: ":" + port,
		TLSConfig: &tls.Config{
			Certificates:             nil,
			MinVersion:               tls.VersionTLS13,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
		},
		ReadTimeout:    24 * time.Hour,
		WriteTimeout:   24 * time.Hour,
		MaxHeaderBytes: 1 << 20,
	}

	return &Server{
		port:            port,
		tlsGenerator:    tlsGen,
		logger:          logger,
		rootHandler:     rootHandler,
		uploadHandler:   uploadHandler,
		httpServer:      server,
	}
}

// Start initializes TLS and starts listening with graceful shutdown.
// SetStaticFileServer configures the server to serve static files from the specified directory.
// If the directory doesn't exist, this is a no-op.
// SetStaticFileServer configures the server to serve static files from the specified directory.
// If the directory doesn't exist, this is a no-op.
func (s *Server) SetStaticFileServer(dir string) {
	s.staticDir = dir
}

// ConfigureTLS enables or disables TLS/HTTPS for the server
func (s *Server) ConfigureTLS(enableTLS bool) {
	s.useTLS = enableTLS
}

func (s *Server) Start() error {
	mux := s.registerRoutes()
	
	// If static directory is set and exists, serve static files
	if s.staticDir != "" {
		if _, err := os.Stat(s.staticDir); !os.IsNotExist(err) {
			fs := http.FileServer(http.Dir(s.staticDir))
			mux.Handle("/", fs)
			s.logger.Info("Serving static files from", "directory", s.staticDir)
		} else {
			s.logger.Warn("Static directory does not exist, not serving static files", "directory", s.staticDir)
		}
	}

	s.httpServer.Handler = mux

	// Only generate and configure TLS if enabled
	if s.useTLS {
		certPEM, keyPEM, err := s.tlsGenerator.GenerateCert()
		if err != nil {
			return fmt.Errorf("failed to generate TLS certificate: %w", err)
		}

		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return fmt.Errorf("failed to parse TLS key pair: %w", err)
		}

		s.httpServer.TLSConfig.Certificates = []tls.Certificate{cert}
	}

	// Create context for shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	go func() {
		// Start with the appropriate protocol
		protocol := "http"
		if s.useTLS {
			protocol = "https"
		}
		s.logger.Info("Server starting",
			"protocol", protocol,
			"address", "0.0.0.0:"+s.port)

		var err error
		if s.useTLS {
			err = s.httpServer.ListenAndServeTLS("", "")
		} else {
			err = s.httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	s.logger.Info("Shutdown signal received, gracefully stopping server...")

	// Give active connections 10 seconds to finish
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	s.logger.Info("Server exited gracefully")
	return nil
}

// registerRoutes maps URL paths to handlers and returns a mux.
// loggingMiddleware adds logging for all requests
func loggingMiddleware(next http.Handler, logger ports.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Request started", 
			"method", r.Method, 
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr)
		
		// Create a response writer that captures the status code
		rw := &responseWriter{
			ResponseWriter: w,
			status:        http.StatusOK,
		}
		
		// Serve the request
		next.ServeHTTP(rw, r)
		
		// Log the response
		logger.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"bytes", rw.bytesWritten)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
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

func (s *Server) registerRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	
	// Set up API routes
	mux.Handle("/api/files", s.rootHandler)
	mux.Handle("/api/files/", s.rootHandler)
	mux.Handle("/api/files/download", s.rootHandler)
	mux.Handle("/api/upload", s.uploadHandler)

	// Swagger documentation
	mux.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		if _, err := w.Write(api.SwaggerSpec); err != nil {
			s.logger.Error("Failed to write swagger spec", "error", err)
		}
	})

	// Swagger UI
	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		html := `<!DOCTYPE html>
    <html>
      <head>
        <title>Swagger UI</title>
        <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4/swagger-ui.css" />
      </head>
      <body>
        <div id="swagger-ui"></div>
        <script src="https://unpkg.com/swagger-ui-dist@4/swagger-ui-bundle.js"></script>
        <script>
          SwaggerUIBundle({
            url: '/swagger.yaml',
            dom_id: '#swagger-ui',
          });
        </script>
      </body>
    </html>`
		if _, err := w.Write([]byte(html)); err != nil {
			s.logger.Error("Failed to write Swagger UI", "error", err)
		}
	})

	// Fallback root handler when no static files are being served
	if s.staticDir == "" {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// Redirect to Swagger UI as a helpful default in development
			http.Redirect(w, r, "/swagger", http.StatusFound)
		})
	}
	return mux
}
