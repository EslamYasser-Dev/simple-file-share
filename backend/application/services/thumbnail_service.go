package services

import (
	"bytes"
	"container/list"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/gif"
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
	mod    int64
	width  int
}

type thumbEntry struct {
	key   thumbKey
	thumb *Thumbnail
}

// ThumbnailService renders small JPEG previews of images the caller may
// read. It mirrors the view path's authZ (path scoper) and keeps a bounded
// in-memory LRU so scrolling a photo folder does not re-decode every row.
type ThumbnailService struct {
	fileRepo ports.FileRepository
	scoper   ports.PathScoper

	mu    sync.Mutex
	items map[thumbKey]*list.Element
	order *list.List
}

func NewThumbnailService(fileRepo ports.FileRepository, scoper ports.PathScoper) *ThumbnailService {
	return &ThumbnailService{
		fileRepo: fileRepo,
		scoper:   scoper,
		items:    make(map[thumbKey]*list.Element),
		order:    list.New(),
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
