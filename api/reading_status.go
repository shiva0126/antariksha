package api

import (
	"net/http"
	"os"

	"github.com/example/panchang/corpus"
)

// Only aggregate readiness and public provenance are exposed, never credentials
// or chart/user data. This is protected by the ordinary application login gate.
func (s *Server) readingStatus(w http.ResponseWriter, r *http.Request) {
	type source struct {
		Title  string `json:"title"`
		Rights string `json:"rights"`
		System string `json:"system"`
	}
	m, err := corpus.LoadManifest()
	if err != nil {
		problem(w, 500, err)
		return
	}
	sources := []source{}
	for _, v := range m.Sources {
		sources = append(sources, source{v.Title, v.Rights, v.System})
	}
	var rows, embedded, compatible int
	databaseOK := false
	if db, ok := s.cache.(PostgresCache); ok {
		model := os.Getenv("EMBEDDING_MODEL")
		if model == "" {
			model = "text-embedding-3-small"
		}
		databaseOK = db.Pool.QueryRow(r.Context(), `SELECT count(*),count(embedding),count(*) FILTER(WHERE embedding IS NOT NULL AND embedding_model=$1 AND language='en' AND system='parashari') FROM astro_corpus`, model).Scan(&rows, &embedded, &compatible) == nil
	}
	writeJSON(w, 200, map[string]any{"llm_configured": s.reading.LLM != nil, "database_available": databaseOK, "corpus_entries": rows, "embedded_entries": embedded, "compatible_embeddings": compatible, "semantic_enabled": os.Getenv("RAG_SEMANTIC_ENABLED") == "true", "semantic_ready": databaseOK && compatible > 0 && os.Getenv("RAG_SEMANTIC_ENABLED") == "true" && os.Getenv("OPENAI_API_KEY") != "", "sources": sources, "note": "RAG is retrieval, not training. Configuration does not guarantee provider availability or source quality. Rights-cleared passages and independent chart validation remain necessary."})
}
