package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/aryannaik/curius-search/internal/embeddings"
	"github.com/aryannaik/curius-search/internal/index"
	"github.com/aryannaik/curius-search/internal/search"
)

type Handlers struct {
	searcher    *search.Searcher
	store       *index.Store
	embedClient *embeddings.Client
	reindexFn   func()
}

func NewHandlers(searcher *search.Searcher, store *index.Store, embedClient *embeddings.Client, reindexFn func()) *Handlers {
	return &Handlers{
		searcher:    searcher,
		store:       store,
		embedClient: embedClient,
		reindexFn:   reindexFn,
	}
}

func (h *Handlers) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing query parameter 'q'"})
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	results, err := h.searcher.Search(query, limit)
	if err != nil {
		log.Printf("Search error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"query":   query,
		"results": results,
		"total":   len(results),
	})
}

type statusResponse struct {
	IndexCount  int    `json:"indexCount"`
	UpdatedAt   string `json:"updatedAt"`
	OllamaOK    bool   `json:"ollamaOk"`
}

func (h *Handlers) HandleStatus(w http.ResponseWriter, r *http.Request) {
	updatedAt := h.store.UpdatedAt()
	updatedStr := ""
	if !updatedAt.IsZero() {
		updatedStr = updatedAt.Format("2006-01-02T15:04:05Z")
	}

	writeJSON(w, http.StatusOK, statusResponse{
		IndexCount: h.store.Count(),
		UpdatedAt:  updatedStr,
		OllamaOK:   h.embedClient.IsHealthy(),
	})
}

func (h *Handlers) HandleReindex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	go h.reindexFn()

	writeJSON(w, http.StatusOK, map[string]string{"status": "reindex started"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
