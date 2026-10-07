package engine

import "fmt"

// Marriage-specific readings used by matching beyond the Moon-based tables:
// Papasamya (the balance of difficult planets between the two charts) and
// each partner's marriage indicators (the 7th house, its ruler, Venus and
// the Navamsha). Both are traditional; schools weigh them differently.

// PapaReport is one chart's Papasamya points.
type PapaReport struct {
	Points float64  `json:"points"`
	Detail []string `json:"detail"`
}

var papaGrahas = []string{"mars", "saturn", "rahu", "ketu", "sun"}

// PapaSamya counts the difficult planets (Mars, Saturn, Rahu, Ketu, and the
// Sun at half weight) in the 1st, 2nd, 4th, 7th, 8th and 12th houses,
// counted from the lagna (full weight), the Moon (half) and Venus
// (quarter). This is the common Kerala and Karnataka scheme.
func PapaSamya(c Chart) PapaReport {
	g := chartMap(c)
	refs := []struct {
		name   string
		sign   int
		weight float64
	}{{"lagna", int(c.Ascendant.Longitude/30) % 12, 1}, {"Moon", signOf(g["moon"]), 0.5}, {"Venus", signOf(g["venus"]), 0.25}}
	rep := PapaReport{Detail: []string{}}
	for _, id := range papaGrahas {
		w := 1.0
		if id == "sun" {
			w = 0.5
		}
		for _, r := range refs {
			h := (signOf(g[id])-r.sign+12)%12 + 1
			if inSet(h, 1, 2, 4, 7, 8, 12) {
				rep.Points += w * r.weight
				rep.Detail = append(rep.Detail, fmt.Sprintf("%s in the %s from the %s", GrahaEnglish(id), ordinal(h), r.name))
			}
		}
	}
	return rep
}

// MarriageReport reads one chart's traditional marriage indicators.
type MarriageReport struct {
	SeventhLord      string   `json:"seventh_lord"`
	SeventhLordHouse int      `json:"seventh_lord_house"`
	Strengths        []string `json:"strengths"`
	Cautions         []string `json:"cautions"`
}

// MarriageIndicators reads the 7th house, its ruler, Venus and the
// Navamsha (D9).
func MarriageIndicators(c Chart) MarriageReport {
	g := chartMap(c)
	dig := Dignities(c)
	comb := CombustionFlags(c)
	asc := c.Ascendant.Longitude
	lord := HouseLord(c, 7)
	rep := MarriageReport{SeventhLord: lord, SeventhLordHouse: houseOf(g[lord].Longitude, asc), Strengths: []string{}, Cautions: []string{}}
	strong := func(st string) bool { return st == "own" || st == "exalted" }

	switch {
	case strong(dig[lord].State):
		rep.Strengths = append(rep.Strengths, fmt.Sprintf("The 7th house ruler, %s, is strong (%s sign).", GrahaEnglish(lord), dignityWord(dig[lord].State)))
	case dig[lord].State == "debilitated" && !dig[lord].NeechaBhanga:
		rep.Cautions = append(rep.Cautions, fmt.Sprintf("The 7th house ruler, %s, is in its weakest sign.", GrahaEnglish(lord)))
	}
	if inSet(rep.SeventhLordHouse, 6, 8, 12) {
		rep.Cautions = append(rep.Cautions, fmt.Sprintf("The 7th house ruler sits in the %s house, traditionally a harder house.", ordinal(rep.SeventhLordHouse)))
	} else if inSet(rep.SeventhLordHouse, 1, 4, 5, 7, 9, 10) {
		rep.Strengths = append(rep.Strengths, fmt.Sprintf("The 7th house ruler sits in the %s house, a supportive house.", ordinal(rep.SeventhLordHouse)))
	}
	for _, id := range GrahasInHouse(c, 7) {
		switch id {
		case "jupiter", "venus", "mercury", "moon":
			rep.Strengths = append(rep.Strengths, fmt.Sprintf("%s, a gentle planet, sits in the 7th house.", GrahaEnglish(id)))
		case "mars", "saturn", "rahu", "ketu", "sun":
			rep.Cautions = append(rep.Cautions, fmt.Sprintf("%s, a demanding planet, sits in the 7th house.", GrahaEnglish(id)))
		}
	}
	if inSet((int(asc/30)+6)%12, aspectedSigns("jupiter", signOf(g["jupiter"]))[1:]...) {
		rep.Strengths = append(rep.Strengths, "Jupiter aspects the 7th house.")
	}
	switch {
	case strong(dig["venus"].State):
		rep.Strengths = append(rep.Strengths, fmt.Sprintf("Venus, the planet of partnership, is strong (%s sign).", dignityWord(dig["venus"].State)))
	case dig["venus"].State == "debilitated" && !dig["venus"].NeechaBhanga:
		rep.Cautions = append(rep.Cautions, "Venus, the planet of partnership, is in its weakest sign.")
	}
	if comb["venus"] {
		rep.Cautions = append(rep.Cautions, "Venus is close to the Sun, which traditionally mutes it.")
	}
	if v, err := VargaChart(c, 9); err == nil {
		vd := Dignities(v)
		for _, id := range []string{"venus", lord} {
			name := GrahaEnglish(id)
			if id == lord {
				name = "the 7th house ruler, " + name + ","
			}
			switch vd[id].State {
			case "exalted", "own":
				rep.Strengths = append(rep.Strengths, fmt.Sprintf("In the Navamsha (D9), the chart read for marriage, %s is strong.", name))
			case "debilitated":
				rep.Cautions = append(rep.Cautions, fmt.Sprintf("In the Navamsha (D9), the chart read for marriage, %s is weak.", name))
			}
			if id == lord && lord == "venus" {
				break
			}
		}
	}
	return rep
}

func dignityWord(st string) string {
	if st == "own" {
		return "its own"
	}
	return "its " + st
}
