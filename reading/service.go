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
	// Distance is the cosine distance of a semantic search hit (0 otherwise).
	Distance float64 `json:"-"`
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

// ChatLLM takes a system instruction and returns plain text. The local
// fine-tuned model answers this way, from compactChart lines.
type ChatLLM interface {
	Chat(ctx context.Context, system, user string) (string, error)
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
	// Compact sends a shorter digest of the facts and rules, for small local
	// models whose speed and attention depend on prompt length.
	Compact bool
}

func NewService(c Corpus, l LLM) *Service { return &Service{Corpus: c, LLM: l} }
func ChartHash(in engine.ChartInput, language string) string {
	b, _ := json.Marshal(struct {
		engine.ChartInput
		Ayanamsa, Language, ReadingVersion string
	}{in, "lahiri", language, "fact-checked-v3"})
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
	// A small local model (Compact) answers chat questions only; a full
	// reading is long structured JSON, beyond it on a CPU in time.
	if s.LLM == nil || s.Compact {
		return grounded(), FallbackModel, nil
	}
	prompt, e := promptFor(facts, rules, s.Compact)
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

func promptFor(f engine.ChartFacts, rules []Rule, compact bool) (string, error) {
	b, e := factsText(f, compact)
	if e != nil {
		return "", e
	}
	var rb strings.Builder
	canonical := fallback(newInsight(context.Background(), nil, f, nil))
	placements, _ := json.Marshal(canonical.Grahas)
	fmt.Fprintf(&rb, "\nCopy each grahas[].graha and grahas[].placement EXACTLY from these canonical fields; write your own meaning. Preserve each yoga strength exactly. Canonical graha fields: %s\n", placements)
	for _, r := range rules {
		if strings.Contains(r.Source, "[public_domain]") {
			// Historical wording stays available in the grounding record, but old
			// predictions are not needed in a user-facing language-generation prompt.
			if !compact {
				fmt.Fprintf(&rb, "\nHISTORICAL REFERENCE %s:%s — %s (%s, %s). Citation metadata only; use the project-authored plain-language rule for this key.\n", r.DocType, r.Key, r.Title, r.Source, r.Ref)
			}
			continue
		}
		if compact {
			fmt.Fprintf(&rb, "\nRULE %s:%s: %s\n", r.DocType, r.Key, ruleBody(r.Body, true))
			continue
		}
		fmt.Fprintf(&rb, "\nRULE %s:%s — %s (%s): %s\n", r.DocType, r.Key, r.Title, r.Source, r.Body)
	}
	return fmt.Sprintf(`You are Astrisk, explaining Vedic astrology to a curious person with no astrology background. Return only valid JSON matching this schema: {"summary":string,"lagna_and_moon":string,"grahas":[{"graha":string,"placement":string,"meaning":string}],"yogas":[{"name":string,"meaning":string,"effect":string,"strength":string}],"dashas":{"current":string,"upcoming":string},"themes":{"career":string,"relationships":string,"strengths":string,"growth_areas":string},"disclaimer":string}. Start each section with an everyday explanation, then give a practical reflection prompt. Use short sentences and define Indian astrology terms in familiar words. Do not assume chart themes are facts about the person's habits, family, work, health or relationships. The authoritative CHART FACTS below come from Swiss Ephemeris and Go; never calculate, alter or infer any position, house, dasha or yoga. Mention only listed yogas and return exactly one object per listed yoga. Strength is a technical label for the engine's rule, never a judgment of the person or certainty of an outcome. Never predict death, illness, divorce, financial ruin or unavoidable events. For medical, legal or financial questions advise a qualified professional. Always include: "For reflection, not certainty; this is not medical, legal or financial advice." Each INTERPRETATION RULE is keyed to a detected fact; use only its matching rule. Historical public-domain wording may be old-fashioned or harmful: explain a humane present-day theme in your own words and do not repeat it.\nCHART FACTS:\n%s\nINTERPRETATION RULES:%s\nWrite a complete natal reading in language a teenager could understand.`, b, rb.String()), nil
}
func validate(r Reading, f engine.ChartFacts) error {
	// Multiple planets may independently form the same named yoga. Compare
	// the exact name/strength multiset, rather than collapsing names to a set.
	allowed := map[string]int{}
	for _, y := range f.Yogas {
		allowed[y.Name+"\x00"+y.Strength]++
	}
	for _, y := range r.Yogas {
		key := y.Name + "\x00" + y.Strength
		if allowed[key] == 0 {
			return fmt.Errorf("undetected, excess or changed yoga %q", y.Name)
		}
		allowed[key]--
	}
	if len(r.Yogas) != len(f.Yogas) {
		return fmt.Errorf("reading yoga count %d does not equal fact count %d", len(r.Yogas), len(f.Yogas))
	}
	if !strings.Contains(strings.ToLower(r.Disclaimer), "reflection") {
		return fmt.Errorf("reading disclaimer missing")
	}
	canonical := fallback(newInsight(context.Background(), nil, f, nil))
	if len(r.Grahas) != len(canonical.Grahas) {
		return fmt.Errorf("graha count differs from facts")
	}
	for i, g := range r.Grahas {
		if g.Graha != canonical.Grahas[i].Graha || g.Placement != canonical.Grahas[i].Placement || strings.TrimSpace(g.Meaning) == "" {
			return fmt.Errorf("graha placement or identity differs from facts at index %d", i)
		}
	}
	if r.Summary == "" || r.LagnaAndMoon == "" || r.Themes.Career == "" || r.Themes.Relationships == "" || r.Themes.Strengths == "" || r.Themes.GrowthAreas == "" {
		return fmt.Errorf("reading sections missing")
	}
	return validateNarrative(r, f)
}

// fallback writes a complete reading from engine facts and corpus entries.
func fallback(in *insight) Reading {
	f := in.f
	lagna := in.signs[in.lagnaSign()]
	moon := in.grahas["moon"]
	r := Reading{Disclaimer: "For reflection, not certainty; this is not medical, legal or financial advice."}

	r.Summary = "Start with how you approach life, what helps you feel settled, and the choices available to you. " + firstSentence(in.entry(engine.DocBhava, "lagna_"+engine.Slug(lagna)), 240)
	r.Summary += " The sections below explain the chart's symbolic themes in everyday language; they are not a verdict about who you are."
	lm := []string{fmt.Sprintf("Your rising sign (also called Lagna) is %s. It is the sign on the eastern horizon at birth.", lagna)}
	if t := in.entry(engine.DocBhava, "lagna_"+engine.Slug(lagna)); t != "" {
		lm = append(lm, t)
	}
	lm = append(lm, fmt.Sprintf("Your Moon sign is %s. Astrology uses the Moon as a symbol for emotional needs and familiar comforts.", moon.Rashi))
	if t := in.entry(engine.DocGrahaInSign, "moon_in_"+engine.Slug(moon.Rashi)); t != "" {
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
		meaning, _ := in.yogaMeaning(y)
		effect := "Formed by " + strings.Join(titleAll(y.Planets), " and ") + "."
		if len(y.Planets) == 0 {
			effect = "Formed by the placements around the Moon."
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
		r.Dashas.Upcoming = fmt.Sprintf("The next main period is associated with %s, from %s to %s.", engine.GrahaEnglish(u), f.Vimshottari.Upcoming.From, f.Vimshottari.Upcoming.To)
		if t := in.entry(engine.DocDasha, "dasha_"+u); t != "" {
			r.Dashas.Upcoming += " " + firstSentence(t, 240)
		}
	}

	r.Themes.Career = in.houseSummary(10)
	r.Themes.Relationships = in.houseSummary(7)
	if s := in.strengths(); len(s) > 0 {
		r.Themes.Strengths = strings.Join(s, "; ") + "."
	} else {
		r.Themes.Strengths = "The selected chart rules do not identify a particularly supported planet. That says nothing about your actual abilities: skills, experience and support matter more than a label."
	}
	if c := in.challenges(); len(c) > 0 {
		r.Themes.GrowthAreas = strings.Join(c, "; ") + ". These describe areas that reward conscious effort, not fixed outcomes."
	} else {
		r.Themes.GrowthAreas = "The selected rules do not flag a challenging pattern. This does not mean life will be free of difficulties; choose the areas you personally want to work on."
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
	return c.send(ctx, map[string]any{"model": c.Model, "temperature": 0.2, "response_format": map[string]string{"type": "json_object"}, "messages": []map[string]string{{"role": "system", "content": "Return JSON only."}, {"role": "user", "content": prompt}}})
}

func (c OpenAIClient) Chat(ctx context.Context, system, user string) (string, error) {
	return c.send(ctx, map[string]any{"model": c.Model, "temperature": 0.2, "max_tokens": 450, "messages": []map[string]string{{"role": "system", "content": system}, {"role": "user", "content": user}}})
}

func (c OpenAIClient) send(ctx context.Context, payload map[string]any) (string, error) {
	body, _ := json.Marshal(payload)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if e != nil {
		return "", e
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
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
