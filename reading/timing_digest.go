package reading

import (
	"fmt"
	"strings"
	"time"

	"github.com/example/panchang/engine"
)

// TimingDigest is the timing side of a chart as a few plain lines for a
// model: the running periods with their end dates, Sade Sati, the slow
// planets' coming sign changes read from the Moon, the year chart, and the
// forecast windows per area of life. A prediction answer may only state
// what these lines give it; the prediction test set (llmdata predict)
// checks exactly these facts.
func TimingDigest(f engine.ChartFacts, cc ChatContext, now time.Time) string {
	var b strings.Builder
	d := f.Vimshottari
	fmt.Fprintf(&b, "Today: %s.\n", dayT(now))
	if c := d.Current; c.Maha != "" {
		fmt.Fprintf(&b, "Main period (mahadasha): %s. Sub-period (antardasha): %s, until %s.\n", engine.GrahaEnglish(c.Maha), engine.GrahaEnglish(c.Antara), day(c.To))
		for _, a := range d.Antaras {
			if a.From == c.To {
				fmt.Fprintf(&b, "Next sub-period: %s, from %s.\n", engine.GrahaEnglish(a.Lord), day(a.From))
			}
		}
		if u := d.Upcoming; u.Lord != "" {
			fmt.Fprintf(&b, "Next main period: %s, from %s.\n", engine.GrahaEnglish(u.Lord), day(u.From))
		}
	}
	if y := f.Yogini.Current; y.Lord != "" {
		fmt.Fprintf(&b, "Yogini dasha: %s (%s), until %s.\n", y.Yogini, engine.GrahaEnglish(y.Lord), day(y.To))
	}
	if cc.Transit != nil {
		r := engine.Transits(f.Chart, *cc.Transit)
		switch {
		case r.SadeSati:
			fmt.Fprintf(&b, "Sade Sati: running, phase %d of 3.\n", r.SadeSatiPhase)
		default:
			b.WriteString("Sade Sati: not running.\n")
		}
	}
	moon := 0
	for _, g := range f.Chart.Grahas {
		if g.ID == "moon" {
			moon = engine.RashiIndex(g.Rashi)
		}
	}
	for _, ev := range cc.Events {
		if ev.Kind != "ingress" || (ev.Graha != "jupiter" && ev.Graha != "saturn" && ev.Graha != "rahu") {
			continue
		}
		h := (engine.RashiIndex(ev.Rashi)-moon+12)%12 + 1
		verdict := ""
		if fav, known := engine.GocharaFavourable(ev.Graha, h); known {
			verdict = map[bool]string{true: ", a good house for it in the classical books", false: ", a harder house for it in the classical books"}[fav]
		}
		fmt.Fprintf(&b, "On %s %s moves into %s, your %s house from the Moon%s.\n", dayT(ev.At), engine.GrahaEnglish(ev.Graha), ev.Rashi, ordinal(h), verdict)
	}
	if v := cc.Varsha; v != nil {
		fmt.Fprintf(&b, "Year chart (Varshaphal) from %s to %s: year lord %s; Muntha in the %s house (%s).\n", dayT(v.ReturnAt), dayT(v.Until), engine.GrahaEnglish(v.YearLord), ordinal(v.Muntha.House), map[string]string{"good": "good", "mixed": "fair", "hard": "harder"}[v.Muntha.Tone])
		for _, m := range v.Mudda {
			if !now.Before(m.From) && now.Before(m.To) {
				fmt.Fprintf(&b, "Mudda dasha now: %s, until %s (%s).\n", engine.GrahaEnglish(m.Lord), dayT(m.To), m.Tone)
			}
		}
	}
	for _, a := range Areas {
		var good, hard []string
		for _, p := range cc.Forecast {
			if af, ok := areaIn(p, a.ID); ok {
				switch af.Tone {
				case "supportive":
					good = append(good, span(p))
				case "challenging":
					hard = append(hard, span(p))
				}
			}
		}
		if len(good)+len(hard) == 0 {
			continue
		}
		fmt.Fprintf(&b, "%s:", upper(a.Name))
		if len(good) > 0 {
			fmt.Fprintf(&b, " supportive %s.", strings.Join(good[:min(2, len(good))], "; "))
		}
		if len(hard) > 0 {
			fmt.Fprintf(&b, " asks for patience %s.", strings.Join(hard[:min(2, len(hard))], "; "))
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func dayT(t time.Time) string { return t.Format("2 January 2006") }

// MatchDigest is a match as a few plain lines for a model: the guna score
// and verdict, the koota scores, the essential poruthams, Kuja dosha and
// Papasamya.
func MatchDigest(m engine.Match) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Guna milan (Ashtakoota): %s of %s (%s).\n", trimNum(m.Total), trimNum(m.Max), m.Verdict)
	for _, k := range m.Kootas {
		fmt.Fprintf(&b, "%s: %s of %s.\n", k.Name, trimNum(k.Score), trimNum(k.Max))
	}
	if len(m.Doshas) > 0 {
		fmt.Fprintf(&b, "Doshas: %s.\n", strings.Join(m.Doshas, "; "))
	}
	if len(m.Exceptions) > 0 {
		fmt.Fprintf(&b, "Exceptions: %s.\n", strings.Join(m.Exceptions, "; "))
	}
	fmt.Fprintf(&b, "Poruthams (South Indian): %d of %d good.\n", m.PoruthamGood, len(m.Poruthams))
	for _, p := range m.Poruthams {
		if p.Essential {
			fmt.Fprintf(&b, "%s (essential): %s.\n", p.Name, p.Status)
		}
	}
	kuja := func(k engine.KujaReport) string {
		switch {
		case k.Effective:
			return "present"
		case len(k.Present) > 0:
			return "present but cancelled"
		}
		return "absent"
	}
	fmt.Fprintf(&b, "Kuja (Mangal) dosha: groom %s, bride %s.\n", kuja(m.BoyKuja), kuja(m.GirlKuja))
	if m.PapaNote != "" {
		fmt.Fprintf(&b, "Papasamya: %s\n", m.PapaNote)
	}
	return strings.TrimSpace(b.String())
}

func trimNum(x float64) string { return strings.TrimSuffix(fmt.Sprintf("%.1f", x), ".0") }

// Forbidden reports wording that states death, divorce or certainty as a
// prediction.
func Forbidden(a string) bool { return forbidden.MatchString(a) }
