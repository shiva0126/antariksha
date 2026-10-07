package reading

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/example/panchang/divination"
	"github.com/example/panchang/engine"
)

// MatchPerson is one side of a compared pair. Names and birth details are
// deliberately absent: answers refer to "the groom" and "the bride".
type MatchPerson struct {
	Facts      engine.ChartFacts
	Rules      []Rule
	Numerology *divination.Numerology
}

// MatchChatContext is everything a matching answer may draw on. All of it is
// computed by the server from the two birth inputs; nothing comes from a client.
type MatchChatContext struct {
	Match     engine.Match
	Boy, Girl MatchPerson
	Lang      string // reply language code; English without an LLM
}

var matchTopics = []topic{
	{"safety", words(`death|die|dies|widow|widowed|widower|divorce|divorced|separation|separate|cheat|cheating|affair|affairs|unfaithful|loyal|loyalty|infertile|infertility|fertility|pregnant|pregnancy|abuse|abusive|violence|violent|dowry|suicide|kill|lifespan|accident`)},
	{"muhurta", words(`wedding date|muhurat|muhurta|muhurtham|auspicious (?:date|day)|date for (?:the )?(?:wedding|marriage)|marriage date|engagement date`)},
	{"score", words(`score|scores|guna|gunas|points?|total|overall|compatible|compatibility|good match|good pair|guna milan|milan|ashtakoota|verdict|result|results`)},
	{"varna", words(`varna`)},
	{"vashya", words(`vashya|vasya`)},
	{"porutham", words(`porutham|poruthams|porutham?s|poruttam|rajju|vedha|mahendra|stree deergha|stree dirgha|south indian|rasyadhipati|10 porutham|ten porutham|thirumana porutham|jathaka porutham`)},
	{"papasamya", words(`papasamya|papa samya|paapa samya|papasamyam|malefic balance|papa points`)},
	{"tara", words(`tara|dina`)},
	{"yoni", words(`yoni|intimacy|physical`)},
	{"graha_maitri", words(`graha maitri|maitri|friendship|communication|understanding|talk|talking|mental`)},
	{"gana", words(`gana|ganas|temperament|nature|attitude`)},
	{"bhakoot", words(`bhakoot|bhakut|bhakuta|rashi koota|money|finances|financial|family|in-laws`)},
	{"nadi", words(`nadi|naadi|health|children|child|kids|progeny|genes|genetic`)},
	{"dosha", words(`dosha|doshas|dosh|cancel|cancels|cancelled|canceled|cancellation|exception|exceptions|parihar|parihara|problem|problems|issue|issues|bad|worry|worried|concern|concerns`)},
	{"mangal", words(`mangal|manglik|mangalik|kuja|mars|chevvai`)},
	{"numerology", words(`numerology|numerological|numbers?|life path|mulank|moolank|bhagyank|bhagyaank|destiny number|root number|lucky numbers?|ank jyotish`)},
	{"moon", words(`moon|moon signs?|rashi|rashis|nakshatra|nakshatras|stars?|birth stars?|emotions?|emotional|feelings|mind`)},
	{"seventh", words(`7th house|seventh house|venus|spouse|married life|marital|marriage house|partner|partners|relationship`)},
	{"timing", words(`dasha|dashas|dasa|mahadasha|period|periods|when|timing|this year|next year|future|now|current|currently`)},
	{"remedy", words(`remedy|remedies|upay|upaya|puja|pooja|gemstones?|mantras?|fix|solution|solutions|remove`)},
}

var kootaTopic = map[string]string{"varna": "Varna", "vashya": "Vashya", "tara": "Tara", "yoni": "Yoni", "graha_maitri": "Graha Maitri", "gana": "Gana", "bhakoot": "Bhakoot", "nadi": "Nadi"}

var matchHousePattern = regexp.MustCompile(`(?i)\b(1[0-2]|[1-9])(st|nd|rd|th)? house\b`)

// nativeMatchTopics recognise matching questions in Hinglish and Indian
// scripts (substring matches; regexp word boundaries are ASCII-only).
var nativeMatchTopics = []topic{
	{"score", regexp.MustCompile(`(?i)\b(kundli milan|kundali milan|gun milan|gunn milan|milan)\b|गुण मिलान|कुंडली मिलान|गुण|ಗುಣ|ಜಾತಕ ಹೊಂದಾಣಿಕೆ|ஜாதகப் பொருத்தம்|గుణ|జాతక పొంతన|गुण मेलन|ജാതകപ്പൊരുത്തം|ગુણ મિલાન|কুষ্ঠি মিলন`)},
	{"porutham", regexp.MustCompile(`பொருத்தம்|ಹೊಂದಾಣಿಕೆ|పొంతన|പൊരുത്തം|ரஜ்ஜு|ರಜ್ಜು|రజ్జు|രജ്ജു|ವೇಧ|வேதை|వేధ`)},
	{"mangal", regexp.MustCompile(`(?i)\b(manglik|mangalik|mangal dosh)\b|मांगलिक|मंगल दोष|ಕುಜ ದೋಷ|செவ்வாய் தோஷம்|కుజ దోషం|मंगळ दोष|ചൊവ്വാ ദോഷം|મંગળ દોષ|মাঙ্গলিক`)},
	{"nadi", regexp.MustCompile(`नाड़ी|नाडी|ನಾಡಿ|நாடி|నాడి|നാഡി|નાડી|নাড়ি`)},
}

func classifyMatch(q string) []string {
	var ts []string
	for _, t := range matchTopics {
		if t.words.MatchString(q) {
			ts = append(ts, t.name)
		}
	}
	for _, t := range nativeMatchTopics {
		if !contains(ts, t.name) && t.words.MatchString(q) {
			ts = append(ts, t.name)
		}
	}
	return ts
}

const matchSafetyNote = "No birth chart, guna score or number can predict divorce, faithfulness, fertility, health, harm or how long anyone will live, so Astrisk does not answer those questions. They are best explored through honest conversation, time together and, where relevant, a qualified doctor, counsellor or lawyer. If you are worried about anyone's safety, including pressure about dowry, please talk to someone you trust; in India you can call the Women Helpline 181 or emergency services on 112.\n\nFor reflection, not certainty."

// AnswerMatch replies to a question about a compared pair. Conversations are
// not stored: the client supplies the previous turns, which are used only to
// carry a topic over to follow-up questions and as LLM context.
func (s *Service) AnswerMatch(ctx context.Context, mc MatchChatContext, question string, history []ChatTurn) (ChatAnswer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return ChatAnswer{}, fmt.Errorf("question is empty")
	}
	if len([]rune(question)) > 600 {
		return ChatAnswer{}, fmt.Errorf("question is too long (600 characters max)")
	}
	boy := newInsight(ctx, s.Corpus, mc.Boy.Facts, mc.Boy.Rules)
	girl := newInsight(ctx, s.Corpus, mc.Girl.Facts, mc.Girl.Rules)
	grounded, ts := composeMatch(mc, boy, girl, question, history)
	sources := append(append([]Rule{}, boy.used...), girl.used...)
	ans := ChatAnswer{Answer: grounded, Topics: ts, Sources: dedupeRules(sources), Model: FallbackModel}
	if s.LLM == nil || contains(ts, "safety") {
		return ans, nil
	}
	b, err := json.Marshal(map[string]any{"match": mc.Match, "groom_moon": mc.Match.BoyMoon, "bride_moon": mc.Match.GirlMoon})
	if err != nil {
		return ans, nil
	}
	var hb strings.Builder
	for _, t := range lastTurns(history, 6) {
		fmt.Fprintf(&hb, "%s: %s\n", strings.ToUpper(t.Role), t.Content)
	}
	prompt := fmt.Sprintf(`You are Astrisk, a careful, friendly guide to Vedic kundali matching for people new to astrology. Answer in 120-220 words of plain everyday English and return JSON only: {"answer": string}.
The DRAFT ANSWER is computed from the two charts and is correct: keep every fact, number and score in it exactly, improve clarity and relevance to the question, and add nothing the facts do not support. Treat every factor as a traditional table, not a trait of either person. Never recommend accepting or rejecting a partner, and never predict marriage success, divorce, fertility, health, harm or wealth. Do not mention caste ranking. End with: "For reflection, not certainty."
MATCH FACTS: %s
CONVERSATION SO FAR:
%s
QUESTION: %s
DRAFT ANSWER: %s%s`, string(b), hb.String(), question, grounded, languageRule(mc.Lang))
	raw, err := s.LLM.Complete(ctx, prompt)
	if err != nil {
		return ans, nil
	}
	var v struct {
		Answer string `json:"answer"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(raw), "```"), "```json"))), &v) != nil || strings.TrimSpace(v.Answer) == "" || unsafeMatchChat(v.Answer) {
		return ans, nil
	}
	ans.Answer, ans.Model = strings.TrimSpace(v.Answer), s.modelName()
	return ans, nil
}

func unsafeMatchChat(s string) bool {
	s = strings.ToLower(s)
	for _, term := range []string{"will divorce", "will die", "infertil", "caste", "superior", "inferior", "guarantee", "destined", "soulmate", "perfect match", "should marry", "should not marry", "do not marry", "don't marry"} {
		if strings.Contains(s, term) {
			return true
		}
	}
	return false
}

func dedupeRules(rs []Rule) []Rule {
	seen := map[string]bool{}
	out := []Rule{}
	for _, r := range rs {
		id := r.DocType + ":" + r.Key + "|" + r.Source
		if !seen[id] {
			seen[id] = true
			out = append(out, r)
		}
	}
	return out
}

func composeMatch(mc MatchChatContext, boy, girl *insight, q string, history []ChatTurn) (string, []string) {
	ts := classifyMatch(q)
	if len(ts) == 0 {
		for i := len(history) - 1; i >= 0; i-- {
			if history[i].Role == "user" {
				if pt := classifyMatch(history[i].Content); len(pt) > 0 {
					ts = pt
					break
				}
			}
		}
	}
	if contains(ts, "safety") {
		return matchSafetyNote, []string{"safety"}
	}
	m := mc.Match
	var parts []string
	add := func(s string) {
		if s = strings.TrimSpace(s); s != "" {
			parts = append(parts, s)
		}
	}
	for _, t := range ts {
		switch t {
		case "muhurta":
			add("Kundali matching compares the two birth charts; it does not choose a wedding date. For that, open the Muhurta page, choose Marriage, and personalise it with either saved profile: it checks the nakshatra, tithi, weekday and the person's tara bala and chandra bala for each day.")
		case "porutham":
			for _, p := range m.Poruthams {
				add(fmt.Sprintf("%s: %s. %s", p.Name, map[string]string{"good": "matches", "medium": "partly matches", "bad": "does not match"}[p.Status], p.Detail))
			}
		case "score":
			add(scoreLine(m))
		case "dosha":
			add(doshaLine(m))
		case "mangal":
			add(mangalPairLine(mc, boy, girl))
		case "numerology":
			add(numerologyPairLine(mc, boy, girl))
		case "moon":
			add(moonPairLine(m, boy, girl))
		case "seventh":
			add(seventhLine("groom", boy))
			add(seventhLine("bride", girl))
		case "timing":
			add(fmt.Sprintf("Current periods (Vimshottari dasha): the groom is in %s, and the bride in %s. Dashas describe each person's own chapter of life; Astrisk does not use them to predict when, or whether, a marriage happens.", periodName(boy), periodName(girl)))
		case "remedy":
			add("Astrisk does not prescribe or sell remedies. Traditional parihara for doshas (such as specific pujas) differ between families and regions, and many traditions already record cancellations, which this report lists. If remedies matter to your families, a trusted astrologer can look at both complete charts; the most reliable step is still an open conversation about expectations.")
		default:
			if name, ok := kootaTopic[t]; ok {
				for _, k := range m.Kootas {
					if k.Name == name {
						add(kootaDetail(k, m.Exceptions))
					}
				}
			}
		}
	}
	if h := matchHousePattern.FindStringSubmatch(q); h != nil && !contains(ts, "seventh") {
		var n int
		fmt.Sscanf(h[1], "%d", &n)
		add(fmt.Sprintf("Groom's %s house: %s", ordinal(n), houseBrief(boy, n)))
		add(fmt.Sprintf("Bride's %s house: %s", ordinal(n), houseBrief(girl, n)))
	}
	if len(parts) == 0 {
		ts = append(ts, "overview")
		add(scoreLine(m))
		add(doshaLine(m))
		add(mangalPairLine(mc, boy, girl))
		if mc.Boy.Numerology != nil && mc.Girl.Numerology != nil {
			add(fmt.Sprintf("Numerology adds a second lens: root numbers %d (%s) and %d (%s). Ask \"What does numerology say about us?\" to see how it lines up with the charts.", mc.Boy.Numerology.Mulank, engine.GrahaEnglish(mc.Boy.Numerology.MulankGraha), mc.Girl.Numerology.Mulank, engine.GrahaEnglish(mc.Girl.Numerology.MulankGraha)))
		}
		add("You can ask about any of the eight factors (Varna, Vashya, Tara, Yoni, Graha Maitri, Gana, Bhakoot, Nadi), doshas and their cancellations, Mangal dosha, the Moon signs and nakshatras, each person's 7th house and Venus, current dashas, or numerology.")
	}
	return matchPlain(mc, boy, girl, ts).render(dedupe(parts)), ts
}

func scoreLine(m engine.Match) string {
	var low, full []string
	for _, k := range m.Kootas {
		switch k.Score {
		case 0:
			low = append(low, k.Name)
		case k.Max:
			full = append(full, k.Name)
		}
	}
	s := fmt.Sprintf("The charts score %g out of %g gunas in the Ashtakoota (eight-factor) tables. Many families treat 18 or more as acceptable, but this is a traditional score, not a percentage chance of a happy marriage.", m.Total, m.Max)
	var ks []string
	for _, k := range m.Kootas {
		ks = append(ks, fmt.Sprintf("%s %g/%g", k.Name, k.Score, k.Max))
	}
	s += " Factor by factor: " + strings.Join(ks, ", ") + "."
	if len(full) > 0 {
		s += " Full points: " + strings.Join(full, ", ") + "."
	}
	if len(low) > 0 {
		s += " No points: " + strings.Join(low, ", ") + ". Ask about any factor by name to see what it compares."
	}
	return s
}

func doshaLine(m engine.Match) string {
	if len(m.Doshas) == 0 {
		return "No koota dosha is flagged by the app's rules (Bhakoot, Nadi or Gana)."
	}
	s := "Flagged by the app's rules: " + strings.Join(m.Doshas, ", ") + "."
	if len(m.Exceptions) > 0 {
		s += " Recorded traditional cancellations: " + strings.Join(m.Exceptions, " ") + " A cancellation does not change the raw score shown."
	} else {
		s += " No cancellation is recorded by the implemented rules; other traditions may consider further exceptions."
	}
	return s + " A dosha is a flag in a traditional table, not a prediction about the two people."
}

func kootaDetail(k engine.Koota, exceptions []string) string {
	g := matchGuides[k.Name]
	s := fmt.Sprintf("%s: %g of %g points. It compares the groom's %s with the bride's %s. %s %s", k.Name, k.Score, k.Max, k.Boy, k.Girl, k.Description, kootaResult(k, exceptions))
	if g[0] != "" {
		s += " " + g[0] + " A question to talk about together: " + g[1]
	}
	return s
}

func mangalPairLine(mc MatchChatContext, boy, girl *insight) string {
	state := func(on bool) string {
		if on {
			return "flagged"
		}
		return "not flagged"
	}
	s := fmt.Sprintf("Mangal dosha (Mars counted from the ascendant): groom %s (%s); bride %s (%s).", state(mc.Match.BoyMangal), boy.placement("mars"), state(mc.Match.GirlMangal), girl.placement("mars"))
	if mc.Match.MangalNote != "" {
		s += " " + mc.Match.MangalNote
	}
	return s + " A Mars flag is not evidence that anyone is angry, unsafe or unsuitable."
}

func moonPairLine(m engine.Match, boy, girl *insight) string {
	s := fmt.Sprintf("Matching is built on the Moon: the groom's Moon is in %s, the bride's in %s.", m.BoyMoon, m.GirlMoon)
	for _, p := range []struct {
		who string
		in  *insight
	}{{"groom", boy}, {"bride", girl}} {
		nak := p.in.grahas["moon"].Nakshatra
		if t := firstSentence(p.in.entry(engine.DocNakshatra, engine.Slug(nak)), 240); t != "" {
			s += fmt.Sprintf(" The %s's birth star %s: %s", p.who, nak, t)
		}
	}
	return s
}

func seventhLine(who string, in *insight) string {
	return fmt.Sprintf("The %s's 7th house (partnership): %s Venus: %s.", who, houseBrief(in, 7), in.placement("venus"))
}

func houseBrief(in *insight, h int) string {
	sign := in.signs[engine.HouseSign(in.f.Chart, h)]
	lord := engine.HouseLord(in.f.Chart, h)
	s := fmt.Sprintf("%s; its lord is %s.", sign, in.placement(lord))
	if occ := engine.GrahasInHouse(in.f.Chart, h); len(occ) > 0 {
		var names []string
		for _, id := range occ {
			names = append(names, engine.GrahaEnglish(id))
		}
		s += " Planets in it: " + strings.Join(names, ", ") + "."
	} else {
		s += " No planet occupies it, which is common and not a lack."
	}
	return s
}

func periodName(in *insight) string {
	c := in.f.Vimshottari.Current
	if c.Maha == "" {
		return "an undetermined period"
	}
	return fmt.Sprintf("%s mahadasha with a %s antardasha", engine.GrahaEnglish(c.Maha), engine.GrahaEnglish(c.Antara))
}

// numerologyPairLine compares both people's numbers and, as with the single
// chart, relates them to the kundali: the root-number grahas are compared by
// the same friendship table that Graha Maitri uses for the Moon-sign lords.
func numerologyPairLine(mc MatchChatContext, boy, girl *insight) string {
	bn, gn := mc.Boy.Numerology, mc.Girl.Numerology
	if bn == nil || gn == nil {
		return "Numerology needs both birth dates."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Numerology from the birth dates: the groom has life path %d, root number (mulank) %d ruled by %s and destiny number (bhagyank) %d ruled by %s; the bride has life path %d, root number %d ruled by %s and destiny number %d ruled by %s.",
		bn.LifePath.Number, bn.Mulank, engine.GrahaEnglish(bn.MulankGraha), bn.Bhagyank, engine.GrahaEnglish(bn.BhagyankGraha),
		gn.LifePath.Number, gn.Mulank, engine.GrahaEnglish(gn.MulankGraha), gn.Bhagyank, engine.GrahaEnglish(gn.BhagyankGraha))
	rootRel, rootScore := grahaRelation(bn.MulankGraha, gn.MulankGraha)
	destRel, _ := grahaRelation(bn.BhagyankGraha, gn.BhagyankGraha)
	fmt.Fprintf(&b, " In Indian numerology the root-number grahas are %s, and the destiny-number grahas are %s.", rootRel, destRel)
	if bn.LifePath.Number == gn.LifePath.Number {
		fmt.Fprintf(&b, " You share life path %d.", bn.LifePath.Number)
	}
	var maitri engine.Koota
	for _, k := range mc.Match.Kootas {
		if k.Name == "Graha Maitri" {
			maitri = k
		}
	}
	if maitri.Max > 0 {
		chart := "friendly"
		if maitri.Score < 3 {
			chart = "strained"
		} else if maitri.Score < 4 {
			chart = "neutral"
		}
		num := map[int]string{2: "friendly", 1: "neutral", 0: "strained"}[rootScore]
		fmt.Fprintf(&b, " How this lines up with the kundalis: Graha Maitri compares the Moon-sign lords (%s and %s) and gives %g of %g, a %s pairing, while the root-number grahas give a %s pairing.", maitri.Boy, maitri.Girl, maitri.Score, maitri.Max, chart, num)
		if chart == num {
			b.WriteString(" The two systems agree here.")
		} else {
			b.WriteString(" The two systems differ here, which is common: they use different starting points (the Moon at birth versus the calendar date).")
		}
	}
	for _, p := range []struct {
		who string
		in  *insight
		n   *divination.Numerology
	}{{"groom", boy, bn}, {"bride", girl, gn}} {
		if e := p.in.grahaEchoes(p.n.MulankGraha, "the "+p.who+"'s"); len(e) > 0 {
			fmt.Fprintf(&b, " In the %s's own chart the root-number graha %s is also emphasised: %s.", p.who, engine.GrahaEnglish(p.n.MulankGraha), strings.Join(e, "; "))
		}
	}
	b.WriteString(" Numerology is a separate symbolic system; neither it nor the guna score measures how two people will get along.")
	return b.String()
}
