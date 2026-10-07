package engine

import (
	"math"
	"testing"
	"time"

	"github.com/example/panchang/engine/swe"
)

func TestSolarReturnPutsTheSunBack(t *testing.T) {
	e := New("../ephe")
	natal, err := e.BirthChart(ChartInput{Date: "1990-05-15", Time: "10:30", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	rep, err := e.Varshaphal(natal, 2026)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Age != 36 {
		t.Fatalf("age %d", rep.Age)
	}
	// The sidereal year is about 20 minutes longer than the tropical one, so
	// the return lands within a day or two of the birthday.
	bday := time.Date(2026, 5, 15, 5, 0, 0, 0, time.UTC)
	if d := rep.ReturnAt.Sub(bday); d < -48*time.Hour || d > 48*time.Hour {
		t.Fatalf("return at %v", rep.ReturnAt)
	}
	defer e.begin()()
	lon := lonAt(t, rep.ReturnAt, swe.Sun)
	if d := math.Abs(math.Mod(lon-natal.Grahas[0].Longitude+540, 360) - 180); d > 1.0/3600 {
		t.Fatalf("Sun off by %.6f degrees", d)
	}
	natalLagna := int(natal.Ascendant.Longitude/30) % 12
	if rep.Muntha.Rashi != rashiNames[(natalLagna+36)%12] {
		t.Fatalf("muntha %s", rep.Muntha.Rashi)
	}
	if len(rep.Offices) != 5 || rep.YearLord == "" {
		t.Fatalf("offices %+v lord %q", rep.Offices, rep.YearLord)
	}
}

func TestYearLordMustAspectTheYearLagna(t *testing.T) {
	// Year lagna Mesha (0). Mars in Mesha (conjunct: aspects), Venus in
	// Vrishabha (2nd: no Tajika aspect) and strongest; Venus must lose.
	vc := Chart{Ascendant: Point{Longitude: 5}, Grahas: []Graha{
		{ID: "sun", Longitude: 200}, {ID: "moon", Longitude: 100}, {ID: "mars", Longitude: 10},
		{ID: "mercury", Longitude: 190}, {ID: "jupiter", Longitude: 250}, {ID: "venus", Longitude: 40}, {ID: "saturn", Longitude: 300},
	}}
	natal := Chart{Ascendant: Point{Longitude: 35}} // Vrishabha: birth lagna lord Venus
	sb := Shadbala{Rows: []ShadbalaRow{{Graha: "venus", Ratio: 2}, {Graha: "mars", Ratio: 1.1}, {Graha: "mercury", Ratio: 1}, {Graha: "jupiter", Ratio: 1}, {Graha: "moon", Ratio: 1}, {Graha: "sun", Ratio: 1}, {Graha: "saturn", Ratio: 1}}}
	rep := readVarsha(natal, vc, 2026, 11, time.Now(), sb)
	// Muntha: Vrishabha + 11 signs = Mesha, the year lagna itself.
	if rep.Muntha.Rashi != "Mesha" || rep.Muntha.House != 1 || rep.Muntha.Lord != "mars" {
		t.Fatalf("muntha %+v", rep.Muntha)
	}
	if rep.YearLord != "mars" {
		t.Fatalf("year lord %s, offices %+v", rep.YearLord, rep.Offices)
	}
}

func TestMuddaDashaFillsTheYear(t *testing.T) {
	e := New("../ephe")
	natal, err := e.BirthChart(ChartInput{Date: "1990-05-15", Time: "10:30", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	rep, err := e.Varshaphal(natal, 2026)
	if err != nil {
		t.Fatal(err)
	}
	m := rep.Mudda
	if len(m) != 9 || !m[0].From.Equal(rep.ReturnAt) || !m[8].To.Equal(rep.Until) {
		t.Fatalf("%d periods, %v..%v", len(m), m[0].From, m[len(m)-1].To)
	}
	var moon float64
	for _, g := range natal.Grahas {
		if g.ID == "moon" {
			moon = g.Longitude
		}
	}
	// (birth star number + age - 2) mod 9, counted from the Sun (0 = Venus).
	fromSun := []string{"venus", "sun", "moon", "mars", "rahu", "jupiter", "saturn", "mercury", "ketu"}
	if want := fromSun[(nakIndex(moon)+1+36-2)%9]; m[0].Lord != want {
		t.Fatalf("first lord %s, want %s", m[0].Lord, want)
	}
	year := rep.Until.Sub(rep.ReturnAt).Hours() / 24
	for i, p := range m {
		if i > 0 && !p.From.Equal(m[i-1].To) {
			t.Fatalf("gap before %s", p.Lord)
		}
		if days := p.To.Sub(p.From).Hours() / 24; math.Abs(days-year*dashaYears[p.Lord]/120) > 0.01 {
			t.Fatalf("%s runs %.2f days", p.Lord, days)
		}
		if p.Detail == "" || p.Tone == "" {
			t.Fatalf("%+v", p)
		}
	}
	// At age 0 the sequence would start with the birth star's own lord.
	if dashaOrder[(nakIndex(moon)+0)%9] != nakLords[nakIndex(moon)%9] {
		t.Fatal("age-0 start is not the birth star lord")
	}
}
