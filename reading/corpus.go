package reading

import (
	"context"
	"strings"

	"github.com/example/panchang/engine"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemoryCorpus struct{ Entries []Rule }

type CompositeCorpus []Corpus

func (c CompositeCorpus) Rules(ctx context.Context, f engine.ChartFacts) ([]Rule, error) {
	seen := map[string]bool{}
	var out []Rule
	for _, source := range c {
		if source == nil {
			continue
		}
		rs, err := source.Rules(ctx, f)
		if err != nil {
			// The embedded corpus remains a safe fallback while migrations or
			// pgvector are unavailable; deterministic facts never depend on RAG.
			continue
		}
		for _, r := range rs {
			if !seen[r.Key] {
				seen[r.Key] = true
				out = append(out, r)
			}
		}
	}
	return out, nil
}

func (m MemoryCorpus) Rules(_ context.Context, f engine.ChartFacts) ([]Rule, error) {
	keys := map[string]bool{}
	for _, y := range f.Yogas {
		keys[strings.ToLower(strings.ReplaceAll(y.Name, " ", "-"))] = true
	}
	for id, d := range f.Dignities {
		keys[id+"_"+d.State] = true
	}
	out := []Rule{}
	for _, r := range m.Entries {
		if keys[r.Key] {
			out = append(out, r)
		}
	}
	return out, nil
}

// PostgresCorpus retrieves only rules whose keys are present in the deterministic
// facts object. It intentionally does not perform fuzzy matching for core facts.
type PostgresCorpus struct{ Pool *pgxpool.Pool }

func (p PostgresCorpus) Rules(ctx context.Context, f engine.ChartFacts) ([]Rule, error) {
	if p.Pool == nil {
		return nil, nil
	}
	keys := map[string]bool{}
	for _, y := range f.Yogas {
		keys[strings.ToLower(strings.ReplaceAll(y.Name, " ", "-"))] = true
	}
	for id, d := range f.Dignities {
		keys[id+"_"+d.State] = true
	}
	for _, g := range f.Chart.Grahas {
		house := (int(g.Longitude/30)-int(f.Chart.Ascendant.Longitude/30)+12)%12 + 1
		keys[g.ID+"_in_"+itoa(house)] = true
	}
	if len(keys) == 0 {
		return nil, nil
	}
	args := []any{"en"}
	clauses := make([]string, 0, len(keys))
	i := 2
	for k := range keys {
		clauses = append(clauses, "key=$"+itoa(i))
		args = append(args, k)
		i++
	}
	q := "SELECT key, COALESCE(title,''), body, COALESCE(source,'') FROM astro_corpus WHERE language=$1 AND key IN (" + strings.Join(placeholders(len(keys), 2), ",") + ")"
	rows, err := p.Pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Rule
	for rows.Next() {
		var r Rule
		if err := rows.Scan(&r.Key, &r.Title, &r.Body, &r.Source); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func placeholders(n, start int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = "$" + itoa(start+i)
	}
	return out
}
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 4)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

var DefaultCorpus = MemoryCorpus{Entries: []Rule{{"gajakesari", "Gajakesari", "Jupiter in a kendra from the Moon is traditionally read as a supportive combination for judgment and learning; its result depends on dignity and affliction.", "BPHS-inspired rule"}, {"budha-aditya", "Budha-Aditya", "Sun and Mercury in one sign is traditionally associated with intellect and communication. Combustion changes emphasis but does not erase the geometric fact.", "BPHS-inspired rule"}, {"ruchaka", "Ruchaka", "Mars in own or exaltation sign in a kendra is a Mahapurusha combination associated with initiative and courage.", "BPHS-inspired rule"}}}
