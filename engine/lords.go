package engine

// Exported chart helpers for the interpretation layer. Houses are whole-sign
// from the lagna, matching the rest of the engine.

// RashiNames returns the twelve sidereal sign names, Mesha first.
func RashiNames() []string { return append([]string(nil), rashiNames...) }

// SignLord returns the graha ruling sign index 0–11.
func SignLord(sign int) string { return rashiLord[((sign%12)+12)%12] }

// HouseOf returns the whole-sign house (1–12) of a longitude from the ascendant.
func HouseOf(longitude, ascendant float64) int { return houseOf(longitude, ascendant) }

// HouseSign returns the sign index occupying house h (1–12).
func HouseSign(c Chart, h int) int { return (int(c.Ascendant.Longitude/30) + h - 1) % 12 }

// HouseLord returns the lord of house h (1–12).
func HouseLord(c Chart, h int) string { return SignLord(HouseSign(c, h)) }

// GrahasInHouse returns the ids of grahas in house h, in engine order.
func GrahasInHouse(c Chart, h int) []string {
	var out []string
	for _, g := range c.Grahas {
		if houseOf(g.Longitude, c.Ascendant.Longitude) == h {
			out = append(out, g.ID)
		}
	}
	return out
}

// GrahaEnglish returns the English name of a graha id ("jupiter" → "Jupiter").
func GrahaEnglish(id string) string {
	if n, ok := grahaEnglish[id]; ok {
		return n
	}
	return id
}

// GrahaSanskrit returns the Sanskrit name used in charts ("jupiter" → "Guru").
func GrahaSanskrit(id string) string {
	if n, ok := planetNames[id]; ok {
		return n
	}
	return id
}

// MangalDosha reports Mars in the 1st, 2nd, 4th, 7th, 8th or 12th house from the
// lagna, the common (Parashari, lagna-based) definition, with the houses found.
func MangalDosha(c Chart) (bool, int) {
	for _, g := range c.Grahas {
		if g.ID == "mars" {
			h := houseOf(g.Longitude, c.Ascendant.Longitude)
			return inSet(h, 1, 2, 4, 7, 8, 12), h
		}
	}
	return false, 0
}

// SadeSati reports whether transiting Saturn is in the 12th, 1st or 2nd sign
// from the natal Moon, and which phase (1 rising, 2 peak, 3 setting).
func SadeSati(natal, transit Chart) (bool, int) {
	var moon, saturn *Graha
	for i := range natal.Grahas {
		if natal.Grahas[i].ID == "moon" {
			moon = &natal.Grahas[i]
		}
	}
	for i := range transit.Grahas {
		if transit.Grahas[i].ID == "saturn" {
			saturn = &transit.Grahas[i]
		}
	}
	if moon == nil || saturn == nil {
		return false, 0
	}
	switch (signOf(*saturn) - signOf(*moon) + 12) % 12 {
	case 11:
		return true, 1
	case 0:
		return true, 2
	case 1:
		return true, 3
	}
	return false, 0
}
