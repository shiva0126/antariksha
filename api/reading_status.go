package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/example/panchang/corpus"
)

// Only aggregate readiness and public provenance are exposed, never credentials
// or chart/user data. This is protected by the ordinary application login gate.
func (s *Server) readingStatus(w http.ResponseWriter, r *http.Request) {
	type source struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Rights  string `json:"rights"`
		System  string `json:"system"`
		Entries int    `json:"entries"`
		Note    string `json:"note"`
	}
	m, err := corpus.LoadManifest()
	if err != nil {
		problem(w, 500, err)
		return
	}
	sources := []source{}
	loaded := map[string]int{}
	var rows, embedded, compatible int
	databaseOK := false
	// The same configuration the API uses for retrieval: the free local model
	// (EMBEDDING_PROVIDER=local) or an OpenAI-compatible embedding key.
	emb, embErr := corpus.ConfiguredEmbedder(time.Second)
	model := ""
	if emb != nil && embErr == nil {
		model = emb.Model()
	}
	if db, ok := s.cache.(PostgresCache); ok {
		databaseOK = db.Pool.QueryRow(r.Context(), `SELECT count(*),count(embedding),count(*) FILTER(WHERE embedding IS NOT NULL AND embedding_model=$1 AND language='en' AND system='parashari') FROM astro_corpus`, model).Scan(&rows, &embedded, &compatible) == nil
		if databaseOK {
			result, err := db.Pool.Query(r.Context(), `SELECT source_id,count(*) FROM astro_corpus GROUP BY source_id`)
			if err != nil {
				problem(w, 503, fmt.Errorf("library status unavailable"))
				return
			}
			for result.Next() {
				var id string
				var n int
				if err = result.Scan(&id, &n); err != nil {
					result.Close()
					problem(w, 503, fmt.Errorf("library status unavailable"))
					return
				}
				loaded[id] = n
			}
			result.Close()
			if err = result.Err(); err != nil {
				problem(w, 503, fmt.Errorf("library status unavailable"))
				return
			}
		}
	}
	for _, v := range m.Sources {
		sources = append(sources, source{v.ID, v.Title, v.Rights, v.System, loaded[v.ID], v.Notes})
	}
	writeJSON(w, 200, map[string]any{"llm_configured": s.reading.LLM != nil, "database_available": databaseOK, "corpus_entries": rows, "embedded_entries": embedded, "compatible_embeddings": compatible, "semantic_enabled": os.Getenv("RAG_SEMANTIC_ENABLED") == "true", "semantic_ready": databaseOK && compatible > 0 && os.Getenv("RAG_SEMANTIC_ENABLED") == "true" && model != "", "embedding_model": model, "sources": sources, "note": "RAG is retrieval, not training. Configuration does not guarantee provider availability or source quality. Rights-cleared passages and independent chart validation remain necessary."})
}
