package ports

// TextDocument is a file path paired with extractable text content used to
// populate the full-text index.
type TextDocument struct {
	Path    string
	Content string
}

// TextHit is a ranked full-text match for a document path.
type TextHit struct {
	Path  string
	Score float64
}

// TextIndex is an in-memory inverted index over extracted file contents.
// Implementations must be safe for concurrent use. Remove/RemovePrefix keep
// postings consistent so deleted files never surface from search.
type TextIndex interface {
	// Index replaces any prior content for path with content.
	Index(path, content string) error
	// Remove drops path from the index if present.
	Remove(path string) error
	// RemovePrefix drops path and every document under path/.
	RemovePrefix(pathPrefix string) error
	// Rebuild atomically replaces the index with docs.
	Rebuild(docs []TextDocument) error
	// Clear empties the index.
	Clear() error
	// Search returns paths containing every query term (AND), ranked by score.
	Search(query string, limit int) ([]TextHit, error)
}
