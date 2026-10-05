package reading

import (
	"fmt"
	"strings"

	"github.com/example/panchang/divination"
	"github.com/example/panchang/engine"
)

// Numerology and the kundali are separate symbolic systems. Indian numerology
// (ank jyotish) gives each root number a ruling graha, which is the bridge used
// here: the number's graha is looked up in the computed chart. Nothing here
// changes a chart fact or a number; it only reports where they echo each other.

var vimshottariOrder = []string{"ketu", "venus", "sun", "moon", "mars", "rahu", "jupiter", "saturn", "mercury"}

func nakshatraLord(name string) string {
	if i := engine.NakshatraIndex(name); i >= 0 {
		return vimshottariOrder[i%9]
	}
	return ""
}

// friendshipProxy maps the shadow grahas to the classical stand-ins
// ("Shanivat Rahu, Kujavat Ketu": Rahu acts like Saturn, Ketu like Mars).
func friendshipProxy(id string) string {
	switch id {
	case "rahu":
		return "saturn"
	case "ketu":
		return "mars"
	}
	return id
}

// NumberGrahaRelation describes how two numerology ruling grahas relate in
// the natural friendship table (Rahu read as Saturn, Ketu as Mars).
func NumberGrahaRelation(a, b string) (string, int) { return grahaRelation(a, b) }

// grahaRelation describes the natural friendship of two grahas both ways.
func grahaRelation(a, b string) (string, int) {
	if a == b {
		return "the same graha", 2
	}
	pa, pb := friendshipProxy(a), friendshipProxy(b)
	if pa == pb {
		return "treated alike (Rahu is read like Saturn and Ketu like Mars)", 2
	}
	ra, rb := engine.NaturalRelation(pa, pb), engine.NaturalRelation(pb, pa)
	switch {
	case ra == 2 && rb == 2:
		return "natural friends", 2
	case ra == 0 && rb == 0:
		return "natural enemies", 0
	case ra+rb >= 3:
		return "friendly one way and neutral the other", 2
	case ra == 1 && rb == 1:
		return "neutral to each other", 1
	default:
		return "mixed (one regards the other as an enemy)", 0
	}
}

func (in *insight) shadbalaFor(cc ChatContext, id string) string {
	if cc.Shadbala == nil {
		return ""
	}
	for _, r := range cc.Shadbala.Rows {
		if r.Graha == id {
			verdict := "meets its classical requirement"
			if r.Ratio < 1 {
				verdict = "is below its classical requirement"
			}
			return fmt.Sprintf("its Shadbala strength is %.0f%% of the required amount, so it %s", r.Ratio*100, verdict)
		}
	}
	return ""
}

// grahaEchoes lists where a graha already carries weight in the chart; whose
// is the possessive used in the text ("your", "the bride's").
func (in *insight) grahaEchoes(id, whose string) []string {
	f := in.f
	var out []string
	if engine.HouseLord(f.Chart, 1) == id {
		out = append(out, "it rules "+whose+" ascendant (lagna lord)")
	}
	if engine.SignLord(in.signIdx("moon")) == id {
		out = append(out, "it rules "+whose+" Moon sign")
	}
	if nakshatraLord(in.grahas["moon"].Nakshatra) == id {
		out = append(out, "it rules "+whose+" birth nakshatra")
	}
	cur := f.Vimshottari.Current
	if cur.Maha == id {
		out = append(out, whose+" current mahadasha (main period) is its")
	} else if cur.Antara == id {
		out = append(out, whose+" current antardasha (sub-period) is its")
	}
	if d, ok := f.Dignities[id]; ok && (d.State == "exalted" || d.State == "own") {
		out = append(out, map[string]string{"exalted": "it is exalted", "own": "it is in its own sign"}[d.State])
	}
	if h := in.house(id); h == 1 || h == 4 || h == 7 || h == 10 {
		out = append(out, fmt.Sprintf("it sits in an angular (kendra) house, the %s", ordinal(h)))
	}
	return out
}

// numerologyLine correlates the birth-date numbers with the kundali.
func numerologyLine(in *insight, cc ChatContext) string {
	n := cc.Numerology
	if n == nil {
		return "Numerology needs your birth date, which is unavailable for this chart."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "From your birth date, numerology (Pythagorean method) gives life path %d, birthday number %d and personal year %d for %d. ", n.LifePath.Number, n.Birthday.Number, n.PersonalYear.Number, n.Year)
	fmt.Fprintf(&b, "Indian numerology (ank jyotish) reduces these to a root number (mulank) of %d, ruled by %s, and a destiny number (bhagyank) of %d, ruled by %s.", n.Mulank, engine.GrahaEnglish(n.MulankGraha), n.Bhagyank, engine.GrahaEnglish(n.BhagyankGraha))
	themes, _ := divination.RootMeaning(n.Mulank)
	b.WriteString(" " + themes)

	b.WriteString("\n\nHow this connects with your kundali:")
	echoed := 0
	for i, id := range []string{n.MulankGraha, n.BhagyankGraha} {
		if i == 1 && id == n.MulankGraha {
			b.WriteString(" Both numbers point to the same graha, so its theme is doubled in numerology.")
			break
		}
		role := map[int]string{0: "root-number", 1: "destiny-number"}[i]
		fmt.Fprintf(&b, " Your %s graha %s is placed as %s", role, engine.GrahaEnglish(id), in.placement(id))
		if s := in.shadbalaFor(cc, id); s != "" {
			b.WriteString("; " + s)
		}
		b.WriteString(".")
		if e := in.grahaEchoes(id, "your"); len(e) > 0 {
			echoed++
			fmt.Fprintf(&b, " The kundali also emphasises it: %s.", strings.Join(e, "; "))
		}
	}
	switch {
	case echoed > 0:
		b.WriteString(" Where the numbers and the chart point to the same graha, the two traditions agree on a theme worth reflecting on.")
	default:
		b.WriteString(" Neither number's graha is especially prominent in your kundali, so the two systems highlight different themes. That is common: use them as two separate lenses rather than one verdict.")
	}
	if n.MulankGraha != n.BhagyankGraha {
		rel, _ := grahaRelation(n.MulankGraha, n.BhagyankGraha)
		fmt.Fprintf(&b, " %s and %s are %s in the classical friendship table.", engine.GrahaEnglish(n.MulankGraha), engine.GrahaEnglish(n.BhagyankGraha), rel)
	}
	if cur := in.f.Vimshottari.Current.Maha; cur != "" && cur != n.MulankGraha && cur != n.BhagyankGraha {
		rel, _ := grahaRelation(n.MulankGraha, cur)
		fmt.Fprintf(&b, " Your current mahadasha lord %s and your root-number graha %s are %s.", engine.GrahaEnglish(cur), engine.GrahaEnglish(n.MulankGraha), rel)
	}
	b.WriteString(" Numerology and astrology are separate symbolic systems; neither proves the other.")
	return b.String()
}
