package engine

import "fmt"

// South Indian matching: the ten poruthams, read from the two birth stars
// and Moon signs, counted from the bride's star to the groom's as Tamil,
// Kannada and Malayalam practice does. Rajju and Vedha are treated as
// essential: when either fails, families usually do not proceed whatever
// the count. The tables are the common traditional ones; schools differ in
// detail, and every result says which rule it applied.

type Porutham struct {
	Name      string `json:"name"`
	Status    string `json:"status"` // good | medium | bad
	Essential bool   `json:"essential"`
	Detail    string `json:"detail"`
}

// Rajju: the 27 stars in five groups, rising from the feet to the head and
// back in three runs of nine.
var rajjuNames = []string{"Pada (feet)", "Kati (waist)", "Nabhi (navel)", "Kantha (neck)", "Shiro (head)"}
var nakRajju = []int{0, 1, 2, 3, 4, 3, 2, 1, 0, 0, 1, 2, 3, 4, 3, 2, 1, 0, 0, 1, 2, 3, 4, 3, 2, 1, 0}

// Vedha: pairs of stars that obstruct each other, and one triple.
var nakVedha = map[int][]int{
	0: {17}, 17: {0}, 1: {16}, 16: {1}, 2: {15}, 15: {2}, 3: {14}, 14: {3},
	5: {21}, 21: {5}, 6: {20}, 20: {6}, 7: {19}, 19: {7}, 8: {18}, 18: {8},
	9: {26}, 26: {9}, 10: {25}, 25: {10}, 11: {24}, 24: {11}, 12: {23}, 23: {12},
	4: {13, 22}, 13: {4, 22}, 22: {4, 13},
}

// rasiVasya lists, for each Moon sign, the signs it holds as vasya.
var rasiVasya = [][]int{
	{4, 7}, {3, 6}, {5}, {7, 8}, {6}, {2, 11}, {9, 5}, {3}, {11}, {0, 10}, {0}, {9},
}

// Poruthams computes the ten poruthams for a groom's and a bride's Moon.
func Poruthams(boyMoon, girlMoon Graha) []Porutham {
	bn, gn := nakIndex(boyMoon.Longitude), nakIndex(girlMoon.Longitude)
	bs, gs := int(boyMoon.Longitude/30)%12, int(girlMoon.Longitude/30)%12
	count := (bn-gn+27)%27 + 1 // the groom's star counted from the bride's
	rasi := (bs-gs+12)%12 + 1  // the groom's sign counted from the bride's
	status := func(good bool) string {
		if good {
			return "good"
		}
		return "bad"
	}
	var out []Porutham

	// 1. Dina: the count's remainder by nine is even (or nine).
	r := count % 9
	out = append(out, Porutham{"Dina", status(r == 0 || r%2 == 0), false,
		fmt.Sprintf("The groom's star is %s from the bride's; remainder %d after dividing by 9 (good when even or 9).", ordinal(count), r)})

	// 2. Gana: same temperament group, or Deva with Manushya.
	gb, gg := nakGana[bn], nakGana[gn]
	gana := "bad"
	switch {
	case gb == gg:
		gana = "good"
	case gb != 2 && gg != 2:
		gana = "medium"
	}
	out = append(out, Porutham{"Gana", gana, false, fmt.Sprintf("Groom %s, bride %s.", ganaNames[gb], ganaNames[gg])})

	// 3. Mahendra: the count is 4, 7, 10, 13, 16, 19, 22 or 25.
	out = append(out, Porutham{"Mahendra", status(inSet(count, 4, 7, 10, 13, 16, 19, 22, 25)), false,
		fmt.Sprintf("The groom's star is %s from the bride's (good at 4, 7, 10, 13, 16, 19, 22, 25).", ordinal(count))})

	// 4. Stree Deergha: the groom's star well beyond the bride's.
	sd := "bad"
	switch {
	case count > 13:
		sd = "good"
	case count > 7:
		sd = "medium"
	}
	out = append(out, Porutham{"Stree Deergha", sd, false, fmt.Sprintf("The groom's star is %s from the bride's (good beyond 13, partly beyond 7).", ordinal(count))})

	// 5. Yoni: the birth-star animals.
	yb, yg := nakYoni[bn], nakYoni[gn]
	ys := yoniScore[yb][yg]
	yoni := "bad"
	switch {
	case ys >= 3:
		yoni = "good"
	case ys == 2:
		yoni = "medium"
	}
	out = append(out, Porutham{"Yoni", yoni, false, fmt.Sprintf("Groom %s, bride %s.", yoniNames[yb], yoniNames[yg])})

	// 6. Rasi: the groom's sign from the bride's; the 2nd to 6th, 8th and 12th fail.
	out = append(out, Porutham{"Rasi", status(rasi == 1 || (rasi >= 7 && rasi != 8 && rasi != 12)), false,
		fmt.Sprintf("The groom's Moon sign %s is %s from the bride's %s (good at the same sign, the 7th, and the 9th to 11th).", rashiNames[bs], ordinal(rasi), rashiNames[gs])})

	// 7. Rasyadhipati: the two Moon-sign lords as friends.
	lb, lg := rashiLord[bs], rashiLord[gs]
	ra := "bad"
	switch {
	case lb == lg || (relation(lb, lg) == 2 && relation(lg, lb) == 2):
		ra = "good"
	case relation(lb, lg) >= 1 && relation(lg, lb) >= 1:
		ra = "medium"
	}
	out = append(out, Porutham{"Rasyadhipati", ra, false, fmt.Sprintf("Moon-sign lords %s and %s.", GrahaEnglish(lb), GrahaEnglish(lg))})

	// 8. Vasya: one Moon sign holds the other as vasya.
	vasya := inSet(bs, rasiVasya[gs]...) || inSet(gs, rasiVasya[bs]...)
	out = append(out, Porutham{"Vasya", status(vasya), false, fmt.Sprintf("Moon signs %s and %s.", rashiNames[bs], rashiNames[gs])})

	// 9. Rajju: different body groups; the same group is Rajju dosha.
	rb, rg := nakRajju[bn], nakRajju[gn]
	out = append(out, Porutham{"Rajju", status(rb != rg), true, fmt.Sprintf("Groom %s, bride %s; the same group is Rajju dosha.", rajjuNames[rb], rajjuNames[rg])})

	// 10. Vedha: the two stars must not obstruct each other.
	out = append(out, Porutham{"Vedha", status(!inSet(gn, nakVedha[bn]...)), true,
		fmt.Sprintf("%s and %s %s a Vedha (obstructing) pair.", nakshatraNames[bn], nakshatraNames[gn], map[bool]string{true: "form", false: "do not form"}[inSet(gn, nakVedha[bn]...)])})
	return out
}

// KujaReport is Mangal (Kuja) dosha read from the lagna, the Moon and Venus.
type KujaReport struct {
	FromLagna int `json:"from_lagna"`
	FromMoon  int `json:"from_moon"`
	FromVenus int `json:"from_venus"`
	// Present lists the references from which Mars falls in the 1st, 2nd,
	// 4th, 7th, 8th or 12th house.
	Present []string `json:"present"`
	// Cancellations are the traditional exceptions that apply.
	Cancellations []string `json:"cancellations"`
	// Effective: present from at least one reference and not cancelled.
	Effective bool `json:"effective"`
}

// KujaDosha reads Mangal dosha from the lagna, Moon and Venus with the
// commonly applied cancellations (later tradition; schools differ).
func KujaDosha(c Chart) KujaReport {
	g := chartMap(c)
	mars := g["mars"]
	ms := signOf(mars)
	rep := KujaReport{
		FromLagna: (ms-int(c.Ascendant.Longitude/30)%12+12)%12 + 1,
		FromMoon:  (ms-signOf(g["moon"])+12)%12 + 1,
		FromVenus: (ms-signOf(g["venus"])+12)%12 + 1,
	}
	for _, x := range []struct {
		name  string
		house int
	}{{"lagna", rep.FromLagna}, {"Moon", rep.FromMoon}, {"Venus", rep.FromVenus}} {
		if inSet(x.house, 1, 2, 4, 7, 8, 12) {
			rep.Present = append(rep.Present, x.name)
		}
	}
	if len(rep.Present) == 0 {
		rep.Present, rep.Cancellations = []string{}, []string{}
		return rep
	}
	var cancel []string
	if inSet(ms, 0, 7, 9) {
		cancel = append(cancel, fmt.Sprintf("Mars is in %s, its own or exaltation sign.", rashiNames[ms]))
	}
	// Sign and house pairs widely held to cancel the dosha.
	exempt := map[int][]int{2: {2, 5}, 4: {0, 7}, 7: {3, 9}, 8: {8, 11}, 12: {1, 6}}
	if inSet(ms, exempt[rep.FromLagna]...) {
		cancel = append(cancel, fmt.Sprintf("Mars in %s in the %s house from the lagna is a traditional exception.", rashiNames[ms], ordinal(rep.FromLagna)))
	}
	if lg := int(c.Ascendant.Longitude/30) % 12; lg == 3 || lg == 4 {
		cancel = append(cancel, fmt.Sprintf("For %s rising, Mars is a yogakaraka (a beneficial ruler), so the dosha is not applied.", rashiNames[lg]))
	}
	js := signOf(g["jupiter"])
	if inSet(ms, aspectedSigns("jupiter", js)...) {
		cancel = append(cancel, "Jupiter joins or aspects Mars, which softens the dosha.")
	}
	if cancel == nil {
		cancel = []string{}
	}
	rep.Cancellations = cancel
	rep.Effective = len(cancel) == 0
	return rep
}
