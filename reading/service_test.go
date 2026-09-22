package reading

import (
	"context"
	"github.com/example/panchang/engine"
	"testing"
	"time"
)

type fakeLLM struct{ body string }

func (f fakeLLM) Complete(context.Context, string) (string, error) { return f.body, nil }
func sampleFacts() engine.ChartFacts {
	return engine.ChartFacts{Chart: engine.Chart{Ayanamsa: "lahiri", Ascendant: engine.Point{Rashi: "Karka"}, Grahas: []engine.Graha{{Name: "Chandra", Rashi: "Meena", Nakshatra: "Revati"}}}, Yogas: []engine.Yoga{{Name: "Gajakesari", Strength: "strong"}}, Vimshottari: engine.Dasha{Current: engine.DashaPeriod{Maha: "saturn", Antara: "mercury"}}}
}
func TestFallbackContainsExactlyDetectedYogas(t *testing.T) {
	r, e := NewService(nil, nil).Generate(context.Background(), sampleFacts(), nil)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Yogas) != 1 || r.Yogas[0].Name != "Gajakesari" {
		t.Fatalf("%+v", r.Yogas)
	}
}
func TestLLMUnknownYogaRejected(t *testing.T) {
	bad := `{"summary":"x","lagna_and_moon":"x","grahas":[],"yogas":[{"name":"Invented","meaning":"x","effect":"x","strength":"strong"}],"dashas":{},"themes":{},"disclaimer":"For reflection, not certainty; this is not medical, legal or financial advice."}`
	_, e := NewService(nil, fakeLLM{bad}).Generate(context.Background(), sampleFacts(), nil)
	if e == nil {
		t.Fatal("unknown yoga accepted")
	}
}
func TestChartHashStable(t *testing.T) {
	in := engine.ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"}
	if ChartHash(in, "en") != ChartHash(in, "en") {
		t.Fatal("hash not stable")
	}
	_ = time.Now()
}
