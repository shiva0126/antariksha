package engine

import (
	"math"
	"testing"
	"time"
)

// Drik Panchang, Bengaluru, 22 September 2026, inspected 2026-09-22:
// https://www.drikpanchang.com/panchang/day-panchang.html?geoname-id=1277333
// This reference is date-specific even though the source URL defaults to today.
func TestBengaluruDrikReference(t *testing.T) {
	e := New("../ephe")
	d, err := e.Calculate(time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC), Location{12.9716, 77.5946, "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Day.Tithi.Name != "Ekadashi" || d.Day.Nakshatra.Name != "Uttara Ashadha" || d.Day.Yoga.Name != "Atiganda" || d.Day.Karana.Name != "Vanija" || d.Day.LunarMonth != "Bhadrapada" {
		t.Fatalf("reference mismatch: %+v", d.Day)
	}
	for _, p := range [][2]string{{d.Day.Sunrise, "06:09"}, {d.Day.Sunset, "18:16"}, {d.Day.Tithi.EndsAt, "21:43"}, {d.Day.Nakshatra.EndsAt, "07:06"}, {d.Day.Yoga.EndsAt, "16:29"}, {d.Day.Karana.EndsAt, "08:55"}} {
		got, e := time.Parse("15:04", p[0])
		if e != nil {
			t.Fatal(e)
		}
		want, _ := time.Parse("15:04", p[1])
		if math.Abs(got.Sub(want).Minutes()) > 1 {
			t.Errorf("got %s want %s ±1 minute", p[0], p[1])
		}
	}
}
func TestWesternTimezonePreservesCivilDate(t *testing.T) {
	e := New("../ephe")
	d, err := e.Calculate(time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC), Location{40.7128, -74.006, "America/New_York"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Day.Date != "2026-09-22" || d.Day.Vaara != "Mangalavara" {
		t.Fatalf("wrong civil date: %+v", d.Day)
	}
}
