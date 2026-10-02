// Package audit persists the security audit trail as JSON Lines under the
// storage root's metadata directory (`.file-share/audit.jsonl`), which the
// search index already skips. The file rotates on size so an unattended
// deployment cannot fill the disk with audit records.
package audit

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

// JSONLLog is a size-rotated JSONL implementation of ports.AuditLog.
type JSONLLog struct {
	path     string
	maxBytes int64
	keep     int
	mu       sync.Mutex
}

var _ ports.AuditLog = (*JSONLLog)(nil)

// NewJSONLLog stores the trail at ROOT_DIR/.file-share/audit.jsonl.
// maxBytes controls rotation of the active file (values < 1KiB are raised to
// 1KiB); keep is the number of rotated generations retained (min 1).
func NewJSONLLog(rootDir string, maxBytes int64, keep int) *JSONLLog {
	if maxBytes < 1024 {
		maxBytes = 1024
	}
	if keep < 1 {
		keep = 1
	}
	return &JSONLLog{
		path:     filepath.Join(rootDir, ".file-share", "audit.jsonl"),
		maxBytes: maxBytes,
		keep:     keep,
	}
}

// Path returns the active file location.
func (l *JSONLLog) Path() string { return l.path }

// Record appends one JSON line, rotating first when the write would exceed
// the size budget.
func (l *JSONLLog) Record(entry ports.AuditEntry) error {
	if entry.At.IsZero() {
		entry.At = time.Now().UTC()
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}
	line = append(line, '\n')

	l.mu.Lock()
	defer l.mu.Unlock()

	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create audit dir: %w", err)
	}
	if info, err := os.Stat(l.path); err == nil && info.Size()+int64(len(line)) > l.maxBytes {
		if err := l.rotateLocked(); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open audit log: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("append audit entry: %w", err)
	}
	return nil
}

// Query returns a filtered page newest first. The cursor is an opaque
// base64 offset into the filtered newest-first stream, so pages stay stable
// while the file only grows (rotated-away tail simply stops matching).
func (l *JSONLLog) Query(query ports.AuditQuery) ([]ports.AuditEntry, string, error) {
	l.mu.Lock()
	files := l.filesLocked()
	l.mu.Unlock()

	limit := query.Limit
	if limit < 1 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	offset := 0
	if query.Cursor != "" {
		n, err := decodeCursor(query.Cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor")
		}
		offset = n
	}

	// Newest file first; lines inside one file are append-ordered (oldest
	// first), so reverse each file's matches to get one global newest-first
	// stream for stable offset cursors.
	var matched []ports.AuditEntry
	for _, name := range files {
		data, err := os.ReadFile(name)
		if err != nil {
			// A rotated file vanishing mid-query is not an error.
			continue
		}
		var fileEntries []ports.AuditEntry
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var entry ports.AuditEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				continue // skip a torn/corrupt line rather than fail the page
			}
			if query.Action != "" && entry.Action != query.Action {
				continue
			}
			if query.Actor != "" && entry.Actor != query.Actor {
				continue
			}
			fileEntries = append(fileEntries, entry)
		}
		for i, j := 0, len(fileEntries)-1; i < j; i, j = i+1, j-1 {
			fileEntries[i], fileEntries[j] = fileEntries[j], fileEntries[i]
		}
		matched = append(matched, fileEntries...)
	}
	if offset >= len(matched) {
		return []ports.AuditEntry{}, "", nil
	}
	end := offset + limit
	next := ""
	if end < len(matched) {
		next = encodeCursor(end)
	} else {
		end = len(matched)
	}
	page := matched[offset:end]
	return page, next, nil
}

// filesLocked lists the active file then rotated generations newest first.
func (l *JSONLLog) filesLocked() []string {
	var names []string
	if _, err := os.Stat(l.path); err == nil {
		names = append(names, l.path)
	}
	for i := 1; i <= l.keep; i++ {
		rotated := fmt.Sprintf("%s.%d", l.path, i)
		if _, err := os.Stat(rotated); err == nil {
			names = append(names, rotated)
		}
	}
	sort.SliceStable(names, func(a, b int) bool {
		// Newest first: active (version 0), then .1 (older), .2 (oldest).
		return versionOf(l.path, names[a]) < versionOf(l.path, names[b])
	})
	return names
}

// versionOf maps audit.jsonl -> 0, audit.jsonl.N -> N.
func versionOf(base, name string) int {
	if name == base {
		return 0
	}
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return 0
	}
	n, err := strconv.Atoi(name[i+1:])
	if err != nil {
		return 0
	}
	return n
}

// rotateLocked shifts audit.jsonl -> audit.jsonl.1 -> ... dropping the
// oldest generation beyond keep.
func (l *JSONLLog) rotateLocked() error {
	// Drop the oldest.
	oldest := fmt.Sprintf("%s.%d", l.path, l.keep)
	if err := os.Remove(oldest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("drop oldest audit file: %w", err)
	}
	// Shift the rest up: keep-1 -> keep, ... 1 -> 2.
	for i := l.keep - 1; i >= 1; i-- {
		from := fmt.Sprintf("%s.%d", l.path, i)
		to := fmt.Sprintf("%s.%d", l.path, i+1)
		if err := os.Rename(from, to); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("rotate audit file: %w", err)
		}
	}
	if err := os.Rename(l.path, l.path+".1"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("rotate active audit file: %w", err)
	}
	return nil
}

func encodeCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeCursor(cursor string) (int, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(string(raw))
	if err != nil || n < 0 {
		return 0, fmt.Errorf("bad offset")
	}
	return n, nil
}
