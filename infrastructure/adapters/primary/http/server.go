package xhttp

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// Server configures and runs the HTTPS server.
type Server struct {
	port            string
	tlsGenerator    ports.TLSCertGenerator
	logger          ports.Logger
	listHandler     http.Handler
	downloadHandler http.Handler
	uploadHandler   http.Handler
}

// NewServer creates a new HTTP server with given dependencies.
func NewServer(
	port string,
	tlsGen ports.TLSCertGenerator,
	logger ports.Logger,
	listHandler, downloadHandler, uploadHandler http.Handler,
) *Server {
	return &Server{
		port:            port,
		tlsGenerator:    tlsGen,
		logger:          logger,
		listHandler:     listHandler,
		downloadHandler: downloadHandler,
		uploadHandler:   uploadHandler,
	}
}

// Start initializes TLS and starts listening.
func (s *Server) Start() error {
	s.registerRoutes()

	certPEM, keyPEM, err := s.tlsGenerator.GenerateCert()
	if err != nil {
		return fmt.Errorf("failed to generate TLS certificate: %w", err)
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("failed to parse TLS key pair: %w", err)
	}

	server := &http.Server{
		Addr: ":" + s.port,
		TLSConfig: &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
		},
		ReadTimeout:    24 * time.Hour,
		WriteTimeout:   24 * time.Hour,
		MaxHeaderBytes: 1 << 20,
	}

	s.logger.Info("HTTPS server starting", "address", "https://0.0.0.0"+s.port)

	return server.ListenAndServeTLS("", "")
}

// registerRoutes maps URL paths to handlers.
func (s *Server) registerRoutes() {
	http.Handle("/", s.listHandler)
	http.Handle("/upload", s.uploadHandler)

	// For download, you might want:
	// - /download/* → downloadHandler
	// But currently we handle / and .zip in list/download handler.
	// In real app, use a router like Chi or Gin for cleaner mapping.

	// For now, since download is triggered by path suffix, we handle it in listHandler.
	// Alternatively, refactor to use a proper router.
}
