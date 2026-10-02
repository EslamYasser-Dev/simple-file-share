package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/application/services"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/primary/http/dto"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/auth"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/memory"
)

func TestParsePagingDefaultsAndClamps(t *testing.T) {
	cases := []struct {
		query string
		limit int
	}{
		{"", 50},
		{"?limit=25", 25},
		{"?limit=0", 50},   // non-positive falls back
		{"?limit=abc", 50}, // malformed falls back, like the old behavior
		{"?limit=9999", 500},
		{"?limit=500", 500},
	}
	for _, c := range cases {
		p, err := parsePaging(httptest.NewRequest(http.MethodGet, "/x"+c.query, nil))
		if err != nil {
			t.Fatalf("parsePaging(%q): %v", c.query, err)
		}
		if p.limit != c.limit || p.offset != 0 {
			t.Errorf("parsePaging(%q) = limit %d offset %d, want limit %d", c.query, p.limit, p.offset, c.limit)
		}
	}
}

func TestParsePagingRejectsBadCursors(t *testing.T) {
	bad := []string{
		"@@not-base64@@",
		base64.RawURLEncoding.EncodeToString([]byte("not json")),
		encodeTestCursor(-1),
	}
	for _, cursor := range bad {
		if _, err := parsePaging(httptest.NewRequest(http.MethodGet, "/x?cursor="+cursor, nil)); err == nil {
			t.Errorf("cursor %q accepted, want error", cursor)
		}
	}
	// A well-formed cursor decodes back to its offset.
	p, err := parsePaging(httptest.NewRequest(http.MethodGet, "/x?cursor="+encodeTestCursor(120), nil))
	if err != nil || p.offset != 120 {
		t.Errorf("valid cursor → offset %d err %v, want 120", p.offset, err)
	}
}

func encodeTestCursor(offset int) string {
	raw, _ := json.Marshal(struct {
		Offset int `json:"o"`
	}{Offset: offset})
	return base64.RawURLEncoding.EncodeToString(raw)
}

func TestWindowSlicesInMemoryLists(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	p := paging{limit: 2}

	page := window(p, items)
	if len(page.Items) != 2 || page.Items[0] != 1 || page.NextCursor == "" {
		t.Fatalf("page1 = %+v", page)
	}

	p.offset = 2
	page = window(p, items)
	if len(page.Items) != 2 || page.Items[0] != 3 || page.NextCursor == "" {
		t.Fatalf("page2 = %+v", page)
	}

	p.offset = 4
	page = window(p, items)
	if len(page.Items) != 1 || page.Items[0] != 5 || page.NextCursor != "" {
		t.Fatalf("last page = %+v", page)
	}

	// Past the end: empty page, no cursor, never a null items array.
	p.offset = 99
	page = window(p, items)
	if page.Items == nil || len(page.Items) != 0 || page.NextCursor != "" {
		t.Fatalf("overflow page = %+v", page)
	}
	page = window(paging{limit: 10}, []int(nil))
	if page.Items == nil {
		t.Fatal("nil backing list must still encode items as []")
	}
}

func pageJSON(t *testing.T, rec *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
}

func getPage(t *testing.T, h http.Handler, target string) dto.Page[dto.FileItem] {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d body=%s", target, rec.Code, rec.Body)
	}
	var page dto.Page[dto.FileItem]
	pageJSON(t, rec, &page)
	return page
}

func TestListHandlerPaginates(t *testing.T) {
	dir := t.TempDir()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), memory.NewFileIndexRepository(), nil)
	handler := NewListHandler(services.NewListFilesService(fileRepo, policy.NewPathScoper()))
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	first := getPage(t, handler, "/api/files?limit=2")
	if len(first.Items) != 2 || first.NextCursor == "" {
		t.Fatalf("page1 = %+v", first)
	}
	second := getPage(t, handler, "/api/files?limit=2&cursor="+first.NextCursor)
	if len(second.Items) != 1 || second.NextCursor != "" {
		t.Fatalf("page2 = %+v", second)
	}

	// Cursor is opaque; a broken one is a 400, not a 500.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/files?cursor=@@@@", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad cursor = %d, want 400", rec.Code)
	}
}

func TestSearchHandlerPaginates(t *testing.T) {
	index := memory.NewFileIndexRepository()
	for i := 1; i <= 5; i++ {
		info := &models.FileInfo{
			Name: "report-" + string(rune('0'+i)) + ".txt",
			Path: "report-" + string(rune('0'+i)) + ".txt",
			Size: 1,
		}
		if err := index.Upsert(info); err != nil {
			t.Fatal(err)
		}
	}
	handler := NewSearchHandler(services.NewSearchFilesService(index, policy.NewPathScoper(), nil))

	first := getPage(t, handler, "/api/files/search?q=report&limit=2")
	if len(first.Items) != 2 || first.NextCursor == "" {
		t.Fatalf("page1 = %+v", first)
	}
	second := getPage(t, handler, "/api/files/search?q=report&limit=2&cursor="+first.NextCursor)
	if len(second.Items) == 0 || second.NextCursor == "" {
		t.Fatalf("page2 = %+v", second)
	}
	// Walk to the end: 5 hits total, cursors must terminate.
	pages, cursor := 1, second.NextCursor
	total := len(first.Items) + len(second.Items)
	for cursor != "" {
		pages++
		if pages > 10 {
			t.Fatal("cursor did not terminate")
		}
		p := getPage(t, handler, "/api/files/search?q=report&limit=2&cursor="+cursor)
		total += len(p.Items)
		cursor = p.NextCursor
	}
	if total != 5 {
		t.Fatalf("served %d hits, want 5", total)
	}
}

func TestShareListPaginates(t *testing.T) {
	f := newShareHandlerFixture(t)
	for i := 0; i < 3; i++ {
		if rec := doJSON(t, f.shares, http.MethodPost, "/api/shares", `{"path":"/hello.txt"}`); rec.Code != http.StatusCreated {
			t.Fatalf("create %d = %d %s", i, rec.Code, rec.Body)
		}
	}

	first := decodeShares(t, doJSON(t, f.shares, http.MethodGet, "/api/shares?limit=2", ""))
	if len(first.Items) != 2 || first.NextCursor == "" {
		t.Fatalf("page1 = %+v", first)
	}
	second := decodeShares(t, doJSON(t, f.shares, http.MethodGet, "/api/shares?limit=2&cursor="+first.NextCursor, ""))
	if len(second.Items) != 1 || second.NextCursor != "" {
		t.Fatalf("page2 = %+v", second)
	}
}

func TestAdminRolesPaginate(t *testing.T) {
	handler := NewAdminRolesHandler(services.NewListRolesService(services.NewRoleCatalog(nil)), nil)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/admin/roles?limit=1", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("page1 = %d body=%s", rec.Code, rec.Body)
	}
	var first dto.Page[dto.RoleResponse]
	pageJSON(t, rec, &first)
	if len(first.Items) != 1 || first.NextCursor == "" {
		t.Fatalf("page1 = %+v", first)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/admin/roles?limit=1&cursor="+first.NextCursor, nil))
	var second dto.Page[dto.RoleResponse]
	pageJSON(t, rec, &second)
	if len(second.Items) != 1 || second.NextCursor != "" {
		t.Fatalf("page2 = %+v (two built-in roles expected)", second)
	}
}

func TestAdminUsersPaginate(t *testing.T) {
	dir := t.TempDir()
	index := memory.NewFileIndexRepository()
	fileRepo := fs.NewIndexedFileRepository(fs.NewLocalFileRepository(dir), index, nil)
	userRepo := fs.NewUserFileRepository(dir)
	scoper := policy.NewPathScoper()
	register := services.NewRegisterUserService(userRepo, auth.NewPBKDF2Hasher(), fileRepo, scoper, true, 0)
	for _, name := range []string{"alice", "bob", "carol"} {
		if _, err := register.Execute(name, "secret-pass-1"); err != nil {
			t.Fatalf("register %s: %v", name, err)
		}
	}

	handler := NewAdminUserItemHandler(
		services.NewListUsersService(userRepo, index, scoper, services.NewRoleCatalog(nil)),
		nil, nil, nil, nil,
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/admin/users?limit=2", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("page1 = %d body=%s", rec.Code, rec.Body)
	}
	var first dto.Page[dto.UserStatsResponse]
	pageJSON(t, rec, &first)
	if len(first.Items) != 2 || first.NextCursor == "" {
		t.Fatalf("page1 = %+v", first)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/admin/users?limit=2&cursor="+first.NextCursor, nil))
	var second dto.Page[dto.UserStatsResponse]
	pageJSON(t, rec, &second)
	if len(second.Items) != 1 || second.NextCursor != "" {
		t.Fatalf("page2 = %+v", second)
	}
}
