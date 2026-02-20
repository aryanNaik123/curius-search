package index

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aryannaik/curius-search/internal/curius"
)

const maxEmbedTextLen = 6000

type Store struct {
	mu      sync.RWMutex
	entries []IndexEntry
	idSet   map[int]bool
	path    string
}

func NewStore(dataDir string) *Store {
	return &Store{
		idSet: make(map[int]bool),
		path:  filepath.Join(dataDir, "index.json"),
	}
}

// LoadFromDisk loads the index from the JSON file. Returns nil if file doesn't exist.
func (s *Store) LoadFromDisk() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read index file: %w", err)
	}

	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return fmt.Errorf("decode index: %w", err)
	}

	s.entries = idx.Entries
	s.idSet = make(map[int]bool, len(idx.Entries))
	for _, e := range idx.Entries {
		s.idSet[e.ID] = true
	}

	return nil
}

// SaveToDisk persists the index to the JSON file.
func (s *Store) SaveToDisk() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	idx := Index{
		Entries:   s.entries,
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(idx)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("write index file: %w", err)
	}

	return nil
}

// Has returns true if a bookmark ID is already indexed.
func (s *Store) Has(id int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.idSet[id]
}

// Add adds an entry to the index.
func (s *Store) Add(entry IndexEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, entry)
	s.idSet[entry.ID] = true
}

// Clear removes all entries from the index.
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = nil
	s.idSet = make(map[int]bool)
}

// Count returns the number of indexed entries.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// UpdatedAt returns the last update time from disk, or zero time if unknown.
func (s *Store) UpdatedAt() time.Time {
	info, err := os.Stat(s.path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// SearchResult is a scored index entry from a search.
type SearchResult struct {
	Entry IndexEntry
	Score float32
}

// Search finds the top-k entries most similar to the query vector.
func (s *Store) Search(queryVec []float32, limit int) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.entries) == 0 {
		return nil
	}

	results := make([]SearchResult, 0, len(s.entries))
	for _, entry := range s.entries {
		score := cosineSimilarity(queryVec, entry.Embedding)
		results = append(results, SearchResult{Entry: entry, Score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results
}

// BuildEmbeddingText creates the text to embed for a bookmark.
func BuildEmbeddingText(link curius.Link) string {
	var b strings.Builder

	b.WriteString(link.Title)
	b.WriteString("\n")
	b.WriteString(link.URL)
	b.WriteString("\n")

	if link.Description != "" {
		b.WriteString(link.Description)
		b.WriteString("\n")
	}

	for _, h := range link.Highlights {
		b.WriteString(h)
		b.WriteString("\n")
	}

	for _, t := range link.Tags {
		b.WriteString(t.Name)
		b.WriteString(" ")
	}
	if len(link.Tags) > 0 {
		b.WriteString("\n")
	}

	if link.Content != "" {
		remaining := maxEmbedTextLen - b.Len()
		if remaining > 0 {
			content := link.Content
			if len(content) > remaining {
				content = content[:remaining]
			}
			b.WriteString(content)
		}
	}

	return b.String()
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}

	return float32(dot / denom)
}
