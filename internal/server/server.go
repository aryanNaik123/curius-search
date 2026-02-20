package server

import (
	"log"
	"net/http"

	"github.com/aryannaik/curius-search/internal/embeddings"
	"github.com/aryannaik/curius-search/internal/index"
	"github.com/aryannaik/curius-search/internal/search"
)

func New(port string, staticDir string, store *index.Store, embedClient *embeddings.Client, reindexFn func()) *http.Server {
	searcher := search.NewSearcher(store, embedClient)
	handlers := NewHandlers(searcher, store, embedClient, reindexFn)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", handlers.HandleSearch)
	mux.HandleFunc("/api/similar", handlers.HandleSimilar)
	mux.HandleFunc("/api/status", handlers.HandleStatus)
	mux.HandleFunc("/api/reindex", handlers.HandleReindex)
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Server listening on http://localhost:%s", port)
	return srv
}
