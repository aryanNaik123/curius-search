package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	host       string
	model      string
	httpClient *http.Client
}

func NewClient(host, model string) *Client {
	return &Client{
		host:  host,
		model: model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Embed returns the embedding vector for the given text.
func (c *Client) Embed(text string) ([]float32, error) {
	req := embedRequest{
		Model: c.model,
		Input: text,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	resp, err := c.httpClient.Post(c.host+"/api/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama embed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama embed: status %d", resp.StatusCode)
	}

	var result embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}

	if len(result.Embeddings) == 0 || len(result.Embeddings[0]) == 0 {
		return nil, fmt.Errorf("ollama returned empty embeddings")
	}

	return result.Embeddings[0], nil
}

// IsHealthy checks if Ollama is reachable and the model is available.
func (c *Client) IsHealthy() bool {
	resp, err := c.httpClient.Get(c.host + "/api/tags")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
