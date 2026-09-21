package dto

// ErrorResponse is the standard JSON error body returned by the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// MessageResponse is a simple JSON acknowledgement.
type MessageResponse struct {
	Message string `json:"message"`
}

// UpdateFileResponse acknowledges a file content update.
type UpdateFileResponse struct {
	Message string `json:"message"`
	Size    int64  `json:"size"`
}
