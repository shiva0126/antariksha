package main

import (
	"context"
	"log"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/example/panchang/api"
	corpusdata "github.com/example/panchang/corpus"
	"github.com/example/panchang/engine"
	"github.com/example/panchang/reading"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ephe := env("EPHE_PATH", "./ephe")
	addr := env("HTTP_ADDR", ":8080")
	e := engine.New(ephe)
	var cache api.Cache = api.NoCache{}
	corpus := reading.Corpus(reading.DefaultCorpus)
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		p, err := pgxpool.New(context.Background(), dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer p.Close()
		if err = p.Ping(context.Background()); err != nil {
			log.Fatal(err)
		}
		cache = api.PostgresCache{Pool: p}
		var embedder corpusdata.Embedder
		if os.Getenv("RAG_SEMANTIC_ENABLED") == "true" {
			embedder, err = corpusdata.ConfiguredEmbedder(10 * time.Second)
			if err != nil {
				log.Fatal(err)
			}
		}
		corpus = reading.CompositeCorpus{reading.PostgresCorpus{Pool: p, Embedder: embedder}, reading.DefaultCorpus}
	}
	var llm reading.LLM
	if base := os.Getenv("LLM_BASE_URL"); base != "" {
		// A local OpenAI-compatible server (llama.cpp, astrisk-llm.service): free, no
		// key. The timeout stays under Cloudflare's 100 s limit; on timeout the
		// grounded answer is served instead.
		timeout, _ := time.ParseDuration(env("LLM_TIMEOUT", "75s"))
		llm = reading.OpenAIClient{BaseURL: base, Model: env("LLM_MODEL", "local"), HTTP: &http.Client{Timeout: timeout}}
	} else if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		llm = reading.OpenAIClient{BaseURL: env("OPENAI_BASE_URL", "https://api.openai.com/v1"), APIKey: key, Model: env("OPENAI_MODEL", "gpt-4o-mini"), HTTP: &http.Client{Timeout: 90 * time.Second}}
	}
	readings := reading.NewService(corpus, llm)
	readings.Compact = os.Getenv("LLM_BASE_URL") != ""
	server := api.NewServerWithReading(e, cache, nil, readings)
	if u := os.Getenv("TRANSLATE_URL"); u != "" {
		// The local translation service (astrisk-translate.service): free, no
		// key. On timeout the English answer is served.
		timeout, _ := time.ParseDuration(env("TRANSLATE_TIMEOUT", "45s"))
		server.SetTranslator(api.NewTranslator(u, &http.Client{Timeout: timeout}))
	}
	go server.RunAlerts(context.Background())
	handler := server.Handler()
	if dir := os.Getenv("WEB_DIST"); dir != "" {
		mux := http.NewServeMux()
		mux.Handle("/api/", handler)
		mux.Handle("/healthz", handler)
		_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
		files := http.FileServer(http.Dir(dir))
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The service worker must update promptly; hashed assets may be cached.
			if r.URL.Path == "/sw.js" || r.URL.Path == "/" || r.URL.Path == "/index.html" {
				w.Header().Set("Cache-Control", "no-cache")
			}
			files.ServeHTTP(w, r)
		}))
		handler = mux
	}
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
