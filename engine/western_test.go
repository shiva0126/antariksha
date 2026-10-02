package engine

import (
	"math"
	"testing"
)

func TestWesternTropicalChartAndHouseSystems(t *testing.T) {
	e := New("../ephe")
	in := ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"}
	vedic, err := e.BirthChart(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, system := range []string{"placidus", "whole_sign"} {
		got, err := e.WesternChart(in, system)
		if err != nil {
			t.Fatal(err)
		}
		if got.Zodiac != "tropical" || len(got.Planets) != 10 || len(got.Houses.Cusps) != 12 {
			t.Fatalf("bad western response: %+v", got)
		}
		if got.Planets[0].ID != "sun" || got.Planets[0].Sign != "Taurus" {
			t.Fatalf("unexpected tropical Sun: %+v", got.Planets[0])
		}
		if math.Abs(got.Planets[0].Longitude-vedic.Grahas[0].Longitude) < 15 {
			t.Fatal("Western longitude appears sidereal")
		}
		if got.Houses.System != system {
			t.Fatalf("house system %s", got.Houses.System)
		}
		for _, p := range got.Planets {
			if p.Longitude < 0 || p.Longitude >= 360 || p.House < 1 || p.House > 12 {
				t.Fatalf("invalid planet %+v", p)
			}
		}
	}
	if _, err := e.WesternChart(in, "invalid"); err == nil {
		t.Fatal("accepted unknown house system")
	}
}

func TestWesternLocalCivilTimeAndAscendantMotion(t *testing.T) {
	e := New("../ephe")
	a := ChartInput{Date: "2024-07-01", Time: "06:00", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"}
	b := a
	b.Time = "12:00"
	x, err := e.WesternChart(a, "whole_sign")
	if err != nil {
		t.Fatal(err)
	}
	y, err := e.WesternChart(b, "whole_sign")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(normalize(x.Ascendant.Longitude-y.Ascendant.Longitude)) < 60 {
		t.Fatal("ascendant did not respond to local birth time")
	}
	nonexistent := ChartInput{Date: "2024-03-10", Time: "02:30", Lat: 40.7, Lon: -74, TZ: "America/New_York"}
	if _, err = e.WesternChart(nonexistent, "placidus"); err == nil {
		t.Fatal("accepted nonexistent DST wall time")
	}
}
