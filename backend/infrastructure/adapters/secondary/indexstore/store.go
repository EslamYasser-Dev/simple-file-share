// Package indexstore persists the full-text index between boots so startup
// can skip re-extracting unchanged files. Format: 4-byte magic, uint32
// version, CRC32 of the gob payload, then the payload — corruption (torn
// writes, bit rot) is detected on load and reported so callers fall back to
// a full rebuild.
package indexstore

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

const (
	magic            = "FSIX"
	formatVersion    = uint32(1)
	snapshotFileName = "index.snapshot"
)

// Store keeps the snapshot at ROOT_DIR/.file-share/index.snapshot, beside the
// other metadata files (same owner-only directory).
type Store struct {
	path string
}

// NewStore returns a store rooted at the server storage root.
func NewStore(rootDir string) *Store {
	return &Store{path: filepath.Join(rootDir, ".file-share", snapshotFileName)}
}

// Load reads and validates the snapshot.
func (s *Store) Load() (*ports.IndexSnapshot, error) {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ports.ErrNoSnapshot
		}
		return nil, fmt.Errorf("read index snapshot: %w", err)
	}
	if len(raw) < 12 {
		return nil, fmt.Errorf("index snapshot truncated (%d bytes)", len(raw))
	}
	if string(raw[:4]) != magic {
		return nil, fmt.Errorf("index snapshot: bad magic %q", raw[:4])
	}
	version := binary.BigEndian.Uint32(raw[4:8])
	if version != formatVersion {
		return nil, fmt.Errorf("index snapshot: unsupported version %d", version)
	}
	payload := raw[12:]
	want := binary.BigEndian.Uint32(raw[8:12])
	if got := crc32.ChecksumIEEE(payload); got != want {
		return nil, fmt.Errorf("index snapshot: checksum mismatch (want %08x got %08x)", want, got)
	}
	var snap ports.IndexSnapshot
	if err := gob.NewDecoder(bytes.NewReader(payload)).Decode(&snap); err != nil {
		return nil, fmt.Errorf("index snapshot decode: %w", err)
	}
	return &snap, nil
}

// Save writes the snapshot atomically: temp file + fsync + rename, matching
// the crash-safety pattern used by the other metadata repositories.
func (s *Store) Save(snapshot *ports.IndexSnapshot) error {
	if snapshot == nil {
		return fmt.Errorf("index snapshot: nil")
	}
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(snapshot); err != nil {
		return fmt.Errorf("encode index snapshot: %w", err)
	}
	payload := buf.Bytes()

	raw := make([]byte, 12+len(payload))
	copy(raw[:4], magic)
	binary.BigEndian.PutUint32(raw[4:8], formatVersion)
	binary.BigEndian.PutUint32(raw[8:12], crc32.ChecksumIEEE(payload))
	copy(raw[12:], payload)

	dir := filepath.Dir(s.path)
	// The metadata directory is created lazily (fresh installs may never
	// have written users/roles before the first boot completes).
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".index-snapshot-*.tmp")
	if err != nil {
		return fmt.Errorf("create snapshot temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write snapshot: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("sync snapshot: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close snapshot: %w", err)
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("replace snapshot: %w", err)
	}
	return nil
}

// Path exposes the snapshot file location (tests and diagnostics).
func (s *Store) Path() string { return s.path }
