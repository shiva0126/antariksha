package festivals

import "github.com/example/panchang/engine"

type Rule struct {
	Slug, DisplayName, RuleType, Paksha, LunarMonth, Region, MonthSystem string
	Tithi                                                                int
}

func Match(r Rule, d engine.Day) bool {
	if r.RuleType != "tithi" && r.RuleType != "lunar_month_tithi" {
		return false
	}
	n := d.Tithi.Number
	if n > 15 {
		n -= 15
	}
	if n != r.Tithi {
		return false
	}
	if r.Paksha != "" && r.Paksha != lower(d.Paksha) {
		return false
	}
	if r.LunarMonth != "" && r.LunarMonth != lower(d.LunarMonth) {
		return false
	}
	return true
}
func lower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}
