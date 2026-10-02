package reading

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/example/panchang/engine"
)

// insight answers questions about one chart from engine facts plus retrieved
// corpus passages. Every sentence it produces is either an engine fact or a
// corpus passage; nothing is invented.
type insight struct {
	f      engine.ChartFacts
	rules  map[engine.CorpusKey][]Rule
	grahas map[string]engine.Graha
	signs  []string
	used   []Rule
	usedID map[string]bool
}

// allKeysFor lists every token an answer might draw on: the chart's detections
// plus all twelve bhavas and the natal placements (already included) of each graha.
func allKeysFor(f engine.ChartFacts) []engine.CorpusKey {
	keys := engine.CorpusKeys(f)
	for h := 1; h <= 12; h++ {
		keys = append(keys, engine.CorpusKey{DocType: engine.DocBhava, Key: fmt.Sprintf("bhava_%d", h)})
	}
	return keys
}

func newInsight(ctx context.Context, c Corpus, f engine.ChartFacts, base []Rule) *insight {
	in := &insight{f: f, rules: map[engine.CorpusKey][]Rule{}, grahas: map[string]engine.Graha{}, signs: engine.RashiNames(), usedID: map[string]bool{}}
	rules := base
	if c != nil {
		if rs, err := c.RulesFor(ctx, allKeysFor(f)); err == nil && len(rs) > 0 {
			rules = rs
		}
	}
	for _, r := range rules {
		k := engine.CorpusKey{DocType: r.DocType, Key: r.Key}
		in.rules[k] = append(in.rules[k], r)
	}
	for _, g := range f.Chart.Grahas {
		in.grahas[g.ID] = g
	}
	return in
}

func classical(r Rule) bool { return strings.Contains(r.Source, "[public_domain]") }

// entry returns the self-authored interpretation for a token (marking it used).
func (in *insight) entry(doc, key string) string {
	for _, r := range in.rules[engine.CorpusKey{DocType: doc, Key: key}] {
		if !classical(r) {
			in.use(r)
			return r.Body
		}
	}
	return ""
}

// classic returns the first classical passage for a token with its citation.
func (in *insight) classic(doc, key string) (string, string) {
	for _, r := range in.rules[engine.CorpusKey{DocType: doc, Key: key}] {
		if classical(r) {
			in.use(r)
			return firstSentence(stripNote(r.Body), 260), cite(r)
		}
	}
	return "", ""
}

func (in *insight) use(r Rule) {
	id := r.DocType + ":" + r.Key + "|" + r.Source
	if !in.usedID[id] {
		in.usedID[id] = true
		in.used = append(in.used, r)
	}
}

func cite(r Rule) string {
	s := r.Source
	if i := strings.Index(s, " ["); i > 0 {
		s = s[:i]
	}
	return strings.Replace(s, "The Brihat Jataka of Varaha Mihira", "Brihat Jataka", 1)
}

func stripNote(s string) string {
	if i := strings.Index(s, "\n[Editor's note:"); i > 0 {
		return strings.TrimSpace(s[i+len("\n[Editor's note:") : strings.LastIndex(s, "]")])
	}
	return s
}

func firstSentence(s string, max int) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) <= max {
		return s
	}
	r := []rune(s)[:max]
	if i := strings.LastIndexAny(string(r), ".;"); i > 40 {
		return string(r[:i+1])
	}
	return strings.TrimSpace(string(r)) + "…"
}

func ordinal(n int) string {
	switch n {
	case 1, 21:
		return fmt.Sprintf("%dst", n)
	case 2, 22:
		return fmt.Sprintf("%dnd", n)
	case 3, 23:
		return fmt.Sprintf("%drd", n)
	}
	return fmt.Sprintf("%dth", n)
}

func (in *insight) house(id string) int {
	return engine.HouseOf(in.grahas[id].Longitude, in.f.Chart.Ascendant.Longitude)
}

func (in *insight) signIdx(id string) int { return int(in.grahas[id].Longitude/30) % 12 }

func (in *insight) lagnaSign() int { return int(in.f.Chart.Ascendant.Longitude/30) % 12 }

// placement describes a graha as "Jupiter (Guru) in Dhanu, 6th house, Purva
// Ashadha pada 4, exalted, retrograde".
func (in *insight) placement(id string) string {
	g := in.grahas[id]
	parts := []string{fmt.Sprintf("%s in %s, %s house", engine.GrahaEnglish(id), in.signs[in.signIdx(id)], ordinal(in.house(id)))}
	if g.Nakshatra != "" {
		parts = append(parts, fmt.Sprintf("%s pada %d", g.Nakshatra, g.NakshatraPada))
	}
	if d, ok := in.f.Dignities[id]; ok && d.State != "neutral" {
		st := d.State
		if st == "own" {
			st = "in its own sign"
		} else if st == "friendly" || st == "enemy" {
			st = "in a " + st + " sign"
		}
		if d.NeechaBhanga {
			st += " (debilitation cancelled)"
		}
		parts = append(parts, st)
	}
	if in.f.Combustion[id] {
		parts = append(parts, "combust")
	}
	for _, r := range in.f.Retrograde {
		if r == id {
			parts = append(parts, "retrograde")
		}
	}
	return strings.Join(parts, ", ")
}

// grahaMeaning combines the house and sign interpretations for a graha.
func (in *insight) grahaMeaning(id string) string {
	h := in.entry(engine.DocGrahaInHouse, fmt.Sprintf("%s_in_%d", id, in.house(id)))
	s := in.entry(engine.DocGrahaInSign, id+"_in_"+engine.Slug(in.signs[in.signIdx(id)]))
	out := firstSentence(h, 420)
	if s != "" {
		out += " " + firstSentence(s, 220)
	}
	if in.f.Combustion[id] {
		if d := in.entry(engine.DocDignity, "combust_"+id); d != "" {
			out += " " + firstSentence(d, 200)
		}
	}
	return strings.TrimSpace(out)
}

// houseSummary explains a house: its sign, its lord's placement and occupants.
func (in *insight) houseSummary(h int) string {
	lord := engine.HouseLord(in.f.Chart, h)
	var b strings.Builder
	if t := in.entry(engine.DocBhava, fmt.Sprintf("bhava_%d", h)); t != "" {
		b.WriteString(t)
	}
	occ := engine.GrahasInHouse(in.f.Chart, h)
	if len(occ) == 0 {
		fmt.Fprintf(&b, " There is no planet in this section of the chart; that does not mean this area of life is missing. The tradition also considers %s, the planet associated with its sign.", engine.GrahaEnglish(lord))
	} else {
		for _, id := range occ {
			fmt.Fprintf(&b, " %s: %s", engine.GrahaEnglish(id), firstSentence(in.entry(engine.DocGrahaInHouse, fmt.Sprintf("%s_in_%d", id, h)), 180))
		}
	}
	return b.String()
}

func (in *insight) yogaNames() []string {
	out := make([]string, len(in.f.Yogas))
	for i, y := range in.f.Yogas {
		out[i] = y.Name
	}
	return out
}

func (in *insight) yogaMeaning(y engine.Yoga) (meaning, classicalLine string) {
	meaning = in.entry(engine.DocYoga, engine.Slug(y.Name))
	if c, src := in.classic(engine.DocYoga, engine.Slug(y.Name)); c != "" {
		classicalLine = fmt.Sprintf("%s — %s", c, src)
	}
	return
}

func (in *insight) dashaLine() string {
	d := in.f.Vimshottari
	if d.Current.Maha == "" {
		return "The current Vimshottari period could not be determined for this date."
	}
	line := fmt.Sprintf("The current main period is associated with %s, with a shorter %s period inside it (%s to %s). These are traditionally called mahadasha and antardasha.", engine.GrahaEnglish(d.Current.Maha), engine.GrahaEnglish(d.Current.Antara), d.Current.From, d.Current.To)
	for _, p := range d.Sequence {
		if p.Lord == d.Current.Maha && p.To > d.Current.From {
			line += fmt.Sprintf(" The main %s period runs until %s.", engine.GrahaEnglish(p.Lord), p.To)
			break
		}
	}
	if d.Upcoming.Lord != "" {
		line += fmt.Sprintf(" The main %s period follows from %s.", engine.GrahaEnglish(d.Upcoming.Lord), d.Upcoming.From)
	}
	return line
}

// strengths lists grahas in exaltation or own sign and the supportive yogas.
func (in *insight) strengths() []string {
	var out []string
	for _, id := range engine.GrahaIDs {
		switch in.f.Dignities[id].State {
		case "exalted":
			out = append(out, engine.GrahaEnglish(id)+" has a placement traditionally considered especially supported. "+firstSentence(in.grahaMeaning(id), 180))
		case "own":
			out = append(out, engine.GrahaEnglish(id)+" is in a sign traditionally associated with it. "+firstSentence(in.grahaMeaning(id), 180))
		}
	}
	for _, y := range in.f.Yogas {
		if y.Type != "caution" {
			out = append(out, y.Name+": "+firstSentence(in.entry(engine.DocYoga, engine.Slug(y.Name)), 180))
		}
	}
	return out
}

func (in *insight) challenges() []string {
	var out []string
	for _, id := range engine.GrahaIDs {
		d := in.f.Dignities[id]
		if d.State == "debilitated" {
			s := engine.GrahaEnglish(id) + " has a placement traditionally described as challenging, not a personal weakness. " + firstSentence(in.grahaMeaning(id), 180)
			if d.NeechaBhanga {
				s += " The engine also detects a condition that offsets that label"
			}
			out = append(out, s)
		}
		if in.f.Combustion[id] {
			out = append(out, engine.GrahaEnglish(id)+" is close to the Sun in the chart (called combustion). This is a symbolic caution, not a loss of ability")
		}
	}
	for _, y := range in.f.Yogas {
		if y.Type == "caution" {
			out = append(out, y.Name+": "+in.entry(engine.DocYoga, engine.Slug(y.Name)))
		}
	}
	sort.Strings(out)
	return out
}
