package engine

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/example/panchang/engine/swe"
)

// Gochara: the planets' current positions read against the birth chart.
// Classical transit results count houses from the natal Moon (Brihat
// Samhita 104); houses from the lagna are reported as well.

// gocharaGood lists, per graha, the houses from the natal Moon whose transit
// is benefic (Brihat Samhita 104.4). Every graha is benefic in the 11th.
// Venus is the exception in form: the verse names its malefic houses (6th,
// 7th and 10th), so all its other houses are benefic. Rahu and Ketu are not
// covered by the verse.
var gocharaGood = map[string][]int{
	"sun":     {3, 6, 10, 11},
	"moon":    {1, 3, 6, 7, 10, 11},
	"mars":    {3, 6, 11},
	"mercury": {2, 4, 6, 8, 10, 11},
	"jupiter": {2, 5, 7, 9, 11},
	"venus":   {1, 2, 3, 4, 5, 8, 9, 11, 12},
	"saturn":  {3, 6, 11},
}

// GocharaFavourable reports whether a graha transiting the given house from
// the natal Moon is benefic by Brihat Samhita 104.4; known is false for
// Rahu and Ketu, which the verse does not cover.
func GocharaFavourable(id string, house int) (favourable, known bool) {
	good, ok := gocharaGood[id]
	if !ok {
		return false, false
	}
	return inSet(house, good...), true
}

// activeHalf reports whether a graha is in the part of its sign where it
// gives its transit result (Brihat Samhita 104.49-50): the Sun and Mars in
// the first half, the Moon and Saturn in the second, Mercury throughout.
// Jupiter and Venus are not specified, so they count throughout.
func activeHalf(id string, degreeInSign float64) bool {
	switch id {
	case "sun", "mars":
		return degreeInSign < 15
	case "moon", "saturn":
		return degreeInSign >= 15
	}
	return true
}

// drishtiOffsets are the signs each graha aspects, counted from its own sign
// (1 = its own sign): every graha aspects the 7th; Mars also the 4th and
// 8th, Jupiter the 5th and 9th, Saturn the 3rd and 10th.
var drishtiOffsets = map[string][]int{
	"sun": {7}, "moon": {7}, "mercury": {7}, "venus": {7},
	"mars": {4, 7, 8}, "jupiter": {5, 7, 9}, "saturn": {3, 7, 10},
	"rahu": {7}, "ketu": {7},
}

// aspectedSigns returns the signs a graha in sign s aspects, with its own sign.
func aspectedSigns(id string, s int) []int {
	out := []int{s}
	for _, o := range drishtiOffsets[id] {
		out = append(out, (s+o-1)%12)
	}
	return out
}

type TransitPlanet struct {
	Graha      string  `json:"graha"`
	Rashi      string  `json:"rashi"`
	Degree     float64 `json:"degree"`
	Nakshatra  string  `json:"nakshatra"`
	Retrograde bool    `json:"retrograde"`
	FromMoon   int     `json:"from_moon"`
	FromLagna  int     `json:"from_lagna"`
	// Favourable is the Brihat Samhita 104.4 verdict for the house from the
	// Moon; absent for Rahu and Ketu.
	Favourable *bool `json:"favourable,omitempty"`
	// Active is false while the graha is in the half of the sign where the
	// verse says it does not yet give its result (104.49-50).
	Active bool `json:"active"`
	// Weakened lists 104.53 conditions under which good results fail:
	// debilitated, enemy sign, combust.
	Weakened []string `json:"weakened"`
	// Bindus is the graha's own ashtakavarga score (0-8) for the transit
	// sign in the birth chart; 4 or more is traditionally supportive. -1 for
	// Rahu and Ketu, which have no ashtakavarga.
	Bindus int `json:"bindus"`
	// Sarva is the sign's total ashtakavarga (28 or more is above average).
	Sarva int `json:"sarva"`
	// Aspects lists natal grahas this transit occupies or aspects.
	Aspects []string `json:"aspects"`
}

type TransitReport struct {
	At      string          `json:"at"`
	Planets []TransitPlanet `json:"planets"`
	// Saturn's transits from the natal Moon traditionally read as testing:
	// Sade Sati (12th, 1st, 2nd), Kantaka (4th) and Ashtama (8th).
	SadeSati      bool `json:"sade_sati"`
	SadeSatiPhase int  `json:"sade_sati_phase"`
	KantakaShani  bool `json:"kantaka_shani"`
	AshtamaShani  bool `json:"ashtama_shani"`
	// DoubleTransit lists houses from the lagna that both Jupiter and Saturn
	// occupy or aspect now. This is a modern technique (K. N. Rao), not
	// in the classical texts.
	DoubleTransit []int `json:"double_transit"`
}

// Transits reads the sky chart against the natal chart.
func Transits(natal, sky Chart) TransitReport {
	nat := chartMap(natal)
	moonSign := signOf(nat["moon"])
	lagnaSign := int(natal.Ascendant.Longitude/30) % 12
	av := ComputeAshtakavarga(natal)
	dig := Dignities(sky)
	comb := CombustionFlags(sky)
	rep := TransitReport{At: sky.Input.Date + " " + sky.Input.Time, DoubleTransit: []int{}}
	for _, g := range sky.Grahas {
		s := signOf(g)
		tp := TransitPlanet{
			Graha: g.ID, Rashi: g.Rashi, Degree: math.Round(g.RashiDegree*100) / 100, Nakshatra: g.Nakshatra, Retrograde: g.Retrograde,
			FromMoon: (s-moonSign+12)%12 + 1, FromLagna: (s-lagnaSign+12)%12 + 1,
			Active: activeHalf(g.ID, g.RashiDegree), Weakened: []string{}, Bindus: -1, Aspects: []string{},
		}
		if fav, known := GocharaFavourable(g.ID, tp.FromMoon); known {
			tp.Favourable = &fav
		}
		switch dig[g.ID].State {
		case "debilitated", "enemy":
			tp.Weakened = append(tp.Weakened, dig[g.ID].State)
		}
		if comb[g.ID] {
			tp.Weakened = append(tp.Weakened, "combust")
		}
		if row, ok := av.Bhinna[g.ID]; ok {
			tp.Bindus = row[s]
		}
		tp.Sarva = av.Sarva[s]
		for _, a := range aspectedSigns(g.ID, s) {
			for _, id := range GrahaIDs {
				if signOf(nat[id]) == a {
					tp.Aspects = append(tp.Aspects, id)
				}
			}
		}
		rep.Planets = append(rep.Planets, tp)
		if g.ID == "saturn" {
			switch tp.FromMoon {
			case 12:
				rep.SadeSati, rep.SadeSatiPhase = true, 1
			case 1:
				rep.SadeSati, rep.SadeSatiPhase = true, 2
			case 2:
				rep.SadeSati, rep.SadeSatiPhase = true, 3
			case 4:
				rep.KantakaShani = true
			case 8:
				rep.AshtamaShani = true
			}
		}
	}
	sk := chartMap(sky)
	jup := map[int]bool{}
	for _, s := range aspectedSigns("jupiter", signOf(sk["jupiter"])) {
		jup[s] = true
	}
	for _, s := range aspectedSigns("saturn", signOf(sk["saturn"])) {
		if jup[s] {
			rep.DoubleTransit = append(rep.DoubleTransit, (s-lagnaSign+12)%12+1)
		}
	}
	sort.Ints(rep.DoubleTransit)
	return rep
}

// TransitEvent is a sign change, a retrograde or direct station, or an
// eclipse, in the sidereal (Lahiri) zodiac.
type TransitEvent struct {
	Graha  string    `json:"graha"`
	Kind   string    `json:"kind"` // ingress | retrograde | direct | solar_eclipse | lunar_eclipse
	At     time.Time `json:"at"`
	From   string    `json:"from,omitempty"` // previous sign, for ingresses
	Rashi  string    `json:"rashi"`
	Degree float64   `json:"degree"` // longitude within the sign at the event
}

var eventBodies = []struct {
	id   string
	body int
}{{"sun", swe.Sun}, {"mars", swe.Mars}, {"mercury", swe.Mercury}, {"jupiter", swe.Jupiter}, {"venus", swe.Venus}, {"saturn", swe.Saturn}, {"rahu", swe.TrueNode}}

func jdOf(t time.Time) float64 {
	u := t.UTC()
	return swe.JulianDay(u.Year(), int(u.Month()), u.Day(), float64(u.Hour())+float64(u.Minute())/60+float64(u.Second())/3600)
}

func timeOfJD(jd float64) time.Time {
	y, m, d, h := swe.ReverseJulian(jd)
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC).Add(time.Duration(h * float64(time.Hour))).Round(time.Minute)
}

// TransitEvents lists sign changes and stations of the Sun to Saturn and of
// Rahu and Ketu, and solar and lunar eclipses, between from and to. The
// Moon, which changes sign every two or three days, is left out. Times are
// found to within a minute.
func (e *Engine) TransitEvents(from, to time.Time) ([]TransitEvent, error) {
	if !to.After(from) || to.Sub(from) > 5*366*24*time.Hour {
		return nil, fmt.Errorf("choose a range of up to five years")
	}
	defer e.begin()()
	j0, j1 := jdOf(from), jdOf(to)
	var out []TransitEvent
	for _, b := range eventBodies {
		pos := func(jd float64) (float64, float64, error) {
			lon, _, _, speed, err := swe.Position3D(jd, b.body)
			return normalize(lon), speed, err
		}
		lon, speed, err := pos(j0)
		if err != nil {
			return nil, err
		}
		for jd := j0; jd < j1; {
			next := math.Min(jd+0.5, j1)
			nlon, nspeed, err := pos(next)
			if err != nil {
				return nil, err
			}
			if int(nlon/30) != int(lon/30) {
				// Bisect to the minute: the moment the sign changes.
				a, z := jd, next
				for z-a > 1.0/1440/2 {
					mid := (a + z) / 2
					ml, _, _ := pos(mid)
					if int(ml/30) == int(lon/30) {
						a = mid
					} else {
						z = mid
					}
				}
				at, _, _ := pos(z)
				ev := TransitEvent{Graha: b.id, Kind: "ingress", At: timeOfJD(z), From: rashiNames[int(lon/30)], Rashi: rashiNames[int(at/30)], Degree: math.Mod(at, 30)}
				// The true node wobbles across a boundary for days; keep only
				// a crossing after which it stays a month.
				if b.id != "rahu" || stays(pos, z, int(at/30)) {
					out = append(out, ev)
					if b.id == "rahu" {
						k := normalize(at + 180)
						out = append(out, TransitEvent{Graha: "ketu", Kind: "ingress", At: ev.At, From: rashiNames[(int(lon/30)+6)%12], Rashi: rashiNames[int(k/30)], Degree: math.Mod(k, 30)})
					}
				}
			}
			// The true node's motion reverses every few days; stations are
			// meaningful for the planets only.
			if b.id != "rahu" && b.id != "sun" && (speed < 0) != (nspeed < 0) {
				a, z := jd, next
				for z-a > 1.0/1440/2 {
					mid := (a + z) / 2
					_, ms, _ := pos(mid)
					if (ms < 0) == (speed < 0) {
						a = mid
					} else {
						z = mid
					}
				}
				at, _, _ := pos(z)
				kind := "retrograde"
				if nspeed > 0 {
					kind = "direct"
				}
				out = append(out, TransitEvent{Graha: b.id, Kind: kind, At: timeOfJD(z), Rashi: rashiNames[int(at/30)], Degree: math.Mod(at, 30)})
			}
			jd, lon, speed = next, nlon, nspeed
		}
	}
	ecl, err := eclipses(j0, j1)
	if err != nil {
		return nil, err
	}
	out = append(out, ecl...)
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out, nil
}

func stays(pos func(float64) (float64, float64, error), jd float64, sign int) bool {
	for d := 1.0; d <= 30; d++ {
		l, _, err := pos(jd + d)
		if err != nil || int(l/30) != sign {
			return false
		}
	}
	return true
}

// eclipses lists solar and lunar eclipses with the sidereal position of the
// eclipsed light: the Sun for a solar eclipse, the Moon for a lunar one.
func eclipses(j0, j1 float64) ([]TransitEvent, error) {
	var out []TransitEvent
	for _, k := range []struct {
		kind string
		next func(float64) (float64, error)
		body int
	}{{"solar_eclipse", swe.NextSolarEclipse, swe.Sun}, {"lunar_eclipse", swe.NextLunarEclipse, swe.Moon}} {
		for jd := j0; ; {
			at, err := k.next(jd)
			if err != nil {
				return nil, err
			}
			if at > j1 {
				break
			}
			lon, _, _, _, err := swe.Position3D(at, k.body)
			if err != nil {
				return nil, err
			}
			lon = normalize(lon)
			out = append(out, TransitEvent{Graha: map[int]string{swe.Sun: "sun", swe.Moon: "moon"}[k.body], Kind: k.kind, At: timeOfJD(at), Rashi: rashiNames[int(lon/30)], Degree: math.Mod(lon, 30)})
			jd = at + 1
		}
	}
	return out, nil
}

// swePosition3D is the sidereal position, for tests outside the swe package.
var swePosition3D = swe.Position3D
