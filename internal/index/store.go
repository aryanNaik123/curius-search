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

// GetByID returns the entry with the given ID, or nil if not found.
func (s *Store) GetByID(id int) *IndexEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.entries {
		if s.entries[i].ID == id {
			return &s.entries[i]
		}
	}
	return nil
}

// SearchResult is a scored index entry from a search.
type SearchResult struct {
	Entry IndexEntry
	Score float32
}

const (
	semanticWeight = 0.7
	keywordWeight  = 0.3
)

// Search finds the top-k entries using hybrid scoring (cosine similarity + keyword match).
func (s *Store) Search(queryVec []float32, query string, limit int) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.entries) == 0 {
		return nil
	}

	queryTerms := tokenize(query)

	results := make([]SearchResult, 0, len(s.entries))
	for _, entry := range s.entries {
		cosine := cosineSimilarity(queryVec, entry.Embedding)
		keyword := keywordScore(entry, queryTerms)
		score := float32(semanticWeight)*cosine + float32(keywordWeight)*keyword
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

// SearchByVector finds the top-k entries most similar to a vector (no keyword component).
func (s *Store) SearchByVector(queryVec []float32, limit int, excludeID int) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.entries) == 0 {
		return nil
	}

	results := make([]SearchResult, 0, len(s.entries))
	for _, entry := range s.entries {
		if entry.ID == excludeID {
			continue
		}
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

// keywordScore returns 0-1 based on what fraction of query terms appear in the entry's text fields.
func keywordScore(entry IndexEntry, queryTerms []string) float32 {
	if len(queryTerms) == 0 {
		return 0
	}

	text := strings.ToLower(entry.Title + " " + entry.Description + " " + strings.Join(entry.Tags, " ") + " " + strings.Join(entry.Highlights, " "))

	matched := 0
	for _, term := range queryTerms {
		if strings.Contains(text, term) {
			matched++
		}
	}

	return float32(matched) / float32(len(queryTerms))
}

// tokenize splits a query into lowercase terms, filtering short ones.
func tokenize(s string) []string {
	words := strings.Fields(strings.ToLower(s))
	terms := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) >= 2 {
			terms = append(terms, w)
		}
	}
	return terms
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
