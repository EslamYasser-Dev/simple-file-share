package ports

// StorageMover renames a physical path inside the storage root (used for
// username rename: users/<old> → users/<new>).
type StorageMover interface {
	// MovePath renames physical oldPath to newPath. Both are storage-root
	// relative physical paths.
	MovePath(oldPath, newPath string) error
}
