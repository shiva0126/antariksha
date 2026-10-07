package reading

import (
	"fmt"
	"strings"

	"github.com/example/panchang/engine"
)

// Timing questions are answered from the forecast periods: windows when an
// area of life is more or less supported, never events or dates on which
// something will happen.

// topicArea maps a question topic to the forecast area it asks about.
var topicArea = map[string]string{
	"career": "career", "marriage": "partnership", "wealth": "money", "education": "learning", "children": "learning",
	"health": "wellbeing", "travel": "travel", "property": "home", "siblings": "effort", "spirituality": "travel",
}

const noEvents = "A chart cannot say whether or when a particular event will happen, such as a wedding, a new job or a move. These windows show when the tradition sees this part of life as more or less supported."

func span(p Period) string {
	return fmt.Sprintf("%s to %s", p.From.Format("2 January 2006"), p.To.Format("2 January 2006"))
}

func areaIn(p Period, id string) (AreaForecast, bool) {
	for _, a := range p.Areas {
		if a.Area.ID == id {
			return a, true
		}
	}
	return AreaForecast{}, false
}

func forecastPlain(in *insight, ts []string, cc ChatContext) plain {
	var p plain
	if len(cc.Forecast) == 0 {
		p.say("The forecast for the coming months is unavailable right now. You can see your running periods under Timing.")
		return p
	}
	var area string
	for _, t := range ts {
		if a, ok := topicArea[t]; ok {
			area = a
			break
		}
	}
	if area != "" {
		var name string
		for _, a := range Areas {
			if a.ID == area {
				name = a.Name
			}
		}
		var good, hard []string
		for _, per := range cc.Forecast {
			a, ok := areaIn(per, area)
			if !ok {
				continue
			}
			// Quote the factor that weighs most in this area's tone.
			why := ""
			best := -1.0
			for _, r := range a.Reasons {
				w := r.Sign
				if a.Tone == "challenging" {
					w = -w
				}
				if w > best {
					best, why = w, ": "+firstSentence(r.Text, 160)
				}
			}
			switch a.Tone {
			case "supportive":
				good = append(good, span(per)+" ("+a.Confidence+" agreement)"+why)
			case "challenging":
				hard = append(hard, span(per)+why)
			}
		}
		if area == "wellbeing" {
			p.say("Astrisk cannot predict health, so it only shows when your daily routine and wellbeing get more or less traditional support. For any health concern, please see a doctor.")
		}
		switch {
		case len(good) > 0:
			p.say(fmt.Sprintf("For %s, the most supportive stretch in the coming months is %s.", name, strings.SplitN(good[0], " (", 2)[0]))
		case len(hard) > 0:
			p.say(fmt.Sprintf("No stretch in the coming months stands out as clearly supportive for %s; some ask for patience.", name))
		default:
			p.say(fmt.Sprintf("%s is not strongly marked in the coming months: neither your periods nor the slow planets single it out.", upper(name)))
		}
		for _, g := range good[:min(2, len(good))] {
			p.point("Supportive: " + g)
		}
		for _, h := range hard[:min(2, len(hard))] {
			p.point("Asks for patience: " + h)
		}
		p.point(noEvents)
		return p
	}
	now := cc.Forecast[0]
	p.say(fmt.Sprintf("Until %s: %s", now.To.Format("2 January 2006"), lowerFirst(now.Summary)))
	if v := cc.Varsha; v != nil {
		p.point(fmt.Sprintf("Your year since your birthday on %s (Varshaphal, the year chart): it is ruled by %s. %s", v.ReturnAt.Format("2 January 2006"), theName(v.YearLord), v.Muntha.Detail))
	}
	for _, per := range cc.Forecast[1:min(3, len(cc.Forecast))] {
		p.point(fmt.Sprintf("From %s: %s", per.From.Format("2 January 2006"), lowerFirst(per.Summary)))
	}
	moonSign := in.signIdx("moon")
	shown := 0
	for _, ev := range cc.Events {
		if ev.Kind != "ingress" || (ev.Graha != "jupiter" && ev.Graha != "saturn") || shown == 2 {
			continue
		}
		h := (engine.RashiIndex(ev.Rashi)-moonSign+12)%12 + 1
		fav, _ := engine.GocharaFavourable(ev.Graha, h)
		p.point(fmt.Sprintf("On %s %s moves into your %s house from the Moon, which the classical books count as %s for it.", ev.At.Format("2 January 2006"), theName(ev.Graha), ordinal(h), map[bool]string{true: "a good house", false: "a harder house"}[fav]))
		shown++
	}
	p.point("These are tendencies for reflection, read from your periods and the slow planets; they are not events that will happen.")
	return p
}

// forecastDetails lists the reasons behind the coming periods for the
// folded chart details.
func forecastDetails(cc ChatContext) []string {
	var out []string
	for _, per := range cc.Forecast[:min(3, len(cc.Forecast))] {
		var b strings.Builder
		fmt.Fprintf(&b, "%s (%s main period, %s sub-period):", span(per), engine.GrahaEnglish(per.Maha), engine.GrahaEnglish(per.Antara))
		for _, a := range per.Areas[:min(2, len(per.Areas))] {
			fmt.Fprintf(&b, " %s, %s.", upper(a.Area.Name), a.Tone)
			for _, r := range a.Reasons {
				fmt.Fprintf(&b, " %s (%s)", r.Text, r.Source)
			}
		}
		out = append(out, b.String())
	}
	return out
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
