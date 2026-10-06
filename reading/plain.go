package reading

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/example/panchang/corpus"
	"github.com/example/panchang/engine"
)

// The plain layer answers first in everyday words: a short answer, then a
// few practical points. The technical grounding (placements, classical
// terms, divisional charts) follows under "Chart details", which the app
// shows folded. Headings are fixed strings so the app can find them in
// stored chat text; answers without them display as before.
const (
	headShort   = "In short"
	headPoints  = "What this means for you"
	headDetails = "Chart details"
)

// plain is the everyday-language part of an answer.
type plain struct {
	short     []string
	points    []string
	bookNoted bool
}

func (p *plain) say(s string) {
	if s = strings.TrimSpace(s); s != "" && len(p.short) < 2 {
		p.short = append(p.short, s)
	}
}

func (p *plain) point(s string) {
	if s = strings.TrimSpace(s); s != "" && len(p.points) < 6 && !contains(p.points, s) {
		p.points = append(p.points, s)
	}
}

// render joins the plain layer and the technical details into one answer.
func (p plain) render(details []string) string {
	var b strings.Builder
	b.WriteString(headShort + "\n" + strings.Join(p.short, " "))
	if len(p.points) > 0 {
		b.WriteString("\n\n" + headPoints)
		for _, x := range p.points {
			b.WriteString("\n• " + x)
		}
	}
	if len(details) > 0 {
		b.WriteString("\n\n" + headDetails + "\n" + strings.Join(details, "\n\n"))
	}
	b.WriteString("\n\nFor reflection, not certainty.")
	return b.String()
}

// What each graha stands for, in everyday words.
var grahaGist = map[string]string{
	"sun": "confidence and sense of self", "moon": "mind and feelings", "mars": "energy and drive",
	"mercury": "thinking and communication", "jupiter": "wisdom, learning and growth", "venus": "love, comfort and beauty",
	"saturn": "discipline, patience and hard work", "rahu": "ambition and new, unusual paths", "ketu": "detachment and the inner life",
}

// What each house (area of the chart) is about.
var houseArea = map[int]string{
	1: "self, body and how you meet the world", 2: "money, speech and family", 3: "courage, effort and siblings",
	4: "home, mother and inner peace", 5: "children, creativity and studies", 6: "daily work, effort against obstacles and health habits",
	7: "marriage and partnerships", 8: "change, research and shared resources", 9: "luck, teachers, faith and long journeys",
	10: "career and public life", 11: "income, gains and friends", 12: "rest, spending and faraway places",
}

var westernName = map[string]string{
	"Mesha": "Aries", "Vrishabha": "Taurus", "Mithuna": "Gemini", "Karka": "Cancer", "Simha": "Leo", "Kanya": "Virgo",
	"Tula": "Libra", "Vrishchika": "Scorpio", "Dhanu": "Sagittarius", "Makara": "Capricorn", "Kumbha": "Aquarius", "Meena": "Pisces",
}

func signPlain(rashi string) string {
	if w := westernName[rashi]; w != "" {
		return rashi + " (" + w + ")"
	}
	return rashi
}

func gname(id string) string { return engine.GrahaEnglish(id) }

// lifeArea names a house by its life theme: "the career and public life part of your chart".
func lifeArea(h int) string {
	return fmt.Sprintf("the %s part of your chart (%s house)", houseArea[h], ordinal(h))
}

// condition says in plain words how comfortable a graha is where it sits.
func (in *insight) condition(id string) string {
	var bits []string
	switch d := in.f.Dignities[id]; d.State {
	case "exalted":
		bits = append(bits, "at its strongest")
	case "own":
		bits = append(bits, "comfortable, in its own sign")
	case "friendly":
		bits = append(bits, "in a friendly sign")
	case "enemy":
		bits = append(bits, "in a less comfortable sign")
	case "debilitated":
		if d.NeechaBhanga {
			bits = append(bits, "in its weakest sign, though another factor in the chart offsets that")
		} else {
			bits = append(bits, "in its weakest sign, so its qualities need more conscious effort")
		}
	}
	if in.f.Combustion[id] {
		bits = append(bits, "close to the Sun, which traditionally mutes it")
	}
	for _, r := range in.f.Retrograde {
		if r == id && id != "rahu" && id != "ketu" {
			bits = append(bits, "retrograde (appearing to move backwards), often read as working in a more inward way")
		}
	}
	return strings.Join(bits, " and ")
}

// where describes one graha in plain words: what it stands for, which part
// of life it sits in, its sign, and how comfortable it is there.
func (in *insight) where(id string) string {
	s := fmt.Sprintf("%s (%s) is in %s, in %s", upper(theName(id)), grahaGist[id], lifeArea(in.house(id)), signPlain(in.signs[in.signIdx(id)]))
	if c := in.condition(id); c != "" {
		s += ", " + c
	}
	return s + "."
}

// theName adds "the" where English needs it: "the Sun", "the Moon", "Mars".
func theName(id string) string {
	if id == "sun" || id == "moon" {
		return "the " + gname(id)
	}
	return gname(id)
}

// joinAnd joins items as "a, b and c".
func joinAnd(xs []string) string {
	switch len(xs) {
	case 0:
		return ""
	case 1:
		return xs[0]
	}
	return strings.Join(xs[:len(xs)-1], ", ") + " and " + xs[len(xs)-1]
}

// day writes an engine date (2035-06-14) as "14 June 2035".
func day(iso string) string {
	if t, err := time.Parse("2006-01-02", iso); err == nil {
		return t.Format("2 January 2006")
	}
	return iso
}

// houseStory answers "what about this area of life?" in one sentence.
func (in *insight) houseStory(h int) string {
	occ := engine.GrahasInHouse(in.f.Chart, h)
	if len(occ) > 0 {
		verb := "sits"
		if len(occ) > 1 {
			verb = "sit"
		}
		named := make([]string, len(occ))
		for i, id := range occ {
			named[i] = theName(id) + " (" + grahaGist[id] + ")"
		}
		return fmt.Sprintf("%s %s in %s, so %s shape this area of your life.", upper(joinAnd(named)), verb, lifeArea(h), map[bool]string{true: "these qualities", false: "its qualities"}[len(occ) > 1])
	}
	lord := engine.HouseLord(in.f.Chart, h)
	return fmt.Sprintf("No planet sits in %s. That does not mean anything is missing; its ruling planet, %s, is read instead, and it is placed in %s.", lifeArea(h), gname(lord), lifeArea(in.house(lord)))
}

func upper(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (in *insight) periodNow() string {
	d := in.f.Vimshottari.Current
	if d.Maha == "" {
		return ""
	}
	until := d.To
	for _, p := range in.f.Vimshottari.Sequence {
		if p.Lord == d.Maha && p.To > d.From {
			until = p.To
			break
		}
	}
	return fmt.Sprintf("You are in %s's main period (mahadasha) until %s, a time traditionally about %s.", theName(d.Maha), day(until), grahaGist[d.Maha])
}

func yogaStrength(s string) string {
	switch s {
	case "strong":
		return "clearly formed"
	case "weak":
		return "only partly formed"
	}
	return s
}

// book returns what the classical text says for a token, in plain words,
// with its citation ("Brihat Jataka 20.3"); empty when no book covers it.
func (in *insight) book(doc, key string) string {
	for _, r := range in.rules[engine.CorpusKey{DocType: doc, Key: key}] {
		if !classical(r) {
			continue
		}
		i := strings.Index(r.Body, corpus.PlainLabel)
		if i < 0 {
			continue
		}
		in.use(r)
		plain := strings.TrimSuffix(strings.TrimSpace(r.Body[i+len(corpus.PlainLabel):]), "]")
		ref := "Brihat Jataka"
		if r.Ref != "" {
			ref += " " + r.Ref
		}
		return fmt.Sprintf("Classical view (%s): %s", ref, plain)
	}
	return ""
}

// bookPoint adds the classical view, with a one-time reminder of how to read
// it: old texts state results as certain.
func (p *plain) bookPoint(s string) {
	if s == "" {
		return
	}
	if !p.bookNoted {
		s += " Old texts state results as certain; read them as traditional tendencies."
		p.bookNoted = true
	}
	p.point(s)
}

// plainFor writes the everyday-language answer for the detected topics.
func plainFor(in *insight, ts, gs []string, house int, cc ChatContext) plain {
	if contains(ts, "forecast") {
		return forecastPlain(in, ts, cc)
	}
	var p plain
	f := in.f
	for _, t := range ts {
		switch t {
		case "career":
			p.say(in.houseStory(10))
			for _, id := range engine.GrahasInHouse(f.Chart, 10) {
				p.point(in.where(id))
				p.bookPoint(in.book(engine.DocGrahaInHouse, fmt.Sprintf("%s_in_10", id)))
			}
			for _, y := range f.Yogas {
				if y.Type == "raja" || y.Type == "mahapurusha" {
					p.point(fmt.Sprintf("You have %s, a pattern (%s) traditionally linked with achievement. %s", y.Name, yogaStrength(y.Strength), firstSentence(in.entry(engine.DocYoga, engine.Slug(y.Name)), 160)))
					break
				}
			}
			p.point(in.periodNow())
			p.point("Try this: pick one skill that fits these themes and practise it for the next few months.")
		case "marriage":
			p.say(in.houseStory(7))
			for _, id := range engine.GrahasInHouse(f.Chart, 7) {
				p.bookPoint(in.book(engine.DocGrahaInHouse, fmt.Sprintf("%s_in_7", id)))
			}
			p.point(in.where("venus"))
			if ok, h := engine.MangalDosha(f.Chart); ok {
				p.point(fmt.Sprintf("Mangal dosha is present by the usual rule (Mars in your %s house). Matching also checks the partner's chart, so this is one factor, not a verdict.", ordinal(h)))
			} else {
				p.point("You do not have Mangal dosha by the usual rule.")
			}
			p.point("Try this: think about what you most want to give and receive in a partnership.")
		case "wealth":
			p.say(in.houseStory(2))
			p.say(in.houseStory(11))
			for _, y := range f.Yogas {
				if y.Type == "wealth" {
					p.point(fmt.Sprintf("You have %s (%s), a pattern traditionally linked with prosperity.", y.Name, yogaStrength(y.Strength)))
				}
			}
			p.point("Astrology cannot forecast returns; for money decisions, rely on a budget and a qualified adviser.")
		case "health":
			p.say("Astrology cannot judge or predict health. For any health concern, please see a doctor.")
			p.point(in.houseStory(6))
		case "education":
			p.say(in.houseStory(5))
			p.point(in.where("mercury"))
			p.point(in.where("jupiter"))
		case "children":
			p.say(in.houseStory(5))
			p.point(in.where("jupiter"))
			p.point("A chart cannot predict fertility or a child's health; please ask a doctor about those.")
		case "mangal_dosha":
			ok, h := engine.MangalDosha(f.Chart)
			if !ok {
				p.say(fmt.Sprintf("No, you do not have Mangal dosha by the usual rule. Mars is in your %s house, and the dosha needs Mars in the 1st, 2nd, 4th, 7th, 8th or 12th.", ordinal(h)))
			} else {
				p.say(fmt.Sprintf("Yes, Mangal dosha is present by the usual rule: Mars is in your %s house.", ordinal(h)))
				if st := f.Dignities["mars"].State; st == "own" || st == "exalted" {
					p.point("Mars is strong in its sign here, which many traditions count as cancelling the dosha.")
				}
				p.point("It is only one factor in matching; the partner's chart matters too, and many couples with it are happily married.")
			}
		case "sade_sati":
			if cc.Transit != nil {
				on, phase := engine.SadeSati(f.Chart, *cc.Transit)
				if on {
					p.say(fmt.Sprintf("Yes, you are in Sade Sati, in its %s phase. It is Saturn's 7½-year passage over your Moon sign.", map[int]string{1: "first", 2: "middle", 3: "last"}[phase]))
					p.point("Traditionally it is a time for responsibility, patience and steady effort, not misfortune.")
				} else {
					p.say("No, you are not in Sade Sati (Saturn's 7½-year passage over your Moon sign) right now.")
				}
			}
		case "dasha":
			p.say(in.periodNow())
			if a := f.Vimshottari.Current.Antara; a != "" && a != f.Vimshottari.Current.Maha {
				p.point(fmt.Sprintf("Inside it, a shorter %s phase runs until %s, adding %s.", gname(a), day(f.Vimshottari.Current.To), grahaGist[a]))
			}
			p.point(in.where(f.Vimshottari.Current.Maha))
			p.bookPoint(in.book(engine.DocDasha, "dasha_"+f.Vimshottari.Current.Maha))
			if u := f.Vimshottari.Upcoming; u.Lord != "" {
				p.point(fmt.Sprintf("Next comes %s's main period, from %s.", theName(u.Lord), day(u.From)))
			}
		case "yoga":
			if len(f.Yogas) == 0 {
				p.say("Your chart has none of the special patterns (yogas) that Astrisk checks for. That is common and says nothing bad about the chart.")
			} else {
				p.say(fmt.Sprintf("Your chart has %d special pattern%s (yogas): %s.", len(f.Yogas), map[bool]string{true: "s", false: ""}[len(f.Yogas) > 1], strings.Join(in.yogaNames(), ", ")))
				for _, y := range f.Yogas {
					p.point(fmt.Sprintf("%s, %s: %s", y.Name, yogaStrength(y.Strength), firstSentence(in.entry(engine.DocYoga, engine.Slug(y.Name)), 150)))
					p.bookPoint(in.book(engine.DocYoga, engine.Slug(y.Name)))
				}
			}
		case "nakshatra":
			moon := in.grahas["moon"]
			p.say(fmt.Sprintf("Your birth star (nakshatra) is %s, part %d of 4. It is the star the Moon was in when you were born.", moon.Nakshatra, moon.NakshatraPada))
			p.point(firstSentence(in.entry(engine.DocNakshatra, engine.Slug(moon.Nakshatra)), 220))
			p.bookPoint(in.book(engine.DocNakshatra, engine.Slug(moon.Nakshatra)))
		case "lagna":
			lagna := in.signs[in.lagnaSign()]
			p.say(fmt.Sprintf("Your rising sign (lagna) is %s. It is the sign that was rising in the east when you were born, and it colours how you meet the world.", signPlain(lagna)))
			p.point(firstSentence(in.entry(engine.DocBhava, "lagna_"+engine.Slug(lagna)), 220))
			p.point("Its ruling planet: " + in.where(engine.HouseLord(f.Chart, 1)))
		case "moon_sign":
			p.say(fmt.Sprintf("Your Moon sign (rashi) is %s. It describes your mind and feelings.", signPlain(in.signs[in.signIdx("moon")])))
			p.point(firstSentence(in.entry(engine.DocGrahaInSign, "moon_in_"+engine.Slug(in.signs[in.signIdx("moon")])), 220))
			p.bookPoint(in.book(engine.DocGrahaInSign, "moon_in_"+engine.Slug(in.signs[in.signIdx("moon")])))
			p.point(in.where("moon"))
		case "strength":
			if cc.Shadbala != nil && len(cc.Shadbala.Rows) > 0 {
				best, worst := cc.Shadbala.Rows[0], cc.Shadbala.Rows[0]
				for _, r := range cc.Shadbala.Rows {
					if r.Rank < best.Rank {
						best = r
					}
					if r.Rank > worst.Rank {
						worst = r
					}
				}
				p.say(fmt.Sprintf("Your strongest planet is %s (%s) and the one needing most support is %s (%s), by the classical six-part strength score (Shadbala).", theName(best.Graha), grahaGist[best.Graha], theName(worst.Graha), grahaGist[worst.Graha]))
				p.point(in.where(best.Graha))
				p.point(in.where(worst.Graha))
			}
		case "numerology":
			if n := cc.Numerology; n != nil {
				p.say(fmt.Sprintf("Your life path number is %d and your root number (mulank) is %d, which is linked with %s.", n.LifePath.Number, n.Mulank, theName(n.MulankGraha)))
				p.point(in.where(n.MulankGraha))
			}
		case "travel":
			p.say(in.houseStory(12))
			p.point(in.houseStory(9))
			p.point(in.where("rahu"))
			p.point("A chart cannot tell whether a visa or move will happen; use it to think about what you want from travel.")
		case "siblings":
			p.say(in.houseStory(3))
			p.point(in.where("mars"))
		case "property":
			p.say(in.houseStory(4))
			p.point(in.where("mars"))
			p.point(in.where("venus"))
		case "spirituality":
			p.say(in.houseStory(12))
			p.point(in.where("jupiter"))
			p.point(in.where("ketu"))
		case "remedy":
			p.say("Astrisk does not prescribe remedies. The simplest traditional one is to live the good side of the planet whose period you are in.")
			p.point(in.periodNow())
			p.point("For gemstones or rituals, ask a trusted astrologer who can see your whole chart.")
		}
	}
	for _, id := range gs {
		p.say(in.where(id))
		p.point(firstSentence(in.grahaMeaning(id), 220))
		p.bookPoint(in.book(engine.DocGrahaInHouse, fmt.Sprintf("%s_in_%d", id, in.house(id))))
	}
	if house > 0 {
		p.say(in.houseStory(house))
	}
	if len(p.short) == 0 {
		p.say(fmt.Sprintf("Here is the core of your chart: your rising sign is %s and your Moon sign is %s.", signPlain(in.signs[in.lagnaSign()]), signPlain(in.signs[in.signIdx("moon")])))
		p.point(in.periodNow())
		if names := in.yogaNames(); len(names) > 0 {
			p.point("Special patterns (yogas) in your chart: " + strings.Join(names, ", ") + ".")
		}
		p.point("You can ask about career, marriage, money, studies, children, health, travel, a planet, a house, your dasha, Mangal dosha, Sade Sati or your numerology.")
	}
	return p
}

// scoreBand says what an Ashtakoota total means in the usual reading.
func scoreBand(total float64) string {
	switch {
	case total < 18:
		return "below the usual minimum of 18"
	case total < 25:
		return "acceptable by the usual rule (18 or more)"
	case total < 33:
		return "a good match by the usual rule"
	}
	return "an excellent match by the usual rule"
}

// matchPlain writes the everyday-language answer for a matching question.
func matchPlain(mc MatchChatContext, boy, girl *insight, ts []string) plain {
	var p plain
	m := mc.Match
	yesNo := map[bool]string{true: "yes", false: "no"}
	score := func() {
		p.say(fmt.Sprintf("Your charts score %g out of %g in traditional eight-part matching (Ashtakoota), which is %s.", m.Total, m.Max, scoreBand(m.Total)))
		var full, none []string
		for _, k := range m.Kootas {
			switch k.Score {
			case k.Max:
				full = append(full, k.Name)
			case 0:
				none = append(none, k.Name)
			}
		}
		if len(full) > 0 {
			p.point("Full points in: " + joinAnd(full) + ".")
		}
		if len(none) > 0 {
			p.point("No points in: " + joinAnd(none) + ". Ask about any of these by name to see what it compares.")
		}
		p.point("The score is a traditional table, not a chance of a happy marriage.")
	}
	doshas := func() {
		if len(m.Doshas) == 0 {
			p.say("No warning flags (doshas) were found in the matching tables.")
			return
		}
		p.say(fmt.Sprintf("The matching tables raise %d warning flag%s (dosha): %s.", len(m.Doshas), map[bool]string{true: "s", false: ""}[len(m.Doshas) > 1], joinAnd(m.Doshas)))
		if len(m.Exceptions) > 0 {
			p.point("A traditional exception applies, which many families treat as cancelling the flag.")
		}
		p.point("A dosha is a flag in a table, not a prediction about the two of you.")
	}
	for _, t := range ts {
		switch t {
		case "score":
			score()
		case "dosha":
			doshas()
		case "mangal":
			p.say(fmt.Sprintf("Mangal dosha: groom %s, bride %s.", yesNo[m.BoyMangal], yesNo[m.GirlMangal]))
			if m.BoyMangal == m.GirlMangal {
				p.point("When both or neither have it, the tradition treats the pair as balanced.")
			}
			p.point("A Mars flag says nothing about anyone's temper or safety.")
		case "moon":
			p.say(fmt.Sprintf("The groom's Moon sign is %s and the bride's is %s. Matching is built on these two Moons and their birth stars.", signPlain(m.BoyMoon), signPlain(m.GirlMoon)))
		case "seventh":
			p.say("The 7th house is the partnership part of each chart.")
			p.point("Groom: " + boy.houseStory(7))
			p.point("Bride: " + girl.houseStory(7))
		case "timing":
			p.say(fmt.Sprintf("The groom is in %s's main period and the bride in %s's. These describe each person's own chapter of life; they do not predict when or whether a marriage happens.", theName(boy.f.Vimshottari.Current.Maha), theName(girl.f.Vimshottari.Current.Maha)))
		case "numerology":
			if mc.Boy.Numerology != nil && mc.Girl.Numerology != nil {
				p.say(fmt.Sprintf("Your root numbers (mulank) are %d, linked with %s, and %d, linked with %s.", mc.Boy.Numerology.Mulank, theName(mc.Boy.Numerology.MulankGraha), mc.Girl.Numerology.Mulank, theName(mc.Girl.Numerology.MulankGraha)))
			}
		case "remedy":
			p.say("Astrisk does not prescribe remedies. If they matter to your families, a trusted astrologer can look at both full charts; an open conversation about expectations matters most.")
		case "muhurta":
			p.say("Matching compares the two charts; it does not pick a wedding date. Use the Muhurta page and choose Marriage for that.")
		default:
			if kn, ok := kootaTopic[t]; ok {
				for _, k := range m.Kootas {
					if k.Name == kn {
						p.say(fmt.Sprintf("%s gives %g of %g points. %s", k.Name, k.Score, k.Max, firstSentence(k.Description, 200)))
					}
				}
			}
		}
	}
	if len(p.short) == 0 {
		score()
		doshas()
		p.point(fmt.Sprintf("Mangal dosha: groom %s, bride %s.", yesNo[m.BoyMangal], yesNo[m.GirlMangal]))
	}
	return p
}

// BookView is what a classical book says for one condition, in plain words.
type BookView struct {
	Ref    string `json:"ref"`
	Source string `json:"source"`
	Plain  string `json:"plain"`
}

// BookViews returns the classical passage, in plain words, for each token
// that a public-domain book covers (keyed "doc:key").
func (s *Service) BookViews(ctx context.Context, keys []engine.CorpusKey) map[string]BookView {
	out := map[string]BookView{}
	if s.Corpus == nil {
		return out
	}
	rules, err := s.Corpus.RulesFor(ctx, keys)
	if err != nil {
		return out
	}
	for _, r := range rules {
		k := r.DocType + ":" + r.Key
		i := strings.Index(r.Body, corpus.PlainLabel)
		if _, done := out[k]; done || !classical(r) || i < 0 {
			continue
		}
		out[k] = BookView{Ref: r.Ref, Source: cite(r), Plain: strings.TrimSuffix(strings.TrimSpace(r.Body[i+len(corpus.PlainLabel):]), "]")}
	}
	return out
}
