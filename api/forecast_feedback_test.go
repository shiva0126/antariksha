package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestForecastFeedbackIsPastOnlyAndAggregated(t *testing.T) {
	dsn := os.Getenv("ACCOUNT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires isolated migrated database")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	s := NewServer(realEngine(t), PostgresCache{Pool: pool}, nil)
	request := func(c *http.Cookie, method, path string, body any) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		r := httptest.NewRequest(method, "http://example.com"+path, bytes.NewReader(raw))
		r.Header.Set("Origin", "http://example.com")
		r.Header.Set("Content-Type", "application/json")
		if c != nil {
			r.AddCookie(c)
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	h := "feedback_" + randomToken()[:8]
	w := request(nil, "POST", "/api/auth/register", map[string]any{"handle": h, "email": h + "@example.com", "password": "feedback-test-password", "birth_date": "1996-05-14", "birth_time": "10:15", "consent": true})
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	c := w.Result().Cookies()[0]
	var me struct{ ID string }
	_ = json.Unmarshal(request(c, "GET", "/api/me", nil).Body.Bytes(), &me)
	t.Cleanup(func() { pool.Exec(context.Background(), `DELETE FROM member_accounts WHERE id=$1`, me.ID) })

	past := time.Now().AddDate(0, -3, 0).Format("2006-01-02")
	to := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	ok := map[string]string{"period_from": past, "period_to": to, "area": "career", "tone": "supportive", "response": "yes"}
	if w := request(c, "PUT", "/api/me/forecast-feedback", ok); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	ok["response"] = "partly" // answering again replaces the answer
	request(c, "PUT", "/api/me/forecast-feedback", ok)
	for name, bad := range map[string]map[string]string{
		"future period": {"period_from": time.Now().AddDate(0, 1, 0).Format("2006-01-02"), "period_to": time.Now().AddDate(0, 2, 0).Format("2006-01-02"), "area": "career", "tone": "supportive", "response": "yes"},
		"unknown area":  {"period_from": past, "period_to": to, "area": "lottery", "tone": "supportive", "response": "yes"},
		"bad answer":    {"period_from": past, "period_to": to, "area": "career", "tone": "supportive", "response": "maybe"},
	} {
		if w := request(c, "PUT", "/api/me/forecast-feedback", bad); w.Code != 400 {
			t.Errorf("%s: %d %s", name, w.Code, w.Body.String())
		}
	}
	var mine []struct{ Area, Response string }
	_ = json.Unmarshal(request(c, "GET", "/api/me/forecast-feedback", nil).Body.Bytes(), &mine)
	if len(mine) != 1 || mine[0].Response != "partly" {
		t.Errorf("own answers: %+v", mine)
	}
	if w := request(c, "GET", "/api/admin/forecast-accuracy", nil); w.Code != 403 {
		t.Errorf("a member read the accuracy report: %d", w.Code)
	}
}
