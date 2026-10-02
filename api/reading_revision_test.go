package api

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/example/panchang/engine"
	"github.com/example/panchang/reading"
)

type revisionCorpus struct{ body string }

func (c *revisionCorpus) Rules(context.Context, engine.ChartFacts) ([]reading.Rule, error) {
	return []reading.Rule{{DocType: "yoga", Key: "gajakesari", Body: c.body, Source: "project authored"}}, nil
}
func (c *revisionCorpus) RulesFor(ctx context.Context, _ []engine.CorpusKey) ([]reading.Rule, error) {
	return c.Rules(ctx, engine.ChartFacts{})
}

type revisionCache struct {
	NoCache
	keys []string
}

func (c *revisionCache) GetReading(_ context.Context, key string) (CachedReading, bool, error) {
	c.keys = append(c.keys, key)
	return CachedReading{}, false, nil
}

func TestReadingCacheInvalidatesAfterCorpusEdit(t *testing.T) {
	corpus := &revisionCorpus{body: "First explanation."}
	cache := &revisionCache{}
	s := NewServerWithReading(realEngine(t), cache, nil, reading.NewService(corpus, nil))
	request := func() {
		t.Helper()
		w := httptest.NewRecorder()
		s.readingHandler(w, httptest.NewRequest("GET", "/api/reading?"+birthQ+"&as_of=2026-10-02T00:00:00Z", nil))
		if w.Code != 200 {
			t.Fatal(w.Body.String())
		}
	}
	request()
	request()
	corpus.body = "A corrected explanation."
	request()
	if len(cache.keys) != 3 || cache.keys[0] != cache.keys[1] || cache.keys[0] == cache.keys[2] {
		t.Fatal("cache ignored changed corpus", cache.keys)
	}
}
