package engine

import (
	"strings"
	"testing"
	"time"
)

// Reference dates from Drik Panchang for 2026 (New Delhi / Bengaluru agree on
// these). Each festival must fall on exactly that date and not the day before
// or after.
func TestFestivalDates2026(t *testing.T) {
	e := New("../ephe")
	loc := Location{28.6139, 77.209, "Asia/Kolkata"}
	cases := map[string]string{
		"Maha Shivaratri":          "2026-02-15",
		"Holi":                     "2026-03-04",
		"Krishna Janmashtami":      "2026-09-04",
		"Ganesh Chaturthi":         "2026-09-14",
		"Sharad Navaratri begins":  "2026-10-11",
		"Vijayadashami (Dussehra)": "2026-10-20",
		"Diwali (Lakshmi Puja)":    "2026-11-08",
		"Makar Sankranti":          "2026-01-14",
		"Ugadi / Gudi Padwa":       "2026-03-19", // kshaya Pratipada
		"Raksha Bandhan":           "2026-08-28",
		"Dhanteras":                "2026-11-06",
	}
	for name, date := range cases {
		d, _ := time.Parse("2006-01-02", date)
		for off := -1; off <= 1; off++ {
			day := d.AddDate(0, 0, off)
			c, err := e.Calculate(day, loc)
			if err != nil {
				t.Fatal(err)
			}
			has := false
			for _, f := range c.Day.Festivals {
				has = has || f == name
			}
			if has != (off == 0) {
				t.Errorf("%s on %s: present=%v (festivals %v)", name, day.Format("2006-01-02"), has, c.Day.Festivals)
			}
		}
	}
}

func TestFestivalYearListing(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	e := New("../ephe")
	loc := Location{12.9716, 77.5946, "Asia/Kolkata"}
	d := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	counts := map[string]int{}
	named := map[string]bool{}
	for d.Year() == 2026 {
		c, err := e.Calculate(d, loc)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range c.Day.Festivals {
			if strings.HasSuffix(f, "Ekadashi") {
				counts["Ekadashi"]++
				named[f] = true
			}
			counts[f]++
			if !isRecurring(f) && testing.Verbose() {
				t.Logf("%s %s", d.Format("2006-01-02"), f)
			}
		}
		d = d.AddDate(0, 0, 1)
	}
	// 12–13 lunations: two Ekadashis each, one Purnima/Amavasya each, 12 sankrantis.
	if counts["Ekadashi"] < 24 || counts["Ekadashi"] > 28 || counts["Purnima"] < 12 || counts["Amavasya"] < 12 {
		t.Fatalf("recurring counts %v", counts)
	}
	// 2026 has an Adhika Jyeshtha, so all 24 regular names plus Padmini and Parama occur.
	for _, n := range []string{"Nirjala Ekadashi", "Devshayani Ekadashi", "Parsva Ekadashi", "Devutthana Ekadashi", "Padmini Ekadashi", "Parama Ekadashi", "Mokshada Ekadashi"} {
		if !named[n] {
			t.Errorf("missing %s (have %v)", n, named)
		}
	}
	for _, r := range Festivals {
		if r.Month != "" && counts[r.Name] != 1 {
			t.Errorf("%s occurs %d times in 2026", r.Name, counts[r.Name])
		}
	}
}

func TestFestivalsNeverNil(t *testing.T) {
	e := New("../ephe")
	c, err := e.Calculate(time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), Location{12.9716, 77.5946, "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Day.Festivals == nil {
		t.Fatal("festivals must be an empty list, not null")
	}
}
