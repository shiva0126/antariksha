package reading

import (
	"fmt"
	"strings"
	"time"

	"github.com/example/panchang/engine"
)

// MatchOverview gathers every matching check into strong points and
// cautions in plain words, with one summary line. It never recommends
// accepting or rejecting a partner.
type MatchOverview struct {
	Summary   string   `json:"summary"`
	Strengths []string `json:"strengths"`
	Cautions  []string `json:"cautions"`
}

// dashaSandhi reports a main-period change within the next year: families
// traditionally avoid the junction of two main periods for the wedding.
func dashaSandhi(who string, f engine.ChartFacts, now time.Time) string {
	cur := f.Vimshottari.Current
	for i, p := range f.Vimshottari.Sequence {
		if p.Lord != cur.Maha {
			continue
		}
		end, err := time.Parse("2006-01-02", p.To)
		if err != nil || end.Before(now) || end.After(now.AddDate(1, 0, 0)) {
			return ""
		}
		next := ""
		if i+1 < len(f.Vimshottari.Sequence) {
			next = f.Vimshottari.Sequence[i+1].Lord
		}
		return fmt.Sprintf("The %s's main period changes from %s to %s on %s. Families traditionally avoid the junction of two main periods (dasha sandhi) for the wedding itself.", who, theName(cur.Maha), theName(next), end.Format("2 January 2006"))
	}
	return ""
}

// Overview reads a match with both charts' facts at the given time.
func Overview(m engine.Match, boy, girl engine.ChartFacts, now time.Time) MatchOverview {
	o := MatchOverview{Strengths: []string{}, Cautions: []string{}}
	if m.Total >= 18 {
		o.Strengths = append(o.Strengths, fmt.Sprintf("The guna score is %g of 36, %s.", m.Total, scoreBand(m.Total)))
	} else {
		o.Cautions = append(o.Cautions, fmt.Sprintf("The guna score is %g of 36, %s.", m.Total, scoreBand(m.Total)))
	}
	var essential []string
	for _, p := range m.Poruthams {
		if p.Essential && p.Status == "bad" {
			essential = append(essential, p.Name)
		}
	}
	switch {
	case len(essential) > 0:
		o.Cautions = append(o.Cautions, fmt.Sprintf("%s porutham does not match; South Indian families treat it as essential.", joinAnd(essential)))
	case m.PoruthamGood >= 6:
		o.Strengths = append(o.Strengths, fmt.Sprintf("%d of the 10 poruthams match, with Rajju and Vedha clear.", m.PoruthamGood))
	default:
		o.Cautions = append(o.Cautions, fmt.Sprintf("Only %d of the 10 poruthams match, though Rajju and Vedha are clear.", m.PoruthamGood))
	}
	for _, d := range m.Doshas {
		if strings.HasSuffix(d, "dosha") && (strings.HasPrefix(d, "Bhakoot") || strings.HasPrefix(d, "Nadi") || strings.HasPrefix(d, "Gana")) {
			name := strings.TrimSuffix(d, " dosha")
			cancelled := false
			for _, x := range m.Exceptions {
				cancelled = cancelled || strings.HasPrefix(x, name)
			}
			if cancelled {
				o.Strengths = append(o.Strengths, d+" is present but traditionally cancelled.")
			} else {
				o.Cautions = append(o.Cautions, d+" is present in the guna tables.")
			}
		}
	}
	switch {
	case m.BoyMangal == m.GirlMangal:
		o.Strengths = append(o.Strengths, map[bool]string{true: "Both charts have Mangal dosha, which tradition treats as balanced.", false: "Mangal dosha is not a concern: neither chart has it after the traditional cancellations."}[m.BoyMangal])
	default:
		o.Cautions = append(o.Cautions, "Only one chart has Mangal dosha after the traditional cancellations.")
	}
	if m.BoyPapa.Points >= m.GirlPapa.Points {
		o.Strengths = append(o.Strengths, "Papasamya, the balance of difficult planets, is in balance.")
	} else {
		o.Cautions = append(o.Cautions, "Papasamya, the balance of difficult planets, is not in balance.")
	}
	for _, p := range []struct {
		who string
		r   engine.MarriageReport
	}{{"groom", m.BoyMarriage}, {"bride", m.GirlMarriage}} {
		switch {
		case len(p.r.Strengths) > len(p.r.Cautions) && len(p.r.Strengths) > 0:
			o.Strengths = append(o.Strengths, fmt.Sprintf("The %s's chart supports partnership: %s", p.who, lowerFirst(p.r.Strengths[0])))
		case len(p.r.Cautions) > 0:
			o.Cautions = append(o.Cautions, fmt.Sprintf("In the %s's chart, %s", p.who, lowerFirst(p.r.Cautions[0])))
		}
	}
	for _, s := range []string{dashaSandhi("groom", boy, now), dashaSandhi("bride", girl, now)} {
		if s != "" {
			o.Cautions = append(o.Cautions, s)
		}
	}
	switch s, c := len(o.Strengths), len(o.Cautions); {
	case c == 0:
		o.Summary = fmt.Sprintf("All %d checks look supportive.", s)
	case s >= 2*c:
		o.Summary = fmt.Sprintf("Mostly supportive: %d strong points and %d to discuss.", s, c)
	case c > s:
		o.Summary = fmt.Sprintf("Several points to discuss: %d cautions and %d strong points.", c, s)
	default:
		o.Summary = fmt.Sprintf("Mixed: %d strong points and %d to discuss.", s, c)
	}
	o.Summary += " Matching compares traditional tables; it is not a verdict on the two people."
	return o
}
