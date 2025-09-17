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
func (s *Server) Start() error {
	mux := s.registerRoutes()
	s.httpServer.Handler = mux

	certPEM, keyPEM, err := s.tlsGenerator.GenerateCert()
	if err != nil {
		return fmt.Errorf("failed to generate TLS certificate: %w", err)
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("failed to parse TLS key pair: %w", err)
	}

	s.httpServer.TLSConfig.Certificates = []tls.Certificate{cert}

	// Create context for shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	go func() {
		s.logger.Info("HTTPS server starting", "address", "https://0.0.0.0:"+s.port)
		if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
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
func (s *Server) registerRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", s.rootHandler)
	mux.Handle("/upload", s.uploadHandler)

	mux.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(api.SwaggerSpec)
	})

	mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
    <!DOCTYPE html>
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
          })
        </script>
      </body>
    </html>`))
	})
	return mux
}
