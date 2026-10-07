package reading

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/example/panchang/divination"
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
	Transit    *engine.Chart          // sky at the time of asking, for Sade Sati and transits
	Shadbala   *engine.Shadbala       // six-fold planetary strength, when available
	Numerology *divination.Numerology // birth-date numbers, correlated with the chart
	Related    []Rule                 // library passages closest to the question (semantic search)
	Lang       string                 // reply language code (en, hi, mr, kn, ta, te, ml, gu, bn)
	// Forecast and Events are filled for timing questions only: the coming
	// periods (see Forecast) and the slow planets' sign changes.
	Forecast []Period
	Events   []engine.TransitEvent
	// Varsha is the Tajika year chart for the running birthday year.
	Varsha *engine.VarshaReport
}

// LanguageNames maps the app's language codes to names an LLM understands.
var LanguageNames = map[string]string{"hi": "Hindi", "mr": "Marathi", "kn": "Kannada", "ta": "Tamil", "te": "Telugu", "ml": "Malayalam", "gu": "Gujarati", "bn": "Bengali"}

// languageRule tells the LLM which language to answer in; empty for English.
func languageRule(lang string) string {
	if name, ok := LanguageNames[lang]; ok {
		return " Write the answer in " + name + " in its native script, keeping Sanskrit astrology terms recognisable."
	}
	return ""
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
	{"travel", words(`travel|travels|travelling|traveling|abroad|foreign|overseas|relocate|relocation|immigrate|immigration|visa|journey|journeys|pilgrimage`)},
	{"siblings", words(`sibling|siblings|brother|brothers|sister|sisters`)},
	{"property", words(`property|properties|land|plot|real estate|vehicle|vehicles|car|bike|buy a house|own a house|home loan|(?:buy|buying|purchase|purchasing|build|building) (?:a |my |our )?(?:house|home|flat|apartment|land)`)},
	{"spirituality", words(`spiritual|spirituality|moksha|meditation|god|devotion|liberation|retreat`)},
	{"numerology", words(`numerology|numerological|numbers?|life path|mulank|moolank|bhagyank|bhagyaank|destiny number|root number|psychic number|lucky numbers?|personal year|ank jyotish`)},
	{"forecast", words(`when|this year|next year|coming months?|coming year|next few months|future|forecast|predict|prediction|predictions|upcoming|what'?s coming|will i|transit|transits|gochar|gochara|good time|right time|best time|varshaphal|varsha ?phal|varshphal|annual chart|year chart|solar return|my year|birthday year|muntha|year lord`)},
	{"dasha", words(`dasha|dashas|dasa|mahadasha|antardasha|period|periods|timing`)},
	{"yoga", words(`yoga|yogas|raja yoga|gajakesari|kemadruma|kemdrum|kemadrum|adhi yoga|combination|combinations`)},
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

// nativeTopics recognise the main topics in Hinglish and in the app's
// Indian-script languages (Hindi, Marathi, Kannada, Tamil, Telugu,
// Malayalam, Gujarati, Bengali). Indian scripts are matched as substrings:
// regexp word boundaries are ASCII-only.
var nativeTopics = []topic{
	{"safety", regexp.MustCompile(`(?i)\b(maut|marunga|marenge)\b|मृत्यु|मौत|मरण|ಸಾವು|ಮರಣ|மரணம்|சாவு|మరణం|చావు|मृत्यू|മരണം|મૃત્યુ|মৃত্যু`)},
	{"career", regexp.MustCompile(`(?i)\b(naukri|naukari|kaam|vyapar|vyavsay|business|karobar)\b|नौकरी|करियर|कैरियर|व्यवसाय|व्यापार|ಉದ್ಯೋಗ|ಕೆಲಸ|ವೃತ್ತಿ|வேலை|தொழில்|ఉద్యోగం|ఉద్యోగ|వృత్తి|नोकरी|व्यवसाय|ജോലി|തൊഴിൽ|નોકરી|ધંધો|চাকরি|ব্যবসা`)},
	{"marriage", regexp.MustCompile(`(?i)\b(shaadi|shadi|vivah|vivaah|byah|biyah|rishta|patni|pati)\b|शादी|विवाह|ब्याह|रिश्ता|पत्नी|पति|ಮದುವೆ|ವಿವಾಹ|திருமணம்|கல்யாணம்|పెళ్లి|వివాహం|വിവാഹം|കല്യാണം|લગ્ન|বিয়ে|বিবাহ`)},
	{"wealth", regexp.MustCompile(`(?i)\b(paisa|paise|dhan|daulat|kamai)\b|पैसा|पैसे|धन|दौलत|कमाई|ಹಣ|ಸಂಪತ್ತು|பணம்|செல்வம்|డబ్బు|ధనం|संपत्ती|പണം|സമ്പത്ത്|પૈસા|ધન|টাকা|অর্থ|সম্পদ`)},
	{"health", regexp.MustCompile(`(?i)\b(sehat|swasthya|bimari|beemari)\b|स्वास्थ्य|सेहत|बीमारी|ಆರೋಗ್ಯ|ஆரோக்கியம்|உடல்நலம்|ఆరోగ్యం|आरोग्य|ആരോഗ്യം|આરોગ્ય|স্বাস্থ্য`)},
	{"education", regexp.MustCompile(`(?i)\b(padhai|pariksha|shiksha)\b|पढ़ाई|पढाई|शिक्षा|परीक्षा|ಶಿಕ್ಷಣ|ಓದು|ಪರೀಕ್ಷೆ|படிப்பு|கல்வி|தேர்வு|చదువు|విద్య|పరీక్ష|शिक्षण|अभ्यास|പഠനം|പരീക്ഷ|ભણતર|શિક્ષણ|પરીક્ષા|পড়াশোনা|শিক্ষা|পরীক্ষা`)},
	{"children", regexp.MustCompile(`(?i)\b(bachche|bacche|santan|aulad|beta|beti)\b|संतान|बच्चे|बच्चा|औलाद|ಮಕ್ಕಳು|ಸಂತಾನ|குழந்தை|பிள்ளை|పిల్లలు|సంతానం|मुले|മക്കൾ|കുട്ടി|બાળક|સંતાન|সন্তান|বাচ্চা`)},
	{"property", regexp.MustCompile(`(?i)\b(ghar|makaan|makan|zameen|jameen|flat)\b|घर|मकान|ज़मीन|जमीन|ಮನೆ|ಆಸ್ತಿ|வீடு|சொத்து|ఇల్లు|ఆస్తి|मालमत्ता|വീട്|സ്വത്ത്|ઘર|મિલકત|বাড়ি|সম্পত্তি`)},
	{"travel", regexp.MustCompile(`(?i)\b(videsh|bahar|abroad|pardes)\b|विदेश|परदेश|ವಿದೇಶ|வெளிநாடு|విదేశం|विदेशात|വിദേശ|વિદેશ|বিদেশ`)},
	{"mangal_dosha", regexp.MustCompile(`(?i)\b(manglik|mangalik|mangal dosh)\b|मांगलिक|मंगल दोष|ಕುಜ ದೋಷ|ಮಾಂಗಲಿಕ|செவ்வாய் தோஷம்|కుజ దోషం|मंगळ दोष|ചൊവ്വാ ദോഷം|મંગળ દોષ|মাঙ্গলিক`)},
	{"sade_sati", regexp.MustCompile(`(?i)\b(sadhe ?sati|sade ?saati|shani ki sadhe)\b|साढ़ेसाती|साढ़े साती|साडेसाती|ಸಾಡೇಸಾತಿ|ஏழரை சனி|ఏలినాటి శని|ഏഴര ശനി|સાડાસાતી|সাড়ে সাতি`)},
	{"remedy", regexp.MustCompile(`(?i)\b(upay|upaay|totka)\b|उपाय|ಪರಿಹಾರ|பரிகாரம்|పరిహారం|उपाय|പരിഹാരം|ઉપાય|প্রতিকার`)},
	{"dasha", regexp.MustCompile(`दशा|ದಶೆ|ದಶಾ|தசை|దశ|ദശ|દશા|দশা`)},
	{"forecast", regexp.MustCompile(`(?i)\b(kab|kabhi|is saal|agle saal|aane wala|aage kya|bhavishya|bhavishyafal|yaavaga|eppo|eppodhu|eppudu|kevha)\b|कब|इस साल|अगले साल|भविष्य|आने वाले|ಯಾವಾಗ|ಈ ವರ್ಷ|ಮುಂದಿನ ವರ್ಷ|ಭವಿಷ್ಯ|எப்போது|இந்த ஆண்டு|அடுத்த ஆண்டு|எதிர்காலம்|ఎప్పుడు|ఈ సంవత్సరం|వచ్చే సంవత్సరం|భవిష్యత్తు|केव्हा|या वर्षी|पुढच्या वर्षी|എപ്പോൾ|ഈ വർഷം|അടുത്ത വർഷം|ഭാവി|ક્યારે|આ વર્ષે|આવતા વર્ષે|ભવિષ્ય|কবে|এই বছর|আগামী বছর|ভবিষ্যৎ|\b20[2-9][0-9]\b`)},
}

func classify(q string) ([]string, []string, int) {
	var ts []string
	for _, t := range topics {
		if t.words.MatchString(q) {
			ts = append(ts, t.name)
		}
	}
	for _, t := range nativeTopics {
		if !contains(ts, t.name) && t.words.MatchString(q) {
			ts = append(ts, t.name)
		}
	}
	// A safety topic must lead, as in the English list.
	if contains(ts, "safety") && ts[0] != "safety" {
		ts = append([]string{"safety"}, removeString(ts, "safety")...)
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
	if cc.Related == nil {
		cc.Related = s.related(ctx, facts, question)
	}
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
	if s.Compact {
		return s.compactAnswer(ctx, in, facts, question, ts, cc, ans), nil
	}
	for _, rule := range cc.Related {
		in.use(rule)
	}
	ans.Sources = in.used
	prompt, err := chatPrompt(facts, in, question, history, grounded, s.Compact)
	if err == nil {
		prompt += languageRule(cc.Lang)
	}
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

// relatedDistance is the largest cosine distance at which a library passage
// is offered as related to the question (bge-small: on-topic passages measure
// about 0.35–0.45, unrelated ones above 0.5).
const relatedDistance = 0.47

// related returns library passages for this chart's own placements that are
// semantically closest to the question. It never blocks an answer: without a
// semantic index, or if the local model is slow or down, it returns nothing.
func (s *Service) related(ctx context.Context, facts engine.ChartFacts, question string) []Rule {
	sc, ok := s.Corpus.(SemanticCorpus)
	if !ok {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	rs, err := sc.Relevant(ctx, facts, question)
	if err != nil {
		return nil
	}
	var out []Rule
	for _, r := range rs {
		if r.Distance > 0 && r.Distance <= relatedDistance {
			out = append(out, r)
		}
	}
	return out
}

// relatedLines turns semantic hits into plain-language lines. Classical
// passages are cited, never quoted, so only self-authored text is shown.
func relatedLines(in *insight, rs []Rule, max int) []string {
	var out []string
	for _, r := range rs {
		if len(out) == max {
			break
		}
		if classical(r) || in.usedID[r.DocType+":"+r.Key+"|"+r.Source] {
			continue
		}
		in.use(r)
		out = append(out, fmt.Sprintf("%s: %s", r.Title, firstSentence(stripNote(r.Body), 320)))
	}
	return out
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func chatPrompt(f engine.ChartFacts, in *insight, q string, history []ChatTurn, grounded string, compact bool) (string, error) {
	b, err := factsText(f, compact)
	if err != nil {
		return "", err
	}
	var rb strings.Builder
	for _, r := range in.used {
		if strings.Contains(r.Source, "[public_domain]") {
			fmt.Fprintf(&rb, "\nHISTORICAL REFERENCE %s:%s (%s, %s); citation metadata only, use the project's plain-language interpretation.\n", r.DocType, r.Key, r.Source, r.Ref)
			continue
		}
		if compact {
			fmt.Fprintf(&rb, "\nRULE %s:%s: %s\n", r.DocType, r.Key, ruleBody(r.Body, true))
			continue
		}
		fmt.Fprintf(&rb, "\nRULE %s:%s (%s): %s\n", r.DocType, r.Key, r.Source, r.Body)
	}
	var hb strings.Builder
	for _, t := range lastTurns(history, 6) {
		fmt.Fprintf(&hb, "%s: %s\n", strings.ToUpper(t.Role), t.Content)
	}
	return fmt.Sprintf(`You are Astrisk, a friendly Vedic astrology guide for someone new to astrology. Answer in 120-220 words of plain everyday English. Explain Indian terms the first time they appear. Start with the practical meaning and use a concrete reflection question when helpful. Return JSON only: {"answer": string}.
Use only the CHART FACTS (computed by Swiss Ephemeris; never recompute or change a position, house, dasha or yoga) and the RULES. The DRAFT ANSWER is already correct and grounded; improve its clarity and relevance to the question, keep every fact in it, and add nothing that the facts or rules do not support. Do not treat a chart as evidence about the person's real life. Classical passages marked public_domain may be old-fashioned or harmful: explain a humane present-day theme in your own words; do not quote them. Never predict death, illness, divorce or financial ruin. End with: "For reflection, not certainty."
CHART FACTS: %s
RULES:%s
CONVERSATION SO FAR:
%s
QUESTION: %s
DRAFT ANSWER: %s`, b, rb.String(), hb.String(), q, grounded), nil
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
				add(firstSentence(in.entry(engine.DocGrahaInHouse, id+"_in_10"), 300))
				add("Chart position details: " + in.placement(id) + ".")
			}
			if lord := engine.HouseLord(f.Chart, 10); in.house(lord) != 10 {
				add("A traditional career indicator is " + engine.GrahaEnglish(lord) + ". " + firstSentence(in.entry(engine.DocGrahaInHouse, fmt.Sprintf("%s_in_%d", lord, in.house(lord))), 280))
				add("Chart position details: " + in.placement(lord) + ".")
			}
			add(careerYogas(in))
			add(vargaLine(in, 10, "sun", "saturn"))
		case "marriage":
			add(in.houseSummary(7))
			for _, id := range engine.GrahasInHouse(f.Chart, 7) {
				add(firstSentence(in.entry(engine.DocGrahaInHouse, id+"_in_7"), 300))
				add("Chart position details: " + in.placement(id) + ".")
			}
			add("Venus is one traditional symbol for affection and partnership. " + firstSentence(in.entry(engine.DocGrahaInHouse, fmt.Sprintf("venus_in_%d", in.house("venus"))), 260))
			add("Chart position details: " + in.placement("venus") + ".")
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
			add("Astrology cannot assess or diagnose health. Please ask a qualified doctor about health concerns.")
		case "education":
			add(in.houseSummary(4))
			add(in.houseSummary(5))
			add("Mercury and Jupiter are traditionally used to reflect on learning and teaching. These are symbolic associations, not measures of ability.")
			add("Chart position details: Mercury — " + in.placement("mercury") + "; Jupiter — " + in.placement("jupiter") + ".")
		case "children":
			add(in.houseSummary(5))
			add("Some astrological traditions include Jupiter as a symbol when discussing family. A chart cannot predict fertility or a child's health; consult a qualified professional for those questions.")
		case "mangal_dosha":
			add(mangal(in))
		case "sade_sati":
			add(sadeSati(in, cc))
		case "dasha":
			add(in.dashaLine())
			add(in.entry(engine.DocDasha, "dasha_"+f.Vimshottari.Current.Maha))
			if a := f.Vimshottari.Current.Antara; a != "" && a != f.Vimshottari.Current.Maha {
				add(fmt.Sprintf("Within the longer period, a shorter phase associated with %s is also considered: %s", engine.GrahaEnglish(a), firstSentence(in.entry(engine.DocDasha, "dasha_"+a), 240)))
			}
			if p := f.Vimshottari.Current.Pratyantara; p != "" {
				add(fmt.Sprintf("The finer pratyantardasha running now is %s.", engine.GrahaEnglish(p)))
			}
			if y := f.Yogini.Current; y.Yogini != "" {
				add(fmt.Sprintf("A second traditional timing method, called Yogini dasha, assigns this phase to %s until %s.", engine.GrahaEnglish(y.Lord), y.To))
			}
		case "yoga":
			if len(f.Yogas) == 0 {
				add("The engine detects none of the yogas in its catalogue for this chart.")
			}
			for _, y := range f.Yogas {
				m, _ := in.yogaMeaning(y)
				line := fmt.Sprintf("%s (%s): %s", y.Name, y.Strength, m)
				add(line)
			}
		case "nakshatra":
			moon := in.grahas["moon"]
			add(fmt.Sprintf("Your birth star, called nakshatra in this tradition, is %s (quarter %d).", moon.Nakshatra, moon.NakshatraPada))
			add(in.entry(engine.DocNakshatra, engine.Slug(moon.Nakshatra)))
		case "lagna":
			lagna := in.signs[in.lagnaSign()]
			add(fmt.Sprintf("Your rising sign (ascendant, or Lagna) is %s. It is the sign on the eastern horizon at birth.", lagna))
			add(in.entry(engine.DocBhava, "lagna_"+engine.Slug(lagna)))
			lord := engine.HouseLord(f.Chart, 1)
			add(fmt.Sprintf("The lagna lord %s: %s.", engine.GrahaEnglish(lord), in.placement(lord)))
		case "moon_sign":
			add("Your Moon sign (rashi): " + in.placement("moon") + ".")
			add(in.entry(engine.DocGrahaInSign, "moon_in_"+engine.Slug(in.signs[in.signIdx("moon")])))
		case "strength":
			add(strengthLine(cc))
		case "numerology":
			add(numerologyLine(in, cc))
		case "travel":
			add(in.houseSummary(9))
			add(in.houseSummary(12))
			add("Rahu is traditionally linked with foreign people and places: " + in.placement("rahu") + ". Astrology cannot tell you whether a visa or a move will happen; use these as prompts for what you want from a journey.")
		case "siblings":
			add(in.houseSummary(3))
			add("Mars is the traditional significator (karaka) of siblings: " + in.placement("mars") + ".")
		case "property":
			add(in.houseSummary(4))
			add("Mars is the traditional significator of land and Venus of vehicles and comforts. Mars: " + in.placement("mars") + "; Venus: " + in.placement("venus") + ". For a purchase, rely on your budget and qualified advice.")
		case "spirituality":
			add(in.houseSummary(12))
			add(in.houseSummary(9))
			add("Jupiter and Ketu are the traditional significators of wisdom and detachment. Jupiter: " + in.placement("jupiter") + "; Ketu: " + in.placement("ketu") + ".")
		case "forecast":
			for _, l := range forecastDetails(cc) {
				add(l)
			}
		case "remedy":
			add("Astrisk describes the chart rather than prescribing remedies. Traditionally, strengthening a graha begins with its significations: for the current dasha lord " + engine.GrahaEnglish(f.Vimshottari.Current.Maha) + ", that means living its qualities consciously. For specific remedies such as gemstones or rituals, consult a trusted astrologer who can see the whole chart.")
		}
	}
	for _, id := range gs {
		if contains(ts, "marriage") && id == "venus" || contains(ts, "moon_sign") && id == "moon" {
			continue
		}
		add(in.grahaMeaning(id))
		add("Chart position details: " + in.placement(id) + ".")
	}
	if house > 0 && !contains(ts, "career") && !contains(ts, "marriage") {
		add(in.houseSummary(house))
	}
	if len(parts) == 0 {
		if rel := relatedLines(in, cc.Related, 2); len(rel) > 0 {
			ts = append(ts, "library")
			add("These are the passages from the Astrisk library that are closest to your question, chosen only from placements in your own chart.")
			for _, l := range rel {
				add(l)
			}
		}
	} else if rel := relatedLines(in, cc.Related, 1); len(rel) > 0 {
		add("Also related to your question in your chart — " + rel[0])
	}
	if len(parts) == 0 {
		ts = append(ts, "overview")
		add(fmt.Sprintf("Here is the core of your chart. Ascendant %s; Moon in %s, %s nakshatra; Sun in %s.", in.signs[in.lagnaSign()], in.grahas["moon"].Rashi, in.grahas["moon"].Nakshatra, in.grahas["sun"].Rashi))
		if y := in.yogaNames(); len(y) > 0 {
			add("Detected yogas: " + strings.Join(y, ", ") + ".")
		}
		add(in.dashaLine())
		add("You can ask about career, marriage, wealth, education, children, health, travel abroad, siblings, property, spirituality, your nakshatra, a planet (for example \"What does my Saturn mean?\"), a house, your yogas, Mangal dosha, Sade Sati, your current dasha, your planetary strength (Shadbala) or how your numerology connects with your chart.")
	}
	return plainFor(in, ts, gs, house, cc).render(dedupe(parts)), ts
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
			m, _ := in.yogaMeaning(y)
			names = append(names, y.Name+": "+m)
		}
	}
	if len(names) == 0 {
		return ""
	}
	return "The chart has patterns traditionally called yogas (combinations). Some schools associate these with work and achievement; use them as reflection prompts, not forecasts: " + strings.Join(names, "; ") + "."
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

var thinkBlock = regexp.MustCompile(`(?s)<think>.*?</think>`)

// compactAnswer asks the local fine-tuned model, which reads compactChart
// lines and answers in the plain format. Its answer is used only if it
// passes the same checks the model was evaluated on; the grounded answer's
// chart details are kept below it. Otherwise the grounded answer stands.
func (s *Service) compactAnswer(ctx context.Context, in *insight, f engine.ChartFacts, question string, ts []string, cc ChatContext, grounded ChatAnswer) ChatAnswer {
	cl, ok := s.LLM.(ChatLLM)
	if !ok || (cc.Lang != "" && cc.Lang != "en") || contains(ts, "forecast") {
		// The local model answers in English only, and its chart lines carry
		// no forecast, so timing questions keep the grounded answer.
		return grounded
	}
	raw, err := cl.Chat(ctx, CompactSystem, "CHART\n"+compactChart(in, f, cc)+"\nQUESTION: "+question)
	if err != nil {
		return grounded
	}
	text := strings.TrimSpace(thinkBlock.ReplaceAllString(raw, ""))
	if !strings.HasPrefix(text, headShort) || len(ScoreText(f, ts, text)) > 0 {
		return grounded
	}
	if i := strings.Index(grounded.Answer, "\n\n"+headDetails+"\n"); i >= 0 {
		details := strings.TrimSuffix(grounded.Answer[i:], "\n\nFor reflection, not certainty.")
		text = strings.TrimSuffix(text, "For reflection, not certainty.")
		text = strings.TrimSpace(text) + details + "\n\nFor reflection, not certainty."
	}
	grounded.Answer, grounded.Model = text, s.modelName()
	return grounded
}

func removeString(xs []string, x string) []string {
	var out []string
	for _, y := range xs {
		if y != x {
			out = append(out, y)
		}
	}
	return out
}
