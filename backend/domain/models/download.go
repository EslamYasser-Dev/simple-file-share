package models

import "io"

// Download describes a resolved, streamable download response.
type Download struct {
	Stream      io.ReadCloser
	Filename    string
	ContentType string
}
