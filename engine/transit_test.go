package engine

import (
	"testing"
	"time"
)

// Brihat Samhita 104.4, read off the verse: benefic houses from the natal
// Moon. "All the planets produce benefic effects when they pass through the
// 11th house from the Moon"; Venus is malefic in the 7th, 6th and 10th.
func TestGocharaTableMatchesVerse(t *testing.T) {
	good := map[string][]int{
		"sun":     {6, 3, 10, 11},
		"moon":    {3, 6, 10, 7, 1, 11},
		"jupiter": {7, 9, 2, 5, 11},
		"mars":    {6, 3, 11},
		"saturn":  {6, 3, 11},
		"mercury": {6, 2, 4, 10, 8, 11},
	}
	for id, hs := range good {
		for h := 1; h <= 12; h++ {
			fav, known := GocharaFavourable(id, h)
			if !known || fav != inSet(h, hs...) {
				t.Errorf("%s in the %dth from the Moon: favourable=%v known=%v", id, h, fav, known)
			}
		}
	}
	for h := 1; h <= 12; h++ {
		fav, _ := GocharaFavourable("venus", h)
		if fav == inSet(h, 7, 6, 10) {
			t.Errorf("venus in the %dth from the Moon: favourable=%v", h, fav)
		}
	}
	if _, known := GocharaFavourable("rahu", 3); known {
		t.Error("the verse does not cover Rahu")
	}
}

// Published sidereal (Lahiri) dates for 2025, in IST.
func TestTransitEventsKnownDates(t *testing.T) {
	e := New("../ephe")
	ev, err := e.TransitEvents(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	ist := time.FixedZone("IST", 5*3600+1800)
	want := []struct{ graha, kind, rashi, at string }{
		{"saturn", "ingress", "Meena", "2025-03-29 21:4"},    // published 21:45
		{"jupiter", "ingress", "Mithuna", "2025-05-14 22:3"}, // published about 22:36
		{"jupiter", "ingress", "Karka", "2025-10-18"},
		{"saturn", "retrograde", "Meena", "2025-07-13"},
		{"saturn", "direct", "Meena", "2025-11-28"},
		{"jupiter", "retrograde", "Karka", "2025-11-11"},
		{"moon", "lunar_eclipse", "Simha", "2025-03-14"},
		{"moon", "lunar_eclipse", "Kumbha", "2025-09-07"},
		{"sun", "solar_eclipse", "Meena", "2025-03-29"},
		{"sun", "solar_eclipse", "Kanya", "2025-09-22"}, // 19:42 UTC on the 21st
	}
	for _, w := range want {
		found := false
		for _, x := range ev {
			if x.Graha == w.graha && x.Kind == w.kind && x.Rashi == w.rashi && x.At.In(ist).Format("2006-01-02 15:04")[:len(w.at)] == w.at {
				found = true
			}
		}
		if !found {
			t.Errorf("missing %s %s into %s at %s", w.graha, w.kind, w.rashi, w.at)
		}
	}
	// Every ingress is exact to the minute: the old sign just before, the new one just after.
	defer e.begin()()
	ids := map[string]int{}
	for _, b := range eventBodies {
		ids[b.id] = b.body
	}
	for _, x := range ev {
		body, ok := ids[x.Graha]
		if x.Kind != "ingress" || !ok {
			continue
		}
		before := lonAt(t, x.At.Add(-2*time.Minute), body)
		after := lonAt(t, x.At.Add(2*time.Minute), body)
		if rashiNames[int(before/30)] != x.From || rashiNames[int(after/30)] != x.Rashi {
			t.Errorf("%s ingress %s at %s: %s before, %s after", x.Graha, x.Rashi, x.At, rashiNames[int(before/30)], rashiNames[int(after/30)])
		}
	}
}

func lonAt(t *testing.T, at time.Time, body int) float64 {
	t.Helper()
	l, _, _, _, err := swePosition3D(jdOf(at), body)
	if err != nil {
		t.Fatal(err)
	}
	return normalize(l)
}

func TestTransitsReadFromTheMoon(t *testing.T) {
	e := New("../ephe")
	natal, err := e.BirthChart(ChartInput{Date: "1994-08-17", Time: "06:42", Lat: 13.34, Lon: 74.75, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	sky, err := e.BirthChart(ChartInput{Date: "2026-10-06", Time: "12:00", Lat: 13.34, Lon: 74.75, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	rep := Transits(natal, sky)
	nat := chartMap(natal)
	if len(rep.Planets) != 9 {
		t.Fatalf("%d planets", len(rep.Planets))
	}
	for _, p := range rep.Planets {
		g := chartMap(sky)[p.Graha]
		if want := (signOf(g)-signOf(nat["moon"])+12)%12 + 1; p.FromMoon != want {
			t.Errorf("%s from Moon %d, want %d", p.Graha, p.FromMoon, want)
		}
		if want := HouseOf(g.Longitude, natal.Ascendant.Longitude); p.FromLagna != want {
			t.Errorf("%s from lagna %d, want %d", p.Graha, p.FromLagna, want)
		}
		if (p.Graha == "rahu" || p.Graha == "ketu") != (p.Favourable == nil) {
			t.Errorf("%s favourable %v", p.Graha, p.Favourable)
		}
		if (p.Graha == "rahu" || p.Graha == "ketu") != (p.Bindus == -1) || p.Bindus > 8 {
			t.Errorf("%s bindus %d", p.Graha, p.Bindus)
		}
	}
	on, phase := SadeSati(natal, sky)
	if on != rep.SadeSati || phase != rep.SadeSatiPhase {
		t.Errorf("sade sati %v %d vs %v %d", rep.SadeSati, rep.SadeSatiPhase, on, phase)
	}
}
