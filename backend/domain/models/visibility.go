package models

import "time"

// VisibilityLevel is the per-file audience selector. It defaults to private:
// nothing is ever visible to others unless the owner opts in.
type VisibilityLevel string

const (
	// VisibilityPrivate restricts the file to its owner.
	VisibilityPrivate VisibilityLevel = "private"
	// VisibilityLink exposes the file to the owner's followers (timeline
	// feed entries, metadata, and media when streaming is allowed).
	VisibilityLink VisibilityLevel = "link"
	// VisibilityPublic exposes the file to every authenticated user.
	VisibilityPublic VisibilityLevel = "public"
)

// Valid reports whether the level is a known audience selector.
func (v VisibilityLevel) Valid() bool {
	switch v {
	case VisibilityPrivate, VisibilityLink, VisibilityPublic:
		return true
	}
	return false
}

// FileVisibility is the owner's per-file privacy setting. Absence of a
// record means VisibilityPrivate: the zero value is the safe default.
type FileVisibility struct {
	Owner string
	Path  string
	Level VisibilityLevel
	// AllowStream gates media playback (image view / video play) for
	// non-owners who may otherwise see the file. Owners always stream.
	AllowStream bool
	UpdatedAt   time.Time
}
