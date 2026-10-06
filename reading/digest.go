package reading

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/example/panchang/engine"
)

// factsDigest is a compact, prompt-sized view of the chart facts for small
// local models: each placement is the same canonical sentence the grounded
// reading uses (house and dignity already worked out), floats are rounded,
// and dasha timelines are one short line per period. Nothing is recomputed;
// it only restates engine output, so the facts stay authoritative.
func factsDigest(f engine.ChartFacts) (string, error) {
	in := newInsight(context.Background(), nil, f, nil)
	period := func(p engine.DashaPeriod) string {
		return fmt.Sprintf("%s %s to %s", engine.GrahaEnglish(p.Lord), p.From, p.To)
	}
	var grahas []string
	for _, g := range f.Chart.Grahas {
		grahas = append(grahas, in.placement(g.ID))
	}
	var mahas, antaras []string
	for _, p := range f.Vimshottari.Sequence {
		mahas = append(mahas, period(p))
	}
	for _, p := range f.Vimshottari.Antaras {
		antaras = append(antaras, period(p))
	}
	type yoga struct {
		Name     string   `json:"name"`
		Type     string   `json:"type"`
		Planets  []string `json:"planets"`
		Houses   []string `json:"houses"`
		Strength string   `json:"strength"`
	}
	yogas := []yoga{}
	for _, y := range f.Yogas {
		yogas = append(yogas, yoga{y.Name, y.Type, y.Planets, y.Houses, y.Strength})
	}
	cur := f.Vimshottari.Current
	b, err := json.Marshal(map[string]any{
		"as_of":  f.AsOf,
		"lagna":  fmt.Sprintf("%s %.1f°", f.Chart.Ascendant.Rashi, f.Chart.Ascendant.Degree),
		"grahas": grahas,
		"current_dasha": fmt.Sprintf("%s mahadasha, %s antardasha, %s pratyantardasha (%s to %s)",
			engine.GrahaEnglish(cur.Maha), engine.GrahaEnglish(cur.Antara), engine.GrahaEnglish(cur.Pratyantara), cur.From, cur.To),
		"mahadashas":         mahas,
		"antardashas_now":    antaras,
		"yogini_now":         fmt.Sprintf("%s (%s) %s to %s", f.Yogini.Current.Yogini, engine.GrahaEnglish(f.Yogini.Current.Lord), f.Yogini.Current.From, f.Yogini.Current.To),
		"yogas":              yogas,
		"sarvashtakavarga":   f.Ashtakavarga.Sarva,
		"houses_whole_sign":  "house 1 is the lagna sign; count signs onward",
		"retrograde_planets": strings.Join(f.Retrograde, ", "),
	})
	return string(b), err
}

// ruleBody shortens a rule for compact prompts, keeping whole sentences.
func ruleBody(body string, compact bool) string {
	if !compact {
		return body
	}
	return firstSentence(body, 260)
}

// factsText is the CHART FACTS block: full JSON, or the digest when compact.
func factsText(f engine.ChartFacts, compact bool) (string, error) {
	if compact {
		return factsDigest(f)
	}
	b, err := json.Marshal(f)
	return string(b), err
}
