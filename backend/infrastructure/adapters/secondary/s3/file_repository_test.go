package s3

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// fakeS3 is an in-memory S3-compatible subset used by repository tests.
type fakeS3 struct {
	mu      sync.RWMutex
	objects map[string]fakeObject
}

type fakeObject struct {
	body    []byte
	modTime time.Time
}

func newFakeS3() *fakeS3 {
	return &fakeS3{objects: map[string]fakeObject{}}
}

func (f *fakeS3) handler(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Errorf("missing Authorization header on %s %s", r.Method, r.URL.Path)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		key := strings.TrimPrefix(r.URL.Path, "/")
		// Path-style: /bucket/key... → strip bucket segment.
		if idx := strings.Index(key, "/"); idx >= 0 {
			key = key[idx+1:]
		} else {
			key = ""
		}
		if unescaped, err := url.PathUnescape(key); err == nil {
			key = unescaped
		}

		f.mu.Lock()
		defer f.mu.Unlock()

		switch r.Method {
		case http.MethodHead:
			obj, ok := f.objects[key]
			if !ok {
				writeS3Error(w, http.StatusNotFound, "NoSuchKey")
				return
			}
			w.Header().Set("Content-Length", strconv.Itoa(len(obj.body)))
			w.Header().Set("Last-Modified", obj.modTime.UTC().Format(http.TimeFormat))
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			if r.URL.Query().Get("list-type") == "2" {
				f.writeList(w, r)
				return
			}
			obj, ok := f.objects[key]
			if !ok {
				writeS3Error(w, http.StatusNotFound, "NoSuchKey")
				return
			}
			w.Header().Set("Content-Length", strconv.Itoa(len(obj.body)))
			w.Header().Set("Last-Modified", obj.modTime.UTC().Format(http.TimeFormat))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(obj.body)
		case http.MethodPut:
			if src := r.Header.Get("x-amz-copy-source"); src != "" {
				src = strings.TrimPrefix(src, "/")
				// src is bucket/key
				if i := strings.Index(src, "/"); i >= 0 {
					src = src[i+1:]
				}
				if unescaped, err := url.PathUnescape(src); err == nil {
					src = unescaped
				}
				obj, ok := f.objects[src]
				if !ok {
					writeS3Error(w, http.StatusNotFound, "NoSuchKey")
					return
				}
				f.objects[key] = fakeObject{body: append([]byte(nil), obj.body...), modTime: time.Now()}
				w.Header().Set("Content-Type", "application/xml")
				fmt.Fprint(w, `<?xml version="1.0"?><CopyObjectResult></CopyObjectResult>`)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			f.objects[key] = fakeObject{body: body, modTime: time.Now()}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			delete(f.objects, key)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func (f *fakeS3) writeList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	prefix := q.Get("prefix")
	delimiter := q.Get("delimiter")

	type content struct {
		Key          string    `xml:"Key"`
		LastModified time.Time `xml:"LastModified"`
		Size         int       `xml:"Size"`
	}
	type cp struct {
		Prefix string `xml:"Prefix"`
	}
	var contents []content
	common := map[string]struct{}{}

	keys := make([]string, 0, len(f.objects))
	for k := range f.objects {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		rest := strings.TrimPrefix(k, prefix)
		if delimiter != "" {
			if i := strings.Index(rest, delimiter); i >= 0 {
				common[prefix+rest[:i+1]] = struct{}{}
				continue
			}
		}
		obj := f.objects[k]
		contents = append(contents, content{
			Key:          k,
			LastModified: obj.modTime.UTC(),
			Size:         len(obj.body),
		})
	}
	var prefixes []cp
	for p := range common {
		prefixes = append(prefixes, cp{Prefix: p})
	}
	sort.Slice(prefixes, func(i, j int) bool { return prefixes[i].Prefix < prefixes[j].Prefix })

	type result struct {
		XMLName        xml.Name  `xml:"ListBucketResult"`
		IsTruncated    bool      `xml:"IsTruncated"`
		Contents       []content `xml:"Contents"`
		CommonPrefixes []cp      `xml:"CommonPrefixes"`
	}
	w.Header().Set("Content-Type", "application/xml")
	_ = xml.NewEncoder(w).Encode(result{Contents: contents, CommonPrefixes: prefixes})
}

func writeS3Error(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	fmt.Fprintf(w, `<?xml version="1.0"?><Error><Code>%s</Code><Message>%s</Message></Error>`, code, code)
}

func newTestRepo(t *testing.T) (*FileRepository, *fakeS3) {
	t.Helper()
	fake := newFakeS3()
	srv := httptest.NewServer(fake.handler(t))
	t.Cleanup(srv.Close)

	repo := NewFileRepository(ports.S3Settings{
		Endpoint:  srv.URL,
		Bucket:    "test-bucket",
		Region:    "us-east-1",
		AccessKey: "test-key",
		SecretKey: "test-secret",
		PathStyle: true,
	})
	return repo, fake
}

func TestS3FileRepository_PathTraversalBlocked(t *testing.T) {
	repo, _ := newTestRepo(t)
	if _, _, err := repo.ServeFile("../../etc/passwd"); err == nil {
		t.Fatal("expected path traversal to be blocked")
	}
	if err := repo.CreateDirectory("../evil"); err == nil {
		t.Fatal("expected mkdir traversal to be blocked")
	}
}

func TestS3FileRepository_WriteReadListDelete(t *testing.T) {
	repo, fake := newTestRepo(t)

	if err := repo.CreateDirectory("docs/nested"); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if !strings.HasSuffix(firstKey(fake, "docs/"), "/") {
		t.Fatalf("expected directory marker, keys=%v", keysOf(fake))
	}

	n, err := repo.WriteFile("docs/hello.txt", io.NopCloser(strings.NewReader("hello s3")))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != 8 {
		t.Fatalf("wrote %d, want 8", n)
	}

	rc, name, err := repo.ServeFile("docs/hello.txt")
	if err != nil {
		t.Fatalf("serve: %v", err)
	}
	body, _ := io.ReadAll(rc)
	rc.Close()
	if string(body) != "hello s3" || name != "hello.txt" {
		t.Fatalf("got %q name=%q", body, name)
	}

	files, err := repo.ListDirectory("docs")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var foundFile, foundDir bool
	for _, f := range files {
		if f.Name == "hello.txt" && !f.IsDir && f.Size == 8 {
			foundFile = true
		}
		if f.Name == "nested" && f.IsDir {
			foundDir = true
		}
		if strings.Contains(f.Name, ".versions") {
			t.Fatalf("versions leaked into listing: %+v", f)
		}
	}
	if !foundFile || !foundDir {
		t.Fatalf("listing incomplete: %+v", files)
	}

	isDir, err := repo.IsDirectory("docs")
	if err != nil || !isDir {
		t.Fatalf("IsDirectory(docs)=%v,%v", isDir, err)
	}
	exists, err := repo.FileExists("docs/hello.txt")
	if err != nil || !exists {
		t.Fatalf("FileExists=%v,%v", exists, err)
	}

	if err := repo.DeletePath("docs"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if exists, _ := repo.FileExists("docs/hello.txt"); exists {
		t.Fatal("file should be gone after dir delete")
	}
}

func TestS3FileRepository_KeepsFileVersions(t *testing.T) {
	repo, _ := newTestRepo(t)

	write := func(content string) {
		t.Helper()
		if _, err := repo.WriteFile("doc.txt", io.NopCloser(strings.NewReader(content))); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	write("v1")
	write("v2")
	write("v3")

	info, err := repo.GetFileInfo("doc.txt")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if info.Version != 2 {
		t.Fatalf("version=%d, want 2", info.Version)
	}

	versions, err := repo.ListVersions("doc.txt")
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("versions=%d, want 2", len(versions))
	}
	if versions[0].Version != 1 || versions[1].Version != 2 {
		t.Fatalf("version numbers = %d,%d", versions[0].Version, versions[1].Version)
	}

	rc, name, err := repo.ServeVersion("doc.txt", 1)
	if err != nil {
		t.Fatalf("serve version: %v", err)
	}
	body, _ := io.ReadAll(rc)
	rc.Close()
	if string(body) != "v1" || name != "doc.txt" {
		t.Fatalf("got %q name=%q", body, name)
	}

	// Versions stay hidden from the directory listing.
	listing, err := repo.ListDirectory("")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range listing {
		if strings.Contains(entry.Name, ".versions") {
			t.Fatalf("versions visible: %+v", entry)
		}
	}

	if err := repo.RestoreVersion("doc.txt", 1); err != nil {
		t.Fatalf("restore: %v", err)
	}
	rc, _, err = repo.ServeFile("doc.txt")
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(rc)
	rc.Close()
	if string(body) != "v1" {
		t.Fatalf("restored content = %q", body)
	}
}

func TestS3FileRepository_VersionKeepPrunes(t *testing.T) {
	repo, _ := newTestRepo(t)
	repo.SetVersionKeep(2)
	for i := 1; i <= 5; i++ {
		if _, err := repo.WriteFile("a.txt", io.NopCloser(strings.NewReader(fmt.Sprintf("v%d", i)))); err != nil {
			t.Fatal(err)
		}
	}
	info, err := repo.GetFileInfo("a.txt")
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != 2 {
		t.Fatalf("version=%d, want 2 (keep=2)", info.Version)
	}
}

func TestS3FileRepository_ZipDirectory(t *testing.T) {
	repo, _ := newTestRepo(t)
	if err := repo.CreateDirectory("folder"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.WriteFile("folder/a.txt", io.NopCloser(strings.NewReader("AAA"))); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.WriteFile("folder/b.txt", io.NopCloser(strings.NewReader("BBB"))); err != nil {
		t.Fatal(err)
	}

	rc, err := repo.ZipDirectory("folder")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip: %v", err)
	}
	got := map[string]string{}
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		fr, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(fr)
		fr.Close()
		got[f.Name] = string(b)
	}
	if got["a.txt"] != "AAA" || got["b.txt"] != "BBB" {
		t.Fatalf("zip contents = %#v", got)
	}
}

func TestS3FileRepository_MovePath(t *testing.T) {
	repo, _ := newTestRepo(t)
	if err := repo.CreateDirectory("users/alice/docs"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.WriteFile("users/alice/docs/f.txt", io.NopCloser(strings.NewReader("data"))); err != nil {
		t.Fatal(err)
	}
	if err := repo.MovePath("users/alice", "users/bob"); err != nil {
		t.Fatalf("move: %v", err)
	}
	if exists, _ := repo.FileExists("users/alice/docs/f.txt"); exists {
		t.Fatal("old path should be gone")
	}
	rc, _, err := repo.ServeFile("users/bob/docs/f.txt")
	if err != nil {
		t.Fatalf("new path missing: %v", err)
	}
	body, _ := io.ReadAll(rc)
	rc.Close()
	if string(body) != "data" {
		t.Fatalf("moved content = %q", body)
	}
}

func TestS3FileRepository_WalkSkipsMetadata(t *testing.T) {
	repo, fake := newTestRepo(t)
	fake.mu.Lock()
	fake.objects[".file-share/users.json"] = fakeObject{body: []byte("{}"), modTime: time.Now()}
	fake.objects["shared/readme.md"] = fakeObject{body: []byte("# hi"), modTime: time.Now()}
	fake.objects["shared/x.txt.versions/1"] = fakeObject{body: []byte("old"), modTime: time.Now()}
	fake.objects["shared/x.txt"] = fakeObject{body: []byte("new"), modTime: time.Now()}
	fake.mu.Unlock()

	entries, err := repo.Walk("")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Path, ".file-share") {
			t.Fatalf("metadata walked: %+v", e)
		}
		if strings.Contains(e.Path, ".versions") {
			t.Fatalf("version walked: %+v", e)
		}
	}
	var sawFile, sawDir bool
	for _, e := range entries {
		if e.Path == "shared/readme.md" && !e.IsDir {
			sawFile = true
		}
		if e.Path == "shared" && e.IsDir {
			sawDir = true
		}
	}
	if !sawFile || !sawDir {
		t.Fatalf("walk entries incomplete: %+v", entries)
	}

	docs, err := repo.WalkTextDocuments("")
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if d.Path != "shared/readme.md" && d.Path != "shared/x.txt" {
			t.Fatalf("unexpected doc %q", d.Path)
		}
	}
}

func TestS3FileRepository_GetFileInfoNotFound(t *testing.T) {
	repo, _ := newTestRepo(t)
	if _, err := repo.GetFileInfo("missing.txt"); err == nil {
		t.Fatal("expected not found")
	}
}

func TestClient_SigV4AuthorizationPresent(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		if r.Header.Get("x-amz-date") == "" || r.Header.Get("x-amz-content-sha256") == "" {
			t.Error("missing SigV4 headers")
		}
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0"?><ListBucketResult></ListBucketResult>`)
	}))
	defer srv.Close()

	c := NewClient(ports.S3Settings{
		Endpoint:  srv.URL,
		Bucket:    "b",
		Region:    "us-east-1",
		AccessKey: "AKID",
		SecretKey: "SECRET",
		PathStyle: true,
	})
	if _, err := c.listObjects(listOptions{}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256 Credential=AKID/") {
		t.Fatalf("auth = %q", auth)
	}
	if !strings.Contains(auth, "SignedHeaders=host;x-amz-content-sha256;x-amz-date") {
		t.Fatalf("signed headers missing: %q", auth)
	}
	if !strings.Contains(auth, "Signature=") {
		t.Fatalf("signature missing: %q", auth)
	}
}

func firstKey(f *fakeS3, prefix string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	keys := keysOf(f)
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) {
			return k
		}
	}
	return ""
}

func keysOf(f *fakeS3) []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]string, 0, len(f.objects))
	for k := range f.objects {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
