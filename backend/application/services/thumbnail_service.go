package services

import (
	"bytes"
	"container/list"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"sync"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
	"github.com/EslamYasser-Dev/simple-file-share/domain/valueobjects"
)

// Thumbnail bounds and budgets: small enough for lists, big enough to look
// sharp on 3x screens. Decoding is capped so a hostile file cannot eat RAM.
const (
	defaultThumbWidth = 320
	minThumbWidth     = 32
	maxThumbWidth     = 512
	maxThumbSource    = 40 << 20 // 40 MiB
	maxThumbEntries   = 128
	thumbQuality      = 70
	// Pixel caps read from DecodeConfig before any full decode. JPEG decodes
	// to ~1.5 bytes/pixel (YCbCr) and everything else to RGBA at 4 bytes, so
	// each cap keeps a worst-case decode near 120 MB while still accepting
	// modern phone photos.
	maxThumbJPEGPixels = 80_000_000
	maxThumbPixels     = 32_000_000
	thumbHeaderBytes   = 1 << 20 // enough for any image header
)

// Thumbnail is a re-encoded JPEG preview. Re-encoding drops EXIF/GPS metadata
// by construction, so previews never leak location tags.
type Thumbnail struct {
	Bytes  []byte
	Width  int
	Height int
}

type thumbKey struct {
	owner string
	path  string
	size  int64
	mod   int64
	width int
}

type thumbEntry struct {
	key   thumbKey
	thumb *Thumbnail
}

// ThumbnailService renders small JPEG previews of images the caller may
// read. It mirrors the view path's authZ (path scoper) and keeps a bounded
// in-memory LRU so scrolling a photo folder does not re-decode every row.
// Concurrent misses for the same image are collapsed to one decode.
type ThumbnailService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper

	mu       sync.Mutex
	items    map[thumbKey]*list.Element
	order    *list.List
	inflight map[thumbKey]chan struct{}
}

func NewThumbnailService(fileRepo ports.FileRepository, scoper ports.PathScoper) *ThumbnailService {
	return &ThumbnailService{
		fileRepo: fileRepo,
		scoper:   scoper,
		items:    make(map[thumbKey]*list.Element),
		order:    list.New(),
		inflight: make(map[thumbKey]chan struct{}),
	}
}

// Execute renders (or serves from cache) a JPEG thumbnail of virtual path.
func (s *ThumbnailService) Execute(user *models.User, path string, width int) (*Thumbnail, error) {
	if width <= 0 {
		width = defaultThumbWidth
	}
	if width < minThumbWidth {
		width = minThumbWidth
	}
	if width > maxThumbWidth {
		width = maxThumbWidth
	}

	fp, err := valueobjects.NewFilePath(requestPath(path))
	if err != nil {
		return nil, err
	}
	virtual := fp.Relative()
	physical, err := s.scoper.ReadPath(user, virtual)
	if err != nil {
		return nil, err
	}
	exists, err := s.fileRepo.FileExists(physical)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, &domainerrors.NotFoundError{Path: path}
	}
	isDir, err := s.fileRepo.IsDirectory(physical)
	if err != nil {
		return nil, err
	}
	if isDir {
		return nil, &domainerrors.IsDirectoryError{Path: path}
	}
	info, err := s.fileRepo.GetFileInfo(physical)
	if err != nil {
		return nil, err
	}
	if info.Size > maxThumbSource {
		return nil, domainerrors.NewValidationError("path", path, "file too large to thumbnail")
	}

	owner := ""
	if user != nil {
		owner = user.Username
	}
	key := thumbKey{owner: owner, path: virtual, size: info.Size, mod: info.Modified.UnixNano(), width: width}
	if cached := s.cached(key); cached != nil {
		return cached, nil
	}

	// Collapse concurrent misses for the same image: the loser waits for the
	// winner instead of starting a second (potentially huge) decode.
	if wait := s.beginGen(key); wait != nil {
		<-wait
		if cached := s.cached(key); cached != nil {
			return cached, nil
		}
	} else {
		defer s.endGen(key)
	}
	return s.generate(path, physical, key, width)
}

// generate performs the bounded decode + downscale. The header is probed with
// DecodeConfig first so absurd pixel dimensions are rejected before a full
// decode can allocate for them.
func (s *ThumbnailService) generate(path, physical string, key thumbKey, width int) (*Thumbnail, error) {
	probe, _, err := s.fileRepo.ServeFile(physical)
	if err != nil {
		return nil, err
	}
	cfg, cfgFormat, cfgErr := image.DecodeConfig(io.LimitReader(probe, thumbHeaderBytes))
	_ = probe.Close()
	if cfgErr == nil {
		pixels := int64(cfg.Width) * int64(cfg.Height)
		limit := int64(maxThumbPixels)
		if cfgFormat == "jpeg" {
			limit = maxThumbJPEGPixels
		}
		if cfg.Width <= 0 || cfg.Height <= 0 || pixels > limit {
			return nil, domainerrors.NewValidationError("path", path, "image too large to thumbnail")
		}
	}

	stream, _, err := s.fileRepo.ServeFile(physical)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	// Bound the decode even if the size raced past the stat.
	src, format, err := image.Decode(io.LimitReader(stream, maxThumbSource+1))
	if err != nil {
		return nil, domainerrors.NewValidationError("path", path, "unsupported image format ("+format+")")
	}
	bounds := src.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return nil, domainerrors.NewValidationError("path", path, "invalid image dimensions")
	}
	height := bounds.Dy() * width / bounds.Dx()
	if height < 1 {
		height = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: thumbQuality}); err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}
	thumb := &Thumbnail{Bytes: buf.Bytes(), Width: width, Height: height}
	s.store(key, thumb)
	return thumb, nil
}

// beginGen registers key as in-flight and returns a wait channel when another
// goroutine is already generating it (caller must wait), or nil when this
// caller owns the generation (pair with endGen).
func (s *ThumbnailService) beginGen(key thumbKey) chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ch, ok := s.inflight[key]; ok {
		return ch
	}
	ch := make(chan struct{})
	s.inflight[key] = ch
	return nil
}

func (s *ThumbnailService) endGen(key thumbKey) {
	s.mu.Lock()
	ch, ok := s.inflight[key]
	if ok {
		delete(s.inflight, key)
	}
	s.mu.Unlock()
	if ok {
		close(ch)
	}
}

func (s *ThumbnailService) cached(key thumbKey) *Thumbnail {
	s.mu.Lock()
	defer s.mu.Unlock()

	if el, ok := s.items[key]; ok {
		s.order.MoveToFront(el)
		return el.Value.(*thumbEntry).thumb
	}
	return nil
}

func (s *ThumbnailService) store(key thumbKey, thumb *Thumbnail) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if el, ok := s.items[key]; ok {
		s.order.MoveToFront(el)
		el.Value.(*thumbEntry).thumb = thumb
		return
	}
	s.items[key] = s.order.PushFront(&thumbEntry{key: key, thumb: thumb})
	for len(s.items) > maxThumbEntries {
		back := s.order.Back()
		if back == nil {
			break
		}
		s.order.Remove(back)
		delete(s.items, back.Value.(*thumbEntry).key)
	}
}
