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

type fakeChat struct{ reply string }

func (f fakeChat) Complete(context.Context, string) (string, error) { return "", nil }
func (f fakeChat) Chat(_ context.Context, system, user string) (string, error) {
	if system != CompactSystem || !strings.Contains(user, "CHART\n") || !strings.Contains(user, "QUESTION: ") {
		return "", nil
	}
	return f.reply, nil
}
func (f fakeChat) ModelName() string { return "astrisk-local" }

func TestCompactAnswerUsesOnlyCheckedModelAnswers(t *testing.T) {
	f := referenceFacts(t)
	rules, _ := DefaultCorpus.Rules(context.Background(), f)
	q := "What does my chart say about my career?"
	good := "In short\nYour career part of the chart holds the Sun and Mars, so leadership and drive shape your work.\n\nWhat this means for you\n• " + strings.Repeat("The Sun is strong here and supports confidence at work. ", 6) + "\n• Try one new responsibility this year.\n\nFor reflection, not certainty."
	wrongSign := strings.Replace(good, "Try one", "Your Moon is in "+map[bool]string{true: "Mesha", false: "Vrishabha"}[f.Chart.Ascendant.Rashi != "Mesha" && newInsight(context.Background(), nil, f, nil).signs[newInsight(context.Background(), nil, f, nil).signIdx("moon")] != "Mesha"]+". Try one", 1)
	for _, c := range []struct {
		name, reply string
		useModel    bool
	}{
		{"good", good, true},
		{"no format", "Your career looks great. For reflection, not certainty.", false},
		{"fatal", strings.Replace(good, "Try one", "You will die young. Try one", 1), false},
		{"wrong sign", wrongSign, false},
		{"thinking stripped", "<think>planning</think>\n" + good, true},
	} {
		s := NewService(DefaultCorpus, fakeChat{c.reply})
		s.Compact = true
		a, err := s.Answer(context.Background(), f, rules, q, nil, ChatContext{})
		if err != nil {
			t.Fatal(err)
		}
		if got := a.Model == "astrisk-local"; got != c.useModel {
			t.Errorf("%s: model used = %v, want %v", c.name, got, c.useModel)
		}
		if c.useModel && (!strings.Contains(a.Answer, "\n\nChart details\n") || !strings.HasSuffix(a.Answer, "For reflection, not certainty.") || strings.Contains(a.Answer, "<think>")) {
			t.Errorf("%s: answer should keep chart details and the closing line: %q", c.name, a.Answer)
		}
	}
}
