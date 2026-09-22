package api

import (
	"github.com/example/panchang/engine"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeCalc struct{}

func (fakeCalc) Calculate(d time.Time, l engine.Location) (engine.Computed, error) {
	return engine.Computed{Day: engine.Day{SchemaVersion: 1, Date: d.Format("2006-01-02"), Location: l, Tithi: engine.Limb{Name: "Ekadashi"}, Festivals: []string{}}}, nil
}
func (fakeCalc) BirthChart(in engine.ChartInput) (engine.Chart, error) {
	return engine.Chart{SchemaVersion: 1, Input: in, Ayanamsa: "lahiri", Ascendant: engine.Point{Longitude: 0, Rashi: "Mesha"}, Grahas: []engine.Graha{{ID: "sun", Name: "Surya", Rashi: "Mesha", Nakshatra: "Ashwini"}, {ID: "moon", Name: "Chandra", Rashi: "Mesha", Nakshatra: "Ashwini"}, {ID: "mars", Name: "Mangala", Rashi: "Mesha"}, {ID: "mercury", Name: "Budha", Rashi: "Mesha"}, {ID: "jupiter", Name: "Guru", Rashi: "Mesha"}, {ID: "venus", Name: "Shukra", Rashi: "Mesha"}, {ID: "saturn", Name: "Shani", Rashi: "Mesha"}, {ID: "rahu", Name: "Rahu", Rashi: "Mesha"}, {ID: "ketu", Name: "Ketu", Rashi: "Mesha"}}}, nil
}
func TestPanchangValidation(t *testing.T) {
	s := NewServer(fakeCalc{}, NoCache{}, nil)
	r := httptest.NewRequest("GET", "/api/panchang?date=nope", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatalf("status %d", w.Code)
	}
}
func TestPanchang(t *testing.T) {
	s := NewServer(fakeCalc{}, NoCache{}, nil)
	r := httptest.NewRequest("GET", "/api/panchang?date=2026-09-22&lat=12.97&lon=77.59&tz=Asia/Kolkata", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
}
func TestChartFacts(t *testing.T) {
	s := NewServer(fakeCalc{}, NoCache{}, nil)
	r := httptest.NewRequest("GET", "/api/chart/facts?date=2026-09-22&time=10:15&lat=12.97&lon=77.59&tz=Asia%2FKolkata&as_of=2026-09-22T00:00:00Z", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
}
func TestReading(t *testing.T) {
	s := NewServer(fakeCalc{}, NoCache{}, nil)
	r := httptest.NewRequest("GET", "/api/reading?date=2026-09-22&time=10:15&lat=12.97&lon=77.59&tz=Asia%2FKolkata&as_of=2026-09-22T00:00:00Z", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
}
