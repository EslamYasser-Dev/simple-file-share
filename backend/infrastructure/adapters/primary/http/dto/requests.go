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
