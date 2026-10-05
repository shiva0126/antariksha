package corpus

import (
	"testing"
	"time"
)

func TestIndependentEmbeddingConfiguration(t *testing.T) {
	for _, k := range []string{"EMBEDDING_PROVIDER", "EMBEDDING_BASE_URL", "EMBEDDING_API_KEY", "EMBEDDING_MODEL", "OPENAI_API_KEY", "OPENAI_BASE_URL"} {
		t.Setenv(k, "")
	}
	if e, err := ConfiguredEmbedder(time.Second); err != nil || e != nil {
		t.Fatalf("unconfigured: %v %v", e, err)
	}
	t.Setenv("EMBEDDING_PROVIDER", "local")
	e, err := ConfiguredEmbedder(time.Second)
	if err != nil || e.Model() != LocalEmbeddingModel || e.(OpenAIEmbedder).APIKey != "" {
		t.Fatalf("local: %v %v", e, err)
	}
	for _, address := range []string{"https://remote.example/v1", "http://127.0.0.1.evil/v1", "http://user:pass@127.0.0.1/v1"} {
		t.Setenv("EMBEDDING_BASE_URL", address)
		if _, err := ConfiguredEmbedder(time.Second); err == nil {
			t.Fatal("accepted non-local URL", address)
		}
	}
	t.Setenv("EMBEDDING_BASE_URL", "")
	t.Setenv("EMBEDDING_PROVIDER", "openai")
	t.Setenv("EMBEDDING_API_KEY", "test-embedding-only")
	e, err = ConfiguredEmbedder(time.Second)
	if err != nil || e.Model() != "text-embedding-3-small" {
		t.Fatalf("embedding-only: %v %v", e, err)
	}
}
