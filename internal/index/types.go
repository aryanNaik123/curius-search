package index

import "time"

// IndexEntry stores a bookmark with its embedding vector.
type IndexEntry struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Highlights  []string  `json:"highlights,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	Embedding   []float32 `json:"embedding"`
}

// Index is the top-level persisted structure.
type Index struct {
	Entries   []IndexEntry `json:"entries"`
	UpdatedAt time.Time    `json:"updatedAt"`
}
