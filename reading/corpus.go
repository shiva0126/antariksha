package reading

import (
	"context"

	"github.com/example/panchang/corpus"
	"github.com/example/panchang/engine"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MemoryCorpus serves entries from memory by exact engine token. Only the
// Parashari namespace is ever consulted.
type MemoryCorpus struct {
	byKey map[engine.CorpusKey][]Rule
}

func NewMemoryCorpus(entries []corpus.Entry) MemoryCorpus {
	m := MemoryCorpus{byKey: map[engine.CorpusKey][]Rule{}}
	for _, e := range entries {
		if e.System != engine.SystemParashari {
			continue
		}
		k := engine.CorpusKey{DocType: e.DocType, Key: e.Key}
		m.byKey[k] = append(m.byKey[k], Rule{DocType: e.DocType, Key: e.Key, Title: e.Title, Body: e.Body, Source: e.Source, Ref: e.Ref})
	}
	return m
}

func (m MemoryCorpus) Rules(ctx context.Context, f engine.ChartFacts) ([]Rule, error) {
	return m.RulesFor(ctx, engine.CorpusKeys(f))
}

func (m MemoryCorpus) RulesFor(_ context.Context, keys []engine.CorpusKey) ([]Rule, error) {
	var out []Rule
	for _, k := range keys {
		out = append(out, m.byKey[k]...)
	}
	return out, nil
}

// PostgresCorpus retrieves every astro_corpus row whose (doc_type, key) is a
// token detected in the facts. It intentionally does not fuzzy-match core facts.
type PostgresCorpus struct{ Pool *pgxpool.Pool }

func (p PostgresCorpus) Rules(ctx context.Context, f engine.ChartFacts) ([]Rule, error) {
	return p.RulesFor(ctx, engine.CorpusKeys(f))
}

func (p PostgresCorpus) RulesFor(ctx context.Context, keys []engine.CorpusKey) ([]Rule, error) {
	if p.Pool == nil {
		return nil, nil
	}
	ps, err := corpus.Retrieve(ctx, p.Pool, keys, "en")
	if err != nil {
		return nil, err
	}
	out := make([]Rule, 0, len(ps))
	for _, x := range ps {
		out = append(out, Rule{DocType: x.DocType, Key: x.Key, Title: x.Title, Body: x.Body, Source: x.Source, Ref: x.Ref})
	}
	return out, nil
}

// CompositeCorpus consults sources in order. A token answered by an earlier
// source (which may return several passages for it) is not answered again by a
// later one, so the embedded corpus only fills gaps.
type CompositeCorpus []Corpus

func (c CompositeCorpus) Rules(ctx context.Context, f engine.ChartFacts) ([]Rule, error) {
	return c.RulesFor(ctx, engine.CorpusKeys(f))
}

func (c CompositeCorpus) RulesFor(ctx context.Context, keys []engine.CorpusKey) ([]Rule, error) {
	answered := map[engine.CorpusKey]bool{}
	var out []Rule
	for _, source := range c {
		if source == nil {
			continue
		}
		rs, err := source.RulesFor(ctx, keys)
		if err != nil {
			// The embedded corpus remains a safe fallback while migrations or
			// pgvector are unavailable; deterministic facts never depend on RAG.
			continue
		}
		fresh := map[engine.CorpusKey]bool{}
		for _, r := range rs {
			k := engine.CorpusKey{DocType: r.DocType, Key: r.Key}
			if answered[k] {
				continue
			}
			fresh[k] = true
			out = append(out, r)
		}
		for k := range fresh {
			answered[k] = true
		}
	}
	return out, nil
}

// DefaultCorpus is the embedded self-authored corpus: full engine coverage
// with no database, so readings are grounded even before ingestion runs.
var DefaultCorpus = func() MemoryCorpus {
	entries, err := corpus.Authored()
	if err != nil {
		panic("embedded corpus is invalid: " + err.Error())
	}
	return NewMemoryCorpus(entries)
}()
