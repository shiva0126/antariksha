package engine

import (
	"math/rand"
	"testing"
	"time"
)

func randomChart(r *rand.Rand) Chart {
	c := Chart{Ayanamsa: "lahiri", Ascendant: Point{Longitude: r.Float64() * 360}}
	c.Ascendant.Rashi = rashiNames[int(c.Ascendant.Longitude/30)]
	for _, id := range GrahaIDs[:8] {
		speed := r.Float64()*2 - 0.5
		if id == "sun" || id == "moon" {
			speed = 1 // luminaries are never retrograde
		}
		if id == "rahu" {
			speed = -0.05
		}
		c.Grahas = append(c.Grahas, makeGraha(id, planetNames[id], r.Float64()*360, 0, 1, speed))
	}
	rahu := c.Grahas[7]
	c.Grahas = append(c.Grahas, makeGraha("ketu", "Ketu", normalize(rahu.Longitude+180), 0, 1, rahu.Speed))
	c.Input = ChartInput{Date: "1990-01-01", Time: "06:00", TZ: "UTC"}
	return c
}

func TestEveryEmittedCorpusKeyIsInVocabulary(t *testing.T) {
	vocab := map[CorpusKey]bool{}
	for _, v := range CorpusVocabulary() {
		if vocab[v.CorpusKey] {
			t.Fatalf("duplicate vocabulary token %s", v)
		}
		vocab[v.CorpusKey] = true
	}
	catalog := map[string]bool{}
	for _, y := range YogaCatalog {
		catalog[y] = true
	}
	r := rand.New(rand.NewSource(7))
	seenYoga := map[string]bool{}
	for i := 0; i < 3000; i++ {
		f, err := Facts(randomChart(r), time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatal(err)
		}
		for _, y := range f.Yogas {
			if !catalog[y.Name] {
				t.Fatalf("yoga %q missing from YogaCatalog", y.Name)
			}
			seenYoga[y.Name] = true
		}
		for _, k := range CorpusKeys(f) {
			if !vocab[k] {
				t.Fatalf("emitted token %s not in vocabulary", k)
			}
		}
	}
	if len(seenYoga) < len(YogaCatalog)-2 {
		t.Logf("random charts exercised %d/%d yogas", len(seenYoga), len(YogaCatalog))
	}
}

func TestSlug(t *testing.T) {
	for in, want := range map[string]string{"Budha-Aditya": "budha_aditya", "Neecha Bhanga Raja Yoga": "neecha_bhanga_raja_yoga", "Purva Phalguni": "purva_phalguni", "Makara": "makara"} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q)=%q want %q", in, got, want)
		}
	}
}

func TestReferenceChartKeys(t *testing.T) {
	e := New("../ephe")
	c, err := e.BirthChart(ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	f, _ := Facts(c, time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC))
	keys := map[string]bool{}
	for _, k := range CorpusKeys(f) {
		keys[k.String()] = true
	}
	for _, want := range []string{"dignity:exalted_sun", "dignity:combust_mercury", "bhava:bhava_1"} {
		if !keys[want] {
			t.Errorf("missing %s in %v", want, keys)
		}
	}
}
