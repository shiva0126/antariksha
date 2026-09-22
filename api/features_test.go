package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func do(t *testing.T, s *Server, method, path, body string) (int, map[string]any, string) {
	t.Helper()
	var r = httptest.NewRequest(method, path, strings.NewReader(body))
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	var v map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &v)
	return w.Code, v, w.Body.String()
}

const birthQ = "date=1996-05-14&time=10:15&lat=12.97&lon=77.59&tz=Asia%2FKolkata"

func TestPlacesVargaMatch(t *testing.T) {
	s := NewServer(realEngine(t), NoCache{}, nil)
	code, v, _ := do(t, s, "GET", "/api/places?q=bangalore", "")
	if code != 200 || v["places"].([]any)[0].(map[string]any)["name"] != "Bengaluru" {
		t.Fatalf("places %d %v", code, v)
	}
	code, v, _ = do(t, s, "GET", "/api/chart/varga?"+birthQ+"&n=9", "")
	if code != 200 || v["name"] != "Navamsha" || v["chart"].(map[string]any)["ascendant"].(map[string]any)["rashi"] != "Karka" {
		t.Fatalf("varga %d %v", code, v)
	}
	if code, _, _ = do(t, s, "GET", "/api/chart/varga?"+birthQ+"&n=5", ""); code != 400 {
		t.Fatalf("D5 should be rejected, got %d", code)
	}
	b := `{"date":"1996-05-14","time":"10:15","lat":12.97,"lon":77.59,"tz":"Asia/Kolkata"}`
	code, v, body := do(t, s, "POST", "/api/match", `{"boy":`+b+`,"girl":`+b+`}`)
	if code != 200 || v["match"].(map[string]any)["total"].(float64) != 28 {
		t.Fatalf("match %d %s", code, body)
	}
}

func TestMuhurtaTodayICSAndDelete(t *testing.T) {
	s := NewServer(realEngine(t), NoCache{}, nil)
	code, v, body := do(t, s, "GET", "/api/muhurta?event=marriage&date=2026-10-01&days=30&lat=12.97&lon=77.59&tz=Asia%2FKolkata&bdate=1996-05-14&btime=10:15&blat=12.97&blon=77.59&btz=Asia%2FKolkata", "")
	if code != 200 || len(v["days"].([]any)) != 30 || v["personalised"] != true {
		t.Fatalf("muhurta %d %s", code, body[:min(300, len(body))])
	}
	for _, d := range v["days"].([]any) {
		m := d.(map[string]any)
		if m["good"] == true && (strings.Contains(m["tithi"].(string), "Amavasya") || len(m["reasons"].([]any)) < 3) {
			t.Fatalf("bad day marked good: %v", m)
		}
	}
	code, v, body = do(t, s, "GET", "/api/today?"+birthQ+"&at="+time.Date(2026, 9, 22, 6, 30, 0, 0, time.UTC).Format(time.RFC3339), "")
	if code != 200 || len(v["transits"].([]any)) != 9 || v["sade_sati"].(map[string]any)["active"] != true {
		t.Fatalf("today %d %s", code, body[:min(400, len(body))])
	}
	code, _, body = do(t, s, "GET", "/api/calendar.ics?year=2026&lat=28.61&lon=77.21&tz=Asia%2FKolkata", "")
	if code != 200 || !strings.Contains(body, "BEGIN:VCALENDAR") || !strings.Contains(body, "SUMMARY:Diwali (Lakshmi Puja)") || !strings.Contains(body, "DTSTART;VALUE=DATE:20261108") {
		t.Fatalf("ics %d", code)
	}
	_, v, _ = do(t, s, "POST", "/api/chat", `{"birth":{"date":"1996-05-14","time":"10:15","lat":12.97,"lon":77.59,"tz":"Asia/Kolkata"},"question":"hi"}`)
	sid := v["session_id"].(string)
	if code, _, _ = do(t, s, "DELETE", "/api/chat/session?session_id="+sid, ""); code != 200 {
		t.Fatalf("delete %d", code)
	}
	if code, _, _ = do(t, s, "GET", "/api/chat/history?session_id="+sid, ""); code != 404 {
		t.Fatalf("history after delete %d", code)
	}
}

func TestRateLimiter(t *testing.T) {
	l := newRateLimiter(3)
	now := time.Now()
	for i := 0; i < 3; i++ {
		if !l.allow("a", now) {
			t.Fatal("burst refused")
		}
	}
	if l.allow("a", now) {
		t.Fatal("4th request allowed")
	}
	if !l.allow("b", now) {
		t.Fatal("other client throttled")
	}
	if !l.allow("a", now.Add(25*time.Second)) {
		t.Fatal("tokens did not refill")
	}
}
