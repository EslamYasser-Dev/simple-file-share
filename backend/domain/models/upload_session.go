package models

import "time"

// UploadSession tracks a resumable upload. Offset is the number of bytes
// already staged; when Offset == Size the session is ready to complete.
type UploadSession struct {
	ID          string
	Owner       string
	Fingerprint string
	Destination string
	Filename    string
	Size        int64
	Offset      int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   time.Time
}
