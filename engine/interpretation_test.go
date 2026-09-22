package engine

import (
	"testing"
	"time"
)

func TestFactsReferenceHasDeterministicFeatures(t *testing.T) {
	e := New("../ephe")
	c, err := e.BirthChart(ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	f, err := Facts(c, time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if f.Dignities["sun"].State != "exalted" {
		t.Errorf("sun dignity=%+v", f.Dignities["sun"])
	}
	if !f.Combustion["mercury"] {
		t.Errorf("mercury should be combust at reference distance")
	}
	if f.Vimshottari.BirthBalance.Lord != "mercury" || f.Vimshottari.BirthBalance.YearsRemaining <= 0 {
		t.Fatalf("dasha=%+v", f.Vimshottari)
	}
	if f.Vimshottari.Current.Maha == "" || f.Vimshottari.Current.Antara == "" {
		t.Fatalf("current dasha=%+v", f.Vimshottari.Current)
	}
	for _, y := range f.Yogas {
		if y.Name == "" || y.Strength == "" || len(y.Geometry) == 0 {
			t.Fatalf("incomplete yoga=%+v", y)
		}
	}
}
func TestYogaRulesUseGeometry(t *testing.T) {
	c := Chart{Ayanamsa: "lahiri", Ascendant: Point{Longitude: 0, Rashi: "Mesha"}, Grahas: []Graha{{ID: "sun", Longitude: 0, Rashi: "Mesha"}, {ID: "moon", Longitude: 30, Rashi: "Vrishabha"}, {ID: "mars", Longitude: 30, Rashi: "Vrishabha"}, {ID: "mercury", Longitude: 0, Rashi: "Mesha"}, {ID: "jupiter", Longitude: 120, Rashi: "Simha"}, {ID: "venus", Longitude: 180, Rashi: "Tula"}, {ID: "saturn", Longitude: 270, Rashi: "Makara"}, {ID: "rahu", Longitude: 60, Rashi: "Mithuna"}, {ID: "ketu", Longitude: 240, Rashi: "Dhanu"}}}
	ys := DetectYogas(c, Dignities(c))
	found := map[string]bool{}
	for _, y := range ys {
		found[y.Name] = true
	}
	for _, name := range []string{"Budha-Aditya", "Chandra-Mangala", "Gajakesari"} {
		if !found[name] {
			t.Errorf("missing %s: %+v", name, ys)
		}
	}
}

func TestVimshottariSequenceAndAntara(t *testing.T) {
	e := New("../ephe")
	c, err := e.BirthChart(ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	d := Vimshottari(c, time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC))
	// Moon 350.52° is 28.9% through Revati (Mercury): 12.08 years of Mercury
	// remain, then Ketu 7, Venus 20, Sun 6 ... in strict Vimshottari order.
	want := []string{"mercury", "ketu", "venus", "sun", "moon", "mars", "rahu", "jupiter", "saturn", "mercury"}
	for i, l := range want {
		if d.Sequence[i].Lord != l {
			t.Fatalf("period %d lord %s want %s (sequence %+v)", i, d.Sequence[i].Lord, l, d.Sequence[:4])
		}
	}
	if d.Sequence[0].To[:7] != "2008-06" || d.Sequence[1].To[:7] != "2015-06" || d.Sequence[2].To[:7] != "2035-06" {
		t.Fatalf("period ends %s %s %s", d.Sequence[0].To, d.Sequence[1].To, d.Sequence[2].To)
	}
	cur := d.Current
	if cur.Maha != "venus" || cur.Antara != "jupiter" || cur.From[:7] != "2025-08" || cur.To[:7] != "2028-04" {
		t.Fatalf("current %+v, want venus–jupiter 2025-08..2028-04", cur)
	}
	if d.Upcoming.Lord != "sun" {
		t.Fatalf("upcoming %+v", d.Upcoming)
	}
}

func TestBirthMahaAntaraUsesTheoreticalStart(t *testing.T) {
	e := New("../ephe")
	c, _ := e.BirthChart(ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	// Mercury maha began ~1991-05 (4.9 years before birth); Mercury/Mercury
	// (2.41y) → Ketu (0.99y) → Venus (2.83y) runs 1994-10..1997-08.
	d := Vimshottari(c, time.Date(1996, 6, 1, 0, 0, 0, 0, time.UTC))
	if d.Current.Maha != "mercury" || d.Current.Antara != "venus" {
		t.Fatalf("at birth: %+v", d.Current)
	}
}

func TestDignityEnemyAndFriend(t *testing.T) {
	c := Chart{Ascendant: Point{Longitude: 0}, Grahas: []Graha{{ID: "sun", Longitude: 45}, {ID: "moon", Longitude: 100}, {ID: "saturn", Longitude: 125}, {ID: "mars", Longitude: 250}, {ID: "venus", Longitude: 70}, {ID: "mercury", Longitude: 200}, {ID: "jupiter", Longitude: 10}}}
	d := Dignities(c)
	for id, want := range map[string]string{"sun": "enemy", "saturn": "enemy", "mars": "friendly", "venus": "friendly", "mercury": "friendly", "jupiter": "friendly", "moon": "own"} {
		if d[id].State != want {
			t.Errorf("%s: %s want %s", id, d[id].State, want)
		}
	}
}

func TestMoonYogasAreExclusive(t *testing.T) {
	// Mars 2nd and Venus 12th from the Moon: Durudhara only.
	c := Chart{Ayanamsa: "lahiri", Ascendant: Point{Longitude: 0}, Grahas: []Graha{{ID: "sun", Longitude: 200}, {ID: "moon", Longitude: 95}, {ID: "mars", Longitude: 125}, {ID: "mercury", Longitude: 215}, {ID: "jupiter", Longitude: 275}, {ID: "venus", Longitude: 65}, {ID: "saturn", Longitude: 300}, {ID: "rahu", Longitude: 10}, {ID: "ketu", Longitude: 190}}}
	names := map[string]bool{}
	for _, y := range DetectYogas(c, Dignities(c)) {
		names[y.Name] = true
	}
	if !names["Durudhara"] || names["Sunapha"] || names["Anapha"] {
		t.Fatalf("%v", names)
	}
}

func TestDashaLevelsYoginiAndAshtakavarga(t *testing.T) {
	e := New("../ephe")
	c, _ := e.BirthChart(ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	f, err := Facts(c, time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	v := f.Vimshottari
	if len(v.Antaras) != 9 || v.Antaras[0].Lord != "venus" || v.Antaras[5].Lord != "jupiter" {
		t.Fatalf("antaras %+v", v.Antaras)
	}
	if v.Antaras[5].From[:7] != v.Current.From[:7] || v.Antaras[8].To[:7] != "2035-06" {
		t.Fatalf("antara bounds %+v vs current %+v", v.Antaras, v.Current)
	}
	// Jupiter pratyantaras start with Jupiter; the running one must be named.
	if len(v.Pratyantaras) != 9 || v.Pratyantaras[0].Lord != "jupiter" || v.Current.Pratyantara == "" {
		t.Fatalf("pratyantaras %+v current %+v", v.Pratyantaras, v.Current)
	}
	// Revati is nakshatra 27: (27+3) mod 8 = 6 → Ulka (Saturn, 6 years).
	if f.Yogini.Sequence[0].Yogini != "Ulka" || f.Yogini.Current.Yogini == "" {
		t.Fatalf("yogini %+v", f.Yogini)
	}
	totals := map[string]int{"sun": 48, "moon": 49, "mars": 39, "mercury": 54, "jupiter": 56, "venus": 52, "saturn": 39}
	sum := 0
	for g, want := range totals {
		got := 0
		for _, b := range f.Ashtakavarga.Bhinna[g] {
			got += b
		}
		if got != want {
			t.Errorf("%s BAV total %d want %d", g, got, want)
		}
	}
	for _, b := range f.Ashtakavarga.Sarva {
		sum += b
	}
	if sum != 337 {
		t.Fatalf("SAV total %d", sum)
	}
}

func TestVargas(t *testing.T) {
	cases := []struct {
		lon  float64
		n    int
		sign int
	}{
		{0.5, 9, 0}, {29.9, 9, 8}, {30.5, 9, 9}, {95, 9, 4}, // Mesha→Mesha…Dhanu; Vrishabha starts Makara; Karka 5° → Simha
		{1, 10, 0}, {5, 10, 1}, {31, 10, 9}, // Mesha from itself (5° is the 2nd part); Vrishabha (even) from its 9th, Makara
		{5, 2, 4}, {20, 2, 3}, {35, 2, 3}, // odd: Simha then Karka; even reversed
		{25, 3, 8}, {61, 12, 2}, {40, 7, 9}, // Vrishabha 10° is the 3rd saptamsha from its 7th (Vrishchika) → Makara
	}
	for _, c := range cases {
		if s, _ := vargaSign(c.lon, c.n); s != c.sign {
			t.Errorf("D%d of %.1f°: sign %d want %d", c.n, c.lon, s, c.sign)
		}
	}
	e := New("../ephe")
	ch, _ := e.BirthChart(ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	d9, err := VargaChart(ch, 9)
	if err != nil || len(d9.Grahas) != 9 {
		t.Fatal(err)
	}
	// Lagna 90.48° is in the first navamsha of Karka → Karka.
	if d9.Ascendant.Rashi != "Karka" {
		t.Fatalf("D9 lagna %s", d9.Ascendant.Rashi)
	}
	if _, err := VargaChart(ch, 5); err == nil {
		t.Fatal("unsupported varga accepted")
	}
}

func TestMatchingKnownCases(t *testing.T) {
	e := New("../ephe")
	a, _ := e.BirthChart(ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	// Identical Moons: the classical result is 28 points with Nadi dosha.
	m, err := MatchCharts(a, a)
	if err != nil {
		t.Fatal(err)
	}
	if m.Total != 28 || m.Kootas[7].Score != 0 {
		t.Fatalf("same moon: %.1f %+v", m.Total, m.Kootas)
	}
	b, _ := e.BirthChart(ChartInput{Date: "1994-11-02", Time: "06:40", Lat: 28.61, Lon: 77.21, TZ: "Asia/Kolkata"})
	m, _ = MatchCharts(b, a)
	if m.Total < 0 || m.Total > 36 || len(m.Kootas) != 8 {
		t.Fatalf("%+v", m)
	}
	var sum float64
	for _, k := range m.Kootas {
		if k.Score < 0 || k.Score > k.Max {
			t.Fatalf("koota out of range %+v", k)
		}
		sum += k.Score
	}
	if sum != m.Total {
		t.Fatal("total mismatch")
	}
	for i := range yoniScore {
		for j := range yoniScore {
			if yoniScore[i][j] != yoniScore[j][i] {
				t.Fatalf("yoni table asymmetric at %d,%d", i, j)
			}
		}
	}
}

func TestMuhurtaWindowsAvoidInauspiciousPeriods(t *testing.T) {
	e := New("../ephe")
	c, err := e.Calculate(time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC), Location{Lat: 12.9716, Lon: 77.5946, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	ev, _ := MuhurtaEventByID("marriage")
	m := EvaluateMuhurta(ev, c.Day, -1, -1, -1)
	if !m.Good {
		t.Skipf("reference day not suitable in general muhurta: %v", m.Cautions)
	}
	for _, w := range m.Windows {
		for _, a := range m.Avoid {
			if w.Start < a.End && a.Start < w.End {
				t.Fatalf("window %v overlaps avoided %v", w, a)
			}
		}
		if w == c.Day.Abhijit {
			t.Fatal("Abhijit offered on a Wednesday")
		}
	}
}
