package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/example/panchang/engine"
)

// VarshaCalculator casts the Tajika year chart; test doubles may omit it.
type VarshaCalculator interface {
	SolarReturn(natal engine.Chart, year int) (time.Time, error)
	Varshaphal(natal engine.Chart, year int) (engine.VarshaReport, error)
}

// varshaphal answers GET /api/varshaphal: the year chart for the running
// birthday year (the return on or before now), or for ?year=YYYY.
func (s *Server) varshaphal(w http.ResponseWriter, r *http.Request) {
	in, err := s.chartInput(r.URL.Query())
	if err != nil {
		problem(w, 400, err)
		return
	}
	natal, err := s.engine.BirthChart(in)
	if err != nil {
		problem(w, 400, err)
		return
	}
	vc, ok := s.engine.(VarshaCalculator)
	if !ok {
		problem(w, 503, fmt.Errorf("the year chart is unavailable"))
		return
	}
	now := time.Now()
	if v := r.URL.Query().Get("at"); v != "" {
		if t, e := time.Parse(time.RFC3339, v); e == nil {
			now = t
		}
	}
	var rep engine.VarshaReport
	if v := r.URL.Query().Get("year"); v != "" {
		y, e := strconv.Atoi(v)
		if e != nil {
			problem(w, 400, fmt.Errorf("year must be a number"))
			return
		}
		if rep, err = vc.Varshaphal(natal, y); err != nil {
			problem(w, 400, err)
			return
		}
	} else if p := s.runningVarsha(natal, now); p != nil {
		rep = *p
	} else {
		problem(w, 400, fmt.Errorf("the year chart could not be cast for this birth date"))
		return
	}
	writeJSON(w, 200, map[string]any{"varsha": rep, "until": rep.Until})
}

// runningVarsha is the year chart for the birthday year that contains now,
// or nil when it cannot be cast.
func (s *Server) runningVarsha(natal engine.Chart, now time.Time) *engine.VarshaReport {
	vc, ok := s.engine.(VarshaCalculator)
	if !ok {
		return nil
	}
	year := now.Year()
	if at, err := vc.SolarReturn(natal, year); err != nil {
		return nil
	} else if at.After(now) {
		year--
	}
	rep, err := vc.Varshaphal(natal, year)
	if err != nil {
		return nil
	}
	return &rep
}
