package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alexandrmotologa/poisonprobe/internal/config"
)

// Vectorizer defines the standard contract for embedding providers.
type Vectorizer interface {
	Embed(text string) ([]float64, error)
	EmbedBatch(texts []string) ([][]float64, error)
	Dimension() int
}

// RemoteVectorizer connects to OpenAI, Ollama, or HuggingFace TEI endpoints.
type RemoteVectorizer struct {
	endpoint   string
	model      string
	apiKey     string
	batchSize  int
	dimension  int
	httpClient *http.Client
}

// NewRemoteVectorizer creates a client for remote embedding APIs.
func NewRemoteVectorizer(cfg config.VectorizerConfig) *RemoteVectorizer {
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 64
	}
	dimension := cfg.Dimension
	if dimension <= 0 {
		dimension = 1536
	}

	return &RemoteVectorizer{
		endpoint:  cfg.Endpoint,
		model:     cfg.Model,
		apiKey:    cfg.APIKey,
		batchSize: batchSize,
		dimension: dimension,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Dimension returns configured vector dimension.
func (r *RemoteVectorizer) Dimension() int {
	return r.dimension
}

// Embed generates an embedding for a single text input.
func (r *RemoteVectorizer) Embed(text string) ([]float64, error) {
	batch, err := r.EmbedBatch([]string{text})
	if err != nil {
		return nil, err
	}
	if len(batch) == 0 {
		return nil, fmt.Errorf("remote API returned zero embeddings")
	}
	return batch[0], nil
}

type openAIEmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// EmbedBatch sends batched embedding requests to the remote endpoint.
func (r *RemoteVectorizer) EmbedBatch(texts []string) ([][]float64, error) {
	var allEmbeddings [][]float64

	for i := 0; i < len(texts); i += r.batchSize {
		end := i + r.batchSize
		if end > len(texts) {
			end = len(texts)
		}
		chunk := texts[i:end]

		reqBody := openAIEmbeddingRequest{
			Input: chunk,
			Model: r.model,
		}

		data, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}

		httpReq, err := http.NewRequest(http.MethodPost, r.endpoint, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		httpReq.Header.Set("Content-Type", "application/json")
		if r.apiKey != "" {
			httpReq.Header.Set("Authorization", "Bearer "+r.apiKey)
		}

		resp, err := r.httpClient.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("remote embedding request failed: %w", err)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("remote embedding API error HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		}

		var apiResp openAIEmbeddingResponse
		if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to parse embedding API response: %w", err)
		}

		if apiResp.Error != nil {
			return nil, fmt.Errorf("embedding API error: %s", apiResp.Error.Message)
		}

		for _, item := range apiResp.Data {
			allEmbeddings = append(allEmbeddings, item.Embedding)
		}
	}

	return allEmbeddings, nil
}

// CreateVectorizer returns the appropriate vectorizer based on configuration.
func CreateVectorizer(cfg config.VectorizerConfig) Vectorizer {
	if cfg.Mode == "remote" && cfg.Endpoint != "" {
		return NewRemoteVectorizer(cfg)
	}
	return NewOfflineVectorizer(cfg.Dimension)
}
