package ports

import "github.com/EslamYasser-Dev/simple-file-share/domain/models"

// APIKeyRepository stores API keys issued by accounts. Implementations must
// never return or accept plaintext secrets — only models.APIKey records.
type APIKeyRepository interface {
	// Create persists a new key, returning an error when the id already
	// exists.
	Create(key models.APIKey) error
	// List returns an owner's keys newest first; empty owner lists every
	// key (system view).
	List(owner string) ([]models.APIKey, error)
	// Find returns the key with the given id, or domainerrors.ErrNotFound.
	Find(id string) (*models.APIKey, error)
	// Save replaces the stored record with the same id (used to update
	// LastUsedAt), or returns domainerrors.ErrNotFound.
	Save(key models.APIKey) error
	// Delete removes the key with the given id, or returns
	// domainerrors.ErrNotFound.
	Delete(id string) error
}
