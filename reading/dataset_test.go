package reading

import (
	"context"
	"strings"
	"testing"
)

// The grounded answers are the training targets, so they must pass the same
// checks a model's answers are scored on.
func TestGroundedAnswersPassScoring(t *testing.T) {
	f := referenceFacts(t)
	rules, _ := DefaultCorpus.Rules(context.Background(), f)
	s := NewService(DefaultCorpus, nil)
	s.Compact = true
	for _, q := range []string{"Which dasha am I running now and what does it mean for me?", "What does my chart say about my career?", "Explain my lagna and personality.", "What does my Moon sign say about my mind and emotions?"} {
		prompt, target, ts, err := s.ChatPair(context.Background(), f, rules, q, nil, ChatContext{})
		if err != nil || prompt == "" {
			t.Fatalf("%q: %v", q, err)
		}
		if _, failed := ScoreAnswer(f, ts, target); len(failed) > 0 {
			// Grounded answers may be short; every other check must hold.
			for _, x := range failed {
				if !strings.HasPrefix(x, "length:") {
					t.Fatalf("%q failed %v: %s", q, failed, target)
				}
			}
		}
	}
}

func TestScoreAnswerCatchesInventedFacts(t *testing.T) {
	f := referenceFacts(t)
	wrong := "Mesha"
	if f.Chart.Ascendant.Rashi == wrong {
		wrong = "Vrishabha"
	}
	raw := `{"answer":"Your lagna is ` + wrong + `. In the Navamsha the ascendant is Mesha. You will die young. ` + strings.Repeat("Reflect on this. ", 30) + `"}`
	_, failed := ScoreAnswer(f, []string{"lagna"}, raw)
	got := strings.Join(failed, ",")
	for _, want := range []string{"disclaimer", "forbidden_prediction", "wrong_lagna_sign"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %s in %v", want, failed)
		}
	}
	if _, failed = ScoreAnswer(f, nil, "not json"); len(failed) != 1 || failed[0] != "json" {
		t.Fatal(failed)
	}
}
