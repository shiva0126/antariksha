package reading

import (
	"sort"
	"time"

	"github.com/example/panchang/engine"
)

// ChartMaker casts a chart; the engine implements it.
type ChartMaker interface {
	BirthChart(engine.ChartInput) (engine.Chart, error)
}

// CutPeriods cuts [from, to) at every sign change of Jupiter, Saturn,
// Rahu and Ketu and every change of antardasha, and takes each period's sky
// and running dasha at its middle.
func CutPeriods(charts ChartMaker, natal engine.Chart, in engine.ChartInput, events []engine.TransitEvent, from, to time.Time) ([]PeriodInput, error) {
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
	var out []PeriodInput
	for i := 0; i+1 < len(cuts); i++ {
		a, b := cuts[i], cuts[i+1]
		if a.Before(from) || b.After(to) || b.Sub(a) < 24*time.Hour {
			continue
		}
		mid := a.Add(b.Sub(a) / 2).In(zone)
		sky, err := charts.BirthChart(engine.ChartInput{Date: mid.Format("2006-01-02"), Time: mid.Format("15:04"), Lat: in.Lat, Lon: in.Lon, TZ: in.TZ})
		if err != nil {
			return nil, err
		}
		f, err := engine.Facts(natal, mid)
		if err != nil {
			return nil, err
		}
		out = append(out, PeriodInput{From: a, To: b, Sky: sky, Dasha: f.Vimshottari.Current, Yogini: f.Yogini.Current.Lord})
	}
	return out, nil
}
