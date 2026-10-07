package engine

import (
	"strings"
	"testing"
)

func moonAt(nak int) Graha {
	lon := float64(nak)*360.0/27 + 1 // inside the star, first pada
	return Graha{ID: "moon", Longitude: lon}
}

func porutham(t *testing.T, ps []Porutham, name string) Porutham {
	t.Helper()
	for _, p := range ps {
		if p.Name == name {
			return p
		}
	}
	t.Fatalf("no %s porutham", name)
	return Porutham{}
}

func TestPoruthamTables(t *testing.T) {
	sizes := map[int]int{}
	for _, r := range nakRajju {
		sizes[r]++
	}
	if len(nakRajju) != 27 || sizes[0] != 6 || sizes[1] != 6 || sizes[2] != 6 || sizes[3] != 6 || sizes[4] != 3 {
		t.Fatalf("rajju groups %v", sizes)
	}
	for n := 0; n < 27; n++ {
		partners := nakVedha[n]
		if len(partners) == 0 {
			t.Errorf("%s has no vedha partner", nakshatraNames[n])
		}
		for _, p := range partners {
			if !inSet(n, nakVedha[p]...) {
				t.Errorf("vedha %s-%s is not mutual", nakshatraNames[n], nakshatraNames[p])
			}
			// The classical pairs sum to 17, 26 or 35 (0-based), apart from the
			// Mrigashira-Chitra-Dhanishtha triple.
			if len(partners) == 1 && !inSet(n+p, 17, 26, 35) {
				t.Errorf("vedha %s-%s breaks the pattern", nakshatraNames[n], nakshatraNames[p])
			}
		}
	}
	if !inSet(13, nakVedha[4]...) || !inSet(22, nakVedha[4]...) {
		t.Error("Mrigashira, Chitra and Dhanishtha form the vedha triple")
	}
}

func TestPoruthamWorkedCases(t *testing.T) {
	// Same star for both (Rohini): same Rajju group, no Vedha, same sign.
	same := Poruthams(moonAt(3), moonAt(3))
	if porutham(t, same, "Rajju").Status != "bad" || porutham(t, same, "Vedha").Status != "good" || porutham(t, same, "Rasi").Status != "good" {
		t.Errorf("same star: %+v", same)
	}
	// Bride Ashwini, groom Jyeshtha: a Vedha pair.
	if p := porutham(t, Poruthams(moonAt(17), moonAt(0)), "Vedha"); p.Status != "bad" || !strings.Contains(p.Detail, "form a Vedha") {
		t.Errorf("Ashwini-Jyeshtha: %+v", p)
	}
	// Bride Ashwini, groom Rohini: the 4th star. Mahendra and Dina good, Stree Deergha not.
	ps := Poruthams(moonAt(3), moonAt(0))
	if porutham(t, ps, "Mahendra").Status != "good" || porutham(t, ps, "Dina").Status != "good" || porutham(t, ps, "Stree Deergha").Status != "bad" {
		t.Errorf("4th star: %+v", ps)
	}
	// Bride Ashwini, groom Hasta: the 13th star. Mahendra good, Stree Deergha partly.
	ps = Poruthams(moonAt(12), moonAt(0))
	if porutham(t, ps, "Mahendra").Status != "good" || porutham(t, ps, "Stree Deergha").Status != "medium" {
		t.Errorf("13th star: %+v", ps)
	}
	// Groom Krittika (Rakshasa), bride Ashwini (Deva): Gana fails.
	if p := porutham(t, Poruthams(moonAt(2), moonAt(0)), "Gana"); p.Status != "bad" {
		t.Errorf("Rakshasa with Deva: %+v", p)
	}
	// Rajju and Vedha are the essential ones.
	for _, p := range same {
		if p.Essential != (p.Name == "Rajju" || p.Name == "Vedha") {
			t.Errorf("%s essential=%v", p.Name, p.Essential)
		}
	}
}

func chartWith(lagna int, signs map[string]int) Chart {
	c := Chart{Ascendant: Point{Longitude: float64(lagna)*30 + 5}}
	for id, s := range signs {
		c.Grahas = append(c.Grahas, Graha{ID: id, Longitude: float64(s)*30 + 10})
	}
	return c
}

func TestKujaDoshaReferencesAndCancellations(t *testing.T) {
	// Mesha rising; Mars in Tula is 7th from the lagna, Moon and Venus; Jupiter
	// in Vrishabha does not aspect it: present three ways and effective.
	k := KujaDosha(chartWith(0, map[string]int{"mars": 6, "moon": 0, "venus": 0, "jupiter": 1}))
	if len(k.Present) != 3 || !k.Effective || len(k.Cancellations) != 0 {
		t.Errorf("three references: %+v", k)
	}
	// Makara rising, Mars in Karka in the 7th: a traditional exception.
	k = KujaDosha(chartWith(9, map[string]int{"mars": 3, "moon": 5, "venus": 5, "jupiter": 1}))
	if k.Effective || len(k.Cancellations) == 0 {
		t.Errorf("Mars in Karka in the 7th: %+v", k)
	}
	// Jupiter in Mithuna aspects Tula (its 5th): softened.
	k = KujaDosha(chartWith(0, map[string]int{"mars": 6, "moon": 0, "venus": 0, "jupiter": 2}))
	if k.Effective {
		t.Errorf("Jupiter aspect: %+v", k)
	}
	// Simha rising: Mars is a yogakaraka.
	k = KujaDosha(chartWith(4, map[string]int{"mars": 10, "moon": 4, "venus": 4, "jupiter": 4}))
	if k.Effective {
		t.Errorf("Simha rising: %+v", k)
	}
	// Mars in the 3rd from everything: no dosha.
	k = KujaDosha(chartWith(0, map[string]int{"mars": 2, "moon": 0, "venus": 0, "jupiter": 0}))
	if len(k.Present) != 0 || k.Effective {
		t.Errorf("no dosha: %+v", k)
	}
}

func TestPapaSamyaWeights(t *testing.T) {
	// Mesha rising, Moon and Venus in Mesha too; Mars alone in the 7th (Tula)
	// counts from all three references: 1 + 0.5 + 0.25. The other difficult
	// planets sit in the 3rd (Mithuna), which does not count.
	c := chartWith(0, map[string]int{"mars": 6, "moon": 0, "venus": 0, "saturn": 2, "rahu": 2, "ketu": 8, "sun": 2, "jupiter": 1})
	// Ketu in Dhanu is the 9th: not counted either.
	if p := PapaSamya(c); p.Points != 1.75 || len(p.Detail) != 3 {
		t.Errorf("Mars in the 7th from all three: %+v", p)
	}
	// The Sun counts at half: Sun alone in the 1st from all three gives 0.875.
	c = chartWith(0, map[string]int{"sun": 0, "moon": 0, "venus": 0, "mars": 2, "saturn": 2, "rahu": 2, "ketu": 8, "jupiter": 1})
	if p := PapaSamya(c); p.Points != 0.875 {
		t.Errorf("Sun in the 1st: %+v", p)
	}
}

func TestMarriageIndicators(t *testing.T) {
	// Mesha rising: Venus rules the 7th (Tula). Venus in Tula (own sign) in
	// the 7th, and Jupiter in Mithuna aspecting Tula with its 5th-sign aspect.
	m := MarriageIndicators(chartWith(0, map[string]int{"venus": 6, "jupiter": 2, "moon": 3, "mars": 9, "saturn": 10, "sun": 4, "mercury": 4, "rahu": 1, "ketu": 7}))
	all := strings.Join(m.Strengths, " ")
	if m.SeventhLord != "venus" || m.SeventhLordHouse != 7 || !strings.Contains(all, "is strong") || !strings.Contains(all, "Jupiter aspects the 7th house") {
		t.Errorf("strong 7th: %+v", m)
	}
	// Venus in Kanya (debilitated) in the 6th: cautions.
	m = MarriageIndicators(chartWith(0, map[string]int{"venus": 5, "jupiter": 1, "moon": 3, "mars": 9, "saturn": 10, "sun": 4, "mercury": 4, "rahu": 2, "ketu": 8}))
	if len(m.Cautions) < 2 {
		t.Errorf("weak 7th ruler in the 6th: %+v", m)
	}
}
