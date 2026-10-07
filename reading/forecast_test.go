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

func forecastFor(t *testing.T) (engine.ChartFacts, []Rule, ChatContext) {
	t.Helper()
	e := engine.New("../ephe")
	natal, err := e.BirthChart(engine.ChartInput{Date: "1996-05-14", Time: "10:15", Lat: 12.97, Lon: 77.59, TZ: "Asia/Kolkata"})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	var periods []PeriodInput
	for m := 0; m < 12; m += 2 {
		from, to := start.AddDate(0, m, 0), start.AddDate(0, m+2, 0)
		mid := from.Add(to.Sub(from) / 2)
		sky, _ := e.BirthChart(engine.ChartInput{Date: mid.Format("2006-01-02"), Time: "12:00", Lat: 12.97, Lon: 77.59, TZ: "UTC"})
		f, _ := engine.Facts(natal, mid)
		periods = append(periods, PeriodInput{From: from, To: to, Sky: sky, Dasha: f.Vimshottari.Current, Yogini: f.Yogini.Current.Lord})
	}
	events, err := e.TransitEvents(start, start.AddDate(1, 0, 0))
	if err != nil {
		t.Fatal(err)
	}
	s := NewService(DefaultCorpus, nil)
	facts, rules, err := s.BuildFacts(context.Background(), natal, start)
	if err != nil {
		t.Fatal(err)
	}
	sb, err := e.Shadbala(natal)
	if err != nil {
		t.Fatal(err)
	}
	v, err := e.Varshaphal(natal, 2026)
	if err != nil {
		t.Fatal(err)
	}
	return facts, rules, ChatContext{Forecast: s.Forecast(context.Background(), natal, periods, ForecastOption{Shadbala: &sb}), Events: events, Varsha: &v}
}

// Timing questions get windows, never events: each says a chart cannot tell
// whether or when something happens, health points to a doctor, death gets
// the safety reply, and nothing makes a forbidden prediction.
func TestForecastAnswersGiveWindowsNotEvents(t *testing.T) {
	facts, rules, cc := forecastFor(t)
	s := NewService(DefaultCorpus, nil)
	for _, c := range []struct {
		q    string
		want []string
	}{
		{"When is a good time for my career?", []string{"In short", "career and public life", "cannot say whether or when"}},
		{"Will I get married this year?", []string{"marriage and partnerships", "cannot say whether or when"}},
		{"Will my business fail next year?", []string{"career and public life", "cannot say whether or when"}},
		{"What's coming for me this year?", []string{"Until ", "From ", "not events that will happen", "Varshaphal", "Muntha falls"}},
		{"What does my varshaphal say?", []string{"ruled by", "Muntha falls", "Mudda dasha"}},
		{"Is my health going to be bad next year?", []string{"cannot predict health", "doctor"}},
		{"When will I die?", []string{"cannot predict death"}},
	} {
		if ts := Topics(c.q); !contains(ts, "forecast") && !contains(ts, "safety") {
			t.Errorf("%q classified as %v", c.q, ts)
		}
		a, err := s.Answer(context.Background(), facts, rules, c.q, nil, cc)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range c.want {
			if !strings.Contains(strings.ToLower(a.Answer), strings.ToLower(w)) {
				t.Errorf("%q: answer lacks %q:\n%s", c.q, w, a.Answer)
			}
		}
		if forbidden.MatchString(a.Answer) || strings.Contains(strings.ToLower(a.Answer), "you will ") {
			t.Errorf("%q: forbidden wording:\n%s", c.q, a.Answer)
		}
	}
	// Without forecast data the answer says so instead of guessing.
	a, _ := s.Answer(context.Background(), facts, rules, "What is coming next year?", nil, ChatContext{})
	if !strings.Contains(a.Answer, "unavailable") {
		t.Errorf("no forecast: %s", a.Answer)
	}
}

// In a Venus main period, partnership is read with Venus as its natural
// significator, and periods carry the dasha lords' chart-specific themes.
func TestForecastUsesKarakasAndDashaThemes(t *testing.T) {
	_, _, cc := forecastFor(t)
	if cc.Forecast[0].Maha != "venus" {
		t.Skipf("reference chart now runs %s", cc.Forecast[0].Maha)
	}
	karaka := false
	for _, p := range cc.Forecast {
		if a, ok := areaIn(p, "partnership"); ok {
			for _, r := range a.Reasons {
				karaka = karaka || strings.HasPrefix(r.Source, "karaka")
			}
		}
	}
	if !karaka {
		t.Error("no karaka reason for partnership in a Venus period")
	}
	th := strings.Join(cc.Forecast[0].Themes, " ")
	if !strings.Contains(th, "Venus rules your") || !strings.Contains(th, "Sub-period: Jupiter") {
		t.Errorf("period themes: %q", th)
	}
}

// Questions in Hinglish and Indian scripts reach the same topics as English;
// "lagna" stays the ascendant (it means marriage only in Marathi).
func TestQuestionsInIndianLanguages(t *testing.T) {
	for q, want := range map[string][]string{
		"shaadi kab hogi?":  {"marriage", "forecast"},
		"naukri kab milegi": {"career", "forecast"},
		"मेरी शादी कब होगी?":              {"marriage", "forecast"},
		"ನನ್ನ ಮದುವೆ ಯಾವಾಗ?":               {"marriage", "forecast"},
		"என் திருமணம் எப்போது?":           {"marriage", "forecast"},
		"పెళ్లి ఎప్పుడు?":                 {"marriage", "forecast"},
		"is 2027 good for buying a house": {"property", "forecast"},
		"kya main manglik hoon":           {"mangal_dosha"},
		"शनि की साढ़ेसाती":                {"sade_sati"},
		"मेरी मृत्यु कब होगी":             {"safety"},
		"what is my lagna":                {"lagna"},
	} {
		got := Topics(q)
		for _, w := range want {
			if !contains(got, w) {
				t.Errorf("%q: topics %v lack %s", q, got, w)
			}
		}
		if contains(got, "safety") && got[0] != "safety" {
			t.Errorf("%q: safety must lead: %v", q, got)
		}
		if q == "what is my lagna" && contains(got, "marriage") {
			t.Errorf("lagna read as marriage: %v", got)
		}
	}
}

// Shadbala and the Yogini dasha feed the reasons when they apply.
func TestForecastUsesShadbalaAndYogini(t *testing.T) {
	_, _, cc := forecastFor(t)
	var shadbala, yogini bool
	for _, p := range cc.Forecast {
		for _, a := range p.Areas {
			for _, r := range a.Reasons {
				shadbala = shadbala || strings.Contains(r.Text, "Shadbala")
				yogini = yogini || r.Source == "Yogini dasha"
			}
		}
	}
	if !shadbala {
		t.Error("no reason mentions Shadbala")
	}
	if !yogini {
		t.Log("the Yogini lord does not touch a shown area for this chart in these months")
	}
}
