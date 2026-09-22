package corpus

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/example/panchang/engine"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CheckStatus string

const (
	Pass CheckStatus = "PASS"
	Fail CheckStatus = "FAIL"
	Skip CheckStatus = "SKIP"
)

type Check struct {
	Name   string      `json:"name"`
	Status CheckStatus `json:"status"`
	Detail string      `json:"detail"`
}

type Validation struct {
	Checks     []Check   `json:"checks"`
	SpotCheck  []Passage `json:"spot_check"`
	SmokeLines []string  `json:"smoke"`
}

func (v Validation) OK() bool {
	for _, c := range v.Checks {
		if c.Status == Fail {
			return false
		}
	}
	return true
}

func (v *Validation) add(name string, ok bool, detail string, args ...any) {
	st := Pass
	if !ok {
		st = Fail
	}
	v.Checks = append(v.Checks, Check{name, st, fmt.Sprintf(detail, args...)})
}

// garbleRatio is the share of tokens that still look like OCR damage: letters
// mixed with digits or symbols, or runs of consonants no English word has.
func garbleRatio(s string) float64 {
	toks := strings.Fields(s)
	if len(toks) == 0 {
		return 0
	}
	bad := 0
	for _, t := range toks {
		t = strings.Trim(t, ".,;:!?()[]'\"—-…")
		letters, other := 0, 0
		for _, r := range t {
			switch {
			case unicode.IsLetter(r):
				letters++
			case r == '-' || r == '\'' || r == '′' || r == '°' || r == '–':
			case unicode.IsDigit(r):
				if letters > 0 && !strings.HasSuffix(t, "th") && !strings.HasSuffix(t, "st") && !strings.HasSuffix(t, "nd") && !strings.HasSuffix(t, "rd") {
					other++
				}
			default:
				other++
			}
		}
		if other > 0 && letters > 0 {
			bad++
		}
	}
	return float64(bad) / float64(len(toks))
}

// Validate runs the ingestion checklist against the live database. samples are
// chart facts for the retrieval smoke test; emb may be nil, which skips the
// semantic checks.
func Validate(ctx context.Context, pool *pgxpool.Pool, samples []engine.ChartFacts, emb Embedder) (Validation, error) {
	var v Validation
	var n, total int

	if err := pool.QueryRow(ctx, `SELECT count(*) FROM astro_corpus`).Scan(&total); err != nil {
		return v, err
	}
	v.add("corpus loaded", total > 0, "%d rows", total)

	if err := pool.QueryRow(ctx, `SELECT count(*) FROM astro_corpus a LEFT JOIN corpus_sources s ON s.id=a.source_id WHERE s.id IS NULL OR btrim(s.rights_basis)=''`).Scan(&n); err != nil {
		return v, err
	}
	var sources int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM corpus_sources`).Scan(&sources)
	v.add("every row has a manifest rights decision", n == 0, "%d rows lack a recorded decision; %d manifest sources", n, sources)

	if err := pool.QueryRow(ctx, `SELECT count(*) FROM astro_corpus a JOIN corpus_sources s ON s.id=a.source_id WHERE s.rights NOT IN ('public_domain','self_authored','licensed') OR a.rights<>s.rights`).Scan(&n); err != nil {
		return v, err
	}
	blocked := 0
	for _, b := range blockedTranslators {
		var c int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM astro_corpus a JOIN corpus_sources s ON s.id=a.source_id WHERE lower(COALESCE(s.translator,'')||' '||COALESCE(s.edition,'')) LIKE '%'||$1||'%'`, b).Scan(&c)
		blocked += c
	}
	v.add("no copyrighted modern translation ingested", n == 0 && blocked == 0, "%d rows from ineligible sources, %d rows from blocked translators", n, blocked)

	if err := pool.QueryRow(ctx, `SELECT count(*) FROM astro_corpus WHERE btrim(source)='' OR btrim(source_id)=''`).Scan(&n); err != nil {
		return v, err
	}
	v.add("provenance on every row", n == 0, "%d rows without source", n)

	rows, err := pool.Query(ctx, `SELECT system,doc_type,key,COALESCE(title,''),body,source,source_id,ref,rights FROM astro_corpus WHERE rights='public_domain' ORDER BY md5(id::text) LIMIT 20`)
	if err != nil {
		return v, err
	}
	v.SpotCheck, err = scanPassages(rows, false)
	if err != nil {
		return v, err
	}
	worst, worstKey := 0.0, ""
	for _, p := range v.SpotCheck {
		if g := garbleRatio(p.Body); g > worst {
			worst, worstKey = g, p.Key
		}
	}
	if len(v.SpotCheck) == 0 {
		v.Checks = append(v.Checks, Check{"OCR spot-check (20 random classical rows)", Skip, "no public-domain rows loaded"})
	} else {
		v.add("OCR spot-check (20 random classical rows)", worst <= 0.02, "%d sampled; worst garble ratio %.3f (%s); samples listed for human review", len(v.SpotCheck), worst, worstKey)
	}

	covered := map[engine.CorpusKey]bool{}
	rows, err = pool.Query(ctx, `SELECT DISTINCT doc_type,key FROM astro_corpus WHERE system=$1`, engine.SystemParashari)
	if err != nil {
		return v, err
	}
	for rows.Next() {
		var k engine.CorpusKey
		if err = rows.Scan(&k.DocType, &k.Key); err != nil {
			rows.Close()
			return v, err
		}
		covered[k] = true
	}
	rows.Close()
	cov := Coverage(covered)
	missing := []string{}
	for i, m := range cov.Missing {
		if i == 10 {
			missing = append(missing, "…")
			break
		}
		missing = append(missing, m.String())
	}
	v.add("coverage map green", cov.Green(), "required %d/%d, optional %d/%d %s", cov.RequiredCovered, cov.Required, cov.OptionalCovered, cov.Optional, strings.Join(missing, " "))

	var jaimini, clash int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM astro_corpus WHERE system='jaimini'`).Scan(&jaimini)
	vocab := vocabulary()
	rows, err = pool.Query(ctx, `SELECT doc_type,key FROM astro_corpus WHERE system='jaimini'`)
	if err != nil {
		return v, err
	}
	for rows.Next() {
		var k engine.CorpusKey
		_ = rows.Scan(&k.DocType, &k.Key)
		if _, ok := vocab[k]; ok {
			clash++
		}
	}
	rows.Close()
	leaked := 0
	emptyKeys, offTopic, retrieved := 0, 0, 0
	for i, f := range samples {
		keys := engine.CorpusKeys(f)
		ps, err := Retrieve(ctx, pool, keys, "en")
		if err != nil {
			return v, err
		}
		got := map[engine.CorpusKey]int{}
		for _, p := range ps {
			if p.System != engine.SystemParashari {
				leaked++
			}
			k := engine.CorpusKey{DocType: p.DocType, Key: p.Key}
			got[k]++
			wanted := false
			for _, w := range keys {
				wanted = wanted || w == k
			}
			if !wanted {
				offTopic++
			}
		}
		for _, k := range keys {
			if got[k] == 0 && vocab[k].Required {
				emptyKeys++
			}
		}
		retrieved += len(ps)
		v.SmokeLines = append(v.SmokeLines, fmt.Sprintf("chart %d: %d tokens → %d passages (%d yogas)", i+1, len(keys), len(ps), len(f.Yogas)))
	}
	v.add("Parashari/Jaimini namespaced", clash == 0 && leaked == 0 && jaimini > 0, "%d jaimini rows, %d collide with engine tokens, %d leaked into Parashari retrieval", jaimini, clash, leaked)
	if len(samples) == 0 {
		v.Checks = append(v.Checks, Check{"retrieval smoke test (exact keys)", Skip, "no sample charts"})
	} else {
		v.add("retrieval smoke test (exact keys)", emptyKeys == 0 && offTopic == 0, "%d charts, %d passages; %d required tokens retrieved nothing, %d passages off-topic", len(samples), retrieved, emptyKeys, offTopic)
	}

	var unembedded int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM astro_corpus WHERE embedding IS NULL`).Scan(&unembedded)
	if emb == nil {
		v.Checks = append(v.Checks, Check{"embeddings present", Skip, fmt.Sprintf("no embedder configured; %d/%d rows unembedded", unembedded, total)})
		v.Checks = append(v.Checks, Check{"semantic smoke test (top-5 on-topic)", Skip, "no embedder configured"})
		return v, nil
	}
	var stale int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM astro_corpus WHERE embedding IS NOT NULL AND embedding_model IS DISTINCT FROM $1`, emb.Model()).Scan(&stale)
	v.add("embeddings present", unembedded == 0 && stale == 0, "%d unembedded, %d from another model (%s expected)", unembedded, stale, emb.Model())
	if unembedded == total {
		v.Checks = append(v.Checks, Check{"semantic smoke test (top-5 on-topic)", Skip, "no rows embedded yet; run: corpus load -embed"})
		return v, nil
	}

	// Semantic smoke test: for each detected yoga and placement, a plain-English
	// query should find a passage for that same token among the top 5.
	hits, tries := 0, 0
	var misses []string
	for _, f := range samples {
		for _, k := range engine.CorpusKeys(f) {
			if k.DocType != engine.DocYoga && k.DocType != engine.DocGrahaInHouse && k.DocType != engine.DocGrahaInSign {
				continue
			}
			tries++
			ps, err := Search(ctx, pool, emb, vocab[k].Label, engine.SystemParashari, 5)
			if err != nil {
				v.add("semantic smoke test (top-5 on-topic)", false, "embedding query failed: %v", err)
				return v, nil
			}
			found := false
			for _, p := range ps {
				found = found || (p.DocType == k.DocType && p.Key == k.Key)
			}
			if found {
				hits++
			} else if len(misses) < 8 {
				misses = append(misses, k.String())
			}
		}
	}
	rate := 0.0
	if tries > 0 {
		rate = float64(hits) / float64(tries)
	}
	v.add("semantic smoke test (top-5 on-topic)", tries > 0 && rate >= 0.8, "%d/%d queries (%.0f%%) found their own token in the top 5; misses: %s", hits, tries, rate*100, strings.Join(misses, " "))
	return v, nil
}
