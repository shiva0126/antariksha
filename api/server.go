package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/example/panchang/engine"
)

type Calculator interface {
	Calculate(time.Time, engine.Location) (engine.Computed, error)
	BirthChart(engine.ChartInput) (engine.Chart, error)
}
type Server struct {
	engine Calculator
	cache  Cache
	mux    *http.ServeMux
	logger *slog.Logger
}

func NewServer(e Calculator, c Cache, l *slog.Logger) *Server {
	if c == nil {
		c = NoCache{}
	}
	if l == nil {
		l = slog.Default()
	}
	s := &Server{e, c, http.NewServeMux(), l}
	s.routes()
	return s
}
func (s *Server) Handler() http.Handler { return recoverer(cors(s.mux), s.logger) }
func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	s.mux.HandleFunc("GET /api/panchang", s.panchang)
	s.mux.HandleFunc("GET /api/month", s.month)
	s.mux.HandleFunc("GET /api/festivals", s.festivals)
	s.mux.HandleFunc("GET /api/chart", s.chart)
}

func (s *Server) chart(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	lat, e := strconv.ParseFloat(q.Get("lat"), 64)
	if e != nil || math.IsNaN(lat) || math.IsInf(lat, 0) || lat < -90 || lat > 90 {
		problem(w, 400, fmt.Errorf("invalid lat"))
		return
	}
	lon, e := strconv.ParseFloat(q.Get("lon"), 64)
	if e != nil || math.IsNaN(lon) || math.IsInf(lon, 0) || lon < -180 || lon > 180 {
		problem(w, 400, fmt.Errorf("invalid lon"))
		return
	}
	if q.Get("ayanamsa") != "" && q.Get("ayanamsa") != "lahiri" {
		problem(w, 400, fmt.Errorf("only lahiri ayanamsa is supported"))
		return
	}
	in := engine.ChartInput{Date: q.Get("date"), Time: q.Get("time"), Lat: lat, Lon: lon, TZ: q.Get("tz")}
	chart, e := s.engine.BirthChart(in)
	if e != nil {
		problem(w, 400, e)
		return
	}
	writeJSON(w, 200, chart)
}

func params(r *http.Request) (time.Time, engine.Location, error) {
	q := r.URL.Query()
	d, e := time.Parse("2006-01-02", q.Get("date"))
	if e != nil {
		return d, engine.Location{}, fmt.Errorf("date must be YYYY-MM-DD")
	}
	lat, e := strconv.ParseFloat(q.Get("lat"), 64)
	if e != nil || math.IsNaN(lat) || math.IsInf(lat, 0) || lat < -90 || lat > 90 {
		return d, engine.Location{}, fmt.Errorf("invalid lat")
	}
	lon, e := strconv.ParseFloat(q.Get("lon"), 64)
	if e != nil || math.IsNaN(lon) || math.IsInf(lon, 0) || lon < -180 || lon > 180 {
		return d, engine.Location{}, fmt.Errorf("invalid lon")
	}
	tz := q.Get("tz")
	if _, e = time.LoadLocation(tz); e != nil || tz == "" {
		return d, engine.Location{}, fmt.Errorf("invalid IANA timezone")
	}
	return d, engine.Location{lat, lon, tz}, nil
}
func (s *Server) get(ctx context.Context, d time.Time, l engine.Location) (engine.Day, error) {
	if x, ok, e := s.cache.Get(ctx, d, l.Lat, l.Lon); e != nil {
		return x, e
	} else if ok && x.Location == l {
		return x, nil
	}
	c, e := s.engine.Calculate(d, l)
	if e != nil {
		return engine.Day{}, e
	}
	if e = s.cache.Put(ctx, d, l.Lat, l.Lon, c.Day); e != nil {
		s.logger.Error("cache write", "error", e)
	}
	return c.Day, nil
}
func (s *Server) panchang(w http.ResponseWriter, r *http.Request) {
	d, l, e := params(r)
	if e != nil {
		problem(w, 400, e)
		return
	}
	day, e := s.get(r.Context(), d, l)
	if e != nil {
		problem(w, 500, e)
		return
	}
	writeJSON(w, 200, day)
}

type monthDay struct {
	Date      string   `json:"date"`
	Tithi     string   `json:"tithi"`
	Paksha    string   `json:"paksha"`
	Festivals []string `json:"festivals"`
}

func (s *Server) month(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	y, e := strconv.Atoi(q.Get("year"))
	if e != nil {
		problem(w, 400, fmt.Errorf("invalid year"))
		return
	}
	m, e := strconv.Atoi(q.Get("month"))
	if e != nil || m < 1 || m > 12 {
		problem(w, 400, fmt.Errorf("invalid month"))
		return
	}
	q.Set("date", fmt.Sprintf("%04d-%02d-01", y, m))
	r.URL.RawQuery = q.Encode()
	_, l, e := params(r)
	if e != nil {
		problem(w, 400, e)
		return
	}
	n := time.Date(y, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	out := make([]monthDay, n)
	jobs := make(chan int)
	var wg sync.WaitGroup
	workers := 4
	if n < workers {
		workers = n
	}
	var firstErr error
	var mu sync.Mutex
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				d := time.Date(y, time.Month(m), i+1, 0, 0, 0, 0, time.UTC)
				x, err := s.get(r.Context(), d, l)
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					continue
				}
				out[i] = monthDay{x.Date, x.Tithi.Name, x.Paksha, x.Festivals}
			}
		}()
	}
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	if firstErr != nil {
		problem(w, 500, firstErr)
		return
	}
	writeJSON(w, 200, out)
}
func (s *Server) festivals(w http.ResponseWriter, r *http.Request) {
	y, e := strconv.Atoi(r.URL.Query().Get("year"))
	if e != nil || y < 1 {
		problem(w, 400, fmt.Errorf("invalid year"))
		return
	}
	store, ok := s.cache.(interface {
		Festivals(context.Context, int, string) ([]Festival, error)
	})
	if !ok {
		problem(w, 503, fmt.Errorf("festival storage is unavailable"))
		return
	}
	items, e := store.Festivals(r.Context(), y, r.URL.Query().Get("region"))
	if e != nil {
		problem(w, 500, e)
		return
	}
	writeJSON(w, 200, items)
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func problem(w http.ResponseWriter, status int, e error) {
	writeJSON(w, status, map[string]string{"error": e.Error()})
}
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}
func recoverer(next http.Handler, l *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				l.Error("panic", "value", x)
				problem(w, 500, fmt.Errorf("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
