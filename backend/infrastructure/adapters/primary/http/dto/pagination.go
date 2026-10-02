package dto

// Page is the shared list envelope for collection endpoints. NextCursor is an
// opaque continuation token and is absent on the last page.
type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
}
