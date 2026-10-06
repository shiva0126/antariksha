package reading

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/example/panchang/engine"
)

// Training and evaluation hooks for the Astrisk language model. A pair is the
// exact prompt production sends plus the grounded answer composed by the
// engine and corpus, so a model trained on these pairs learns to write from
// the facts without inventing any.

// ReadingPair returns the natal-reading prompt and its grounded JSON answer.
func (s *Service) ReadingPair(ctx context.Context, f engine.ChartFacts, rules []Rule) (prompt, target string, err error) {
	if prompt, err = promptFor(f, rules, s.Compact); err != nil {
		return "", "", err
	}
	b, err := json.Marshal(fallback(newInsight(ctx, s.Corpus, f, rules)))
	return prompt, string(b), err
}

// ChatPair returns the chat prompt and its grounded JSON answer. Safety
// questions return an empty prompt: production never sends them to a model.
func (s *Service) ChatPair(ctx context.Context, f engine.ChartFacts, rules []Rule, question string, history []ChatTurn, cc ChatContext) (prompt, target string, topics []string, err error) {
	in := newInsight(ctx, s.Corpus, f, rules)
	grounded, ts := compose(in, question, history, cc)
	b, _ := json.Marshal(map[string]string{"answer": grounded})
	if contains(ts, "safety") {
		return "", string(b), ts, nil
	}
	for _, r := range cc.Related {
		in.use(r)
	}
	prompt, err = chatPrompt(f, in, question, history, grounded, s.Compact)
	return prompt + languageRule(cc.Lang), string(b), ts, err
}

// ValidateReading checks a model's reading against the engine facts: every
// graha placement copied exactly, exactly the detected yogas and strengths,
// and the narrative free of contradicted facts.
func ValidateReading(r Reading, f engine.ChartFacts) error { return validate(r, f) }

var westernSigns = []string{"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo", "Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces"}

var (
	sentenceSplit = regexp.MustCompile(`[.!?\n]+`)
	otherChart    = regexp.MustCompile(`(?i)\b(D\d+|navamsh?a|dashamsh?a|varga|divisional|transit|transiting|currently in|now in|sade sati)\b`)
	signClaim     = regexp.MustCompile(`(?i)\byour\s+(moon sign|moon|rashi|lagna|ascendant|rising sign)\s+(?:is\s+(?:in\s+)?|in\s+|falls in\s+)(\pL+)`)
)

// signIndex maps a Sanskrit or Western sign name to 0-11, or -1.
func signIndex(name string) int {
	for i, s := range engine.RashiNames() {
		if strings.EqualFold(s, name) || strings.EqualFold(westernSigns[i], name) {
			return i
		}
	}
	return -1
}

// forbidden are predictions the product never makes, in any answer.
var forbidden = regexp.MustCompile(`(?i)\b(you will (die|divorce|get divorced|fall ill|lose (all|everything))|death is|early death|guaranteed|certainly will|100% sure|will definitely)\b`)

// ScoreAnswer grades a chat answer on rules every answer must follow. It
// returns the failed checks; an empty list is a pass.
func ScoreAnswer(f engine.ChartFacts, topics []string, raw string) (answer string, failed []string) {
	var v struct {
		Answer string `json:"answer"`
	}
	raw = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(raw), "```"), "```json"))
	if json.Unmarshal([]byte(raw), &v) != nil || strings.TrimSpace(v.Answer) == "" {
		return "", []string{"json"}
	}
	a := v.Answer
	n := len(strings.Fields(a))
	if n < 60 || n > 320 {
		failed = append(failed, fmt.Sprintf("length:%d", n))
	}
	if !strings.Contains(strings.ToLower(a), "for reflection, not certainty") {
		failed = append(failed, "disclaimer")
	}
	if forbidden.MatchString(a) {
		failed = append(failed, "forbidden_prediction")
	}
	if contains(topics, "dasha") {
		maha := engine.GrahaEnglish(f.Vimshottari.Current.Maha)
		if !strings.Contains(strings.ToLower(a), strings.ToLower(maha)) {
			failed = append(failed, "dasha_lord")
		}
	}
	// A wrong sign for the Moon or the lagna is the most common invented fact.
	// Only direct claims count ("your Moon is in X"); sentences about
	// divisional charts or transits name other signs legitimately.
	in := newInsight(context.Background(), nil, f, nil)
	actual := map[string]int{"moon": in.signIdx("moon"), "lagna": in.lagnaSign()}
	for _, sentence := range sentenceSplit.Split(a, -1) {
		if otherChart.MatchString(sentence) {
			continue
		}
		for _, m := range signClaim.FindAllStringSubmatch(sentence, -1) {
			who := "lagna"
			if strings.EqualFold(m[1], "moon") || strings.EqualFold(m[1], "moon sign") || strings.EqualFold(m[1], "rashi") {
				who = "moon"
			}
			if i := signIndex(m[2]); i >= 0 && i != actual[who] && !contains(failed, "wrong_"+who+"_sign") {
				failed = append(failed, "wrong_"+who+"_sign")
			}
		}
	}
	return a, failed
}
