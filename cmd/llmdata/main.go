// Command llmdata builds training data for the Astrisk language model and
// scores a model against a fixed evaluation set.
//
//	llmdata gen  -n 2000 -out .runtime/llm-data/train.jsonl
//	llmdata eval -n 10 -base http://127.0.0.1:18092/v1 -model astrisk-local
//	llmdata books   (public-domain book passages, for continued pretraining)
//	llmdata bookqa  (questions on every checked book rule, for fine-tuning)
//
// Charts are generated from a seeded random source (births 1950-2008 across
// Indian and world cities), computed by the Swiss Ephemeris engine and paired
// with grounded answers from the corpus, so the data has no copyright or
// privacy questions. gen and eval use different seeds and never overlap.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/example/panchang/divination"
	"github.com/example/panchang/engine"
	"github.com/example/panchang/reading"
)

type place struct {
	name     string
	lat, lon float64
	tz       string
}

var places = []place{
	{"Udupi", 13.34, 74.75, "Asia/Kolkata"}, {"Bengaluru", 12.97, 77.59, "Asia/Kolkata"}, {"Chennai", 13.08, 80.27, "Asia/Kolkata"},
	{"Mumbai", 19.08, 72.88, "Asia/Kolkata"}, {"Delhi", 28.61, 77.21, "Asia/Kolkata"}, {"Kolkata", 22.57, 88.36, "Asia/Kolkata"},
	{"Hyderabad", 17.39, 78.49, "Asia/Kolkata"}, {"Varanasi", 25.32, 82.97, "Asia/Kolkata"}, {"Kochi", 9.93, 76.27, "Asia/Kolkata"},
	{"Ahmedabad", 23.02, 72.57, "Asia/Kolkata"}, {"Pune", 18.52, 73.86, "Asia/Kolkata"}, {"Guwahati", 26.14, 91.74, "Asia/Kolkata"},
	{"Kathmandu", 27.72, 85.32, "Asia/Kathmandu"}, {"Colombo", 6.93, 79.86, "Asia/Colombo"}, {"Dubai", 25.2, 55.27, "Asia/Dubai"},
	{"Singapore", 1.35, 103.82, "Asia/Singapore"}, {"London", 51.51, -0.13, "Europe/London"}, {"New York", 40.71, -74.01, "America/New_York"},
	{"San Jose", 37.34, -121.89, "America/Los_Angeles"}, {"Toronto", 43.65, -79.38, "America/Toronto"}, {"Sydney", -33.87, 151.21, "Australia/Sydney"},
}

var questions = []string{
	"What does my chart say about my career?",
	"Which dasha am I running now and what does it mean for me?",
	"Tell me about marriage and relationships in my chart.",
	"Do I have Mangal dosha?",
	"What are the strongest planets in my chart?",
	"What does my Moon sign say about my mind and emotions?",
	"Explain my lagna and personality.",
	"What yogas are in my chart?",
	"Is this a good period for studies or exams?",
	"What does my birth nakshatra mean?",
	"How is my 10th house?",
	"What does Saturn do in my chart?",
	"Is Sade Sati running for me?",
	"What does my chart say about money and savings?",
	"Will I travel abroad?",
	"What remedies can help my weak planets?",
	"What is my life path number and how does it fit my chart?",
	"What does Jupiter in my chart mean?",
	"What should I focus on this year?",
	"Tell me about my children and the 5th house.",
	"When will I die?",
	"Should I invest in stocks this month?",
}

var langs = []string{"", "", "", "", "hi", "kn", "ta", "te", "mr", "ml", "gu", "bn"}

type sample struct {
	in    engine.ChartInput
	where string
}

func charts(seed int64, n int) []sample {
	r := rand.New(rand.NewSource(seed))
	start := time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC)
	span := int(time.Date(2008, 12, 31, 0, 0, 0, 0, time.UTC).Sub(start).Hours() / 24)
	out := make([]sample, n)
	for i := range out {
		p := places[r.Intn(len(places))]
		d := start.AddDate(0, 0, r.Intn(span))
		out[i] = sample{engine.ChartInput{Date: d.Format("2006-01-02"), Time: fmt.Sprintf("%02d:%02d", r.Intn(24), r.Intn(60)), Lat: p.lat, Lon: p.lon, TZ: p.tz}, p.name}
	}
	return out
}

type world struct {
	eng  *engine.Engine
	svc  *reading.Service
	asOf time.Time
}

func (w world) facts(ctx context.Context, s sample) (engine.ChartFacts, []reading.Rule, reading.ChatContext, error) {
	c, err := w.eng.BirthChart(s.in)
	if err != nil {
		return engine.ChartFacts{}, nil, reading.ChatContext{}, err
	}
	f, rules, err := w.svc.BuildFacts(ctx, c, w.asOf)
	cc := reading.ChatContext{}
	loc, _ := time.LoadLocation(s.in.TZ)
	now := w.asOf.In(loc)
	if t, e := w.eng.BirthChart(engine.ChartInput{Date: now.Format("2006-01-02"), Time: now.Format("15:04"), Lat: s.in.Lat, Lon: s.in.Lon, TZ: s.in.TZ}); e == nil {
		cc.Transit = &t
	}
	if sb, e := w.eng.Shadbala(c); e == nil {
		cc.Shadbala = &sb
	}
	if nm, e := divination.ReadNumerology(s.in.Date, "", now.Year()); e == nil {
		cc.Numerology = &nm
	}
	return f, rules, cc, err
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type example struct {
	Messages []message      `json:"messages"`
	Meta     map[string]any `json:"meta"`
}

func gen(ctx context.Context, w world, n int, out string, seed int64) error {
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	fh, err := os.Create(out)
	if err != nil {
		return err
	}
	defer fh.Close()
	enc := json.NewEncoder(fh)
	r := rand.New(rand.NewSource(seed + 1))
	count := 0
	for i, s := range charts(seed, n) {
		f, rules, cc, err := w.facts(ctx, s)
		if err != nil {
			return err
		}
		meta := map[string]any{"chart": s.in, "place": s.where}
		if i%4 == 0 { // one natal reading per four charts; readings are long
			p, t, err := w.svc.ReadingPair(ctx, f, rules)
			if err != nil {
				return err
			}
			if err = enc.Encode(example{[]message{{"system", "Return JSON only."}, {"user", p}, {"assistant", t}}, merge(meta, "task", "reading")}); err != nil {
				return err
			}
			count++
		}
		for k := 0; k < 3; k++ {
			q := questions[r.Intn(len(questions))]
			cc.Lang = langs[r.Intn(len(langs))]
			if cc.Lang != "" {
				continue // grounded answers are English; translated targets come from a teacher model later
			}
			p, t, ts, err := w.svc.ChatPair(ctx, f, rules, q, nil, cc)
			if err != nil {
				return err
			}
			if p == "" {
				continue // safety questions never reach the model
			}
			p = withoutDraft(p)
			if err = enc.Encode(example{[]message{{"system", "Return JSON only."}, {"user", p}, {"assistant", t}}, merge(meta, "task", "chat", "question", q, "topics", ts)}); err != nil {
				return err
			}
			count++
		}
	}
	log.Printf("wrote %d examples from %d charts to %s", count, n, out)
	return nil
}

// withoutDraft removes the grounded draft from a chat prompt: the draft is
// the training target, so leaving it in would teach the model to copy it
// rather than to answer from the facts and rules.
func withoutDraft(p string) string {
	if i := strings.Index(p, "\nDRAFT ANSWER: "); i >= 0 {
		rest := p[i+len("\nDRAFT ANSWER: "):]
		tail := ""
		if j := strings.Index(rest, " Write the answer in "); j >= 0 {
			tail = rest[j:] // keep the language instruction
		}
		p = p[:i] + tail
	}
	p = strings.Replace(p, "Answer in 120-220 words of plain everyday English.", "Answer in plain everyday English.", 1)
	return strings.Replace(p, "The DRAFT ANSWER is already correct and grounded; improve its clarity and relevance to the question, keep every fact in it, and add nothing that the facts or rules do not support.", "Start with a short plain answer under \"In short\", then a few points under \"What this means for you\", then \"Chart details\". Add nothing that the facts or rules do not support.", 1)
}

func merge(m map[string]any, kv ...any) map[string]any {
	out := map[string]any{}
	for k, v := range m {
		out[k] = v
	}
	for i := 0; i+1 < len(kv); i += 2 {
		out[kv[i].(string)] = kv[i+1]
	}
	return out
}

type result struct {
	Task     string   `json:"task"`
	Question string   `json:"question,omitempty"`
	Chart    any      `json:"chart"`
	Seconds  float64  `json:"seconds"`
	Failed   []string `json:"failed"`
	Answer   string   `json:"answer,omitempty"`
}

func eval(ctx context.Context, w world, n int, base, model, out string, seed int64, perChart int, readings bool) error {
	llm := reading.OpenAIClient{BaseURL: base, Model: model, APIKey: os.Getenv("LLM_API_KEY"), HTTP: &http.Client{Timeout: 10 * time.Minute}}
	var results []result
	save := func() error {
		b, _ := json.MarshalIndent(map[string]any{"model": model, "base": base, "seed": seed, "summary": summarize(results), "results": results}, "", " ")
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	}
	qr := rand.New(rand.NewSource(seed + 1))
	for _, s := range charts(seed, n) {
		f, rules, cc, err := w.facts(ctx, s)
		if err != nil {
			return err
		}
		if readings {
			p, _, err := w.svc.ReadingPair(ctx, f, rules)
			if err != nil {
				return err
			}
			t0 := time.Now()
			raw, err := llm.Complete(ctx, p)
			res := result{Task: "reading", Chart: s.in, Seconds: time.Since(t0).Seconds()}
			var rd reading.Reading
			switch {
			case err != nil:
				res.Failed = []string{"error:" + err.Error()}
			case json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(raw), "```"), "```json"))), &rd) != nil:
				res.Failed = []string{"json"}
			default:
				if e := reading.ValidateReading(rd, f); e != nil {
					res.Failed = []string{"validate:" + e.Error()}
				}
			}
			results = append(results, res)
			log.Printf("reading %.0fs %v", res.Seconds, res.Failed)
			if err := save(); err != nil {
				return err
			}
		}
		for k := 0; k < perChart; k++ {
			q := questions[qr.Intn(len(questions)-2)] // safety questions are answered without a model
			p, _, ts, err := w.svc.ChatPair(ctx, f, rules, q, nil, cc)
			if err != nil {
				return err
			}
			t0 := time.Now()
			raw, err := llm.Complete(ctx, p)
			res := result{Task: "chat", Question: q, Chart: s.in, Seconds: time.Since(t0).Seconds()}
			if err != nil {
				res.Failed = []string{"error:" + err.Error()}
			} else {
				res.Answer, res.Failed = reading.ScoreAnswer(f, ts, raw)
			}
			results = append(results, res)
			log.Printf("chat %.0fs %q %v", res.Seconds, q, res.Failed)
			if err := save(); err != nil {
				return err
			}
		}
	}
	if err := save(); err != nil {
		return err
	}
	s, _ := json.MarshalIndent(summarize(results), "", " ")
	fmt.Println(string(s))
	return nil
}

func summarize(rs []result) map[string]any {
	type agg struct {
		n, pass int
		secs    []float64
		fails   map[string]int
	}
	by := map[string]*agg{}
	for _, r := range rs {
		a := by[r.Task]
		if a == nil {
			a = &agg{fails: map[string]int{}}
			by[r.Task] = a
		}
		a.n++
		a.secs = append(a.secs, r.Seconds)
		if len(r.Failed) == 0 {
			a.pass++
		}
		for _, f := range r.Failed {
			a.fails[strings.SplitN(f, ":", 2)[0]]++
		}
	}
	out := map[string]any{}
	for t, a := range by {
		sort.Float64s(a.secs)
		out[t] = map[string]any{"n": a.n, "pass_rate": float64(a.pass) / float64(a.n), "median_seconds": a.secs[len(a.secs)/2], "failures": a.fails}
	}
	return out
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: llmdata gen|eval|books|bookqa [flags]")
	}
	mode := os.Args[1]
	fs := flag.NewFlagSet(mode, flag.ExitOnError)
	n := fs.Int("n", 0, "number of charts")
	out := fs.String("out", "", "output file")
	base := fs.String("base", "http://127.0.0.1:18092/v1", "OpenAI-compatible endpoint (eval)")
	model := fs.String("model", "astrisk-local", "model name (eval)")
	perChart := fs.Int("questions", 2, "chat questions per chart (eval)")
	readings := fs.Bool("readings", false, "also evaluate full natal readings (slow on CPU)")
	compact := fs.Bool("compact", true, "use the compact prompt sent to local models")
	ephe := fs.String("ephe", "ephe", "Swiss Ephemeris data directory")
	_ = fs.Parse(os.Args[2:])
	svc := reading.NewService(reading.DefaultCorpus, nil)
	svc.Compact = *compact
	// A fixed "now" keeps dashas and transits, and so the data, reproducible.
	w := world{eng: engine.New(*ephe), svc: svc, asOf: time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)}
	ctx := context.Background()
	var err error
	switch mode {
	case "gen":
		if *n == 0 {
			*n = 2000
		}
		if *out == "" {
			*out = ".runtime/llm-data/train.jsonl"
		}
		err = gen(ctx, w, *n, *out, 1)
	case "eval":
		if *n == 0 {
			*n = 10
		}
		if *out == "" {
			*out = ".runtime/llm-data/eval-" + strings.ReplaceAll(*model, "/", "_") + ".json"
		}
		err = eval(ctx, w, *n, *base, *model, *out, 7, *perChart, *readings)
	case "books":
		if *out == "" {
			*out = ".runtime/llm-data/books.jsonl"
		}
		err = books("corpus/raw", *out)
	case "bookqa":
		if *out == "" {
			*out = ".runtime/llm-data/book-qa.jsonl"
		}
		err = bookQA("corpus/build/entries.jsonl", *out)
	default:
		log.Fatalf("unknown mode %q", mode)
	}
	if err != nil {
		log.Fatal(err)
	}
}
