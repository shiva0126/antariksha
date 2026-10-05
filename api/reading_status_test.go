package api

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestReadingStatusWithoutDatabaseDoesNotClaimLoadedBooks(t *testing.T) {
	s := NewServer(nil, NoCache{}, nil)
	w := httptest.NewRecorder()
	s.readingStatus(w, httptest.NewRequest("GET", "/api/reading/status", nil))
	var out struct {
		Database bool `json:"database_available"`
		LLM      bool `json:"llm_configured"`
		Sources  []struct {
			ID      string `json:"id"`
			Entries int    `json:"entries"`
			Rights  string `json:"rights"`
		} `json:"sources"`
	}
	if json.Unmarshal(w.Body.Bytes(), &out) != nil || w.Code != 200 || out.Database || out.LLM {
		t.Fatal(w.Body.String())
	}
	if len(out.Sources) < 10 {
		t.Fatal("source manifest missing")
	}
	for _, s := range out.Sources {
		if s.Entries != 0 {
			t.Fatal("source list mistaken for loaded corpus")
		}
		if s.ID == "lal-kitab-original-1939-1952" && s.Rights != "not_acquired" {
			t.Fatal("Lal Kitab falsely marked loaded")
		}
	}
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/reading/status", nil))
	if w.Code != 401 {
		t.Fatal("library endpoint bypasses login")
	}
}

func TestReadingStatusDatabaseCounts(t *testing.T) {
	dsn := os.Getenv("ACCOUNT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("isolated migrated database required")
	}
	p, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	s := NewServer(nil, PostgresCache{Pool: p}, nil)
	w := httptest.NewRecorder()
	s.readingStatus(w, httptest.NewRequest("GET", "/api/reading/status", nil))
	var out struct {
		Database bool `json:"database_available"`
		Entries  int  `json:"corpus_entries"`
		Sources  []struct {
			ID      string `json:"id"`
			Entries int    `json:"entries"`
		} `json:"sources"`
	}
	if json.Unmarshal(w.Body.Bytes(), &out) != nil || w.Code != 200 || !out.Database {
		t.Fatal(w.Body.String())
	}
	total := 0
	for _, s := range out.Sources {
		total += s.Entries
	}
	var actual int
	if err = p.QueryRow(context.Background(), `SELECT count(*) FROM astro_corpus`).Scan(&actual); err != nil {
		t.Fatal(err)
	}
	if total != actual || out.Entries != actual {
		t.Fatalf("aggregate %d source counts %d database %d", out.Entries, total, actual)
	}
}

func TestReadingStatusReportsLocalEmbeddingModel(t *testing.T) {
	t.Setenv("EMBEDDING_PROVIDER", "local")
	t.Setenv("EMBEDDING_BASE_URL", "")
	t.Setenv("EMBEDDING_MODEL", "")
	s := NewServer(nil, NoCache{}, nil)
	w := httptest.NewRecorder()
	s.readingStatus(w, httptest.NewRequest("GET", "/api/reading/status", nil))
	var out struct {
		Model string `json:"embedding_model"`
		Ready bool   `json:"semantic_ready"`
	}
	if json.Unmarshal(w.Body.Bytes(), &out) != nil || out.Model != "bge-small-en-v1.5-onnx-q-chunk400-pad1536-v1" || out.Ready {
		t.Fatal(w.Body.String()) // no database here, so never ready
	}
}
