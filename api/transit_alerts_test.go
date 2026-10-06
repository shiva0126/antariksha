package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTransitAlertsAreOptInAndOncePerEvent(t *testing.T) {
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
	register := func(prefix string) (*http.Cookie, string) {
		h := prefix + "_" + randomToken()[:8]
		w := request(nil, "POST", "/api/auth/register", map[string]any{"handle": h, "email": h + "@example.com", "password": "transit-alert-password",
			"birth_date": "1996-05-14", "birth_time": "10:15", "consent": true,
			"birth_place": map[string]any{"name": "Bengaluru, Karnataka", "lat": 12.97, "lon": 77.59, "tz": "Asia/Kolkata"}})
		if w.Code != 201 {
			t.Fatal(w.Code, w.Body.String())
		}
		c := w.Result().Cookies()[0]
		var me struct{ ID string }
		_ = json.Unmarshal(request(c, "GET", "/api/me", nil).Body.Bytes(), &me)
		t.Cleanup(func() { pool.Exec(context.Background(), `DELETE FROM member_accounts WHERE id=$1`, me.ID) })
		return c, me.ID
	}
	on, onID := register("transit_on")
	_, offID := register("transit_off")

	var setting struct {
		Enabled      bool `json:"enabled"`
		BirthDetails bool `json:"birth_details"`
	}
	_ = json.Unmarshal(request(on, "GET", "/api/me/transit-alerts", nil).Body.Bytes(), &setting)
	if setting.Enabled || !setting.BirthDetails {
		t.Fatalf("new member: %+v", setting)
	}
	if w := request(on, "PUT", "/api/me/transit-alerts", map[string]bool{"enabled": true}); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}

	// Saturn entered Meena on 29 March 2025; a week ahead of that, the job queues it.
	at := time.Date(2025, 3, 25, 0, 0, 0, 0, time.UTC)
	if _, err := s.queueTransitAlerts(context.Background(), at); err != nil {
		t.Fatal(err)
	}
	if n, err := s.queueTransitAlerts(context.Background(), at); err != nil || n != 0 {
		t.Fatalf("second run added %d (%v); each event is queued once", n, err)
	}
	var subjects []string
	rows, _ := pool.Query(context.Background(), `SELECT recipient, subject FROM member_notifications WHERE kind='transit' AND recipient IN ($1,$2)`, onID, offID)
	for rows.Next() {
		var who, subject string
		_ = rows.Scan(&who, &subject)
		if who == offID {
			t.Errorf("a member without alerts got %q", subject)
		}
		subjects = append(subjects, subject)
	}
	rows.Close()
	saturn := ""
	for _, x := range subjects {
		if strings.HasPrefix(x, "saturn|Meena|2025-03-29|") {
			saturn = x
		}
	}
	if saturn == "" {
		t.Fatalf("no Saturn alert among %v", subjects)
	}
	if text := transitAlertText(saturn); !strings.Contains(text, "On 29 March Saturn moves into Meena: your ") || !strings.Contains(text, "house from the Moon") {
		t.Errorf("alert text %q", text)
	}
	var list []struct {
		Kind, Subject string
	}
	_ = json.Unmarshal(request(on, "GET", "/api/me/notifications", nil).Body.Bytes(), &list)
	found := false
	for _, n := range list {
		found = found || (n.Kind == "transit" && n.Subject == saturn)
	}
	if !found {
		t.Errorf("the member's notifications do not list the alert: %+v", list)
	}
	// Turning alerts off hides them.
	request(on, "PUT", "/api/me/transit-alerts", map[string]bool{"enabled": false})
	list = nil
	_ = json.Unmarshal(request(on, "GET", "/api/me/notifications", nil).Body.Bytes(), &list)
	for _, n := range list {
		if n.Kind == "transit" {
			t.Errorf("alerts off, but %+v is still listed", n)
		}
	}
}
