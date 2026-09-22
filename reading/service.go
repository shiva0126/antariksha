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

type Rule struct{ Key, Title, Body, Source string }
type Corpus interface {
	Rules(context.Context, engine.ChartFacts) ([]Rule, error)
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
func (s *Service) Generate(ctx context.Context, facts engine.ChartFacts, rules []Rule) (Reading, error) {
	if s.LLM == nil {
		return fallback(facts), nil
	}
	prompt, e := promptFor(facts, rules)
	if e != nil {
		return Reading{}, e
	}
	raw, e := s.LLM.Complete(ctx, prompt)
	if e != nil {
		return Reading{}, e
	}
	raw = strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(raw, "```"), "```json"))
	var out Reading
	if e = json.Unmarshal([]byte(raw), &out); e != nil {
		return Reading{}, fmt.Errorf("LLM returned invalid reading JSON: %w", e)
	}
	if e = validate(out, facts); e != nil {
		return Reading{}, e
	}
	return out, nil
}
func promptFor(f engine.ChartFacts, rules []Rule) (string, error) {
	b, e := json.Marshal(f)
	if e != nil {
		return "", e
	}
	var rb strings.Builder
	for _, r := range rules {
		fmt.Fprintf(&rb, "\nRULE %s (%s): %s\n", r.Key, r.Source, r.Body)
	}
	return fmt.Sprintf(`You are a careful Vedic astrology interpreter. Return only valid JSON matching this schema: {"summary":string,"lagna_and_moon":string,"grahas":[{"graha":string,"placement":string,"meaning":string}],"yogas":[{"name":string,"meaning":string,"effect":string,"strength":string}],"dashas":{"current":string,"upcoming":string},"themes":{"career":string,"relationships":string,"strengths":string,"growth_areas":string},"disclaimer":string}. The authoritative CHART FACTS below are computed by Swiss Ephemeris and Go. Never calculate, change or infer a position, house, dasha or yoga. Mention only listed yogas, and output exactly one yoga object per listed yoga. Use supportive non-deterministic language. Do not predict death, terminal illness, divorce or financial ruin. For medical, legal or financial questions advise a qualified professional. Always include: "For reflection, not certainty; this is not medical, legal or financial advice."\nCHART FACTS:\n%s\nINTERPRETATION RULES:%s\nWrite a complete natal reading.`, string(b), rb.String()), nil
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
func fallback(f engine.ChartFacts) Reading {
	moon := engine.Graha{Rashi: "unknown", Nakshatra: "unknown"}
	for _, g := range f.Chart.Grahas {
		if g.ID == "moon" || g.Name == "Chandra" {
			moon = g
			break
		}
	}
	r := Reading{Summary: fmt.Sprintf("A %s ascendant chart with the Moon in %s. The chart contains %d deterministic yoga flags.", f.Chart.Ascendant.Rashi, moon.Rashi, len(f.Yogas)), LagnaAndMoon: fmt.Sprintf("Lagna is %s at %.3f°. Chandra is in %s, %s.", f.Chart.Ascendant.Rashi, f.Chart.Ascendant.Degree, moon.Rashi, moon.Nakshatra), Disclaimer: "For reflection, not certainty; this is not medical, legal or financial advice."}
	for _, g := range f.Chart.Grahas {
		r.Grahas = append(r.Grahas, struct {
			Graha     string `json:"graha"`
			Placement string `json:"placement"`
			Meaning   string `json:"meaning"`
		}{g.Name, fmt.Sprintf("%s %.3f°; %s", g.Rashi, g.RashiDegree, g.Nakshatra), "This placement is an engine fact for reflective interpretation."})
	}
	for _, y := range f.Yogas {
		r.Yogas = append(r.Yogas, struct {
			Name     string `json:"name"`
			Meaning  string `json:"meaning"`
			Effect   string `json:"effect"`
			Strength string `json:"strength"`
		}{y.Name, "Detected by the deterministic geometry shown in chart facts.", "Consider this as a theme, not a certainty.", y.Strength})
	}
	r.Dashas.Current = fmt.Sprintf("%s–%s (%s to %s)", f.Vimshottari.Current.Maha, f.Vimshottari.Current.Antara, f.Vimshottari.Current.From, f.Vimshottari.Current.To)
	r.Dashas.Upcoming = f.Vimshottari.Upcoming.Lord
	return r
}

type OpenAIClient struct {
	BaseURL, APIKey, Model string
	HTTP                   *http.Client
}

func (c OpenAIClient) Complete(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{"model": c.Model, "temperature": 0.2, "response_format": map[string]string{"type": "json_object"}, "messages": []map[string]string{{"role": "system", "content": "Return JSON only."}, {"role": "user", "content": prompt}}})
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if e != nil {
		return "", e
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, e := c.HTTP.Do(req)
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
