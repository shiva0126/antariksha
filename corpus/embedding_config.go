package corpus

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

const LocalEmbeddingModel = "bge-small-en-v1.5-onnx-q-chunk400-pad1536-v1"

// ConfiguredEmbedder deliberately separates retrieval from the paid reading
// model. Local embeddings never require (or enable) an OPENAI_API_KEY.
func ConfiguredEmbedder(timeout time.Duration) (Embedder, error) {
	provider := os.Getenv("EMBEDDING_PROVIDER")
	base, key, model := os.Getenv("EMBEDDING_BASE_URL"), os.Getenv("EMBEDDING_API_KEY"), os.Getenv("EMBEDDING_MODEL")
	switch provider {
	case "local":
		if base == "" {
			base = "http://127.0.0.1:18091/v1"
		}
		u, err := url.Parse(base)
		if err != nil || u.Scheme != "http" || u.User != nil || (u.Hostname() != "127.0.0.1" && u.Hostname() != "::1") {
			return nil, fmt.Errorf("local embedding endpoint must be loopback HTTP")
		}
		if model == "" {
			model = LocalEmbeddingModel
		}
	case "", "openai":
		if key == "" {
			key = os.Getenv("OPENAI_API_KEY")
		}
		if key == "" {
			return nil, nil
		}
		if base == "" {
			base = os.Getenv("OPENAI_BASE_URL")
		}
		if base == "" {
			base = "https://api.openai.com/v1"
		}
		if model == "" {
			model = "text-embedding-3-small"
		}
	default:
		return nil, fmt.Errorf("unknown EMBEDDING_PROVIDER %q", provider)
	}
	return OpenAIEmbedder{BaseURL: base, APIKey: key, ModelName: model, HTTP: &http.Client{Timeout: timeout}}, nil
}
