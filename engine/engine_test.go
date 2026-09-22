package engine

import (
	"math"
	"testing"
	"time"

	"github.com/example/panchang/engine/swe"
)

func TestJulianJ2000(t *testing.T) {
	got := swe.JulianDay(2000, 1, 1, 12)
	if math.Abs(got-2451545) > 1e-8 {
		t.Fatalf("got %.9f", got)
	}
}
func TestKaranaSequence(t *testing.T) {
	cases := map[int]string{0: "Kimstughna", 1: "Bava", 2: "Balava", 7: "Vishti", 8: "Bava", 57: "Shakuni", 58: "Chatushpada", 59: "Naga"}
	for n, want := range cases {
		if got := karanaName(n); got != want {
			t.Errorf("%d: %s != %s", n, got, want)
		}
	}
}
func TestDaytimeWindows(t *testing.T) {
	z := time.UTC
	r := time.Date(2026, 9, 22, 6, 0, 0, 0, z)
	s := time.Date(2026, 9, 22, 18, 0, 0, 0, z)
	d := Day{}
	applyWindows(&d, r, s, z, time.Tuesday)
	if d.RahuKaal.Start != "15:00" || d.RahuKaal.End != "16:30" {
		t.Fatalf("rahu: %+v", d.RahuKaal)
	}
	if len(d.Choghadiya) != 16 {
		t.Fatalf("choghadiya count %d", len(d.Choghadiya))
	}
}
func TestBangaloreSmoke(t *testing.T) {
	e := New("../ephe")
	d, err := e.Calculate(time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC), Location{12.9716, 77.5946, "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Day.Sunrise < "05:30" || d.Day.Sunrise > "06:45" {
		t.Fatalf("implausible sunrise %s", d.Day.Sunrise)
	}
	if d.Day.Tithi.Number < 1 || d.Day.Tithi.Number > 30 {
		t.Fatal(d.Day.Tithi)
	}
}

func TestBirthChartNodesAreOpposite(t *testing.T) {
	e := New("../ephe")
	c, err := e.BirthChart(ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Grahas) != 9 {
		t.Fatalf("grahas: %d", len(c.Grahas))
	}
	var rahu, ketu Graha
	for _, g := range c.Grahas {
		if g.ID == "rahu" {
			rahu = g
		}
		if g.ID == "ketu" {
			ketu = g
		}
	}
	if math.Abs(normalize(ketu.Longitude-rahu.Longitude)-180) > 1e-9 {
		t.Fatalf("nodes are not opposite: %.9f %.9f", rahu.Longitude, ketu.Longitude)
	}
	if c.Ascendant.Rashi == "" || len(c.Houses.Cusps) != 12 {
		t.Fatalf("invalid houses: %+v", c.Houses)
	}
}
