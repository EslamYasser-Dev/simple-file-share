package dto

// Request structs shared by the HTTP handlers. Keeping them in the DTO package
// (with the response types) means every adapter ships the same wire contract.

// RegisterRequest is the body of POST /api/auth/register.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DeleteRequest is the body of DELETE /api/files.
type DeleteRequest struct {
	Path string `json:"path"`
}

// CreateDirectoryRequest is the body of POST /api/directories.
type CreateDirectoryRequest struct {
	Path string `json:"path"`
}

// UpdateContentRequest is the body of PUT /api/files/content.
type UpdateContentRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// RevokeShareRequest is the body of DELETE /api/shares.
type RevokeShareRequest struct {
	Token string `json:"token"`
}

// SetQuotaRequest is the body of PUT /api/admin/users/{username}/quota. Quota
// accepts a byte count or a human size ("2GB"); "0" or "unlimited" removes the
// cap.
type SetQuotaRequest struct {
	Quota string `json:"quota"`
}

// RestoreVersionRequest is the body of POST /api/files/version/restore.
type RestoreVersionRequest struct {
	Path string `json:"path"`
	N    int    `json:"n"`
}

// TokenRequest is the body of POST /api/auth/token.
type TokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshRequest is the body of POST /api/auth/refresh and /api/auth/revoke
// when Authorization is not used.
type RefreshRequest struct {
	AccessToken string `json:"accessToken"`
}

// CreateUploadSessionRequest is the body of POST /api/uploads. Path is the
// destination directory (virtual path); Filename is the base name. Size is
// the total byte length the client will send. Fingerprint is an optional
// client-computed identity (name|size|mtime|path) used to resume after reload.
type CreateUploadSessionRequest struct {
	Path        string `json:"path"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	Fingerprint string `json:"fingerprint,omitempty"`
}
