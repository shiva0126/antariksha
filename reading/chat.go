package reading

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/example/panchang/engine"
)

// ChatTurn is one prior message in a conversation.
type ChatTurn struct {
	Role    string `json:"role"` // "user" | "assistant"
	Content string `json:"content"`
}

// ChatAnswer is a reply grounded in the chart and the corpus.
type ChatAnswer struct {
	Answer  string   `json:"answer"`
	Topics  []string `json:"topics"`
	Sources []Rule   `json:"sources"`
	Model   string   `json:"model"`
}

// ChatContext carries facts that depend on "now" rather than on birth.
type ChatContext struct {
	Transit  *engine.Chart    // sky at the time of asking, for Sade Sati and transits
	Shadbala *engine.Shadbala // six-fold planetary strength, when available
}

type topic struct {
	name  string
	words *regexp.Regexp
}

// Topics are matched in order; a question may touch several.
func words(ws string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b(?:` + ws + `)\b`)
}

// Topics are matched in order; a question may touch several. Patterns are
// whole words with their inflections spelled out, so "diet" is not "die" and
// "Sunapha" is not "Sun".
var topics = []topic{
	{"safety", words(`death|die|dies|dying|lifespan|longevity|suicide|kill|killed|accident|accidents|cancer|terminal|pregnant|pregnancy|abortion|lawsuit|court case|lottery|gamble|gambling|stock tips?|invest in`)},
	{"career", words(`career|careers|job|jobs|work|working|profession|professional|business|promotion|office|boss|employment|employed|occupation|10th house|tenth house`)},
	{"marriage", words(`marriage|marriages|marry|married|spouse|wife|husband|partner|partners|relationship|relationships|love|romance|wedding|7th house|seventh house|divorce`)},
	{"wealth", words(`money|wealth|wealthy|finance|finances|financial|income|rich|salary|savings|property|earn|earning|earnings|dhana|2nd house|11th house`)},
	{"health", words(`health|healthy|disease|diseases|illness|sick|sickness|fitness|6th house`)},
	{"education", words(`education|educational|study|studies|studying|exam|exams|college|degree|learning|school|4th house|5th house`)},
	{"children", words(`child|children|kids|son|sons|daughter|daughters|progeny|baby`)},
	{"mangal_dosha", words(`mangal dosha|mangal dosh|manglik|kuja dosha|mangal`)},
	{"sade_sati", words(`sade ?sati|shani dasha|shani transit|saturn transit`)},
	{"dasha", words(`dasha|dashas|dasa|mahadasha|antardasha|period|periods|timing|when|this year|next year|future|now|current|currently`)},
	{"yoga", words(`yoga|yogas|raja yoga|gajakesari|combination|combinations`)},
	{"nakshatra", words(`nakshatra|nakshatras|star|birth star|janma nakshatra|pada`)},
	{"lagna", words(`lagna|ascendant|rising|personality|nature|who am i|temperament`)},
	{"moon_sign", words(`rashi|moon sign|mind|emotions?|emotional`)},
	{"strength", words(`strength|strengths|strong|strongest|weak|weakest|shadbala|powerful|power`)},
	{"remedy", words(`remedy|remedies|upay|upaya|gemstones?|mantras?|puja|fix|improve`)},
}

var grahaWords = map[string]*regexp.Regexp{
	"sun":     words(`sun|surya|ravi`),
	"moon":    words(`moon|chandra|soma`),
	"mars":    words(`mars|mangala|kuja`),
	"mercury": words(`mercury|budha`),
	"jupiter": words(`jupiter|guru|brihaspati`),
	"venus":   words(`venus|shukra`),
	"saturn":  words(`saturn|shani`),
	"rahu":    words(`rahu`),
	"ketu":    words(`ketu`),
}

var houseWord = regexp.MustCompile(`(?i)\b(1[0-2]|[1-9])(st|nd|rd|th)? house\b`)

func classify(q string) ([]string, []string, int) {
	var ts []string
	for _, t := range topics {
		if t.words.MatchString(q) {
			ts = append(ts, t.name)
		}
	}
	var gs []string
	for _, id := range engine.GrahaIDs {
		if grahaWords[id].MatchString(q) {
			gs = append(gs, id)
		}
	}
	h := 0
	if m := houseWord.FindStringSubmatch(q); m != nil {
		fmt.Sscanf(m[1], "%d", &h)
	}
	return ts, gs, h
}

const safetyNote = "Astrology can describe tendencies for reflection, but it cannot predict death, illness, legal outcomes or financial returns. For those questions please consult a qualified doctor, lawyer or financial adviser."

// Answer replies to a question about the chart. An LLM, when configured,
// writes the reply from the same grounded material; otherwise, or if it fails,
// the reply is composed directly from engine facts and corpus passages.
func (s *Service) Answer(ctx context.Context, facts engine.ChartFacts, rules []Rule, question string, history []ChatTurn, cc ChatContext) (ChatAnswer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return ChatAnswer{}, fmt.Errorf("question is empty")
	}
	if len([]rune(question)) > 600 {
		return ChatAnswer{}, fmt.Errorf("question is too long (600 characters max)")
	}
	in := newInsight(ctx, s.Corpus, facts, rules)
	grounded, ts := compose(in, question, history, cc)
	sources := in.used
	if sources == nil {
		sources = []Rule{}
	}
	if ts == nil {
		ts = []string{}
	}
	ans := ChatAnswer{Answer: grounded, Topics: ts, Sources: sources, Model: FallbackModel}
	if s.LLM == nil || contains(ts, "safety") {
		return ans, nil
	}
	prompt, err := chatPrompt(facts, in, question, history, grounded)
	if err != nil {
		return ans, nil
	}
	raw, err := s.LLM.Complete(ctx, prompt)
	if err != nil {
		return ans, nil
	}
	var v struct {
		Answer string `json:"answer"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(raw), "```"), "```json"))), &v) != nil || strings.TrimSpace(v.Answer) == "" {
		return ans, nil
	}
	ans.Answer, ans.Model = strings.TrimSpace(v.Answer), s.modelName()
	return ans, nil
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func chatPrompt(f engine.ChartFacts, in *insight, q string, history []ChatTurn, grounded string) (string, error) {
	b, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	var rb strings.Builder
	for _, r := range in.used {
		fmt.Fprintf(&rb, "\nRULE %s:%s (%s): %s\n", r.DocType, r.Key, r.Source, r.Body)
	}
	var hb strings.Builder
	for _, t := range lastTurns(history, 6) {
		fmt.Fprintf(&hb, "%s: %s\n", strings.ToUpper(t.Role), t.Content)
	}
	return fmt.Sprintf(`You are Antariksha, a careful Vedic astrology assistant. Answer the user's question about their own birth chart in 120-220 words of plain text, warm and non-deterministic. Return JSON only: {"answer": string}.
Use only the CHART FACTS (computed by Swiss Ephemeris; never recompute or change a position, house, dasha or yoga) and the RULES. The DRAFT ANSWER is already correct and grounded; improve its clarity and relevance to the question, keep every fact in it, and add nothing that the facts or rules do not support. Classical passages marked public_domain are historical and archaic: convey their theme, never repeat fatalistic, derogatory, gendered or bodily predictions literally. Never predict death, illness, divorce or financial ruin. End with: "For reflection, not certainty."
CHART FACTS: %s
RULES:%s
CONVERSATION SO FAR:
%s
QUESTION: %s
DRAFT ANSWER: %s`, string(b), rb.String(), hb.String(), q, grounded), nil
}

func lastTurns(h []ChatTurn, n int) []ChatTurn {
	if len(h) > n {
		return h[len(h)-n:]
	}
	return h
}

// compose builds the grounded answer. Follow-up questions without a topic of
// their own ("and when?", "tell me more") inherit the previous question's topic.
func compose(in *insight, q string, history []ChatTurn, cc ChatContext) (string, []string) {
	ts, gs, house := classify(q)
	if len(ts) == 0 && len(gs) == 0 && house == 0 {
		for i := len(history) - 1; i >= 0; i-- {
			if history[i].Role == "user" {
				if pt, pg, ph := classify(history[i].Content); len(pt)+len(pg) > 0 || ph > 0 {
					ts, gs, house = pt, pg, ph
					break
				}
			}
		}
	}
	if contains(ts, "safety") {
		return safetyNote, ts
	}
	var parts []string
	add := func(s string) {
		if s = strings.TrimSpace(s); s != "" {
			parts = append(parts, s)
		}
	}
	f := in.f
	for _, t := range ts {
		switch t {
		case "career":
			add(in.houseSummary(10))
			for _, id := range engine.GrahasInHouse(f.Chart, 10) {
				add(fmt.Sprintf("%s: %s", in.placement(id), firstSentence(in.entry(engine.DocGrahaInHouse, id+"_in_10"), 300)))
			}
			if lord := engine.HouseLord(f.Chart, 10); in.house(lord) != 10 {
				add(fmt.Sprintf("The 10th lord %s is in the %s house: %s", engine.GrahaEnglish(lord), ordinal(in.house(lord)), firstSentence(in.entry(engine.DocGrahaInHouse, fmt.Sprintf("%s_in_%d", lord, in.house(lord))), 280)))
			}
			add(careerYogas(in))
			add(vargaLine(in, 10, "sun", "saturn"))
		case "marriage":
			add(in.houseSummary(7))
			for _, id := range engine.GrahasInHouse(f.Chart, 7) {
				add(fmt.Sprintf("%s: %s", in.placement(id), firstSentence(in.entry(engine.DocGrahaInHouse, id+"_in_7"), 300)))
			}
			add("Venus, the natural significator of partnership: " + in.placement("venus") + ". " + firstSentence(in.entry(engine.DocGrahaInHouse, fmt.Sprintf("venus_in_%d", in.house("venus"))), 260))
			add(vargaLine(in, 9, "venus", "jupiter"))
			add(mangal(in))
		case "wealth":
			add(in.houseSummary(2))
			add(in.houseSummary(11))
			for _, y := range f.Yogas {
				if y.Type == "wealth" {
					m, _ := in.yogaMeaning(y)
					add(y.Name + " is present: " + firstSentence(m, 260))
				}
			}
		case "health":
			add(in.houseSummary(1))
			add(in.houseSummary(6))
			add("For any health concern, please rely on a qualified doctor; the chart only describes constitution and tendencies.")
		case "education":
			add(in.houseSummary(4))
			add(in.houseSummary(5))
			add("Mercury, significator of learning: " + in.placement("mercury") + ". Jupiter, significator of wisdom: " + in.placement("jupiter") + ".")
		case "children":
			add(in.houseSummary(5))
			add("Jupiter, the natural significator of children: " + in.placement("jupiter") + ".")
		case "mangal_dosha":
			add(mangal(in))
		case "sade_sati":
			add(sadeSati(in, cc))
		case "dasha":
			add(in.dashaLine())
			add(in.entry(engine.DocDasha, "dasha_"+f.Vimshottari.Current.Maha))
			if a := f.Vimshottari.Current.Antara; a != "" && a != f.Vimshottari.Current.Maha {
				add(fmt.Sprintf("Within it, the %s antardasha colours events: %s", engine.GrahaEnglish(a), firstSentence(in.entry(engine.DocDasha, "dasha_"+a), 240)))
			}
			if p := f.Vimshottari.Current.Pratyantara; p != "" {
				add(fmt.Sprintf("The finer pratyantardasha running now is %s.", engine.GrahaEnglish(p)))
			}
			if y := f.Yogini.Current; y.Yogini != "" {
				add(fmt.Sprintf("In the Yogini dasha system you are in %s (ruled by %s) until %s.", y.Yogini, engine.GrahaEnglish(y.Lord), y.To))
			}
		case "yoga":
			if len(f.Yogas) == 0 {
				add("The engine detects none of the yogas in its catalogue for this chart.")
			}
			for _, y := range f.Yogas {
				m, c := in.yogaMeaning(y)
				line := fmt.Sprintf("%s (%s): %s", y.Name, y.Strength, m)
				if c != "" {
					line += " Classical text: " + c
				}
				add(line)
			}
		case "nakshatra":
			moon := in.grahas["moon"]
			add(fmt.Sprintf("Your janma nakshatra (the Moon's star) is %s, pada %d.", moon.Nakshatra, moon.NakshatraPada))
			add(in.entry(engine.DocNakshatra, engine.Slug(moon.Nakshatra)))
			if c, src := in.classic(engine.DocNakshatra, engine.Slug(moon.Nakshatra)); c != "" {
				add(fmt.Sprintf("Classical text: \"%s\" — %s", c, src))
			}
		case "lagna":
			lagna := in.signs[in.lagnaSign()]
			add(fmt.Sprintf("Your ascendant is %s at %.2f°.", lagna, f.Chart.Ascendant.Degree))
			add(in.entry(engine.DocBhava, "lagna_"+engine.Slug(lagna)))
			lord := engine.HouseLord(f.Chart, 1)
			add(fmt.Sprintf("The lagna lord %s: %s.", engine.GrahaEnglish(lord), in.placement(lord)))
		case "moon_sign":
			add("Your Moon sign (rashi): " + in.placement("moon") + ".")
			add(in.entry(engine.DocGrahaInSign, "moon_in_"+engine.Slug(in.signs[in.signIdx("moon")])))
		case "strength":
			add(strengthLine(cc))
		case "remedy":
			add("Antariksha describes the chart rather than prescribing remedies. Traditionally, strengthening a graha begins with its significations: for the current dasha lord " + engine.GrahaEnglish(f.Vimshottari.Current.Maha) + ", that means living its qualities consciously. For specific remedies such as gemstones or rituals, consult a trusted astrologer who can see the whole chart.")
		}
	}
	for _, id := range gs {
		if contains(ts, "marriage") && id == "venus" || contains(ts, "moon_sign") && id == "moon" {
			continue
		}
		add(in.placement(id) + ".")
		add(in.grahaMeaning(id))
		if c, src := in.classic(engine.DocGrahaInHouse, fmt.Sprintf("%s_in_%d", id, in.house(id))); c != "" {
			add(fmt.Sprintf("Classical text: \"%s\" — %s", c, src))
		}
	}
	if house > 0 && !contains(ts, "career") && !contains(ts, "marriage") {
		add(in.houseSummary(house))
	}
	if len(parts) == 0 {
		ts = append(ts, "overview")
		add(fmt.Sprintf("Here is the core of your chart. Ascendant %s; Moon in %s, %s nakshatra; Sun in %s.", in.signs[in.lagnaSign()], in.grahas["moon"].Rashi, in.grahas["moon"].Nakshatra, in.grahas["sun"].Rashi))
		if y := in.yogaNames(); len(y) > 0 {
			add("Detected yogas: " + strings.Join(y, ", ") + ".")
		}
		add(in.dashaLine())
		add("You can ask about career, marriage, wealth, education, children, health, your nakshatra, a planet (for example \"What does my Saturn mean?\"), a house, your yogas, Mangal dosha, Sade Sati or your current dasha.")
	}
	parts = dedupe(parts)
	return strings.Join(parts, "\n\n") + "\n\nFor reflection, not certainty.", ts
}

func dedupe(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		if !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

func careerYogas(in *insight) string {
	var names []string
	for _, y := range in.f.Yogas {
		switch y.Type {
		case "raja", "mahapurusha":
			names = append(names, y.Name)
		}
	}
	if len(names) == 0 {
		return ""
	}
	return "Yogas that support status and achievement: " + strings.Join(names, ", ") + "."
}

func mangal(in *insight) string {
	ok, h := engine.MangalDosha(in.f.Chart)
	if !ok {
		return fmt.Sprintf("Mangal dosha (lagna-based): not present. Mars is in the %s house, which is not one of the 1st, 2nd, 4th, 7th, 8th or 12th.", ordinal(h))
	}
	s := fmt.Sprintf("Mangal dosha (lagna-based): present, with Mars in the %s house.", ordinal(h))
	if st := in.f.Dignities["mars"].State; st == "own" || st == "exalted" {
		s += " Mars is " + map[string]string{"own": "in its own sign", "exalted": "exalted"}[st] + ", which many traditions treat as a cancellation."
	}
	s += " Matching traditions also check the dosha from the Moon and Venus and in the partner's chart, so treat this as one input, not a verdict."
	return s
}

func sadeSati(in *insight, cc ChatContext) string {
	if cc.Transit == nil {
		return "Sade Sati depends on Saturn's current transit, which is unavailable right now."
	}
	on, phase := engine.SadeSati(in.f.Chart, *cc.Transit)
	var sat engine.Graha
	for _, g := range cc.Transit.Grahas {
		if g.ID == "saturn" {
			sat = g
		}
	}
	moon := in.grahas["moon"].Rashi
	if !on {
		return fmt.Sprintf("You are not in Sade Sati. Transiting Saturn is in %s, and Sade Sati runs while it passes the 12th, 1st and 2nd signs from your Moon sign %s.", sat.Rashi, moon)
	}
	names := map[int]string{1: "rising (first) phase, Saturn in the 12th from the Moon", 2: "peak (second) phase, Saturn over the natal Moon", 3: "setting (third) phase, Saturn in the 2nd from the Moon"}
	return fmt.Sprintf("You are in Sade Sati: the %s. Transiting Saturn is in %s and your Moon sign is %s. Classically this is a period of responsibility, restructuring and maturity rather than misfortune; its tone depends on Saturn's strength in your chart (%s).", names[phase], sat.Rashi, moon, in.placement("saturn"))
}

// vargaLine summarises a divisional chart for the given significators.
func vargaLine(in *insight, n int, ids ...string) string {
	v, err := engine.VargaChart(in.f.Chart, n)
	if err != nil {
		return ""
	}
	parts := []string{}
	for _, id := range ids {
		for _, g := range v.Grahas {
			if g.ID == id {
				parts = append(parts, fmt.Sprintf("%s in %s (%s house)", engine.GrahaEnglish(id), g.Rashi, ordinal(engine.HouseOf(g.Longitude, v.Ascendant.Longitude))))
			}
		}
	}
	return fmt.Sprintf("In the %s (D%d, the chart of %s) the ascendant is %s, with %s.", engine.VargaName(n), n, engine.VargaTheme(n), v.Ascendant.Rashi, strings.Join(parts, " and "))
}

// strengthLine summarises Shadbala: strongest and weakest grahas against
// their classical minimums.
func strengthLine(cc ChatContext) string {
	if cc.Shadbala == nil || len(cc.Shadbala.Rows) == 0 {
		return ""
	}
	rows := append([]engine.ShadbalaRow(nil), cc.Shadbala.Rows...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Rank < rows[j].Rank })
	var strong, weak []string
	for _, r := range rows {
		s := fmt.Sprintf("%s %.2f rupas (%.0f%% of the %.1f required)", engine.GrahaEnglish(r.Graha), r.Rupas, r.Ratio*100, r.Required)
		if r.Ratio >= 1 {
			strong = append(strong, s)
		} else {
			weak = append(weak, s)
		}
	}
	out := fmt.Sprintf("By Shadbala (six-fold strength), your strongest graha is %s and the weakest is %s.", engine.GrahaEnglish(rows[0].Graha), engine.GrahaEnglish(rows[len(rows)-1].Graha))
	if len(strong) > 0 {
		out += " Meeting their required strength: " + strings.Join(strong, "; ") + "."
	}
	if len(weak) > 0 {
		out += " Below their requirement: " + strings.Join(weak, "; ") + ". Strong grahas deliver their significations and dashas more fully; weaker ones need more conscious effort."
	}
	return out
}
