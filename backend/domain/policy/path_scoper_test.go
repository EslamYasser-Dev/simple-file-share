package policy

import (
	"errors"
	"testing"

	domainerrors "github.com/EslamYasser-Dev/simple-file-share/domain/errors"
	"github.com/EslamYasser-Dev/simple-file-share/domain/models"
)

func TestPathScoperAdminIsIdentity(t *testing.T) {
	s := NewPathScoper()
	admin := &models.User{Username: "root", IsAdmin: true}

	for _, in := range []string{"", "docs/a.txt", "users/alice/x", "shared/y"} {
		got, err := s.ReadPath(admin, in)
		if err != nil {
			t.Fatalf("ReadPath(%q): %v", in, err)
		}
		if got != in {
			t.Errorf("ReadPath(%q) = %q, want identity", in, got)
		}
		if back := s.PhysicalToVirtual(admin, in); back != in {
			t.Errorf("PhysicalToVirtual(%q) = %q, want identity", in, back)
		}
	}
	if prefixes := s.ReadPrefixes(admin); prefixes != nil {
		t.Errorf("ReadPrefixes(admin) = %v, want nil (unrestricted)", prefixes)
	}
	if s.IsPrivateRoot(admin, "users/root") {
		t.Error("IsPrivateRoot(admin) = true, want false for system view")
	}
}

func TestPathScoperUserPrivateNamespace(t *testing.T) {
	s := NewPathScoper()
	user := &models.User{Username: "alice"}

	cases := map[string]string{
		"":            "users/alice",
		"notes.md":    "users/alice/notes.md",
		"docs/a.txt":  "users/alice/docs/a.txt",
		"shared":      "shared",
		"shared/x.md": "shared/x.md",
	}
	for in, want := range cases {
		got, err := s.ReadPath(user, in)
		if err != nil {
			t.Fatalf("ReadPath(%q): %v", in, err)
		}
		if got != want {
			t.Errorf("ReadPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPathScoperUserCannotReachOtherUsers(t *testing.T) {
	s := NewPathScoper()
	user := &models.User{Username: "alice"}

	for _, in := range []string{"users", "users/bob", "users/bob/secret.txt"} {
		_, err := s.ReadPath(user, in)
		var forbidden *domainerrors.ForbiddenError
		if !errors.As(err, &forbidden) {
			t.Errorf("ReadPath(%q) error = %v, want ForbiddenError", in, err)
		}
	}
}

func TestPathScoperSharedIsReadOnlyForUsers(t *testing.T) {
	s := NewPathScoper()
	user := &models.User{Username: "alice"}

	if _, err := s.WritePath(user, "shared"); err == nil {
		t.Error("WritePath(shared) succeeded, want forbidden")
	}
	if _, err := s.WritePath(user, "shared/report.pdf"); err == nil {
		t.Error("WritePath(shared/report.pdf) succeeded, want forbidden")
	}
	// Private writes still work.
	if got, err := s.WritePath(user, "draft.txt"); err != nil || got != "users/alice/draft.txt" {
		t.Errorf("WritePath(draft.txt) = %q, %v", got, err)
	}
}

func TestPathScoperPhysicalToVirtual(t *testing.T) {
	s := NewPathScoper()
	user := &models.User{Username: "alice"}

	cases := map[string]string{
		"users/alice":           "",
		"users/alice/a.txt":     "a.txt",
		"users/alice/docs/b.md": "docs/b.md",
		"shared/c.txt":          "shared/c.txt",
	}
	for in, want := range cases {
		if got := s.PhysicalToVirtual(user, in); got != want {
			t.Errorf("PhysicalToVirtual(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPathScoperReadPrefixesAndPrivateRoot(t *testing.T) {
	s := NewPathScoper()
	user := &models.User{Username: "alice"}

	prefixes := s.ReadPrefixes(user)
	if len(prefixes) != 2 || prefixes[0] != "users/alice" || prefixes[1] != SharedRoot {
		t.Errorf("ReadPrefixes = %v, want [users/alice shared]", prefixes)
	}
	if !s.IsPrivateRoot(user, "users/alice") {
		t.Error("IsPrivateRoot(users/alice) = false, want true")
	}
	if !s.IsPrivateRoot(user, "/users/alice/") {
		t.Error("IsPrivateRoot(/users/alice/) = false, want true")
	}
	if s.IsPrivateRoot(user, "users/alice/a.txt") {
		t.Error("IsPrivateRoot(users/alice/a.txt) = true, want false")
	}
}
