package corpus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/panchang/engine"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Embedder turns entry text into vectors whose width matches astro_corpus.embedding.
type Embedder interface {
	Model() string
	Embed(context.Context, []string) ([][]float32, error)
}

const EmbeddingDims = 1536

// OpenAIEmbedder calls an OpenAI-compatible /embeddings endpoint.
type OpenAIEmbedder struct {
	BaseURL, APIKey, ModelName string
	HTTP                       *http.Client
}

func (o OpenAIEmbedder) Model() string { return o.ModelName }

func (o OpenAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	body, _ := json.Marshal(map[string]any{"model": o.ModelName, "input": texts, "dimensions": EmbeddingDims})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(o.BaseURL, "/")+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.APIKey)
	hc := o.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		var e struct {
			Error struct{ Code, Message string } `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return nil, fmt.Errorf("embeddings HTTP status %d: %s %s", resp.StatusCode, e.Error.Code, e.Error.Message)
	}
	var v struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	out := make([][]float32, len(texts))
	for _, d := range v.Data {
		if d.Index < 0 || d.Index >= len(out) || len(d.Embedding) != EmbeddingDims {
			return nil, fmt.Errorf("embeddings: bad item index %d dims %d", d.Index, len(d.Embedding))
		}
		out[d.Index] = d.Embedding
	}
	for i, e := range out {
		if e == nil {
			return nil, fmt.Errorf("embeddings: missing vector %d", i)
		}
	}
	return out, nil
}

func vectorLiteral(v []float32) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, x := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(x), 'f', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}

type SyncReport struct {
	Upserted, Unchanged, Deleted, Embedded, Unembedded int
}

// Sync makes the database match a build: the manifest is mirrored into
// corpus_sources, entries are upserted by identity, stale rows are removed, and
// rows whose content changed lose their embedding until re-embedded.
func Sync(ctx context.Context, pool *pgxpool.Pool, m Manifest, entries []Entry, emb Embedder) (SyncReport, error) {
	var rep SyncReport
	tx, err := pool.Begin(ctx)
	if err != nil {
		return rep, err
	}
	defer tx.Rollback(ctx)
	for _, s := range m.Sources {
		_, err = tx.Exec(ctx, `INSERT INTO corpus_sources(id,title,author,edition,translator,year,url,sha256,rights,rights_basis,system,updated_at)
VALUES($1,$2,$3,$4,$5,NULLIF($6,0),$7,$8,$9,$10,$11,now())
ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,author=EXCLUDED.author,edition=EXCLUDED.edition,translator=EXCLUDED.translator,year=EXCLUDED.year,url=EXCLUDED.url,sha256=EXCLUDED.sha256,rights=EXCLUDED.rights,rights_basis=EXCLUDED.rights_basis,system=EXCLUDED.system,updated_at=now()`,
			s.ID, s.Title, s.Author, s.Edition, s.Translator, s.Year, s.URL, s.SHA256, s.Rights, s.RightsBasis, s.System)
		if err != nil {
			return rep, fmt.Errorf("upsert source %s: %w", s.ID, err)
		}
	}
	_, err = tx.Exec(ctx, `CREATE TEMP TABLE build_ids(system TEXT, doc_type TEXT, key TEXT, language TEXT, source_id TEXT, ref TEXT) ON COMMIT DROP`)
	if err != nil {
		return rep, err
	}
	rows := make([][]any, len(entries))
	for i, e := range entries {
		rows[i] = []any{e.System, e.DocType, e.Key, e.Language, e.SourceID, e.Ref}
	}
	if _, err = tx.CopyFrom(ctx, pgx.Identifier{"build_ids"}, []string{"system", "doc_type", "key", "language", "source_id", "ref"}, pgx.CopyFromRows(rows)); err != nil {
		return rep, err
	}
	tag, err := tx.Exec(ctx, `DELETE FROM astro_corpus a WHERE NOT EXISTS (SELECT 1 FROM build_ids b WHERE (b.system,b.doc_type,b.key,b.language,b.source_id,b.ref)=(a.system,a.doc_type,a.key,a.language,a.source_id,a.ref))`)
	if err != nil {
		return rep, err
	}
	rep.Deleted = int(tag.RowsAffected())
	for _, e := range entries {
		tag, err = tx.Exec(ctx, `INSERT INTO astro_corpus(system,doc_type,key,language,title,body,source,source_id,ref,rights,content_hash)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT(system,doc_type,key,language,source_id,ref) DO UPDATE SET title=EXCLUDED.title,body=EXCLUDED.body,source=EXCLUDED.source,rights=EXCLUDED.rights,content_hash=EXCLUDED.content_hash,embedding=NULL,embedding_model=NULL,updated_at=now()
WHERE astro_corpus.content_hash IS DISTINCT FROM EXCLUDED.content_hash`,
			e.System, e.DocType, e.Key, e.Language, e.Title, e.Body, e.Source, e.SourceID, e.Ref, e.Rights, e.ContentHash())
		if err != nil {
			return rep, fmt.Errorf("upsert %s: %w", e.ID(), err)
		}
		if tag.RowsAffected() > 0 {
			rep.Upserted++
		} else {
			rep.Unchanged++
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return rep, err
	}
	if emb != nil {
		n, err := EmbedPending(ctx, pool, emb, 64)
		rep.Embedded = n
		if err != nil {
			return rep, err
		}
	}
	err = pool.QueryRow(ctx, `SELECT count(*) FROM astro_corpus WHERE embedding IS NULL`).Scan(&rep.Unembedded)
	return rep, err
}

// EmbedPending embeds rows with no vector, or a vector from a different model.
func EmbedPending(ctx context.Context, pool *pgxpool.Pool, emb Embedder, batch int) (int, error) {
	done := 0
	for {
		rows, err := pool.Query(ctx, `SELECT id, COALESCE(title,''), body FROM astro_corpus WHERE embedding IS NULL OR embedding_model IS DISTINCT FROM $1 ORDER BY id LIMIT $2`, emb.Model(), batch)
		if err != nil {
			return done, err
		}
		var ids []int64
		var texts []string
		for rows.Next() {
			var id int64
			var title, body string
			if err = rows.Scan(&id, &title, &body); err != nil {
				rows.Close()
				return done, err
			}
			ids = append(ids, id)
			texts = append(texts, Entry{Title: title, Body: body}.EmbeddingText())
		}
		rows.Close()
		if err = rows.Err(); err != nil {
			return done, err
		}
		if len(ids) == 0 {
			return done, nil
		}
		vecs, err := emb.Embed(ctx, texts)
		if err != nil {
			return done, err
		}
		for i, id := range ids {
			if _, err = pool.Exec(ctx, `UPDATE astro_corpus SET embedding=$1::vector, embedding_model=$2 WHERE id=$3`, vectorLiteral(vecs[i]), emb.Model(), id); err != nil {
				return done, err
			}
		}
		done += len(ids)
	}
}

// Passage is a retrieved corpus row.
type Passage struct {
	System   string  `json:"system"`
	DocType  string  `json:"doc_type"`
	Key      string  `json:"key"`
	Title    string  `json:"title"`
	Body     string  `json:"body"`
	Source   string  `json:"source"`
	SourceID string  `json:"source_id"`
	Ref      string  `json:"ref"`
	Rights   string  `json:"rights"`
	Distance float64 `json:"distance,omitempty"`
}

// Retrieve is the reading path: an exact fetch of every row for the detected
// tokens, restricted to the Parashari namespace. It never does fuzzy matching.
// Self-authored rows sort first per key, then classical passages by reference.
func Retrieve(ctx context.Context, pool *pgxpool.Pool, keys []engine.CorpusKey, language string) ([]Passage, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	docs := make([]string, len(keys))
	ks := make([]string, len(keys))
	for i, k := range keys {
		docs[i], ks[i] = k.DocType, k.Key
	}
	rows, err := pool.Query(ctx, `SELECT a.system,a.doc_type,a.key,COALESCE(a.title,''),a.body,a.source,a.source_id,a.ref,a.rights
FROM astro_corpus a JOIN unnest($1::text[],$2::text[]) AS k(doc_type,key) ON a.doc_type=k.doc_type AND a.key=k.key
WHERE a.system=$3 AND a.language=$4
ORDER BY a.doc_type,a.key,(a.rights<>'self_authored'),a.source_id,a.ref`, docs, ks, engine.SystemParashari, language)
	if err != nil {
		return nil, err
	}
	return scanPassages(rows, false)
}

// Search is semantic nearest-neighbour retrieval, used for smoke tests and
// exploration. system "" searches every namespace.
func Search(ctx context.Context, pool *pgxpool.Pool, emb Embedder, query, system string, k int) ([]Passage, error) {
	v, err := emb.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	rows, err := pool.Query(ctx, `SELECT system,doc_type,key,COALESCE(title,''),body,source,source_id,ref,rights, embedding <=> $1::vector
FROM astro_corpus WHERE embedding IS NOT NULL AND embedding_model=$2 AND ($3='' OR system=$3)
ORDER BY embedding <=> $1::vector LIMIT $4`, vectorLiteral(v[0]), emb.Model(), system, k)
	if err != nil {
		return nil, err
	}
	return scanPassages(rows, true)
}

func scanPassages(rows pgx.Rows, withDistance bool) ([]Passage, error) {
	defer rows.Close()
	var out []Passage
	for rows.Next() {
		var p Passage
		dest := []any{&p.System, &p.DocType, &p.Key, &p.Title, &p.Body, &p.Source, &p.SourceID, &p.Ref, &p.Rights}
		if withDistance {
			dest = append(dest, &p.Distance)
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
