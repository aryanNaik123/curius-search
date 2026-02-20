# Curius Search

Personal semantic search engine for your [Curius.app](https://curius.app) bookmarks. Uses local [Ollama](https://ollama.com) embeddings for offline, private vector search.

Inspired by [apollo](https://github.com/amirgamil/apollo) and its [curius-search variant](https://github.com/amirgamil/curius-search), replacing TF-IDF/inverted-index with semantic vector search.

## How it works

```
[Curius API] → [Go Indexer] → [Ollama embeddings] → [In-memory vector store + JSON file]
                                                              ↓
[Browser] → [Go HTTP Server] → [Embed query] → [Cosine similarity] → [Ranked results]
```

- Fetches all your Curius bookmarks via the public API
- Embeds each bookmark (title + URL + highlights + tags + snippet) using `nomic-embed-text` (768 dims)
- Stores vectors in memory with JSON file persistence
- **Hybrid search** — blends semantic cosine similarity (70%) with keyword matching (30%)
- **Find similar** — discover related bookmarks using a bookmark's own embedding
- **Search history** — recent queries saved locally with keyboard-navigable dropdown
- Incremental updates — only embeds new bookmarks on subsequent runs

## Prerequisites

- **Go** 1.22+
- **Ollama** with `nomic-embed-text`:
  ```
  ollama pull nomic-embed-text
  ```
- **Curius User ID**: Visit your Curius profile, open DevTools Network tab, find a request to `/api/users/{ID}/links` — the number is your ID

## Setup

```bash
cp .env.example .env
# Edit .env and set your CURIUS_USER_ID
```

## Usage

```bash
# Build and run (indexes then starts server)
make run

# Just build the index without starting the server
make index-only

# Force full re-index (discard existing embeddings)
make reindex

# Build only
make build
```

The server starts at **http://localhost:8990**. Search-as-you-type with 300ms debounce, keyboard friendly (Cmd/Ctrl+K to focus).

### API

| Endpoint | Method | Description |
|---|---|---|
| `/api/search?q={query}&limit={n}` | GET | Hybrid semantic + keyword search, returns ranked results |
| `/api/similar?id={id}&limit={n}` | GET | Find bookmarks similar to a given bookmark |
| `/api/status` | GET | Index stats and Ollama health |
| `/api/reindex` | POST | Trigger background re-index |

## Configuration

Set in `.env` or as environment variables:

| Variable | Default | Description |
|---|---|---|
| `CURIUS_USER_ID` | *(required)* | Your numeric Curius user ID |
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama API endpoint |
| `EMBED_MODEL` | `nomic-embed-text` | Ollama embedding model |
| `PORT` | `8990` | Server port |
| `DATA_DIR` | `data` | Directory for index persistence |

## Project structure

```
cmd/curius-search/main.go     # Entry point, CLI flags, indexing pipeline
internal/
  curius/                      # Curius API client (paginated fetching)
  embeddings/                  # Ollama embedding client
  index/                       # Vector store, cosine search, persistence
  search/                      # Search orchestration
  server/                      # HTTP server and handlers
static/                        # Frontend (vanilla HTML/JS/CSS)
```
