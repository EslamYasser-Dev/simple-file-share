package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRegisterRoutesLimitsPublicShare verifies the /api/share/ chain actually
// contains the rate limiter by exercising the real mux with a stub handler.
func TestRegisterRoutesLimitsPublicShare(t *testing.T) {
	server := NewServer("0", noopTLS{}, dumbLogger{}, RouteHandlers{
		Share: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}),
	}, nil, false)
	// Shape the budget without the (now removed) public setter, since this test
	// lives in the same package.
	server.shareLimiter = NewIPLimiter(1, 5)

	mux := server.registerRoutes()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	limited, ok := false, false
	for i := 0; i < 40; i++ {
		resp, err := http.Get(srv.URL + "/api/share/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		if err != nil {
			t.Fatalf("req %d: %v", i, err)
		}
		code := resp.StatusCode
		resp.Body.Close()
		switch code {
		case http.StatusNoContent:
			ok = true
		case http.StatusTooManyRequests:
			limited = true
		}
	}
	if !ok {
		t.Fatal("handler was never reached (route not matched?)")
	}
	if !limited {
		t.Fatal("expected at least one 429 from the public share route")
	}
}

type noopTLS struct{}

func (noopTLS) GenerateCert() ([]byte, []byte, error) { return nil, nil, nil }

type dumbLogger struct{}

func (dumbLogger) Info(string, ...any)  {}
func (dumbLogger) Warn(string, ...any)  {}
func (dumbLogger) Error(string, ...any) {}
func (dumbLogger) Fatal(string, ...any) {}
