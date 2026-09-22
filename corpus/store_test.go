package corpus

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/example/panchang/engine"
	"github.com/jackc/pgx/v5/pgxpool"
)

// fakeEmbeddings serves an OpenAI-compatible /embeddings endpoint with
// deterministic hashed bag-of-words vectors, so the embed → pgvector → search
// path can be exercised without a paid API.
func fakeEmbeddings(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" || r.Header.Get("Authorization") != "Bearer test" {
			http.Error(w, `{"error":{"code":"bad","message":"bad request"}}`, 400)
			return
		}
		var in struct {
			Input []string `json:"input"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		type item struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		}
		out := struct {
			Data []item `json:"data"`
		}{}
		for i, s := range in.Input {
			v := make([]float32, EmbeddingDims)
			for _, tok := range strings.Fields(strings.ToLower(strings.NewReplacer(",", " ", ".", " ", ";", " ").Replace(s))) {
				h := fnv.New32a()
				h.Write([]byte(tok))
				v[h.Sum32()%EmbeddingDims]++
			}
			var n float64
			for _, x := range v {
				n += float64(x * x)
			}
			for j := range v {
				v[j] = float32(float64(v[j]) / math.Max(math.Sqrt(n), 1e-9))
			}
			out.Data = append(out.Data, item{i, v})
		}
		_ = json.NewEncoder(w).Encode(out)
	}))
}

func TestOpenAIEmbedderSurfacesAPIError(t *testing.T) {
	srv := fakeEmbeddings(t)
	defer srv.Close()
	_, err := OpenAIEmbedder{BaseURL: srv.URL, APIKey: "wrong", ModelName: "m"}.Embed(context.Background(), []string{"x"})
	if err == nil || !strings.Contains(err.Error(), "bad request") {
		t.Fatalf("err=%v", err)
	}
}

// TestPostgresSyncEmbedRetrieve needs a disposable database with migrations
// applied: CORPUS_TEST_DATABASE_URL=postgres://... go test ./corpus
func TestPostgresSyncEmbedRetrieve(t *testing.T) {
	dsn := os.Getenv("CORPUS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("CORPUS_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	srv := fakeEmbeddings(t)
	defer srv.Close()
	emb := OpenAIEmbedder{BaseURL: srv.URL, APIKey: "test", ModelName: "fake-bow-v2"}
	m, _ := LoadManifest()
	entries, _, err := Build(BuildOptions{RawDir: t.TempDir(), AllowMissingRaw: true})
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Sync(ctx, pool, m, entries, emb)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Unembedded != 0 || rep.Embedded+rep.Unchanged == 0 {
		t.Fatalf("sync %+v", rep)
	}
	// Editing one entry must re-embed only that row.
	entries[0].Body += " (revised)"
	rep, err = Sync(ctx, pool, m, entries, emb)
	if err != nil || rep.Upserted != 1 || rep.Embedded != 1 {
		t.Fatalf("resync %+v %v", rep, err)
	}
	ps, err := Retrieve(ctx, pool, []engine.CorpusKey{{DocType: "yoga", Key: "gajakesari"}, {DocType: "karaka", Key: "atmakaraka"}}, "en")
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 1 || ps[0].Key != "gajakesari" {
		t.Fatalf("retrieve returned %+v (jaimini must never answer)", ps)
	}
	hits, err := Search(ctx, pool, emb, "principal Mangala dosha placement partnership", engine.SystemParashari, 5)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, h := range hits {
		found = found || h.Key == "mars_in_7"
	}
	if !found {
		t.Fatalf("semantic search missed mars_in_7: %+v", hits)
	}
}
