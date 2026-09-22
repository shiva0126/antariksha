package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/example/panchang/api/places"
	"github.com/example/panchang/engine"
)

func (s *Server) featureRoutes() {
	s.mux.HandleFunc("GET /api/places", s.placeSearch)
	s.mux.HandleFunc("GET /api/chart/varga", s.varga)
	s.mux.HandleFunc("GET /api/chart/shadbala", s.shadbala)
	s.mux.Handle("POST /api/match", s.limit(s.match, 30))
	s.mux.HandleFunc("GET /api/muhurta/events", s.muhurtaEvents)
	s.mux.Handle("GET /api/muhurta", s.limit(s.muhurta, 30))
	s.mux.HandleFunc("GET /api/today", s.today)
	s.mux.HandleFunc("GET /api/calendar.ics", s.calendarICS)
	s.mux.HandleFunc("DELETE /api/chat/session", s.deleteChat)
}

// ---- rate limiting --------------------------------------------------------

type bucket struct {
	tokens float64
	last   time.Time
}

// rateLimiter is a per-client token bucket (perMinute requests, burst of the
// same size). It protects the chat endpoint, which may call a paid LLM, and
// the expensive multi-day calculations.
type rateLimiter struct {
	mu        sync.Mutex
	perMinute float64
	clients   map[string]*bucket
}

func newRateLimiter(perMinute int) *rateLimiter {
	return &rateLimiter{perMinute: float64(perMinute), clients: map[string]*bucket{}}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (l *rateLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.clients) > 50000 {
		for k, b := range l.clients {
			if now.Sub(b.last) > 10*time.Minute {
				delete(l.clients, k)
			}
		}
	}
	b, ok := l.clients[key]
	if !ok {
		b = &bucket{tokens: l.perMinute, last: now}
		l.clients[key] = b
	}
	b.tokens += now.Sub(b.last).Minutes() * l.perMinute
	if b.tokens > l.perMinute {
		b.tokens = l.perMinute
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (s *Server) limit(h http.HandlerFunc, perMinute int) http.Handler {
	l := newRateLimiter(perMinute)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r), time.Now()) {
			w.Header().Set("Retry-After", "20")
			problem(w, http.StatusTooManyRequests, fmt.Errorf("too many requests; please wait a moment"))
			return
		}
		h(w, r)
	})
}

// ---- places ---------------------------------------------------------------

func (s *Server) placeSearch(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 25 {
		limit = 10
	}
	writeJSON(w, 200, map[string]any{"places": places.Search(r.URL.Query().Get("q"), limit), "attribution": places.Attribution})
}

// ---- divisional charts ----------------------------------------------------

func (s *Server) varga(w http.ResponseWriter, r *http.Request) {
	in, e := s.chartInput(r.URL.Query())
	if e != nil {
		problem(w, 400, e)
		return
	}
	n, _ := strconv.Atoi(r.URL.Query().Get("n"))
	c, e := s.engine.BirthChart(in)
	if e != nil {
		problem(w, 400, e)
		return
	}
	v, e := engine.VargaChart(c, n)
	if e != nil {
		problem(w, 400, e)
		return
	}
	writeJSON(w, 200, map[string]any{"varga": n, "name": engine.VargaName(n), "theme": engine.VargaTheme(n), "chart": v})
}

// ---- shadbala ---------------------------------------------------------------

// ShadbalaCalculator is implemented by the real engine; test doubles may omit it.
type ShadbalaCalculator interface {
	Shadbala(engine.Chart) (engine.Shadbala, error)
}

func (s *Server) shadbala(w http.ResponseWriter, r *http.Request) {
	sc, ok := s.engine.(ShadbalaCalculator)
	if !ok {
		problem(w, 501, fmt.Errorf("shadbala unavailable"))
		return
	}
	in, e := s.chartInput(r.URL.Query())
	if e != nil {
		problem(w, 400, e)
		return
	}
	c, e := s.engine.BirthChart(in)
	if e != nil {
		problem(w, 400, e)
		return
	}
	sb, e := sc.Shadbala(c)
	if e != nil {
		problem(w, 500, e)
		return
	}
	writeJSON(w, 200, sb)
}

// ---- kundli matching -------------------------------------------------------

type matchRequest struct {
	Boy  engine.ChartInput `json:"boy"`
	Girl engine.ChartInput `json:"girl"`
}

func (s *Server) match(w http.ResponseWriter, r *http.Request) {
	var req matchRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&req); err != nil {
		problem(w, 400, fmt.Errorf("invalid JSON body"))
		return
	}
	boy, err := s.engine.BirthChart(req.Boy)
	if err != nil {
		problem(w, 400, fmt.Errorf("groom's details: %w", err))
		return
	}
	girl, err := s.engine.BirthChart(req.Girl)
	if err != nil {
		problem(w, 400, fmt.Errorf("bride's details: %w", err))
		return
	}
	m, err := engine.MatchCharts(boy, girl)
	if err != nil {
		problem(w, 500, err)
		return
	}
	_, bh := engine.MangalDosha(boy)
	_, gh := engine.MangalDosha(girl)
	writeJSON(w, 200, map[string]any{"match": m, "boy": summary(boy, bh), "girl": summary(girl, gh)})
}

func summary(c engine.Chart, marsHouse int) map[string]any {
	moon := engine.Graha{}
	for _, g := range c.Grahas {
		if g.ID == "moon" {
			moon = g
		}
	}
	return map[string]any{"lagna": c.Ascendant.Rashi, "moon_sign": moon.Rashi, "nakshatra": moon.Nakshatra, "pada": moon.NakshatraPada, "mars_house": marsHouse}
}

// ---- muhurta --------------------------------------------------------------

func (s *Server) muhurtaEvents(w http.ResponseWriter, _ *http.Request) {
	type ev struct{ ID, Name, Description string }
	out := []ev{}
	for _, e := range engine.MuhurtaEvents {
		out = append(out, ev{e.ID, e.Name, e.Description})
	}
	writeJSON(w, 200, out)
}

func (s *Server) muhurta(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ev, ok := engine.MuhurtaEventByID(q.Get("event"))
	if !ok {
		problem(w, 400, fmt.Errorf("unknown event %q", q.Get("event")))
		return
	}
	from, l, e := params(r) // date = start date; lat/lon/tz = where the event happens
	if e != nil {
		problem(w, 400, e)
		return
	}
	days, _ := strconv.Atoi(q.Get("days"))
	if days <= 0 || days > 90 {
		days = 30
	}
	janma, moonSign := -1, -1
	if q.Get("bdate") != "" {
		bq := map[string][]string{"date": {q.Get("bdate")}, "time": {q.Get("btime")}, "lat": {q.Get("blat")}, "lon": {q.Get("blon")}, "tz": {q.Get("btz")}}
		in, err := s.chartInput(bq)
		if err != nil {
			problem(w, 400, fmt.Errorf("birth details: %w", err))
			return
		}
		bc, err := s.engine.BirthChart(in)
		if err != nil {
			problem(w, 400, err)
			return
		}
		for _, g := range bc.Grahas {
			if g.ID == "moon" {
				janma, moonSign = engine.NakshatraIndex(g.Nakshatra), engine.RashiIndex(g.Rashi)
			}
		}
	}
	out := make([]engine.MuhurtaDay, 0, days)
	for i := 0; i < days; i++ {
		d := from.AddDate(0, 0, i)
		day, err := s.get(r.Context(), d, l)
		if err != nil {
			problem(w, 500, err)
			return
		}
		transitMoon := -1
		if moonSign >= 0 {
			if c, err := s.engine.BirthChart(engine.ChartInput{Date: day.Date, Time: day.Sunrise, Lat: l.Lat, Lon: l.Lon, TZ: l.TZ}); err == nil {
				for _, g := range c.Grahas {
					if g.ID == "moon" {
						transitMoon = engine.RashiIndex(g.Rashi)
					}
				}
			}
		}
		out = append(out, engine.EvaluateMuhurta(ev, day, janma, moonSign, transitMoon))
	}
	writeJSON(w, 200, map[string]any{"event": map[string]string{"id": ev.ID, "name": ev.Name, "description": ev.Description}, "personalised": janma >= 0, "days": out})
}

// ---- today for you ---------------------------------------------------------

type transit struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Rashi      string  `json:"rashi"`
	Degree     float64 `json:"degree"`
	Nakshatra  string  `json:"nakshatra"`
	Retrograde bool    `json:"retrograde"`
	FromLagna  int     `json:"house_from_lagna"`
	FromMoon   int     `json:"house_from_moon"`
}

func (s *Server) today(w http.ResponseWriter, r *http.Request) {
	in, e := s.chartInput(r.URL.Query())
	if e != nil {
		problem(w, 400, e)
		return
	}
	natal, e := s.engine.BirthChart(in)
	if e != nil {
		problem(w, 400, e)
		return
	}
	zone, _ := time.LoadLocation(in.TZ)
	now := time.Now().In(zone)
	if v := r.URL.Query().Get("at"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			now = t.In(zone)
		}
	}
	loc := engine.Location{Lat: in.Lat, Lon: in.Lon, TZ: in.TZ}
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	day, e := s.get(r.Context(), date, loc)
	if e != nil {
		problem(w, 500, e)
		return
	}
	sky, e := s.engine.BirthChart(engine.ChartInput{Date: now.Format("2006-01-02"), Time: now.Format("15:04"), Lat: in.Lat, Lon: in.Lon, TZ: in.TZ})
	if e != nil {
		problem(w, 500, e)
		return
	}
	var natalMoon engine.Graha
	for _, g := range natal.Grahas {
		if g.ID == "moon" {
			natalMoon = g
		}
	}
	ts := []transit{}
	var skyMoon engine.Graha
	for _, g := range sky.Grahas {
		if g.ID == "moon" {
			skyMoon = g
		}
		ts = append(ts, transit{g.ID, g.Name, g.Rashi, g.RashiDegree, g.Nakshatra, g.Retrograde,
			engine.HouseOf(g.Longitude, natal.Ascendant.Longitude), engine.HouseOf(g.Longitude, natalMoon.Longitude)})
	}
	tn, tname, tgood := engine.TaraBala(engine.NakshatraIndex(natalMoon.Nakshatra), engine.NakshatraIndex(skyMoon.Nakshatra))
	ch, cgood := engine.ChandraBala(engine.RashiIndex(natalMoon.Rashi), engine.RashiIndex(skyMoon.Rashi))
	sade, phase := engine.SadeSati(natal, sky)
	upcoming := []map[string]any{}
	for i := 1; i <= 14 && len(upcoming) < 8; i++ {
		d, err := s.get(r.Context(), date.AddDate(0, 0, i), loc)
		if err != nil {
			break
		}
		if len(d.Festivals) > 0 {
			upcoming = append(upcoming, map[string]any{"date": d.Date, "festivals": d.Festivals})
		}
	}
	facts, _ := engine.Facts(natal, now)
	writeJSON(w, 200, map[string]any{
		"now": now.Format(time.RFC3339), "panchang": day, "transits": ts,
		"tara_bala":    map[string]any{"number": tn, "name": tname, "favourable": tgood, "moon_nakshatra": skyMoon.Nakshatra},
		"chandra_bala": map[string]any{"house": ch, "favourable": cgood, "moon_sign": skyMoon.Rashi},
		"sade_sati":    map[string]any{"active": sade, "phase": phase},
		"dasha":        facts.Vimshottari.Current, "upcoming": upcoming,
	})
}

// ---- iCalendar export -----------------------------------------------------

func icsEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, ";", `\;`, ",", `\,`, "\n", `\n`).Replace(s)
}

func (s *Server) calendarICS(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	year, err := strconv.Atoi(q.Get("year"))
	if err != nil || year < 1900 || year > 2100 {
		problem(w, 400, fmt.Errorf("year must be 1900–2100"))
		return
	}
	q.Set("date", fmt.Sprintf("%04d-01-01", year))
	r.URL.RawQuery = q.Encode()
	_, l, e := params(r)
	if e != nil {
		problem(w, 400, e)
		return
	}
	includeVrat := q.Get("vrat") != "0"
	var b strings.Builder
	b.WriteString("BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//Antariksha//Panchang//EN\r\nCALSCALE:GREGORIAN\r\nX-WR-CALNAME:Hindu festivals " + strconv.Itoa(year) + "\r\n")
	stamp := time.Now().UTC().Format("20060102T150405Z")
	for d := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC); d.Year() == year; d = d.AddDate(0, 0, 1) {
		day, err := s.get(context.Background(), d, l)
		if err != nil {
			problem(w, 500, err)
			return
		}
		for _, f := range day.Festivals {
			vrat := strings.HasSuffix(f, "Ekadashi") || f == "Pradosh Vrat" || f == "Sankashti Chaturthi" || f == "Purnima" || f == "Amavasya"
			if vrat && !includeVrat {
				continue
			}
			uid := fmt.Sprintf("%s-%s@antariksha", d.Format("20060102"), strings.ReplaceAll(strings.ToLower(f), " ", "-"))
			fmt.Fprintf(&b, "BEGIN:VEVENT\r\nUID:%s\r\nDTSTAMP:%s\r\nDTSTART;VALUE=DATE:%s\r\nDTEND;VALUE=DATE:%s\r\nSUMMARY:%s\r\nDESCRIPTION:%s\r\nTRANSP:TRANSPARENT\r\nEND:VEVENT\r\n",
				icsEscape(uid), stamp, d.Format("20060102"), d.AddDate(0, 0, 1).Format("20060102"), icsEscape(f),
				icsEscape(fmt.Sprintf("%s paksha %s, %s month (Amanta). Computed for %.2f, %.2f (%s).", day.Paksha, day.Tithi.Name, day.LunarMonth, l.Lat, l.Lon, l.TZ)))
		}
	}
	b.WriteString("END:VCALENDAR\r\n")
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="hindu-festivals-%d.ics"`, year))
	_, _ = w.Write([]byte(b.String()))
}

// ---- data deletion ---------------------------------------------------------

func (s *Server) deleteChat(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query().Get("session_id")
	if !sessionIDPattern.MatchString(sid) {
		problem(w, 400, fmt.Errorf("invalid session_id"))
		return
	}
	if err := s.chats.Delete(r.Context(), sid); err != nil {
		problem(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"deleted": sid})
}
