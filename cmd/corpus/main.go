// Command corpus is the one-time (and re-runnable) corpus ingestion job:
//
//	corpus acquire              download rights-cleared sources, verify sha256
//	corpus build                clean → segment → map → gate; writes corpus/build/
//	corpus load [-embed]        build, then sync into astro_corpus (+ embeddings)
//	corpus coverage [-db]       engine tokens with no corpus entry
//	corpus validate             run the ingestion checklist against the database
//	corpus search -q "..."      semantic nearest neighbours (needs embeddings)
//
// DATABASE_URL selects the database. OPENAI_API_KEY (with optional
// OPENAI_BASE_URL and EMBEDDING_MODEL) enables embeddings.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/example/panchang/corpus"
	"github.com/example/panchang/engine"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	cmd, args := os.Args[1], os.Args[2:]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	raw := fs.String("raw", "corpus/raw", "directory for downloaded public-domain sources")
	out := fs.String("out", "corpus/build", "directory for build artefacts")
	allowMissing := fs.Bool("allow-missing-raw", false, "build without sources that are not yet acquired")
	embed := fs.Bool("embed", false, "embed new or changed rows after loading")
	useDB := fs.Bool("db", false, "coverage: read the database instead of the build")
	ephe := fs.String("ephe", "ephe", "Swiss Ephemeris directory for validation charts")
	query := fs.String("q", "", "search: query text")
	k := fs.Int("k", 5, "search: number of results")
	system := fs.String("system", engine.SystemParashari, "search: namespace (empty for all)")
	_ = fs.Parse(args)
	ctx := context.Background()

	switch cmd {
	case "acquire":
		m := must(corpus.LoadManifest())
		hc := &http.Client{Timeout: 2 * time.Minute}
		for _, s := range m.Sources {
			if s.Rights != corpus.RightsPublicDomain || s.URL == "" {
				fmt.Printf("skip     %-28s %s\n", s.ID, s.Rights)
				continue
			}
			p, err := corpus.Acquire(ctx, s, *raw, hc)
			check(err)
			fmt.Printf("verified %-28s %s\n", s.ID, p)
		}
	case "build":
		entries, rep := build(*raw, *out, *allowMissing)
		cov := corpus.Coverage(corpus.CoveredKeys(entries))
		fmt.Printf("built %d entries: %d self-authored, %d public-domain, %d jaimini; coverage %d/%d required\n", len(entries), rep.Authored, rep.PublicDomain, rep.Jaimini, cov.RequiredCovered, cov.Required)
		for _, s := range rep.SkippedSources {
			fmt.Printf("warning: source %s not acquired; its passages were skipped\n", s)
		}
	case "load":
		entries, _ := build(*raw, *out, *allowMissing)
		pool := db(ctx)
		defer pool.Close()
		var emb corpus.Embedder
		if *embed {
			emb = embedder()
			if emb == nil {
				check(fmt.Errorf("-embed requires OPENAI_API_KEY"))
			}
		}
		rep, err := corpus.Sync(ctx, pool, must(corpus.LoadManifest()), entries, emb)
		check(err)
		fmt.Printf("synced: %d upserted, %d unchanged, %d stale removed, %d embedded, %d rows still unembedded\n", rep.Upserted, rep.Unchanged, rep.Deleted, rep.Embedded, rep.Unembedded)
	case "coverage":
		covered := map[engine.CorpusKey]bool{}
		if *useDB {
			pool := db(ctx)
			defer pool.Close()
			rows, err := pool.Query(ctx, `SELECT DISTINCT doc_type,key FROM astro_corpus WHERE system=$1`, engine.SystemParashari)
			check(err)
			for rows.Next() {
				var c engine.CorpusKey
				check(rows.Scan(&c.DocType, &c.Key))
				covered[c] = true
			}
			rows.Close()
		} else {
			entries, _ := build(*raw, "", *allowMissing)
			covered = corpus.CoveredKeys(entries)
		}
		cov := corpus.Coverage(covered)
		docs := make([]string, 0, len(cov.ByDocType))
		for d := range cov.ByDocType {
			docs = append(docs, d)
		}
		sort.Strings(docs)
		for _, d := range docs {
			c := cov.ByDocType[d]
			fmt.Printf("%-16s %4d/%-4d\n", d, c[0], c[1])
		}
		fmt.Printf("required %d/%d, optional (nakshatra padas, minor dignities) %d/%d\n", cov.RequiredCovered, cov.Required, cov.OptionalCovered, cov.Optional)
		for _, m := range cov.Missing {
			fmt.Println("missing", m.String(), "—", m.Label)
		}
		if !cov.Green() {
			os.Exit(1)
		}
	case "validate":
		pool := db(ctx)
		defer pool.Close()
		v, err := corpus.Validate(ctx, pool, sampleFacts(*ephe), embedder())
		check(err)
		fmt.Println("OCR spot-check sample (read these):")
		for _, p := range v.SpotCheck {
			fmt.Printf("  [%s %s] %s\n", p.Ref, p.Key, trunc(p.Body, 150))
		}
		fmt.Println("Retrieval smoke test:")
		for _, l := range v.SmokeLines {
			fmt.Println("  " + l)
		}
		fmt.Println("Checklist:")
		for _, c := range v.Checks {
			fmt.Printf("  [%s] %s — %s\n", c.Status, c.Name, c.Detail)
		}
		if !v.OK() {
			os.Exit(1)
		}
	case "search":
		emb := embedder()
		if emb == nil || *query == "" {
			check(fmt.Errorf("search needs OPENAI_API_KEY and -q"))
		}
		pool := db(ctx)
		defer pool.Close()
		ps, err := corpus.Search(ctx, pool, emb, *query, *system, *k)
		check(err)
		for _, p := range ps {
			fmt.Printf("%.3f %s:%s [%s %s]\n      %s\n", p.Distance, p.DocType, p.Key, p.SourceID, p.Ref, trunc(p.Body, 160))
		}
	default:
		usage()
	}
}

func build(raw, out string, allowMissing bool) ([]corpus.Entry, corpus.BuildReport) {
	audit := ""
	if out != "" {
		audit = filepath.Join(out, "audit")
	}
	entries, rep, err := corpus.Build(corpus.BuildOptions{RawDir: raw, AllowMissingRaw: allowMissing, AuditDir: audit})
	check(err)
	if out != "" {
		check(os.MkdirAll(out, 0o755))
		f, err := os.Create(filepath.Join(out, "entries.jsonl"))
		check(err)
		enc := json.NewEncoder(f)
		for _, e := range entries {
			check(enc.Encode(e))
		}
		check(f.Close())
		b, _ := json.MarshalIndent(rep, "", "  ")
		check(os.WriteFile(filepath.Join(out, "report.json"), b, 0o644))
	}
	return entries, rep
}

// sampleFacts computes the validation charts with the real engine so the smoke
// test exercises exactly what the reading backend will ask for.
func sampleFacts(ephe string) []engine.ChartFacts {
	e := engine.New(ephe)
	asOf := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	ins := []engine.ChartInput{
		{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"},
		{Date: "1984-11-02", Time: "04:40", Lat: 28.61, Lon: 77.21, TZ: "Asia/Kolkata"},
		{Date: "1971-02-19", Time: "21:05", Lat: 19.08, Lon: 72.88, TZ: "Asia/Kolkata"},
		{Date: "2003-08-27", Time: "13:30", Lat: 51.51, Lon: -0.13, TZ: "Europe/London"},
		{Date: "1990-12-31", Time: "23:50", Lat: 40.71, Lon: -74.01, TZ: "America/New_York"},
	}
	var out []engine.ChartFacts
	for _, in := range ins {
		c, err := e.BirthChart(in)
		check(err)
		f, err := engine.Facts(c, asOf)
		check(err)
		out = append(out, f)
	}
	return out
}

func embedder() corpus.Embedder {
	key := os.Getenv("OPENAI_API_KEY")
	if key == "" {
		return nil
	}
	return corpus.OpenAIEmbedder{BaseURL: env("OPENAI_BASE_URL", "https://api.openai.com/v1"), APIKey: key, ModelName: env("EMBEDDING_MODEL", "text-embedding-3-small"), HTTP: &http.Client{Timeout: 60 * time.Second}}
}

func db(ctx context.Context) *pgxpool.Pool {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		check(fmt.Errorf("DATABASE_URL is required"))
	}
	p, err := pgxpool.New(ctx, dsn)
	check(err)
	check(p.Ping(ctx))
	return p
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func must[T any](v T, err error) T {
	check(err)
	return v
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "corpus:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: corpus acquire|build|load|coverage|validate|search [flags]")
	os.Exit(2)
}
