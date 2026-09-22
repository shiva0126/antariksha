package reading

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/example/panchang/engine"
)

// Rule is one retrieved corpus passage grounding a detected engine token.
type Rule struct {
	DocType string `json:"doc_type"`
	Key     string `json:"key"`
	Title   string `json:"title"`
	Body    string `json:"-"`
	Source  string `json:"source"`
	Ref     string `json:"ref,omitempty"`
}
type Corpus interface {
	// Rules returns passages for every token detected in the facts.
	Rules(context.Context, engine.ChartFacts) ([]Rule, error)
	// RulesFor returns passages for explicit tokens (chat needs houses and
	// placements beyond the chart's own detections).
	RulesFor(context.Context, []engine.CorpusKey) ([]Rule, error)
}
type LLM interface {
	Complete(context.Context, string) (string, error)
}
type Reading struct {
	Summary      string `json:"summary"`
	LagnaAndMoon string `json:"lagna_and_moon"`
	Grahas       []struct {
		Graha     string `json:"graha"`
		Placement string `json:"placement"`
		Meaning   string `json:"meaning"`
	} `json:"grahas"`
	Yogas []struct {
		Name     string `json:"name"`
		Meaning  string `json:"meaning"`
		Effect   string `json:"effect"`
		Strength string `json:"strength"`
	} `json:"yogas"`
	Dashas struct {
		Current  string `json:"current"`
		Upcoming string `json:"upcoming"`
	} `json:"dashas"`
	Themes struct {
		Career        string `json:"career"`
		Relationships string `json:"relationships"`
		Strengths     string `json:"strengths"`
		GrowthAreas   string `json:"growth_areas"`
	} `json:"themes"`
	Disclaimer string `json:"disclaimer"`
}
type Service struct {
	Corpus Corpus
	LLM    LLM
}

func NewService(c Corpus, l LLM) *Service { return &Service{Corpus: c, LLM: l} }
func ChartHash(in engine.ChartInput, language string) string {
	b, _ := json.Marshal(struct {
		engine.ChartInput
		Ayanamsa, Language string
	}{in, "lahiri", language})
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func (s *Service) BuildFacts(ctx context.Context, c engine.Chart, asOf time.Time) (engine.ChartFacts, []Rule, error) {
	facts, e := engine.Facts(c, asOf)
	if e != nil {
		return facts, nil, e
	}
	if s.Corpus != nil {
		rules, e := s.Corpus.Rules(ctx, facts)
		if e != nil {
			return facts, nil, e
		}
		return facts, rules, nil
	}
	return facts, nil, nil
}

// FallbackModel labels readings composed from engine facts and the corpus
// without an LLM.
const FallbackModel = "grounded-corpus"

// Generate produces a reading and names the model that wrote it. When no LLM
// is configured, or the LLM fails or keeps failing validation, the reading is
// composed deterministically from the facts and retrieved passages instead.
func (s *Service) Generate(ctx context.Context, facts engine.ChartFacts, rules []Rule) (Reading, string, error) {
	grounded := func() Reading { return fallback(newInsight(ctx, s.Corpus, facts, rules)) }
	if s.LLM == nil {
		return grounded(), FallbackModel, nil
	}
	prompt, e := promptFor(facts, rules)
	if e != nil {
		return Reading{}, "", e
	}
	for attempt := 0; attempt < 2; attempt++ {
		raw, err := s.LLM.Complete(ctx, prompt)
		if err != nil {
			break
		}
		raw = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(raw, "```"), "```json"))
		var out Reading
		if err = json.Unmarshal([]byte(raw), &out); err != nil {
			continue
		}
		if err = validate(out, facts); err != nil {
			prompt += "\nYour previous response failed validation: " + err.Error() + ". Return corrected JSON only."
			continue
		}
		return out, s.modelName(), nil
	}
	return grounded(), FallbackModel, nil
}

func (s *Service) modelName() string {
	if m, ok := s.LLM.(interface{ ModelName() string }); ok {
		return m.ModelName()
	}
	return "llm"
}

func promptFor(f engine.ChartFacts, rules []Rule) (string, error) {
	b, e := json.Marshal(f)
	if e != nil {
		return "", e
	}
	var rb strings.Builder
	for _, r := range rules {
		fmt.Fprintf(&rb, "\nRULE %s:%s — %s (%s): %s\n", r.DocType, r.Key, r.Title, r.Source, r.Body)
	}
	return fmt.Sprintf(`You are a careful Vedic astrology interpreter. Return only valid JSON matching this schema: {"summary":string,"lagna_and_moon":string,"grahas":[{"graha":string,"placement":string,"meaning":string}],"yogas":[{"name":string,"meaning":string,"effect":string,"strength":string}],"dashas":{"current":string,"upcoming":string},"themes":{"career":string,"relationships":string,"strengths":string,"growth_areas":string},"disclaimer":string}. The authoritative CHART FACTS below are computed by Swiss Ephemeris and Go. Never calculate, change or infer a position, house, dasha or yoga. Mention only listed yogas, and output exactly one yoga object per listed yoga. Use supportive non-deterministic language. Do not predict death, terminal illness, divorce or financial ruin. For medical, legal or financial questions advise a qualified professional. Always include: "For reflection, not certainty; this is not medical, legal or financial advice." Each INTERPRETATION RULE is keyed to a detected fact; ground your interpretation of that fact in its rules and use no rule for a fact it is not keyed to. Rules marked public_domain are historical classical passages in archaic language: convey their underlying theme, and never repeat their fatalistic, derogatory, gendered or bodily predictions literally.\nCHART FACTS:\n%s\nINTERPRETATION RULES:%s\nWrite a complete natal reading.`, string(b), rb.String()), nil
}
func validate(r Reading, f engine.ChartFacts) error {
	allowed := map[string]bool{}
	for _, y := range f.Yogas {
		allowed[y.Name] = true
	}
	seen := map[string]bool{}
	for _, y := range r.Yogas {
		if !allowed[y.Name] {
			return fmt.Errorf("reading mentions undetected yoga %q", y.Name)
		}
		seen[y.Name] = true
	}
	if len(r.Yogas) != len(f.Yogas) {
		return fmt.Errorf("reading yoga count %d does not equal fact count %d", len(r.Yogas), len(f.Yogas))
	}
	for _, y := range f.Yogas {
		if !seen[y.Name] {
			return fmt.Errorf("reading omitted detected yoga %q", y.Name)
		}
	}
	if !strings.Contains(strings.ToLower(r.Disclaimer), "reflection") {
		return fmt.Errorf("reading disclaimer missing")
	}
	return nil
}

// fallback writes a complete reading from engine facts and corpus entries.
func fallback(in *insight) Reading {
	f := in.f
	lagna := in.signs[in.lagnaSign()]
	moon := in.grahas["moon"]
	r := Reading{Disclaimer: "For reflection, not certainty; this is not medical, legal or financial advice."}

	yogas := in.yogaNames()
	r.Summary = fmt.Sprintf("%s ascendant with the Moon in %s (%s nakshatra) and the Sun in %s.", lagna, moon.Rashi, moon.Nakshatra, in.grahas["sun"].Rashi)
	if len(yogas) > 0 {
		r.Summary += fmt.Sprintf(" The engine detects %d yoga%s: %s.", len(yogas), map[bool]string{true: "s", false: ""}[len(yogas) != 1], strings.Join(yogas, ", "))
	}
	if f.Vimshottari.Current.Maha != "" {
		r.Summary += fmt.Sprintf(" Currently running %s–%s Vimshottari dasha.", engine.GrahaEnglish(f.Vimshottari.Current.Maha), engine.GrahaEnglish(f.Vimshottari.Current.Antara))
	}

	lm := []string{fmt.Sprintf("Lagna is %s at %.2f°, ruled by %s (placed in the %s house).", lagna, f.Chart.Ascendant.Degree, engine.GrahaEnglish(engine.HouseLord(f.Chart, 1)), ordinal(in.house(engine.HouseLord(f.Chart, 1))))}
	if t := in.entry(engine.DocBhava, "lagna_"+engine.Slug(lagna)); t != "" {
		lm = append(lm, t)
	}
	lm = append(lm, fmt.Sprintf("The Moon, your janma rashi, is in %s in %s pada %d.", moon.Rashi, moon.Nakshatra, moon.NakshatraPada))
	if t := in.entry(engine.DocNakshatra, engine.Slug(moon.Nakshatra)); t != "" {
		lm = append(lm, t)
	}
	r.LagnaAndMoon = strings.Join(lm, " ")

	for _, g := range f.Chart.Grahas {
		r.Grahas = append(r.Grahas, struct {
			Graha     string `json:"graha"`
			Placement string `json:"placement"`
			Meaning   string `json:"meaning"`
		}{g.Name + " (" + engine.GrahaEnglish(g.ID) + ")", in.placement(g.ID), in.grahaMeaning(g.ID)})
	}
	for _, y := range f.Yogas {
		meaning, classic := in.yogaMeaning(y)
		effect := "Formed by " + strings.Join(titleAll(y.Planets), " and ") + "."
		if len(y.Planets) == 0 {
			effect = "Formed by the placements around the Moon."
		}
		if classic != "" {
			effect += " Classical text: " + classic
		}
		r.Yogas = append(r.Yogas, struct {
			Name     string `json:"name"`
			Meaning  string `json:"meaning"`
			Effect   string `json:"effect"`
			Strength string `json:"strength"`
		}{y.Name, meaning, effect, y.Strength})
	}

	r.Dashas.Current = in.dashaLine()
	if t := in.entry(engine.DocDasha, "dasha_"+f.Vimshottari.Current.Maha); t != "" {
		r.Dashas.Current += " " + t
	}
	if u := f.Vimshottari.Upcoming.Lord; u != "" {
		r.Dashas.Upcoming = fmt.Sprintf("%s mahadasha from %s to %s.", engine.GrahaEnglish(u), f.Vimshottari.Upcoming.From, f.Vimshottari.Upcoming.To)
		if t := in.entry(engine.DocDasha, "dasha_"+u); t != "" {
			r.Dashas.Upcoming += " " + firstSentence(t, 240)
		}
	}

	r.Themes.Career = in.houseSummary(10)
	r.Themes.Relationships = in.houseSummary(7)
	if s := in.strengths(); len(s) > 0 {
		r.Themes.Strengths = strings.Join(s, "; ") + "."
	} else {
		r.Themes.Strengths = "No graha is exalted or in its own sign; strength comes from house placements and aspects rather than sign dignity."
	}
	if c := in.challenges(); len(c) > 0 {
		r.Themes.GrowthAreas = strings.Join(c, "; ") + ". These describe areas that reward conscious effort, not fixed outcomes."
	} else {
		r.Themes.GrowthAreas = "No debilitated or combust graha and no caution yoga is detected."
	}
	return r
}

func titleAll(ids []string) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = engine.GrahaEnglish(id)
	}
	return out
}

type OpenAIClient struct {
	BaseURL, APIKey, Model string
	HTTP                   *http.Client
}

func (c OpenAIClient) ModelName() string { return c.Model }

func (c OpenAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{"model": c.Model, "temperature": 0.2, "response_format": map[string]string{"type": "json_object"}, "messages": []map[string]string{{"role": "system", "content": "Return JSON only."}, {"role": "user", "content": prompt}}})
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if e != nil {
		return "", e
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, e := hc.Do(req)
	if e != nil {
		return "", e
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return "", fmt.Errorf("LLM HTTP status %d", resp.StatusCode)
	}
	var v struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if e = json.NewDecoder(resp.Body).Decode(&v); e != nil {
		return "", e
	}
	if len(v.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}
	return v.Choices[0].Message.Content, nil
}
