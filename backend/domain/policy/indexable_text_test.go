package policy

import (
	"strings"
	"testing"
)

func TestIsIndexableTextPath(t *testing.T) {
	yes := []string{
		"notes/todo.md", "data/export.csv", "app.go", "config.yaml",
		"README", "Dockerfile", "script.sh", "page.html", ".env",
	}
	no := []string{
		"photo.jpg", "archive.zip", "movie.mp4", "doc.pdf",
		"binary.bin", "lib.so", "randomfile",
	}
	for _, p := range yes {
		if !IsIndexableTextPath(p) {
			t.Errorf("IsIndexableTextPath(%q) = false, want true", p)
		}
	}
	for _, p := range no {
		if IsIndexableTextPath(p) {
			t.Errorf("IsIndexableTextPath(%q) = true, want false", p)
		}
	}
}

func TestReadIndexableText(t *testing.T) {
	text, ok := ReadIndexableText("a.txt", strings.NewReader("hello world"))
	if !ok || text != "hello world" {
		t.Fatalf("text=%q ok=%v", text, ok)
	}

	// Binary with NUL is rejected even with a text extension.
	_, ok = ReadIndexableText("a.txt", strings.NewReader("ab\x00cd"))
	if ok {
		t.Fatal("NUL bytes should not index")
	}

	// Non-indexable extension.
	_, ok = ReadIndexableText("a.jpg", strings.NewReader("not really"))
	if ok {
		t.Fatal("jpg should not index")
	}
}
