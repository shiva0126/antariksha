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
