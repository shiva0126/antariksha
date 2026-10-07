package api

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/example/panchang/engine"
	"github.com/example/panchang/reading"
)

// TransitCalculator finds sign changes, stations and eclipses over a range.
type TransitCalculator interface {
	TransitEvents(from, to time.Time) ([]engine.TransitEvent, error)
}

type personalEvent struct {
	engine.TransitEvent
	// For sign changes: the new sign counted from the natal Moon and lagna,
	// and the Brihat Samhita 104.4 verdict for that house (nil for Rahu and
	// Ketu, which the verse does not cover).
	FromMoon   int   `json:"from_moon,omitempty"`
	FromLagna  int   `json:"from_lagna,omitempty"`
	Favourable *bool `json:"favourable,omitempty"`
	// For eclipses: whether the eclipsed light falls in the natal Moon's or
	// the lagna's sign, and how far from the natal Moon in degrees.
	OnMoonSign    bool              `json:"on_moon_sign,omitempty"`
	OnLagnaSign   bool              `json:"on_lagna_sign,omitempty"`
	DegreesToMoon float64           `json:"degrees_to_moon,omitempty"`
	Book          *reading.BookView `json:"book,omitempty"`
}

// transits answers GET /api/transits: the reading of today's sky against
// the birth chart, and the personal events of the coming months (default
// 24, at most 60): when each planet changes sign, turns retrograde or
// direct, and eclipses.
func (s *Server) transits(w http.ResponseWriter, r *http.Request) {
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
	tc, ok := s.engine.(TransitCalculator)
	if !ok {
		problem(w, 503, fmt.Errorf("transit calculation is unavailable"))
		return
	}
	zone, _ := time.LoadLocation(in.TZ)
	now := time.Now().In(zone)
	if v := r.URL.Query().Get("at"); v != "" {
		if t, e := time.Parse(time.RFC3339, v); e == nil {
			now = t.In(zone)
		}
	}
	months := 24
	if v := r.URL.Query().Get("months"); v != "" {
		if m, e := strconv.Atoi(v); e == nil && m >= 1 && m <= 60 {
			months = m
		}
	}
	sky, err := s.engine.BirthChart(engine.ChartInput{Date: now.Format("2006-01-02"), Time: now.Format("15:04"), Lat: in.Lat, Lon: in.Lon, TZ: in.TZ})
	if err != nil {
		problem(w, 500, err)
		return
	}
	report := engine.Transits(natal, sky)
	events, err := tc.TransitEvents(now, now.AddDate(0, months, 0))
	if err != nil {
		problem(w, 500, err)
		return
	}
	var moonLon float64
	for _, g := range natal.Grahas {
		if g.ID == "moon" {
			moonLon = g.Longitude
		}
	}
	moonSign, lagnaSign := int(moonLon/30)%12, int(natal.Ascendant.Longitude/30)%12
	signIndex := func(rashi string) int { return engine.RashiIndex(rashi) }

	// The book's view for every planet's current house and every sign change.
	keys := []engine.CorpusKey{}
	for _, p := range report.Planets {
		keys = append(keys, engine.CorpusKey{DocType: engine.DocTransit, Key: engine.TransitKey(p.Graha, p.FromMoon)})
	}
	out := make([]personalEvent, 0, len(events))
	for _, ev := range events {
		pe := personalEvent{TransitEvent: ev}
		sg := signIndex(ev.Rashi)
		switch ev.Kind {
		case "ingress":
			pe.FromMoon, pe.FromLagna = (sg-moonSign+12)%12+1, (sg-lagnaSign+12)%12+1
			if fav, known := engine.GocharaFavourable(ev.Graha, pe.FromMoon); known {
				pe.Favourable = &fav
			}
			keys = append(keys, engine.CorpusKey{DocType: engine.DocTransit, Key: engine.TransitKey(ev.Graha, pe.FromMoon)})
		case "solar_eclipse", "lunar_eclipse":
			pe.OnMoonSign, pe.OnLagnaSign = sg == moonSign, sg == lagnaSign
			lon := float64(sg)*30 + ev.Degree
			d := math.Abs(math.Mod(lon-moonLon+540, 360) - 180)
			pe.DegreesToMoon = math.Round(d*10) / 10
		}
		out = append(out, pe)
	}
	views := s.reading.BookViews(r.Context(), keys)
	type planetView struct {
		engine.TransitPlanet
		Book *reading.BookView `json:"book,omitempty"`
	}
	planets := make([]planetView, len(report.Planets))
	for i, p := range report.Planets {
		planets[i] = planetView{TransitPlanet: p}
		if v, ok := views[engine.DocTransit+":"+engine.TransitKey(p.Graha, p.FromMoon)]; ok {
			planets[i].Book = &v
		}
	}
	for i := range out {
		if out[i].Kind == "ingress" {
			if v, ok := views[engine.DocTransit+":"+engine.TransitKey(out[i].Graha, out[i].FromMoon)]; ok {
				out[i].Book = &v
			}
		}
	}
	periods, err := s.forecastPeriods(natal, in, events, now, now.AddDate(0, months, 0))
	if err != nil {
		problem(w, 500, err)
		return
	}
	// Recent past periods, for "did this feel true?" feedback.
	past := []reading.Period{}
	if v := r.URL.Query().Get("past"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n >= 1 && n <= 12 {
			start := now.AddDate(0, -n, 0)
			if pev, e := tc.TransitEvents(start, now); e == nil {
				if pp, e := s.forecastPeriods(natal, in, pev, start, now); e == nil {
					past = s.reading.Forecast(r.Context(), natal, pp)
				}
			}
		}
	}
	writeJSON(w, 200, map[string]any{
		"now": now.Format(time.RFC3339), "months": months,
		"periods": s.reading.Forecast(r.Context(), natal, periods), "past_periods": past,
		"planets": planets, "sade_sati": report.SadeSati, "sade_sati_phase": report.SadeSatiPhase,
		"kantaka_shani": report.KantakaShani, "ashtama_shani": report.AshtamaShani,
		"double_transit": report.DoubleTransit, "events": out,
		"notes": map[string]string{
			"houses":         "Classical transit results count houses from the natal Moon (Brihat Samhita 104); houses from the lagna are given too.",
			"nodes":          "Rahu and Ketu use the true node, as the birth chart does; many panchangs publish mean-node dates, which can differ by a week or two.",
			"double_transit": "Double transit (Jupiter and Saturn both occupying or aspecting a house) is a modern technique, not in the classical texts.",
		},
	})
}

// forecastPeriods cuts [from, to) at every sign change of Jupiter, Saturn,
// Rahu and Ketu and every change of antardasha, and takes each period's sky
// and running dasha at its middle.
func (s *Server) forecastPeriods(natal engine.Chart, in engine.ChartInput, events []engine.TransitEvent, from, to time.Time) ([]reading.PeriodInput, error) {
	cuts := []time.Time{from, to}
	for _, ev := range events {
		if ev.Kind == "ingress" && (ev.Graha == "jupiter" || ev.Graha == "saturn" || ev.Graha == "rahu") {
			cuts = append(cuts, ev.At)
		}
	}
	for t := from; t.Before(to); {
		f, err := engine.Facts(natal, t)
		if err != nil {
			return nil, err
		}
		end, err := time.Parse("2006-01-02", f.Vimshottari.Current.To)
		if err != nil || !end.After(t) {
			break
		}
		cuts = append(cuts, end)
		t = end.Add(time.Hour)
	}
	sort.Slice(cuts, func(i, j int) bool { return cuts[i].Before(cuts[j]) })
	zone, _ := time.LoadLocation(in.TZ)
	var out []reading.PeriodInput
	for i := 0; i+1 < len(cuts); i++ {
		a, b := cuts[i], cuts[i+1]
		if a.Before(from) || b.After(to) || b.Sub(a) < 24*time.Hour {
			continue
		}
		mid := a.Add(b.Sub(a) / 2).In(zone)
		sky, err := s.engine.BirthChart(engine.ChartInput{Date: mid.Format("2006-01-02"), Time: mid.Format("15:04"), Lat: in.Lat, Lon: in.Lon, TZ: in.TZ})
		if err != nil {
			return nil, err
		}
		f, err := engine.Facts(natal, mid)
		if err != nil {
			return nil, err
		}
		out = append(out, reading.PeriodInput{From: a, To: b, Sky: sky, Dasha: f.Vimshottari.Current})
	}
	return out, nil
}

// forecastContext is the forecast for chat: the coming periods and the slow
// planets' events over the next months. Failures leave it empty; the
// answer then says the forecast is unavailable.
func (s *Server) forecastContext(ctx context.Context, natal engine.Chart, in engine.ChartInput, now time.Time, months int) ([]reading.Period, []engine.TransitEvent) {
	tc, ok := s.engine.(TransitCalculator)
	if !ok {
		return nil, nil
	}
	end := now.AddDate(0, months, 0)
	events, err := tc.TransitEvents(now, end)
	if err != nil {
		return nil, nil
	}
	periods, err := s.forecastPeriods(natal, in, events, now, end)
	if err != nil {
		return nil, nil
	}
	return s.reading.Forecast(ctx, natal, periods), events
}
