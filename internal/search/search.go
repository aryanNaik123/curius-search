package search

import (
	"fmt"
	"strings"

	"github.com/aryannaik/curius-search/internal/embeddings"
	"github.com/aryannaik/curius-search/internal/index"
)

// Result is a search result returned to the frontend.
type Result struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Score       float32  `json:"score"`
	Snippet     string   `json:"snippet"`
	Tags        []string `json:"tags"`
	Highlights  []string `json:"highlights,omitempty"`
	CreatedAt   string   `json:"createdAt"`
}

type Searcher struct {
	store       *index.Store
	embedClient *embeddings.Client
}

func NewSearcher(store *index.Store, embedClient *embeddings.Client) *Searcher {
	return &Searcher{
		store:       store,
		embedClient: embedClient,
	}
}

// Search embeds the query and returns the top results.
func (s *Searcher) Search(query string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 20
	}

	queryVec, err := s.embedClient.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	hits := s.store.Search(queryVec, limit)

	results := make([]Result, 0, len(hits))
	for _, hit := range hits {
		r := Result{
			ID:         hit.Entry.ID,
			Title:      hit.Entry.Title,
			URL:        hit.Entry.URL,
			Score:      hit.Score,
			Tags:       hit.Entry.Tags,
			Highlights: hit.Entry.Highlights,
			CreatedAt:  hit.Entry.CreatedAt.Format("2006-01-02"),
		}

		r.Snippet = buildSnippet(hit.Entry)
		results = append(results, r)
	}

	return results, nil
}

func buildSnippet(entry index.IndexEntry) string {
	if entry.Description != "" {
		s := entry.Description
		if len(s) > 200 {
			s = s[:200] + "..."
		}
		return s
	}

	if len(entry.Highlights) > 0 {
		s := strings.Join(entry.Highlights, " ")
		if len(s) > 200 {
			s = s[:200] + "..."
		}
		return s
	}

	return ""
}
