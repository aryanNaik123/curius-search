package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/aryannaik/curius-search/internal/curius"
	"github.com/aryannaik/curius-search/internal/embeddings"
	"github.com/aryannaik/curius-search/internal/index"
	"github.com/aryannaik/curius-search/internal/server"
)

type config struct {
	CuriusUserID string
	OllamaHost   string
	EmbedModel   string
	Port         string
	DataDir      string
	StaticDir    string
}

func loadConfig() config {
	_ = godotenv.Load()

	cfg := config{
		CuriusUserID: os.Getenv("CURIUS_USER_ID"),
		OllamaHost:   envOrDefault("OLLAMA_HOST", "http://localhost:11434"),
		EmbedModel:   envOrDefault("EMBED_MODEL", "nomic-embed-text"),
		Port:         envOrDefault("PORT", "8990"),
		DataDir:      envOrDefault("DATA_DIR", "data"),
		StaticDir:    envOrDefault("STATIC_DIR", "static"),
	}

	if cfg.CuriusUserID == "" {
		log.Fatal("CURIUS_USER_ID is required. Set it in .env or as an environment variable.")
	}

	return cfg
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	reindexFlag := flag.Bool("reindex", false, "Force full re-index (discard existing embeddings)")
	indexOnlyFlag := flag.Bool("index-only", false, "Build index and exit (don't start server)")
	flag.Parse()

	cfg := loadConfig()

	embedClient := embeddings.NewClient(cfg.OllamaHost, cfg.EmbedModel)
	store := index.NewStore(cfg.DataDir)

	// Load existing index
	if err := store.LoadFromDisk(); err != nil {
		log.Printf("Warning: could not load existing index: %v", err)
	}

	if *reindexFlag {
		store.Clear()
		log.Println("Cleared existing index for full re-index")
	}

	// Run indexing
	runIndex(cfg, store, embedClient)

	if *indexOnlyFlag {
		log.Println("Index-only mode: exiting")
		return
	}

	// Start server
	reindexFn := func() {
		log.Println("Re-index triggered")
		runIndex(cfg, store, embedClient)
	}

	srv := server.New(cfg.Port, cfg.StaticDir, store, embedClient, reindexFn)

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	// Background periodic re-index every 24h
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			log.Println("Periodic re-index starting")
			runIndex(cfg, store, embedClient)
		}
	}()

	<-done
	ticker.Stop()
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Goodbye")
}

func runIndex(cfg config, store *index.Store, embedClient *embeddings.Client) {
	curiusClient := curius.NewClient(cfg.CuriusUserID)

	log.Println("Fetching bookmarks from Curius...")
	links, err := curiusClient.FetchAllLinks()
	if err != nil {
		log.Printf("Error fetching bookmarks: %v", err)
		return
	}
	log.Printf("Fetched %d bookmarks", len(links))

	// Find new bookmarks to embed
	var toEmbed []curius.Link
	for _, link := range links {
		if !store.Has(link.ID) {
			toEmbed = append(toEmbed, link)
		}
	}

	if len(toEmbed) == 0 {
		log.Println("Index is up to date, no new bookmarks to embed")
		return
	}

	log.Printf("Embedding %d new bookmarks...", len(toEmbed))

	for i, link := range toEmbed {
		text := index.BuildEmbeddingText(link)
		vec, err := embedClient.Embed(text)
		if err != nil {
			log.Printf("Error embedding bookmark %d (%s): %v", link.ID, link.Title, err)
			continue
		}

		tags := make([]string, len(link.Tags))
		for j, t := range link.Tags {
			tags[j] = t.Name
		}

		entry := index.IndexEntry{
			ID:          link.ID,
			Title:       link.Title,
			URL:         link.URL,
			Highlights:  link.Highlights,
			Tags:        tags,
			Description: link.Description,
			CreatedAt:   link.CreatedAt,
			Embedding:   vec,
		}

		store.Add(entry)

		if (i+1)%10 == 0 || i+1 == len(toEmbed) {
			log.Printf("  Embedded %d/%d", i+1, len(toEmbed))
		}
	}

	if err := store.SaveToDisk(); err != nil {
		log.Printf("Error saving index: %v", err)
		return
	}

	log.Printf("Index saved: %d total entries", store.Count())
}
