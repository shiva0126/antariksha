package engine

import "strconv"

func yoga(name, typ string, planets []string, houses []string, strength string, geometry map[string]any) Yoga {
	return Yoga{name, typ, planets, houses, strength, geometry}
}
func strength(c Chart, ids ...string) string {
	gs := chartMap(c)
	score := 0
	for _, id := range ids {
		g := gs[id]
		if inSet(houseOf(g.Longitude, c.Ascendant.Longitude), 1, 4, 7, 10) {
			score++
		}
		if g.Retrograde {
			score--
		}
	}
	if score >= len(ids) {
		return "strong"
	}
	if score > 0 {
		return "moderate"
	}
	return "weak"
}
func conjunction(gs map[string]Graha, a, b string) bool { return signOf(gs[a]) == signOf(gs[b]) }
func kendraFrom(a, b Graha) bool                        { return inSet((signOf(a)-signOf(b)+12)%12, 0, 3, 6, 9) }
func DetectYogas(c Chart, d map[string]Dignity) []Yoga {
	gs := chartMap(c)
	out := []Yoga{}
	if kendraFrom(gs["jupiter"], gs["moon"]) {
		out = append(out, yoga("Gajakesari", "raja", []string{"jupiter", "moon"}, []string{"kendra_from_moon"}, strength(c, "jupiter", "moon"), map[string]any{"offset": (signOf(gs["jupiter"]) - signOf(gs["moon"]) + 12) % 12}))
	}
	if conjunction(gs, "mercury", "sun") {
		out = append(out, yoga("Budha-Aditya", "buddhi", []string{"mercury", "sun"}, []string{strconv.Itoa(houseOf(gs["sun"].Longitude, c.Ascendant.Longitude))}, strength(c, "mercury", "sun"), map[string]any{"same_sign": gs["sun"].Rashi}))
	}
	if conjunction(gs, "moon", "mars") {
		out = append(out, yoga("Chandra-Mangala", "wealth", []string{"moon", "mars"}, []string{strconv.Itoa(houseOf(gs["moon"].Longitude, c.Ascendant.Longitude))}, strength(c, "moon", "mars"), map[string]any{"same_sign": gs["moon"].Rashi}))
	}
	mah := map[string]string{"mars": "Ruchaka", "mercury": "Bhadra", "jupiter": "Hamsa", "venus": "Malavya", "saturn": "Sasa"}
	for id, name := range mah {
		if dg := d[id]; (dg.State == "own" || dg.State == "exalted") && inSet(houseOf(gs[id].Longitude, c.Ascendant.Longitude), 1, 4, 7, 10) {
			out = append(out, yoga(name, "mahapurusha", []string{id}, []string{strconv.Itoa(houseOf(gs[id].Longitude, c.Ascendant.Longitude))}, strength(c, id), map[string]any{"dignity": dg.State, "house": houseOf(gs[id].Longitude, c.Ascendant.Longitude)}))
		}
	}
	if y := rajaYoga(c, gs); y.Name != "" {
		out = append(out, y)
	}
	if y := dhanaYoga(c, gs); y.Name != "" {
		out = append(out, y)
	}
	if y := vipreet(c, gs); y.Name != "" {
		out = append(out, y)
	}
	if y := adhi(c, gs); y.Name != "" {
		out = append(out, y)
	}
	if y := moonPattern(c, gs, "Sunapha", 1); y.Name != "" {
		out = append(out, y)
	}
	if y := moonPattern(c, gs, "Anapha", -1); y.Name != "" {
		out = append(out, y)
	}
	if y := moonPattern(c, gs, "Durudhara", 0); y.Name != "" {
		out = append(out, y)
	}
	if kemadruma(c, gs) {
		out = append(out, yoga("Kemadruma", "caution", nil, []string{"from_moon"}, "caution", map[string]any{"exception_checks": []string{"no_planets_2nd_or_12th", "no_kendra_from_moon"}}))
	}
	for id, dg := range d {
		if dg.NeechaBhanga {
			out = append(out, yoga("Neecha Bhanga Raja Yoga", "raja", []string{id}, []string{"cancellation_geometry"}, "moderate", map[string]any{"planet": id, "debilitation_cancelled": true}))
		}
	}
	if kalaSarpa(c, gs) {
		out = append(out, yoga("Kala Sarpa", "caution", []string{"rahu", "ketu"}, []string{"between_nodes"}, "caution", map[string]any{"axis": "rahu_to_ketu"}))
	}
	return out
}

// lordsOf returns the distinct lords of the given houses from the lagna.
func lordsOf(c Chart, houses ...int) []string {
	asc := int(c.Ascendant.Longitude / 30)
	seen := map[string]bool{}
	var out []string
	for _, h := range houses {
		l := rashiLord[(asc+h-1)%12]
		if !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}
	return out
}

// associated reports conjunction, mutual 7th-house aspect or sign exchange
// (parivartana), the three classical forms of sambandha used for Raja Yoga.
func associated(gs map[string]Graha, a, b string) (string, bool) {
	sa, sb := signOf(gs[a]), signOf(gs[b])
	switch {
	case sa == sb:
		return "conjunction", true
	case (sa-sb+12)%12 == 6:
		return "mutual_aspect", true
	case rashiLord[sa] == b && rashiLord[sb] == a:
		return "exchange", true
	}
	return "", false
}

func rajaYoga(c Chart, gs map[string]Graha) Yoga {
	kendra, trikona := lordsOf(c, 1, 4, 7, 10), lordsOf(c, 1, 5, 9)
	// A single planet ruling both a kendra (other than the lagna) and a trikona
	// is a yogakaraka and forms Raja Yoga by itself.
	for _, k := range lordsOf(c, 4, 7, 10) {
		for _, t := range lordsOf(c, 5, 9) {
			if k == t {
				return yoga("Raja Yoga", "raja", []string{k}, []string{"yogakaraka"}, strength(c, k), map[string]any{"yogakaraka": k})
			}
		}
	}
	for _, k := range kendra {
		for _, t := range trikona {
			if k == t {
				continue
			}
			if how, ok := associated(gs, k, t); ok {
				return yoga("Raja Yoga", "raja", []string{k, t}, []string{"kendra_trikona_association"}, strength(c, k, t), map[string]any{"kendra_lord": k, "trikona_lord": t, "association": how})
			}
		}
	}
	return Yoga{}
}
func dhanaYoga(c Chart, gs map[string]Graha) Yoga {
	asc := int(c.Ascendant.Longitude / 30)
	lords := []string{}
	for s, id := range rashiLord {
		h := (s-asc+12)%12 + 1
		if inSet(h, 2, 5, 9, 11) {
			lords = append(lords, id)
		}
	}
	for i := range lords {
		for j := i + 1; j < len(lords); j++ {
			if signOf(gs[lords[i]]) == signOf(gs[lords[j]]) {
				return yoga("Dhana Yoga", "wealth", []string{lords[i], lords[j]}, []string{"2_5_9_11_lords"}, strength(c, lords[i], lords[j]), map[string]any{"associated_lords": lords})
			}
		}
	}
	return Yoga{}
}
func vipreet(c Chart, gs map[string]Graha) Yoga {
	asc := int(c.Ascendant.Longitude / 30)
	for s, id := range rashiLord {
		h := (s-asc+12)%12 + 1
		if inSet(h, 6, 8, 12) && inSet(houseOf(gs[id].Longitude, c.Ascendant.Longitude), 6, 8, 12) {
			return yoga("Vipreet Raja Yoga", "raja", []string{id}, []string{"dusthana_lord_in_dusthana"}, strength(c, id), map[string]any{"lord_house": h, "placed_house": houseOf(gs[id].Longitude, c.Ascendant.Longitude)})
		}
	}
	return Yoga{}
}
func adhi(c Chart, gs map[string]Graha) Yoga {
	moon := signOf(gs["moon"])
	found := []string{}
	for _, id := range []string{"jupiter", "venus", "mercury"} {
		if inSet((signOf(gs[id])-moon+12)%12, 5, 6, 7) {
			found = append(found, id)
		}
	}
	if len(found) > 0 {
		return yoga("Adhi Yoga", "raja", found, []string{"6_7_8_from_moon"}, strength(c, found...), map[string]any{"benefics": found})
	}
	return Yoga{}
}
func moonPattern(c Chart, gs map[string]Graha, name string, offset int) Yoga {
	m := signOf(gs["moon"])
	left, right := false, false
	for _, g := range gs {
		if g.ID == "sun" || g.ID == "moon" || g.ID == "rahu" || g.ID == "ketu" {
			continue
		}
		switch (signOf(g) - m + 12) % 12 {
		case 1:
			left = true
		case 11:
			right = true
		}
	}
	// Sunapha and Anapha are one-sided; planets on both sides form Durudhara instead.
	ok := (offset == 1 && left && !right) || (offset == -1 && right && !left) || (offset == 0 && left && right)
	if !ok {
		return Yoga{}
	}
	return yoga(name, "chandra", nil, []string{"from_moon"}, "moderate", map[string]any{"planets_in_2nd": left, "planets_in_12th": right})
}
func kemadruma(c Chart, gs map[string]Graha) bool {
	m := signOf(gs["moon"])
	hasAdjacent := false
	hasKendra := false
	for _, g := range gs {
		if g.ID == "sun" || g.ID == "moon" || g.ID == "rahu" || g.ID == "ketu" {
			continue
		}
		d := (signOf(g) - m + 12) % 12
		if d == 1 || d == 11 {
			hasAdjacent = true
		}
		if inSet(d, 0, 3, 6, 9) {
			hasKendra = true
		}
	}
	return !hasAdjacent && !hasKendra
}

// kalaSarpa: all seven planets strictly on one side of the Rahu–Ketu axis
// (either hemisphere; the Ketu-to-Rahu side is sometimes called Kala Amrita).
func kalaSarpa(c Chart, gs map[string]Graha) bool {
	r := gs["rahu"].Longitude
	side := 0
	for _, g := range gs {
		if g.ID == "rahu" || g.ID == "ketu" {
			continue
		}
		d := normalize(g.Longitude - r)
		s := 1
		if d > 180 {
			s = -1
		}
		if d == 0 || d == 180 {
			return false
		}
		if side == 0 {
			side = s
		} else if side != s {
			return false
		}
	}
	return side != 0
}
