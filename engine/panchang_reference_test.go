package engine

import (
	"fmt"
	"math"
	"strings"
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

// Independent published values inspected 2026-10-01, Bengaluru city selected:
// https://www.drikpanchang.com/panchang/month-panchang.html?geoname-id=1277333
// The URL defaults to the current month; this fixture freezes October 2026.
// The month grid lists sunrise tithi, sunrise/sunset and nakshatra end. It does
// NOT provide all tithi end times, so this test does not claim that coverage.
func TestBengaluruOctober2026ExternalReference(t *testing.T) {
	e := New("../ephe")
	cases := []struct {
		day                           int
		tithi, nakshatra, sunset, end string
		nextDay                       bool
	}{
		{1, "Panchami", "Rohini", "18:10", "04:27", true},
		{2, "Shashthi", "Mrigashira", "18:09", "02:55", true},
		{3, "Saptami", "Ardra", "18:08", "01:29", true},
		{4, "Navami", "Punarvasu", "18:08", "00:13", true},
		{5, "Dashami", "Pushya", "18:07", "23:09", false},
		{6, "Ekadashi", "Ashlesha", "18:06", "22:17", false},
		{7, "Dwadashi", "Magha", "18:06", "21:40", false},
		{8, "Trayodashi", "Purva Phalguni", "18:05", "21:20", false},
		{9, "Chaturdashi", "Uttara Phalguni", "18:04", "21:19", false},
		{10, "Amavasya", "Hasta", "18:04", "21:42", false},
	}
	for _, c := range cases {
		t.Run(fmt.Sprint(c.day), func(t *testing.T) {
			date := time.Date(2026, 10, c.day, 0, 0, 0, 0, time.UTC)
			result, err := e.Calculate(date, Location{12.9716, 77.5946, "Asia/Kolkata"})
			if err != nil {
				t.Fatal(err)
			}
			d := result.Day
			if d.Tithi.Name != c.tithi || d.Nakshatra.Name != c.nakshatra || d.Paksha != "Krishna" {
				t.Fatalf("names differ: tithi=%s nakshatra=%s paksha=%s", d.Tithi.Name, d.Nakshatra.Name, d.Paksha)
			}
			for _, p := range [][2]string{{d.Sunrise, "06:09"}, {d.Sunset, c.sunset}, {strings.Split(d.Nakshatra.EndsAt, " · ")[0], c.end}} {
				a, err := time.Parse("15:04", p[0])
				if err != nil {
					t.Fatal(err)
				}
				b, _ := time.Parse("15:04", p[1])
				if math.Abs(a.Sub(b).Minutes()) > 1 {
					t.Errorf("got %s want %s ±1 minute", p[0], p[1])
				}
			}
			if c.nextDay && !strings.HasSuffix(d.Nakshatra.EndsAt, date.AddDate(0, 0, 1).Format("02 Jan")) {
				t.Errorf("next-day end lost: %s", d.Nakshatra.EndsAt)
			}
			if c.day == 1 {
				for _, p := range [][2]string{{d.Tithi.EndsAt, "12:35"}, {d.Yoga.EndsAt, "21:18"}, {d.Karana.EndsAt, "12:35"}} {
					a, err := time.Parse("15:04", p[0])
					if err != nil {
						t.Fatal(err)
					}
					b, _ := time.Parse("15:04", p[1])
					if math.Abs(a.Sub(b).Minutes()) > 1 {
						t.Errorf("got %s want %s ±1 minute", p[0], p[1])
					}
				}
			}
		})
	}
}
