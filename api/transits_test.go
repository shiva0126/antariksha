package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/panchang/corpus"
	"github.com/example/panchang/engine"
	"github.com/example/panchang/reading"
)

func TestTransitsReadsTheSkyAndListsEvents(t *testing.T) {
	// The book passages come from the corpus build; without the scanned
	// books on disk the test still checks everything but the book views.
	entries, _, err := corpus.Build(corpus.BuildOptions{RawDir: "../corpus/raw", AllowMissingRaw: true})
	if err != nil {
		t.Fatal(err)
	}
	withBooks := false
	for _, e := range entries {
		withBooks = withBooks || e.DocType == engine.DocTransit
	}
	s := NewServerWithReading(realEngine(t), NoCache{}, nil, reading.NewService(reading.NewMemoryCorpus(entries), nil))
	w := httptest.NewRecorder()
	s.transits(w, httptest.NewRequest("GET", "/api/transits?"+birthQ+"&at=2026-10-06T06:30:00Z&months=24", nil))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var out struct {
		Planets []struct {
			Graha      string `json:"graha"`
			FromMoon   int    `json:"from_moon"`
			Favourable *bool  `json:"favourable"`
			Book       *struct{ Ref, Plain string }
		} `json:"planets"`
		Events []struct {
			Graha, Kind, Rashi string
			FromMoon           int   `json:"from_moon"`
			Favourable         *bool `json:"favourable"`
			Book               *struct{ Ref, Plain string }
		} `json:"events"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Planets) != 9 {
		t.Fatalf("%d planets", len(out.Planets))
	}
	for _, p := range out.Planets {
		classical := p.Graha != "rahu" && p.Graha != "ketu"
		if classical != (p.Favourable != nil) {
			t.Errorf("%s favourable %v", p.Graha, p.Favourable)
		}
		if withBooks && classical && (p.Book == nil || !strings.HasPrefix(p.Book.Ref, "104.") || p.Book.Plain == "") {
			t.Errorf("%s in the %dth from the Moon: no Brihat Samhita view %+v", p.Graha, p.FromMoon, p.Book)
		}
	}
	kinds := map[string]int{}
	for _, e := range out.Events {
		kinds[e.Kind]++
		if e.Kind == "ingress" && (e.FromMoon < 1 || e.FromMoon > 12) {
			t.Errorf("%s ingress into %s: house %d", e.Graha, e.Rashi, e.FromMoon)
		}
		if e.Kind == "ingress" && withBooks && e.Favourable != nil && e.Book == nil {
			t.Errorf("%s ingress into the %dth from the Moon has no book view", e.Graha, e.FromMoon)
		}
	}
	// Two years always hold sign changes of Jupiter and Saturn's neighbours,
	// stations and eclipses.
	if kinds["ingress"] < 20 || kinds["retrograde"] < 4 || kinds["direct"] < 4 || kinds["solar_eclipse"] < 2 || kinds["lunar_eclipse"] < 2 {
		t.Errorf("events by kind %v", kinds)
	}
}
