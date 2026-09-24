package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	xhttp "github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/logging"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

type noopTLSCertGenerator struct{}

func (noopTLSCertGenerator) GenerateCert() ([]byte, []byte, error) { return nil, nil, nil }

// TestShareRouteIsRateLimited drives the full route table exactly as production
// wires it, then bursts the public share endpoint to confirm the limiter is
// attached to the RouteHandlers.Share handler.
func TestShareRouteIsRateLimited(t *testing.T) {
	dir := t.TempDir()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memory.NewFileIndexRepository(), nil)
	scoper := policy.NewPathScoper()
	shareRepo := fs.NewShareFileRepository(dir)
	downloadService := services.NewDownloadService(
		services.NewDownloadFileService(fileRepo, scoper),
		services.NewDownloadZipService(fileRepo, scoper),
	)
	server := xhttp.NewServer("0", noopTLSCertGenerator{}, logging.NewStdLoggerPlain(), xhttp.RouteHandlers{
		Share: NewShareDownloadHandler(services.NewResolveShareService(shareRepo, scoper, downloadService)),
	}, nil, nil, false)

	mux := server.RegisterRoutesForTest()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	limited, ok := false, false
	seen := map[int]int{}
	for i := 0; i < 40; i++ {
		resp, err := http.Get(srv.URL + "/api/share/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		code := resp.StatusCode
		resp.Body.Close()
		seen[code]++
		// The real handler answers 404 for an unknown token; any response
		// other than 429 means the limiter let the request through.
		ok = ok || code != http.StatusTooManyRequests
		limited = limited || code == http.StatusTooManyRequests
	}
	t.Logf("status codes: %v", seen)
	if !ok {
		t.Fatal("handler was never reached (route not matched?)")
	}
	if !limited {
		t.Fatal("expected at least one 429 response, none observed")
	}
}
