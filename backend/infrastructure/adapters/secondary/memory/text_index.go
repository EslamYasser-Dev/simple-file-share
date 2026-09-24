package memory

import (
	"math"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/EslamYasser-Dev/simple-file-share/domain/ports"
)

const defaultTextSearchLimit = 50

// bm25 parameters tuned for short file documents (names are scored separately
// by FileIndexRepository).
const (
	bm25K1 = 1.2
	bm25B  = 0.75
)

// TextIndex is a pure-Go inverted index over extracted file contents using
// BM25 ranking. Paths map to term-frequency maps; postings map terms to paths.
type TextIndex struct {
	mu sync.RWMutex
	// path -> term -> term frequency in that document
	tf map[string]map[string]int
	// term -> set of paths (implicit via tf keys; postings for intersection)
	postings map[string]map[string]struct{}
	// path -> document length in tokens
	lengths map[string]int
	// corpus stats
	totalDocs int
	totalLen  int
}

func NewTextIndex() *TextIndex {
	return &TextIndex{
		tf:       make(map[string]map[string]int),
		postings: make(map[string]map[string]struct{}),
		lengths:  make(map[string]int),
	}
}

// Tokenize lowercases and splits on non-letter/non-digit runes. Numbers and
// Unicode letters (including Arabic) are kept as single tokens.
func Tokenize(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func (x *TextIndex) Index(path, content string) error {
	path = normalizeIndexPath(path)
	if path == "" {
		return nil
	}
	tokens := Tokenize(content)
	if len(tokens) == 0 {
		// Empty/unindexable content: drop any prior entry so stale terms vanish.
		return x.Remove(path)
	}

	tf := make(map[string]int, len(tokens))
	for _, tok := range tokens {
		if tok == "" {
			continue
		}
		tf[tok]++
	}

	x.mu.Lock()
	defer x.mu.Unlock()
	x.removeLocked(path)

	doc := make(map[string]int, len(tf))
	length := 0
	for term, n := range tf {
		doc[term] = n
		length += n
		post, ok := x.postings[term]
		if !ok {
			post = make(map[string]struct{})
			x.postings[term] = post
		}
		post[path] = struct{}{}
	}
	x.tf[path] = doc
	x.lengths[path] = length
	x.totalDocs++
	x.totalLen += length
	return nil
}

func (x *TextIndex) Remove(path string) error {
	path = normalizeIndexPath(path)
	x.mu.Lock()
	defer x.mu.Unlock()
	x.removeLocked(path)
	return nil
}

func (x *TextIndex) removeLocked(path string) {
	doc, ok := x.tf[path]
	if !ok {
		return
	}
	for term := range doc {
		if post, exists := x.postings[term]; exists {
			delete(post, path)
			if len(post) == 0 {
				delete(x.postings, term)
			}
		}
	}
	x.totalDocs--
	x.totalLen -= x.lengths[path]
	delete(x.tf, path)
	delete(x.lengths, path)
}

func (x *TextIndex) RemovePrefix(pathPrefix string) error {
	pathPrefix = normalizeIndexPath(pathPrefix)
	if pathPrefix == "" {
		return x.Clear()
	}

	x.mu.Lock()
	defer x.mu.Unlock()
	for path := range x.tf {
		if path == pathPrefix || strings.HasPrefix(path, pathPrefix+"/") {
			x.removeLocked(path)
		}
	}
	return nil
}

func (x *TextIndex) Rebuild(docs []ports.TextDocument) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	x.tf = make(map[string]map[string]int, len(docs))
	x.postings = make(map[string]map[string]struct{})
	x.lengths = make(map[string]int)
	x.totalDocs = 0
	x.totalLen = 0

	for _, d := range docs {
		path := normalizeIndexPath(d.Path)
		if path == "" {
			continue
		}
		tokens := Tokenize(d.Content)
		if len(tokens) == 0 {
			continue
		}
		doc := make(map[string]int, len(tokens))
		length := 0
		for _, tok := range tokens {
			if tok == "" {
				continue
			}
			doc[tok]++
			length++
		}
		if length == 0 {
			continue
		}
		// Re-indexing the same path twice: drop the first copy first.
		if _, exists := x.tf[path]; exists {
			// inline remove without re-lock
			for term := range x.tf[path] {
				if post, ok := x.postings[term]; ok {
					delete(post, path)
					if len(post) == 0 {
						delete(x.postings, term)
					}
				}
			}
			x.totalDocs--
			x.totalLen -= x.lengths[path]
		}
		x.tf[path] = doc
		x.lengths[path] = length
		x.totalDocs++
		x.totalLen += length
		for term := range doc {
			post, ok := x.postings[term]
			if !ok {
				post = make(map[string]struct{})
				x.postings[term] = post
			}
			post[path] = struct{}{}
		}
	}
	return nil
}

func (x *TextIndex) Clear() error {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.tf = make(map[string]map[string]int)
	x.postings = make(map[string]map[string]struct{})
	x.lengths = make(map[string]int)
	x.totalDocs = 0
	x.totalLen = 0
	return nil
}

// Search ANDs query terms (every term must appear) and ranks with BM25.
func (x *TextIndex) Search(query string, limit int) ([]ports.TextHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = defaultTextSearchLimit
	}
	terms := uniqueNonEmpty(Tokenize(query))
	if len(terms) == 0 {
		return nil, nil
	}

	x.mu.RLock()
	defer x.mu.RUnlock()

	// Intersect postings across terms for AND semantics.
	var candidates map[string]struct{}
	for i, term := range terms {
		post := x.postings[term]
		if len(post) == 0 {
			return nil, nil
		}
		if i == 0 {
			candidates = make(map[string]struct{}, len(post))
			for p := range post {
				candidates[p] = struct{}{}
			}
			continue
		}
		for p := range candidates {
			if _, ok := post[p]; !ok {
				delete(candidates, p)
			}
		}
		if len(candidates) == 0 {
			return nil, nil
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	avgLen := 1.0
	if x.totalDocs > 0 {
		avgLen = float64(x.totalLen) / float64(x.totalDocs)
	}
	if avgLen <= 0 {
		avgLen = 1
	}

	hits := make([]ports.TextHit, 0, len(candidates))
	for path := range candidates {
		var score float64
		for _, term := range terms {
			f, ok := x.tf[path][term]
			if !ok {
				continue
			}
			df := float64(len(x.postings[term]))
			idf := math.Log(1 + (float64(x.totalDocs)-df+0.5)/(df+0.5))
			tf := float64(f)
			docLen := float64(x.lengths[path])
			denom := tf + bm25K1*(1-bm25B+bm25B*docLen/avgLen)
			score += idf * (tf * (bm25K1 + 1)) / denom
		}
		if score > 0 {
			hits = append(hits, ports.TextHit{Path: path, Score: score})
		}
	}

	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].Path < hits[j].Path
	})
	if len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, nil
}

func uniqueNonEmpty(tokens []string) []string {
	seen := make(map[string]struct{}, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		if tok == "" {
			continue
		}
		if _, ok := seen[tok]; ok {
			continue
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}
	return out
}

var _ ports.TextIndex = (*TextIndex)(nil)
