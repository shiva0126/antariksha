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
