package reading

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/example/panchang/engine"
)

type promptLLM struct {
	prompt string
	body   string
	calls  int
}

// Regression coverage across births, NOT independent kundli-tool validation.
func TestFiveBirthReadingRegression(t *testing.T) {
	e := engine.New("../ephe")
	for _, date := range []string{"1985-01-15", "1990-07-28", "1996-05-14", "2000-02-29", "2004-11-02"} {
		t.Run(date, func(t *testing.T) {
			c, err := e.BirthChart(engine.ChartInput{Date: date, Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
			if err != nil {
				t.Fatal(err)
			}
			f, rules, err := NewService(DefaultCorpus, nil).BuildFacts(context.Background(), c, time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC))
			if err != nil {
				t.Fatal(err)
			}
			r := fallback(newInsight(context.Background(), DefaultCorpus, f, rules))
			if err := validate(r, f); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func (p *promptLLM) Complete(_ context.Context, s string) (string, error) {
	p.prompt = s
	p.calls++
	return p.body, nil
}

func TestMatchExplanationsAndFallback(t *testing.T) {
	f := referenceFacts(t)
	m, err := engine.MatchCharts(f.Chart, f.Chart)
	if err != nil {
		t.Fatal(err)
	}
	r := ChartMatchExplanation(m)
	if len(r.Factors) != 9 || !strings.Contains(r.Summary, "not a percentage") {
		t.Fatal(r)
	}
	for _, v := range r.Factors {
		if len(v.Explanation) < 80 || v.Source == "" || v.Question == "" || v.Result == "" {
			t.Fatal(v)
		}
	}
	llm := &promptLLM{body: `{"sections":[]}`}
	s := NewService(nil, llm)
	if out := s.ExplainMatch(context.Background(), r, false); out.AIStatus != "not_requested" || llm.calls != 0 {
		t.Fatal(out)
	}
	if out := s.ExplainMatch(context.Background(), r, true); out.AIStatus != "rejected" || out.Factors[0] != r.Factors[0] {
		t.Fatal(out)
	}
	if out := NewService(nil, nil).ExplainMatch(context.Background(), r, true); out.AIStatus != "unavailable" {
		t.Fatal(out)
	}
	for _, secret := range []string{"1996-05-14", "10:15", "Brahmin", "Shudra"} {
		if strings.Contains(llm.prompt, secret) {
			t.Fatalf("sensitive fact in prompt: %s", secret)
		}
	}
}

func TestMatchAIValidAndInvalidSections(t *testing.T) {
	base := ProfileMatchExplanation(map[string]string{"city": "same", "values": "different"}, 2)
	if len(base.Factors) != 8 || !strings.Contains(base.Factors[1].Evidence, "Not shared") {
		t.Fatal(base)
	}
	sections := []map[string]string{}
	for _, f := range base.Factors {
		sections = append(sections, map[string]string{"id": f.ID, "explanation": "Treat this topic as an invitation to understand each other. Ask what the information means in daily life, and leave room for either person to clarify or decline.", "question": "What would you like the other person to understand about this topic?"})
	}
	original := append([]MatchFactor(nil), base.Factors...)
	for _, test := range []string{"valid", "duplicate", "invented", "unsafe", "score"} {
		t.Run(test, func(t *testing.T) {
			b, _ := json.Marshal(map[string]any{"sections": sections})
			var obj map[string][]map[string]string
			json.Unmarshal(b, &obj)
			switch test {
			case "duplicate":
				obj["sections"][1]["id"] = obj["sections"][0]["id"]
			case "invented":
				obj["sections"][0]["id"] = "invented"
			case "unsafe":
				obj["sections"][0]["explanation"] += " You will divorce."
			case "score":
				obj["sections"][0]["explanation"] += " You score 99%."
			}
			b, _ = json.Marshal(obj)
			out := NewService(nil, fakeLLM{string(b)}).ExplainMatch(context.Background(), base, true)
			if (out.AIStatus == "generated") != (test == "valid") {
				t.Fatal(out.AIStatus)
			}
			for i, f := range out.Factors {
				if f.Evidence != base.Factors[i].Evidence || f.Source != base.Factors[i].Source || f.Result != base.Factors[i].Result {
					t.Fatal("model changed evidence")
				}
				if base.Factors[i] != original[i] {
					t.Fatal("AI call mutated reusable deterministic guide")
				}
			}
		})
	}
}

func TestPairSpecificMatchResults(t *testing.T) {
	for _, tc := range []struct {
		score float64
		want  string
	}{{0, "no points"}, {1.5, "part of"}, {3, "all available"}} {
		got := kootaResult(engine.Koota{Name: "Tara", Score: tc.score, Max: 3}, nil)
		if !strings.Contains(got, tc.want) {
			t.Fatal(got)
		}
	}
	nadi := engine.Koota{Name: "Nadi", Max: 8}
	if got := kootaResult(nadi, nil); !strings.Contains(got, "No exception is recorded") {
		t.Fatal(got)
	}
	note := "Some traditions cancel Nadi dosha when the nakshatra is shared but the padas differ."
	got := kootaResult(nadi, []string{note, "Bhakoot dosha is traditionally cancelled because the Moon-sign lords are the same or mutual friends."})
	if !strings.Contains(got, note) || !strings.Contains(got, "total are unchanged") || strings.Contains(got, "Bhakoot") {
		t.Fatal(got)
	}
	nadi.Score = 8
	if strings.Contains(kootaResult(nadi, []string{note}), "cancel") {
		t.Fatal("exception displayed without flag")
	}
	for _, tc := range []struct {
		boy, girl bool
		want      string
	}{{false, false, "Neither chart"}, {true, false, "Only one chart"}, {false, true, "Only one chart"}, {true, true, "Both charts"}} {
		r := ChartMatchExplanation(engine.Match{BoyMangal: tc.boy, GirlMangal: tc.girl})
		if !strings.Contains(r.Factors[0].Result, tc.want) {
			t.Fatal(r)
		}
	}
	for _, topic := range []string{"city", "timeline", "children", "relocation", "lifestyle", "values", "hobbies"} {
		for _, status := range []string{"same", "different", "", "unrecognized"} {
			got := profileResult(topic, status)
			if len(got) < 60 {
				t.Fatalf("%s/%s: %s", topic, status, got)
			}
			if (status == "" || status == "unrecognized") && !strings.Contains(got, "unknown") {
				t.Fatal(got)
			}
		}
	}
	for _, count := range []int{0, 2} {
		r := ProfileMatchExplanation(nil, count)
		got := r.Factors[len(r.Factors)-1].Result
		if (count == 0) != strings.Contains(got, "no shared") {
			t.Fatal(got)
		}
	}
}

func TestReadingFactGuards(t *testing.T) {
	f := referenceFacts(t)
	rules, _ := DefaultCorpus.Rules(context.Background(), f)
	for _, test := range []string{"baseline", "placement", "strength", "duplicate", "narrative_yoga", "dasha", "date", "house"} {
		t.Run(test, func(t *testing.T) {
			r := fallback(newInsight(context.Background(), DefaultCorpus, f, rules))
			switch test {
			case "placement":
				r.Grahas[0].Placement = "invented"
			case "strength":
				r.Yogas[0].Strength = "invented"
			case "duplicate":
				r.Yogas = append(r.Yogas, r.Yogas[0])
			case "narrative_yoga":
				found := map[string]bool{}
				for _, y := range f.Yogas {
					found[y.Name] = true
				}
				for _, name := range engine.YogaCatalog {
					if !found[name] {
						r.Summary += " " + name
						break
					}
				}
			case "dasha":
				r.Dashas.Current = "Sun–Moon period"
			case "date":
				r.Dashas.Current = "Until 1901-01-01"
			case "house":
				r.Summary += " Mars is in house 99."
			}
			if err := validate(r, f); (err == nil) != (test == "baseline") {
				t.Fatalf("%s: %v", test, err)
			}
		})
	}
}
