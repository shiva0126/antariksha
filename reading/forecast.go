package reading

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/example/panchang/engine"
)

// Forecasts combine the running dasha with the slow transits, because a
// transit gives its results only according to the dasha (Brihat Samhita
// 104.46). The coming months are cut into periods at every sign change of
// Jupiter, Saturn, Rahu or Ketu and every change of antardasha; within a
// period nothing slow changes, so each gets one reading. Each life area is
// read from the factors that touch it, and its tone is reported with how
// strongly the factors agree. Nothing here predicts an event.

// Area is a part of life read from houses counted from the lagna.
type Area struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Houses []int  `json:"houses"`
}

// Areas are worded for everyday reading; the 6th and 8th are deliberately
// phrased without illness or death.
var Areas = []Area{
	{"career", "career and public life", []int{10}},
	{"money", "money and gains", []int{2, 11}},
	{"partnership", "marriage and partnerships", []int{7}},
	{"home", "home, family and inner peace", []int{4}},
	{"learning", "studies, creativity and children", []int{5}},
	{"effort", "courage, effort and siblings", []int{3}},
	{"wellbeing", "daily routine and wellbeing", []int{1, 6}},
	{"change", "change and shared resources", []int{8}},
	{"travel", "travel, faith and faraway places", []int{9, 12}},
}

type Reason struct {
	Text   string `json:"text"`
	Source string `json:"source"`
	// Sign is this factor's contribution to the tone: +1 supportive,
	// -1 challenging, 0 emphasis only.
	Sign float64 `json:"sign"`
}

type AreaForecast struct {
	Area       Area     `json:"area"`
	Tone       string   `json:"tone"`       // supportive | mixed | challenging
	Confidence string   `json:"confidence"` // strong | moderate | light
	Score      float64  `json:"score"`
	Reasons    []Reason `json:"reasons"`
}

type Period struct {
	From    time.Time      `json:"from"`
	To      time.Time      `json:"to"`
	Maha    string         `json:"maha"`
	Antara  string         `json:"antara"`
	Summary string         `json:"summary"`
	Areas   []AreaForecast `json:"areas"`
}

// PeriodInput is one period's sky (taken at its middle) and running dasha.
type PeriodInput struct {
	From, To time.Time
	Sky      engine.Chart
	Dasha    engine.DashaPeriod
}

// Forecast reads each period against the natal chart.
func (s *Service) Forecast(ctx context.Context, natal engine.Chart, periods []PeriodInput) []Period {
	nat := map[string]engine.Graha{}
	for _, g := range natal.Grahas {
		nat[g.ID] = g
	}
	dig := engine.Dignities(natal)
	comb := engine.CombustionFlags(natal)
	av := engine.ComputeAshtakavarga(natal)
	asc := natal.Ascendant.Longitude
	lagna := int(asc/30) % 12
	moonSign := int(nat["moon"].Longitude/30) % 12

	// The book's view for each slow planet's house from the Moon, per period.
	var keys []engine.CorpusKey
	for _, p := range periods {
		for _, g := range p.Sky.Grahas {
			if g.ID == "jupiter" || g.ID == "saturn" {
				keys = append(keys, engine.CorpusKey{DocType: engine.DocTransit, Key: engine.TransitKey(g.ID, (int(g.Longitude/30)-moonSign+24)%12+1)})
			}
		}
	}
	views := s.BookViews(ctx, keys)

	out := make([]Period, 0, len(periods))
	for _, p := range periods {
		sky := map[string]engine.Graha{}
		for _, g := range p.Sky.Grahas {
			sky[g.ID] = g
		}
		skyDig := engine.Dignities(p.Sky)
		skyComb := engine.CombustionFlags(p.Sky)
		per := Period{From: p.From, To: p.To, Maha: p.Dasha.Maha, Antara: p.Dasha.Antara}
		for _, a := range Areas {
			af := AreaForecast{Area: a}
			dashaTouches := false
			// 1. The dasha lords: placed in or ruling the area's houses.
			for _, l := range []struct {
				id, label string
				weight    float64
			}{{p.Dasha.Maha, "main period (mahadasha)", 1}, {p.Dasha.Antara, "sub-period (antardasha)", 0.6}} {
				if l.id == "" {
					continue
				}
				placed := engine.HouseOf(nat[l.id].Longitude, asc)
				var rules []int
				for _, h := range a.Houses {
					if engine.HouseLord(natal, h) == l.id {
						rules = append(rules, h)
					}
				}
				if !inInts(placed, a.Houses) && len(rules) == 0 {
					continue
				}
				dashaTouches = true
				sign, why := lordTone(l.id, placed, dig[l.id], comb[l.id], !inInts(placed, a.Houses))
				link := fmt.Sprintf("sits in your %s house", ordinal(placed))
				if len(rules) > 0 {
					link = fmt.Sprintf("rules your %s house", ordinalList(rules))
					if inInts(placed, a.Houses) {
						link += " and sits there"
					}
				}
				af.Reasons = append(af.Reasons, Reason{
					Text:   fmt.Sprintf("Your %s belongs to %s, which %s%s.", l.label, theName(l.id), link, why),
					Source: "Vimshottari dasha", Sign: sign * l.weight,
				})
			}
			// 2. Slow transits occupying or aspecting the area's houses.
			touched := map[string]bool{}
			for _, id := range []string{"jupiter", "saturn", "rahu", "ketu"} {
				g, ok := sky[id]
				if !ok {
					continue
				}
				s := int(g.Longitude/30) % 12
				var hit []int
				for _, x := range engine.AspectedSigns(id, s) {
					if h := (x-lagna+12)%12 + 1; inInts(h, a.Houses) {
						hit = append(hit, h)
					}
				}
				if len(hit) == 0 {
					continue
				}
				touched[id] = true
				fromMoon := (s-moonSign+12)%12 + 1
				how := "passes through"
				if !inInts((s-lagna+12)%12+1, a.Houses) {
					how = "aspects"
				}
				r := Reason{Text: fmt.Sprintf("%s %s your %s house (%s from the Moon).", upper(theName(id)), how, ordinalList(hit), ordinal(fromMoon)), Source: "transit"}
				if fav, known := engine.GocharaFavourable(id, fromMoon); known {
					r.Sign = map[bool]float64{true: 1, false: -1}[fav]
					weak := skyDig[id].State == "debilitated" || skyDig[id].State == "enemy" || skyComb[id]
					if fav && weak {
						r.Sign = 0
						r.Text += " Its good results are weakened (Brihat Samhita 104.53)."
					}
					if b := av.Bhinna[id][s]; b >= 5 {
						r.Sign += 0.5
						r.Text += fmt.Sprintf(" Your chart gives this sign strong support (%d of 8 ashtakavarga points).", b)
					} else if b <= 2 {
						r.Sign -= 0.5
						r.Text += fmt.Sprintf(" Your chart gives this sign little support (%d of 8 ashtakavarga points).", b)
					}
					// A transit gives its results according to the dasha (104.46).
					if !dashaTouches {
						r.Sign *= 0.5
					}
					r.Source = "Brihat Samhita 104.4"
					if v, ok := views[engine.DocTransit+":"+engine.TransitKey(id, fromMoon)]; ok {
						r.Text += " The book: " + strings.TrimSuffix(v.Plain, ".") + "."
						r.Source = "Brihat Samhita " + v.Ref
					}
				} else {
					r.Text += " The nodes bring emphasis and change rather than a set result (later tradition)."
					r.Source = "later tradition"
				}
				af.Reasons = append(af.Reasons, r)
			}
			// 3. Double transit: Jupiter and Saturn both touch the area (modern).
			if touched["jupiter"] && touched["saturn"] {
				af.Reasons = append(af.Reasons, Reason{Text: "Jupiter and Saturn both touch this area at once, which modern astrologers read as a time when it becomes active.", Source: "double transit (modern technique)"})
			}
			// Dasha lords and Jupiter and Saturn are the factors; the nodes and
			// double transit add emphasis only. One factor alone does not
			// make an area a theme.
			factors := 0
			for _, r := range af.Reasons {
				if r.Source != "later tradition" && !strings.HasPrefix(r.Source, "double transit") {
					factors++
				}
			}
			if factors < 2 && !(factors == 1 && touched["jupiter"] && touched["saturn"]) {
				continue
			}
			pos, neg := 0, 0
			for _, r := range af.Reasons {
				af.Score += r.Sign
				if r.Sign > 0 {
					pos++
				} else if r.Sign < 0 {
					neg++
				}
			}
			af.Score = math.Round(af.Score*10) / 10
			switch {
			case af.Score >= 1:
				af.Tone = "supportive"
			case af.Score <= -1:
				af.Tone = "challenging"
			default:
				af.Tone = "mixed"
			}
			// Confidence counts only factors that lean one way, and how many
			// of them agree with the tone.
			lean := max(pos, neg)
			switch {
			case (pos == 0 || neg == 0) && lean >= 3:
				af.Confidence = "strong"
			case lean >= 2 && lean > min(pos, neg):
				af.Confidence = "moderate"
			default:
				af.Confidence = "light"
			}
			per.Areas = append(per.Areas, af)
		}
		sort.SliceStable(per.Areas, func(i, j int) bool {
			if len(per.Areas[i].Reasons) != len(per.Areas[j].Reasons) {
				return len(per.Areas[i].Reasons) > len(per.Areas[j].Reasons)
			}
			return math.Abs(per.Areas[i].Score) > math.Abs(per.Areas[j].Score)
		})
		per.Summary = periodSummary(per)
		out = append(out, per)
	}
	return out
}

// lordTone is a dasha lord's strength in the birth chart, in plain words.
func lordTone(id string, house int, d engine.Dignity, combust, sayHouse bool) (float64, string) {
	sign, parts := 0.0, []string{}
	switch d.State {
	case "exalted":
		sign, parts = sign+1, append(parts, "at its strongest")
	case "own":
		sign, parts = sign+1, append(parts, "comfortable in its own sign")
	case "friendly":
		sign, parts = sign+0.5, append(parts, "in a friendly sign")
	case "enemy":
		sign, parts = sign-0.5, append(parts, "in a less comfortable sign")
	case "debilitated":
		if d.NeechaBhanga {
			parts = append(parts, "in its weakest sign, though that is offset")
		} else {
			sign, parts = sign-1, append(parts, "in its weakest sign")
		}
	}
	if inInts(house, []int{6, 8, 12}) {
		sign = sign - 0.5
		if sayHouse {
			parts = append(parts, fmt.Sprintf("placed in the %s house, traditionally a harder house", ordinal(house)))
		} else {
			parts = append(parts, "a traditionally harder house")
		}
	}
	if combust {
		sign, parts = sign-0.5, append(parts, "close to the Sun")
	}
	if len(parts) == 0 {
		return 0, ""
	}
	return sign, ", " + strings.Join(parts, ", ")
}

func periodSummary(p Period) string {
	if len(p.Areas) == 0 {
		return fmt.Sprintf("A quieter stretch in %s's period: no area of life stands out.", theName(p.Maha))
	}
	var parts []string
	for _, a := range p.Areas[:min(2, len(p.Areas))] {
		word := map[string]string{"supportive": "looks supportive", "challenging": "asks for patience", "mixed": "is mixed"}[a.Tone]
		parts = append(parts, a.Area.Name+" "+word)
	}
	return upper(strings.Join(parts, "; ")) + "."
}

func inInts(x int, xs []int) bool {
	for _, y := range xs {
		if x == y {
			return true
		}
	}
	return false
}

func ordinalList(hs []int) string {
	s := make([]string, len(hs))
	for i, h := range hs {
		s[i] = ordinal(h)
	}
	return joinAnd(s)
}
