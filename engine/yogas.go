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
		if inSet(signOf(g), 1, 4, 7, 10) {
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

func rajaYoga(c Chart, gs map[string]Graha) Yoga {
	asc := int(c.Ascendant.Longitude / 30)
	var k, t string
	for s, id := range rashiLord {
		h := ((s - asc + 12) % 12) + 1
		if inSet(h, 1, 4, 7, 10) {
			k = id
		}
		if inSet(h, 1, 5, 9) {
			t = id
		}
		if k != "" && t != "" && k != t && (signOf(gs[k]) == signOf(gs[t]) || houseOf(gs[k].Longitude, c.Ascendant.Longitude)-houseOf(gs[t].Longitude, c.Ascendant.Longitude) == 6 || signOf(gs[k]) == int(c.Ascendant.Longitude/30)) {
			return yoga("Raja Yoga", "raja", []string{k, t}, []string{"kendra_trikona_association"}, strength(c, k, t), map[string]any{"kendra_lord": k, "trikona_lord": t})
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
		d := (signOf(g) - m + 12) % 12
		if offset == 1 && d == 1 {
			left = true
		}
		if offset == -1 && d == 11 {
			right = true
		}
		if offset == 0 && (d == 1 || d == 11) {
			if d == 1 {
				left = true
			} else {
				right = true
			}
		}
	}
	ok := (offset == 1 && left) || (offset == -1 && right) || (offset == 0 && left && right)
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
func kalaSarpa(c Chart, gs map[string]Graha) bool {
	r, k := signOf(gs["rahu"]), signOf(gs["ketu"])
	for _, g := range gs {
		if g.ID == "rahu" || g.ID == "ketu" {
			continue
		}
		s := signOf(g)
		if s == r || s == k {
			return false
		}
		if (s-r+12)%12 >= (k-r+12)%12 {
			return false
		}
	}
	return true
}
