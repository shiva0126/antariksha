package reading

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/example/panchang/engine"
)

// These checks cover explicit named combinations, common placement assertions,
// and the structured timing summary. They are not a semantic truth guarantee.
func validateNarrative(r Reading, f engine.ChartFacts) error {
	b, _ := json.Marshal(r)
	text := strings.ToLower(string(b))
	allowed := map[string]bool{}
	for _, y := range f.Yogas {
		allowed[strings.ToLower(y.Name)] = true
	}
	names := append([]string{}, engine.YogaCatalog...)
	sort.Slice(names, func(i, j int) bool { return len(names[i]) > len(names[j]) })
	for _, name := range names {
		pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(name) + `\b`)
		if pattern.MatchString(text) && !allowed[strings.ToLower(name)] {
			return fmt.Errorf("narrative names undetected yoga %s", name)
		}
		text = pattern.ReplaceAllString(text, "")
	}
	placements := regexp.MustCompile(`(?i)\b(sun|moon|mars|mercury|jupiter|venus|saturn|rahu|ketu) (?:is |sits |lies )?in (?:the )?(?:house (\d{1,2})|(\d{1,2})(?:st|nd|rd|th)? house)`)
	for _, match := range placements.FindAllStringSubmatch(string(b), -1) {
		id := strings.ToLower(match[1])
		h := match[2] + match[3]
		var actual int
		for _, g := range f.Chart.Grahas {
			if g.ID == id {
				actual = (int(g.Longitude/30)-int(f.Chart.Ascendant.Longitude/30)+12)%12 + 1
			}
		}
		if h != fmt.Sprint(actual) {
			return fmt.Errorf("narrative house differs for %s", id)
		}
	}
	// The current-period section must not introduce a different maha/antara pair.
	period := regexp.MustCompile(`(?i)\b(sun|moon|mars|mercury|jupiter|venus|saturn|rahu|ketu)\s*(?:–|—|-|/|mahadasha\s*(?:and|with))\s*(sun|moon|mars|mercury|jupiter|venus|saturn|rahu|ketu)\b`)
	for _, m := range period.FindAllStringSubmatch(r.Dashas.Current, -1) {
		if !strings.EqualFold(m[1], f.Vimshottari.Current.Maha) || !strings.EqualFold(m[2], f.Vimshottari.Current.Antara) {
			return fmt.Errorf("current period differs from facts")
		}
	}
	date := regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}\b`)
	currentDates := f.Vimshottari.Current.From + " " + f.Vimshottari.Current.To + " " + f.Vimshottari.Upcoming.From
	for _, p := range f.Vimshottari.Sequence {
		if p.Lord == f.Vimshottari.Current.Maha && p.To > f.Vimshottari.Current.From {
			currentDates += " " + p.From + " " + p.To
			break
		}
	}
	for _, pair := range [][2]string{{r.Dashas.Current, currentDates}, {r.Dashas.Upcoming, f.Vimshottari.Upcoming.From + " " + f.Vimshottari.Upcoming.To}} {
		for _, v := range date.FindAllString(pair[0], -1) {
			if !strings.Contains(pair[1], v) {
				return fmt.Errorf("dasha date differs from facts")
			}
		}
	}
	for _, bad := range []string{"will die", "will divorce", "guaranteed wealth", "will get cancer", "will become infertile"} {
		if strings.Contains(strings.ToLower(string(b)), bad) {
			return fmt.Errorf("unsafe prediction")
		}
	}
	return nil
}
