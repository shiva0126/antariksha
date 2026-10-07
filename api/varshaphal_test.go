package api

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVarshaphalPicksTheRunningYear(t *testing.T) {
	s := NewServer(realEngine(t), NoCache{}, nil)
	w := httptest.NewRecorder()
	s.varshaphal(w, httptest.NewRequest("GET", "/api/varshaphal?"+birthQ+"&at=2026-10-06T06:30:00Z", nil))
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var out struct {
		Varsha struct {
			Year     int       `json:"year"`
			ReturnAt time.Time `json:"return_at"`
			YearLord string    `json:"year_lord"`
			Offices  []any     `json:"offices"`
		}
		Until time.Time `json:"until"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 6, 6, 30, 0, 0, time.UTC)
	if out.Varsha.ReturnAt.After(now) || !out.Until.After(now) || out.Varsha.YearLord == "" || len(out.Varsha.Offices) != 5 {
		t.Fatalf("%+v", out)
	}
}
