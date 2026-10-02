package api

import (
	"encoding/json"
	"fmt"
	"github.com/example/panchang/divination"
	"github.com/example/panchang/engine"
	"io"
	"net/http"
)

// Readings are private, authenticated calculations, not public profile data.
func (s *Server) divinationRoutes() {
	s.memberRoute("POST /api/numerology", s.numerology)
	s.memberRoute("POST /api/tarot", s.tarot)
	s.memberRoute("POST /api/western", s.western)
}
func readingInput(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		problem(w, 400, fmt.Errorf("invalid reading request"))
		return false
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		problem(w, 400, fmt.Errorf("send one JSON object"))
		return false
	}
	return true
}
func (s *Server) numerology(w http.ResponseWriter, r *http.Request, _ string) {
	var in struct {
		Date string `json:"date"`
		Name string `json:"name"`
		Year int    `json:"year"`
	}
	if !readingInput(w, r, &in) {
		return
	}
	out, err := divination.ReadNumerology(in.Date, in.Name, in.Year)
	if err != nil {
		problem(w, 400, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, out)
}
func (s *Server) tarot(w http.ResponseWriter, r *http.Request, _ string) {
	var in struct {
		Count     int  `json:"count"`
		Reversals bool `json:"reversals"`
	}
	if !readingInput(w, r, &in) {
		return
	}
	if in.Count != 1 && in.Count != 3 {
		problem(w, 400, fmt.Errorf("choose one or three cards"))
		return
	}
	out, err := divination.DrawTarot(in.Count, in.Reversals)
	if err != nil {
		problem(w, 500, fmt.Errorf("card draw unavailable"))
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, out)
}
func (s *Server) western(w http.ResponseWriter, r *http.Request, _ string) {
	var in struct {
		engine.ChartInput
		System string `json:"system"`
	}
	if !readingInput(w, r, &in) {
		return
	}
	calc, ok := s.engine.(interface {
		WesternChart(engine.ChartInput, string) (engine.WesternReading, error)
	})
	if !ok {
		problem(w, 503, fmt.Errorf("Western chart engine unavailable"))
		return
	}
	out, err := calc.WesternChart(in.ChartInput, in.System)
	if err != nil {
		problem(w, 400, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, out)
}
