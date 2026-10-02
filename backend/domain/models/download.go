package models

import (
	"io"
	"time"
)

// Download describes a resolved, streamable download response.
type Download struct {
	Stream      io.ReadCloser
	Filename    string
	ContentType string
	// Inline reports whether the file is safe to stream inline to a browser
	// (content type inside the allowlist). Non-inline responses must be forced
	// to download so uploaded HTML/SVG cannot execute on the origin.
	Inline bool
	// Size and ModTime describe the underlying object when known (zero
	// ModTime = unknown). They drive ETag/Last-Modified revalidation; synthetic
	// streams (zip archives) leave them zero and stay unvalidated.
	Size    int64
	ModTime time.Time
}
