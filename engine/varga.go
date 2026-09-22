package engine

import (
	"fmt"
	"math"
)

// Vargas are the Parashari divisional charts supported by the engine.
var Vargas = []int{1, 2, 3, 7, 9, 10, 12}

var vargaNames = map[int]string{1: "Rashi", 2: "Hora", 3: "Drekkana", 7: "Saptamsha", 9: "Navamsha", 10: "Dashamsha", 12: "Dwadashamsha"}
var vargaThemes = map[int]string{1: "body and overall life", 2: "wealth", 3: "siblings and courage", 7: "children", 9: "marriage, dharma and inner strength", 10: "career and public life", 12: "parents"}

// VargaName and VargaTheme describe a divisional chart.
func VargaName(n int) string  { return vargaNames[n] }
func VargaTheme(n int) string { return vargaThemes[n] }

// vargaSign maps a sidereal longitude to its sign in the D-n chart (BPHS rules)
// and returns the position within that division scaled to 0–30°.
func vargaSign(lon float64, n int) (int, float64) {
	lon = normalize(lon)
	sign := int(lon / 30)
	deg := math.Mod(lon, 30)
	size := 30.0 / float64(n)
	part := int(deg / size)
	if part >= n {
		part = n - 1
	}
	within := (deg - float64(part)*size) / size * 30
	odd := sign%2 == 0 // Mesha (index 0) is an odd sign
	var s int
	switch n {
	case 1:
		return sign, deg
	case 2: // Parashara hora: odd signs Sun (Simha) then Moon (Karka); even signs reversed
		if (part == 0) == odd {
			s = 4
		} else {
			s = 3
		}
	case 3: // Drekkana: the sign, its 5th, its 9th
		s = sign + part*4
	case 7: // Saptamsha: odd from the sign, even from the 7th
		if odd {
			s = sign + part
		} else {
			s = sign + 6 + part
		}
	case 9: // Navamsha: continuous from Mesha across the zodiac
		s = int(lon / (30.0 / 9))
	case 10: // Dashamsha: odd from the sign, even from the 9th
		if odd {
			s = sign + part
		} else {
			s = sign + 8 + part
		}
	case 12: // Dwadashamsha: from the sign itself
		s = sign + part
	default:
		return sign, deg
	}
	return ((s % 12) + 12) % 12, within
}

// VargaChart returns the D-n chart as an ordinary Chart: each graha (and the
// ascendant) keeps its identity but is placed in its divisional sign, with the
// degree scaled within the division, so every chart renderer works unchanged.
func VargaChart(c Chart, n int) (Chart, error) {
	if _, ok := vargaNames[n]; !ok {
		return Chart{}, fmt.Errorf("unsupported varga D%d; supported: %v", n, Vargas)
	}
	out := c
	asc, ad := vargaSign(c.Ascendant.Longitude, n)
	out.Ascendant = Point{Longitude: float64(asc)*30 + ad, Rashi: rashiNames[asc], Degree: ad}
	out.Grahas = make([]Graha, len(c.Grahas))
	for i, g := range c.Grahas {
		s, d := vargaSign(g.Longitude, n)
		g.Longitude = float64(s)*30 + d
		g.Rashi = rashiNames[s]
		g.RashiDegree = d
		out.Grahas[i] = g
	}
	cusps := make([]float64, 12)
	for i := range cusps {
		cusps[i] = float64((asc+i)%12) * 30
	}
	out.Houses = Houses{System: "whole_sign", Cusps: cusps}
	return out, nil
}
