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
	return engine.Chart{SchemaVersion: 1, Input: in, Ayanamsa: "lahiri"}, nil
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
