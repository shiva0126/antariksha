package reading

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/example/panchang/engine"
)

type fakeLLM struct{ body string }

func (f fakeLLM) Complete(context.Context, string) (string, error) { return f.body, nil }
func sampleFacts() engine.ChartFacts {
	return engine.ChartFacts{Chart: engine.Chart{Ayanamsa: "lahiri", Ascendant: engine.Point{Rashi: "Karka"}, Grahas: []engine.Graha{{Name: "Chandra", Rashi: "Meena", Nakshatra: "Revati"}}}, Yogas: []engine.Yoga{{Name: "Gajakesari", Strength: "strong"}}, Vimshottari: engine.Dasha{Current: engine.DashaPeriod{Maha: "saturn", Antara: "mercury"}}}
}
func TestFallbackContainsExactlyDetectedYogas(t *testing.T) {
	r, _, e := NewService(nil, nil).Generate(context.Background(), sampleFacts(), nil)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Yogas) != 1 || r.Yogas[0].Name != "Gajakesari" {
		t.Fatalf("%+v", r.Yogas)
	}
}
func TestLLMUnknownYogaRejected(t *testing.T) {
	bad := `{"summary":"x","lagna_and_moon":"x","grahas":[],"yogas":[{"name":"Invented","meaning":"x","effect":"x","strength":"strong"}],"dashas":{},"themes":{},"disclaimer":"For reflection, not certainty; this is not medical, legal or financial advice."}`
	r, model, e := NewService(nil, fakeLLM{bad}).Generate(context.Background(), sampleFacts(), nil)
	if e != nil || model != FallbackModel {
		t.Fatalf("invalid LLM reading should fall back: %v %s", e, model)
	}
	for _, y := range r.Yogas {
		if y.Name == "Invented" {
			t.Fatal("unknown yoga accepted")
		}
	}
}
func TestChartHashStable(t *testing.T) {
	in := engine.ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"}
	if ChartHash(in, "en") != ChartHash(in, "en") {
		t.Fatal("hash not stable")
	}
	_ = time.Now()
}

func referenceFacts(t *testing.T) engine.ChartFacts {
	e := engine.New("../ephe")
	c, err := e.BirthChart(engine.ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	f, err := engine.Facts(c, time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestDefaultCorpusGroundsEveryDetectedToken(t *testing.T) {
	f := referenceFacts(t)
	rules, err := DefaultCorpus.Rules(context.Background(), f)
	if err != nil {
		t.Fatal(err)
	}
	got := map[engine.CorpusKey]bool{}
	for _, r := range rules {
		got[engine.CorpusKey{DocType: r.DocType, Key: r.Key}] = true
		if r.Source == "" {
			t.Fatalf("rule without provenance %+v", r)
		}
	}
	for _, k := range engine.CorpusKeys(f) {
		if strings.Contains(k.Key, "_pada") || strings.HasPrefix(k.Key, "friendly_") || strings.HasPrefix(k.Key, "enemy_") || strings.HasPrefix(k.Key, "neutral_") {
			continue // optional tokens
		}
		if !got[k] {
			t.Errorf("detected token %s retrieved nothing", k)
		}
	}
	for _, y := range f.Yogas {
		if !got[engine.CorpusKey{DocType: engine.DocYoga, Key: engine.Slug(y.Name)}] {
			t.Errorf("yoga %s ungrounded", y.Name)
		}
	}
}

type staticCorpus []Rule

func (s staticCorpus) Rules(context.Context, engine.ChartFacts) ([]Rule, error) { return s, nil }
func (s staticCorpus) RulesFor(context.Context, []engine.CorpusKey) ([]Rule, error) {
	return s, nil
}

func TestCompositeKeepsAllPassagesOfFirstSourcePerToken(t *testing.T) {
	a := staticCorpus{{DocType: "yoga", Key: "gajakesari", Source: "db1"}, {DocType: "yoga", Key: "gajakesari", Source: "db2"}}
	b := staticCorpus{{DocType: "yoga", Key: "gajakesari", Source: "mem"}, {DocType: "yoga", Key: "sasa", Source: "mem"}}
	rs, _ := CompositeCorpus{a, b}.Rules(context.Background(), engine.ChartFacts{})
	if len(rs) != 3 || rs[0].Source != "db1" || rs[1].Source != "db2" || rs[2].Key != "sasa" {
		t.Fatalf("%+v", rs)
	}
}

func TestPromptCarriesKeyedRules(t *testing.T) {
	p, err := promptFor(sampleFacts(), []Rule{{DocType: "yoga", Key: "gajakesari", Title: "Gajakesari Yoga", Body: "body", Source: "src [public_domain]"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p, "RULE yoga:gajakesari") || !strings.Contains(p, "never repeat their fatalistic") {
		t.Fatal(p)
	}
}

func TestGroundedReadingIsComplete(t *testing.T) {
	f := referenceFacts(t)
	rules, _ := DefaultCorpus.Rules(context.Background(), f)
	r, model, err := NewService(DefaultCorpus, nil).Generate(context.Background(), f, rules)
	if err != nil || model != FallbackModel {
		t.Fatalf("%v %s", err, model)
	}
	if len(r.Grahas) != 9 || len(r.Yogas) != len(f.Yogas) {
		t.Fatalf("grahas %d yogas %d", len(r.Grahas), len(r.Yogas))
	}
	for _, g := range r.Grahas {
		if len(g.Meaning) < 80 || strings.Contains(g.Meaning, "engine fact for reflective") {
			t.Fatalf("thin graha meaning %+v", g)
		}
	}
	for _, y := range r.Yogas {
		if len(y.Meaning) < 60 {
			t.Fatalf("thin yoga meaning %+v", y)
		}
	}
	if r.Themes.Career == "" || r.Themes.Relationships == "" || r.Themes.Strengths == "" || r.Themes.GrowthAreas == "" {
		t.Fatalf("empty themes %+v", r.Themes)
	}
	if !strings.Contains(r.Dashas.Current, "Venus mahadasha with Jupiter antardasha") {
		t.Fatalf("dasha %q", r.Dashas.Current)
	}
	if err := validate(r, f); err != nil {
		t.Fatalf("grounded reading fails the LLM validator: %v", err)
	}
}

type failingLLM struct{}

func (failingLLM) Complete(context.Context, string) (string, error) {
	return "", context.DeadlineExceeded
}

func TestLLMFailureFallsBackToGroundedReading(t *testing.T) {
	f := referenceFacts(t)
	r, model, err := NewService(DefaultCorpus, failingLLM{}).Generate(context.Background(), f, nil)
	if err != nil || model != FallbackModel || len(r.Grahas) != 9 {
		t.Fatalf("%v %s", err, model)
	}
}

func TestChatAnswersByTopic(t *testing.T) {
	f := referenceFacts(t)
	svc := NewService(DefaultCorpus, nil)
	ask := func(q string, h []ChatTurn) ChatAnswer {
		a, err := svc.Answer(context.Background(), f, nil, q, h, ChatContext{})
		if err != nil {
			t.Fatal(err)
		}
		if testing.Verbose() {
			t.Logf("Q: %s\n[%v]\n%s\n", q, a.Topics, a.Answer)
		}
		return a
	}
	if a := ask("How will my career go?", nil); !contains(a.Topics, "career") || !strings.Contains(a.Answer, "10th house is Mesha") || len(a.Sources) == 0 {
		t.Fatalf("career: %+v", a)
	}
	if a := ask("Do I have mangal dosha?", nil); !strings.Contains(a.Answer, "Mangal dosha") {
		t.Fatalf("mangal: %s", a.Answer)
	}
	if a := ask("What does my Saturn mean?", nil); !strings.Contains(a.Answer, "Saturn in Meena, 9th house") {
		t.Fatalf("saturn: %s", a.Answer)
	}
	if a := ask("tell me more", []ChatTurn{{"user", "what about marriage?"}, {"assistant", "..."}}); !contains(a.Topics, "marriage") {
		t.Fatalf("follow-up did not inherit topic: %+v", a.Topics)
	}
	if a := ask("When will I die?", nil); a.Answer != safetyNote {
		t.Fatalf("safety: %s", a.Answer)
	}
	if a := ask("What diet suits my Sunapha yoga?", nil); contains(a.Topics, "safety") || strings.Contains(a.Answer, "Sun in Mesha, 10th house") {
		t.Fatalf("whole-word matching: %v %s", a.Topics, a.Answer)
	}
	if a := ask("hello", nil); !contains(a.Topics, "overview") {
		t.Fatalf("overview: %+v", a.Topics)
	}
	if _, err := svc.Answer(context.Background(), f, nil, "   ", nil, ChatContext{}); err == nil {
		t.Fatal("empty question accepted")
	}
}
