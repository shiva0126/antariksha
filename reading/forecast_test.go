package reading

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/example/panchang/engine"
)

// Forecasts over two years for varied charts: every reason cites a source,
// no text makes a forbidden prediction, tones follow scores, and no area is
// shown on a single factor.
func TestForecastIsSourcedAndSafe(t *testing.T) {
	e := engine.New("../ephe")
	s := NewService(DefaultCorpus, nil)
	births := []engine.ChartInput{
		{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"},
		{Date: "1971-11-02", Time: "23:40", Lat: 28.61, Lon: 77.21, TZ: "Asia/Kolkata"},
		{Date: "1988-02-29", Time: "04:05", Lat: 51.51, Lon: -0.13, TZ: "Europe/London"},
		{Date: "2003-08-19", Time: "13:30", Lat: 19.08, Lon: 72.88, TZ: "Asia/Kolkata"},
	}
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for _, b := range births {
		natal, err := e.BirthChart(b)
		if err != nil {
			t.Fatal(err)
		}
		var periods []PeriodInput
		for m := 0; m < 24; m += 3 { // quarterly periods are enough to exercise the rules
			from, to := start.AddDate(0, m, 0), start.AddDate(0, m+3, 0)
			mid := from.Add(to.Sub(from) / 2)
			sky, err := e.BirthChart(engine.ChartInput{Date: mid.Format("2006-01-02"), Time: "12:00", Lat: b.Lat, Lon: b.Lon, TZ: "UTC"})
			if err != nil {
				t.Fatal(err)
			}
			f, err := engine.Facts(natal, mid)
			if err != nil {
				t.Fatal(err)
			}
			periods = append(periods, PeriodInput{From: from, To: to, Sky: sky, Dasha: f.Vimshottari.Current})
		}
		out := s.Forecast(context.Background(), natal, periods)
		if len(out) != len(periods) {
			t.Fatalf("%d periods for %d", len(out), len(periods))
		}
		for _, p := range out {
			if p.Summary == "" || p.Maha == "" {
				t.Errorf("%s: empty period %+v", b.Date, p)
			}
			texts := []string{p.Summary}
			for _, a := range p.Areas {
				want := map[bool]string{true: "supportive", false: ""}[a.Score >= 1]
				if a.Score <= -1 {
					want = "challenging"
				} else if want == "" {
					want = "mixed"
				}
				if a.Tone != want {
					t.Errorf("%s %s: tone %s for score %.1f", b.Date, a.Area.ID, a.Tone, a.Score)
				}
				factors := 0
				for _, r := range a.Reasons {
					texts = append(texts, r.Text)
					if r.Source == "" {
						t.Errorf("%s %s: reason without a source: %q", b.Date, a.Area.ID, r.Text)
					}
					if r.Source != "later tradition" && !strings.HasPrefix(r.Source, "double transit") {
						factors++
					}
				}
				if factors == 0 {
					t.Errorf("%s %s shown without a dasha or Jupiter/Saturn factor", b.Date, a.Area.ID)
				}
			}
			for _, x := range texts {
				if forbidden.MatchString(x) || strings.Contains(strings.ToLower(x), "death") || strings.Contains(strings.ToLower(x), "will die") {
					t.Errorf("forbidden wording: %q", x)
				}
			}
		}
	}
}
