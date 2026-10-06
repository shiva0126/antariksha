package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/example/panchang/reading"
)

func writeResults(out, model, base string, results []result) error {
	b, _ := json.MarshalIndent(map[string]any{"model": model, "base": base, "summary": summarize(results), "results": results}, "", " ")
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(out, b, 0o644)
}

// evalCompact scores a model on short chart questions from charts it never
// saw (eval seed), exactly as production asks a compact local model.
func evalCompact(ctx context.Context, w world, n int, base, model, out string, seed int64, perChart int) error {
	llm := reading.OpenAIClient{BaseURL: base, Model: model, HTTP: &http.Client{Timeout: 10 * time.Minute}}
	var results []result
	qr := rand.New(rand.NewSource(seed + 1))
	for _, s := range charts(seed, n) {
		f, rules, cc, err := w.facts(ctx, s)
		if err != nil {
			return err
		}
		for k := 0; k < perChart; k++ {
			q := questions[qr.Intn(len(questions)-2)]
			p, _, ok := w.svc.CompactPair(ctx, f, rules, q, cc)
			if !ok {
				continue
			}
			t0 := time.Now()
			raw, err := llm.Chat(ctx, reading.CompactSystem, p)
			res := result{Task: "compact", Question: q, Chart: s.in, Seconds: time.Since(t0).Seconds()}
			if err != nil {
				res.Failed = []string{"error:" + err.Error()}
			} else {
				res.Answer = strings.TrimSpace(regexp.MustCompile(`(?s)<think>.*?</think>`).ReplaceAllString(raw, ""))
				_, ts := classifyTopics(q)
				res.Failed = reading.ScoreText(f, ts, res.Answer)
				if !strings.HasPrefix(res.Answer, "In short") {
					res.Failed = append(res.Failed, "format")
				}
			}
			results = append(results, res)
			log.Printf("compact %.0fs %q %v", res.Seconds, q, res.Failed)
			if err := writeResults(out, model, base, results); err != nil {
				return err
			}
		}
	}
	return nil
}

func classifyTopics(q string) (string, []string) { return q, reading.Topics(q) }

var wordRx = regexp.MustCompile(`[a-z]+`)

func keyWords(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range wordRx.FindAllString(strings.ToLower(s), -1) {
		if len(w) > 3 {
			out[w] = true
		}
	}
	return out
}

// evalBook asks the book-benchmark questions (.runtime/llm-data/book-bench.jsonl)
// and passes an answer when it cites the right verse and carries at least
// half of the book's meaning (key words of the plain summary).
func evalBook(ctx context.Context, n int, base, model, out string) error {
	llm := reading.OpenAIClient{BaseURL: base, Model: model, HTTP: &http.Client{Timeout: 10 * time.Minute}}
	f, err := os.Open(".runtime/llm-data/book-bench.jsonl")
	if err != nil {
		return err
	}
	defer f.Close()
	var items []struct{ Token, Question, Ref, Plain string }
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var it struct{ Token, Question, Ref, Plain string }
		if json.Unmarshal(sc.Bytes(), &it) == nil {
			items = append(items, it)
		}
	}
	rand.New(rand.NewSource(3)).Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
	seen := map[string]bool{}
	var results []result
	const system = `You are Astrisk. Answer in plain language and cite the classical source. Return JSON only: {"answer": string}.`
	for _, it := range items {
		if len(results) >= n || seen[it.Question] {
			continue
		}
		seen[it.Question] = true
		t0 := time.Now()
		raw, err := llm.Chat(ctx, system, it.Question)
		res := result{Task: "book", Question: it.Question, Chart: it.Token, Seconds: time.Since(t0).Seconds()}
		if err != nil {
			res.Failed = []string{"error:" + err.Error()}
		} else {
			res.Answer = strings.TrimSpace(regexp.MustCompile(`(?s)<think>.*?</think>`).ReplaceAllString(raw, ""))
			if !strings.Contains(res.Answer, it.Ref) {
				res.Failed = append(res.Failed, "verse")
			}
			want, got, hit := keyWords(it.Plain), keyWords(res.Answer), 0
			for w := range want {
				if got[w] {
					hit++
				}
			}
			if len(want) > 0 && hit*2 < len(want) {
				res.Failed = append(res.Failed, "meaning")
			}
		}
		results = append(results, res)
		log.Printf("book %.0fs %q %v", res.Seconds, it.Question, res.Failed)
		if err := writeResults(out, model, base, results); err != nil {
			return err
		}
	}
	return nil
}
