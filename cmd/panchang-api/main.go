package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/example/panchang/api"
	"github.com/example/panchang/engine"
	"github.com/example/panchang/reading"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ephe := env("EPHE_PATH", "./ephe")
	addr := env("HTTP_ADDR", ":8080")
	e := engine.New(ephe)
	var cache api.Cache = api.NoCache{}
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
	}
	var llm reading.LLM
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		llm = reading.OpenAIClient{BaseURL: env("OPENAI_BASE_URL", "https://api.openai.com/v1"), APIKey: key, Model: env("OPENAI_MODEL", "gpt-4o-mini"), HTTP: &http.Client{Timeout: 90 * time.Second}}
	}
	handler := api.NewServerWithReading(e, cache, nil, reading.NewService(reading.DefaultCorpus, llm)).Handler()
	if dir := os.Getenv("WEB_DIST"); dir != "" {
		mux := http.NewServeMux()
		mux.Handle("/api/", handler)
		mux.Handle("/healthz", handler)
		mux.Handle("/", http.FileServer(http.Dir(dir)))
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
