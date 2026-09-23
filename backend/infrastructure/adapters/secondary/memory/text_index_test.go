package memory

import (
	"strings"
	"testing"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

func TestTextIndexTokenize(t *testing.T) {
	got := Tokenize("Hello, World! 123 مرحبا")
	want := []string{"hello", "world", "123", "مرحبا"}
	if len(got) != len(want) {
		t.Fatalf("tokenize = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("token[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestTextIndexSearchFindsContent(t *testing.T) {
	x := NewTextIndex()
	if err := x.Index("docs/report.md", "quarterly revenue increased sharply"); err != nil {
		t.Fatal(err)
	}
	if err := x.Index("notes/todo.txt", "buy milk and eggs"); err != nil {
		t.Fatal(err)
	}

	hits, err := x.Search("revenue", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Path != "docs/report.md" {
		t.Fatalf("hits = %+v, want docs/report.md", hits)
	}
}

func TestTextIndexANDSemantics(t *testing.T) {
	x := NewTextIndex()
	_ = x.Index("a.txt", "alpha beta")
	_ = x.Index("b.txt", "alpha only")
	_ = x.Index("c.txt", "beta gamma")

	hits, err := x.Search("alpha beta", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Path != "a.txt" {
		t.Fatalf("AND hits = %+v, want only a.txt", hits)
	}
}

func TestTextIndexRemoveAndPrefix(t *testing.T) {
	x := NewTextIndex()
	_ = x.Index("docs/a.txt", "shared secret token")
	_ = x.Index("docs/nested/b.txt", "shared secret token")
	_ = x.Index("other/c.txt", "shared secret token")

	if err := x.RemovePrefix("docs"); err != nil {
		t.Fatal(err)
	}
	hits, err := x.Search("secret", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Path != "other/c.txt" {
		t.Fatalf("after RemovePrefix hits = %+v", hits)
	}

	if err := x.Remove("other/c.txt"); err != nil {
		t.Fatal(err)
	}
	hits, _ = x.Search("secret", 10)
	if len(hits) != 0 {
		t.Fatalf("after Remove hits = %+v, want empty", hits)
	}
}

func TestTextIndexReindexReplacesTerms(t *testing.T) {
	x := NewTextIndex()
	_ = x.Index("note.txt", "original content here")
	_ = x.Index("note.txt", "replacement words only")

	if hits, _ := x.Search("original", 10); len(hits) != 0 {
		t.Fatalf("stale term still hits: %+v", hits)
	}
	if hits, _ := x.Search("replacement", 10); len(hits) != 1 {
		t.Fatalf("new term missing: %+v", hits)
	}
}

func TestTextIndexRebuild(t *testing.T) {
	x := NewTextIndex()
	_ = x.Index("stale.txt", "should disappear")

	docs := []ports.TextDocument{
		{Path: "a.md", Content: "alpha bravo"},
		{Path: "b.md", Content: "charlie delta"},
	}
	if err := x.Rebuild(docs); err != nil {
		t.Fatal(err)
	}
	if hits, _ := x.Search("stale", 10); len(hits) != 0 {
		t.Fatalf("rebuild kept stale doc: %+v", hits)
	}
	if hits, _ := x.Search("alpha", 10); len(hits) != 1 || hits[0].Path != "a.md" {
		t.Fatalf("rebuild alpha hits = %+v", hits)
	}
}

func TestTextIndexRankingPrefersStrongerMatch(t *testing.T) {
	x := NewTextIndex()
	// Repeating the term raises TF in the first document.
	_ = x.Index("strong.txt", "needle needle needle filler")
	_ = x.Index("weak.txt", "needle filler words")

	hits, err := x.Search("needle", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("hits = %+v", hits)
	}
	if hits[0].Path != "strong.txt" {
		t.Fatalf("rank order = %+v, want strong.txt first", hits)
	}
	if !strings.Contains(hits[0].Path, "strong") {
		t.Fatal("unexpected top hit")
	}
}

func TestTextIndexEmptyQuery(t *testing.T) {
	x := NewTextIndex()
	_ = x.Index("a.txt", "hello")
	hits, err := x.Search("   ", 10)
	if err != nil || len(hits) != 0 {
		t.Fatalf("empty query hits=%v err=%v", hits, err)
	}
}
