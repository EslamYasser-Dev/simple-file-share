package services

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/policy"
	"github.com/EslamYasser-Dev/simple-file-share/infrastructure/adapters/secondary/fs"
)

func newThumbFixture(t *testing.T) (*ThumbnailService, *models.User) {
	t.Helper()
	dir := t.TempDir()
	fileRepo := fs.NewLocalFileRepository(dir)
	scoper := policy.NewPathScoper()
	alice := &models.User{Username: "alice"}

	// 800x400 red PNG.
	img := image.NewRGBA(image.Rect(0, 0, 800, 400))
	for y := 0; y < 400; y++ {
		for x := 0; x < 800; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 30, B: 30, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	if _, err := fileRepo.WriteFile("users/alice/pic.png", io.NopCloser(bytes.NewReader(buf.Bytes()))); err != nil {
		t.Fatalf("seed png: %v", err)
	}
	if _, err := fileRepo.WriteFile("users/alice/notes.txt", io.NopCloser(strings.NewReader("hello"))); err != nil {
		t.Fatalf("seed txt: %v", err)
	}
	return NewThumbnailService(fileRepo, scoper), alice
}

// TestThumbnailRendersJPEG verifies aspect-preserving resize, JPEG output,
// width clamping, and the error surface (non-image, missing, directory).
func TestThumbnailRendersJPEG(t *testing.T) {
	svc, alice := newThumbFixture(t)

	thumb, err := svc.Execute(alice, "pic.png", 160)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if thumb.Width != 160 || thumb.Height != 80 {
		t.Fatalf("dimensions = %dx%d, want 160x80", thumb.Width, thumb.Height)
	}
	if len(thumb.Bytes) < 4 || thumb.Bytes[0] != 0xFF || thumb.Bytes[1] != 0xD8 {
		t.Fatal("output is not a JPEG")
	}

	// Cache hit returns identical bytes without re-decoding.
	again, err := svc.Execute(alice, "pic.png", 160)
	if err != nil {
		t.Fatalf("cached: %v", err)
	}
	if !bytes.Equal(thumb.Bytes, again.Bytes) {
		t.Fatal("cache must return identical bytes")
	}

	// Width clamps to the supported range.
	wide, err := svc.Execute(alice, "pic.png", 9999)
	if err != nil || wide.Width != maxThumbWidth {
		t.Fatalf("clamp: width=%d err=%v", wide.Width, err)
	}

	if _, err := svc.Execute(alice, "notes.txt", 160); err == nil {
		t.Fatal("non-image must be rejected")
	}
	if _, err := svc.Execute(alice, "missing.png", 160); err == nil {
		t.Fatal("missing file must be rejected")
	}
	if _, err := svc.Execute(alice, "", 160); err == nil {
		t.Fatal("directory must be rejected")
	}
	// Another user cannot reach alice's namespace.
	if _, err := svc.Execute(&models.User{Username: "bob"}, "pic.png", 160); err == nil {
		t.Fatal("cross-user thumbnail must be rejected")
	}
}
