package embeddings

// embedRequest is the request body for Ollama's /api/embed endpoint.
type embedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// embedResponse is the response from Ollama's /api/embed endpoint.
type embedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}
