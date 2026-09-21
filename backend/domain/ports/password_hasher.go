package ports

// PasswordHasher encodes and verifies passwords. Implementations must store a
// per-password salt so every call yields a distinct encoded hash.
type PasswordHasher interface {
	// Hash encodes a plaintext password for storage.
	Hash(password string) (string, error)
	// Verify reports whether the plaintext password matches an encoded hash.
	Verify(password, encoded string) bool
}
