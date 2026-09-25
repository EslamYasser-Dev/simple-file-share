package handlers

import "net/http"

// GRPCCertSource exposes the certificate the gRPC listener serves. ok is
// false when no key pair exists yet (gRPC TLS disabled).
type GRPCCertSource interface {
	CertPEM() ([]byte, bool)
}

// GRPCCertHandler serves the gRPC TLS certificate so mobile clients can pin
// it over the already-trusted HTTPS API before opening the proxied gRPC
// channel.
type GRPCCertHandler struct {
	source GRPCCertSource
}

func NewGRPCCertHandler(source GRPCCertSource) *GRPCCertHandler {
	return &GRPCCertHandler{source: source}
}

func (h *GRPCCertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	certPEM, ok := h.source.CertPEM()
	if !ok {
		http.Error(w, "grpc tls disabled", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(certPEM)
}
