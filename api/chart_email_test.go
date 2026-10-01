package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type recordingMail struct{ messages []string }

func (m *recordingMail) Send(_ context.Context, to, subject, body string) error {
	m.messages = append(m.messages, body)
	return nil
}
func TestChartStorageAndEmailRecovery(t *testing.T) {
	dsn := os.Getenv("ACCOUNT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires isolated migrated database")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	s := NewServer(nil, PostgresCache{Pool: pool}, nil)
	sender := &recordingMail{}
	s.mailer = sender
	t.Setenv("PUBLIC_ORIGIN", "https://example.com")
	request := func(method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		r := httptest.NewRequest(method, "https://example.com"+path, strings.NewReader(string(raw)))
		r.Header.Set("Origin", "https://example.com")
		r.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	check := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("got %d want %d: %s", w.Code, status, w.Body.String())
		}
	}
	var ids, emails []string
	var cookies []*http.Cookie
	for i := 0; i < 2; i++ {
		email := "batch_" + randomToken()[:12] + "@example.com"
		w := request("POST", "/api/auth/register", map[string]any{"email": email, "password": "initial-password-long", "birth_date": "1996-05-14", "birth_time": "10:15", "consent": true}, nil)
		check(w, 201)
		cookie := w.Result().Cookies()[0]
		cookies = append(cookies, cookie)
		emails = append(emails, email)
		var me struct {
			ID string `json:"id"`
		}
		w = request("GET", "/api/me", nil, cookie)
		check(w, 200)
		_ = json.Unmarshal(w.Body.Bytes(), &me)
		ids = append(ids, me.ID)
		defer pool.Exec(context.Background(), `DELETE FROM member_accounts WHERE id=$1`, me.ID)
	}
	chart := map[string]any{"id": "chart-a", "name": "Private test chart", "place": "Bengaluru", "date": "1996-05-14", "time": "10:15", "lat": 12.97, "lon": 77.59, "tz": "Asia/Kolkata"}
	body := map[string]any{"profiles": []any{chart}, "revision": 0, "account_id": ids[0]}
	check(request("PUT", "/api/me/charts", body, cookies[0]), 200)
	check(request("PUT", "/api/me/charts", body, cookies[0]), 409)
	check(request("PUT", "/api/me/charts", body, cookies[1]), 400)
	other := request("GET", "/api/me/charts", nil, cookies[1])
	check(other, 200)
	if strings.Contains(other.Body.String(), "Private test chart") {
		t.Fatal("chart leaked across accounts")
	}
	body["revision"] = 1
	chart["tz"] = "bad/timezone"
	check(request("PUT", "/api/me/charts", body, cookies[0]), 400)
	chart["tz"] = "Asia/Kolkata"
	exported := request("GET", "/api/me/export", nil, cookies[0])
	check(exported, 200)
	if !strings.Contains(exported.Body.String(), "Private test chart") {
		t.Fatal("charts missing in export")
	}
	// Verification doesn't grant a session and links are single-use.
	check(request("POST", "/api/me/email/verification", map[string]any{}, cookies[0]), 202)
	if len(sender.messages) != 1 {
		t.Fatal("verification mail not captured")
	}
	token := strings.Split(strings.Split(sender.messages[0], "/#verify/")[1], "\r\n")[0]
	check(request("POST", "/api/auth/email/reset", map[string]any{"token": token, "password": "replacement-password"}, nil), 400)
	check(request("POST", "/api/auth/email/verify", map[string]any{"token": token}, nil), 200)
	check(request("POST", "/api/auth/email/verify", map[string]any{"token": token}, nil), 400)
	me := request("GET", "/api/me", nil, cookies[0])
	if !strings.Contains(me.Body.String(), `"email_verified":true`) {
		t.Fatal("email not verified")
	}
	// Public response cannot distinguish absent accounts; cooldown survives requests.
	known := request("POST", "/api/auth/email/reset-request", map[string]string{"email": emails[0]}, nil)
	check(known, 202)
	unknown := request("POST", "/api/auth/email/reset-request", map[string]string{"email": "missing@example.com"}, nil)
	check(unknown, 202)
	if known.Body.String() != unknown.Body.String() {
		t.Fatal("account enumeration response")
	}
	check(request("POST", "/api/auth/email/reset-request", map[string]string{"email": emails[0]}, nil), 202)
	if len(sender.messages) != 2 {
		t.Fatal("cooldown bypassed")
	}
	token = strings.Split(strings.Split(sender.messages[1], "/#reset/")[1], "\r\n")[0]
	var stored string
	pool.QueryRow(context.Background(), `SELECT token_hash FROM member_email_tokens WHERE account_id=$1 AND purpose='reset'`, ids[0]).Scan(&stored)
	if stored == token || stored != sessionHash(token) {
		t.Fatal("reset token not hashed")
	}
	_, _ = pool.Exec(context.Background(), `UPDATE member_email_tokens SET expires_at=now()-interval '1 minute' WHERE account_id=$1`, ids[0])
	reset := map[string]string{"token": token, "password": "replacement-password-long"}
	check(request("POST", "/api/auth/email/reset", reset, nil), 400)
	_, _ = pool.Exec(context.Background(), `UPDATE member_email_tokens SET expires_at=now()+interval '10 minutes' WHERE account_id=$1`, ids[0])
	check(request("POST", "/api/auth/email/reset", reset, nil), 200)
	check(request("POST", "/api/auth/email/reset", reset, nil), 400)
	check(request("GET", "/api/me", nil, cookies[0]), 401)
	check(request("POST", "/api/auth/login", map[string]string{"email": emails[0], "password": "initial-password-long"}, nil), 401)
	login := request("POST", "/api/auth/login", map[string]string{"email": emails[0], "password": "replacement-password-long"}, nil)
	check(login, 200)
	cookie := login.Result().Cookies()[0]
	saved := request("GET", "/api/me/charts", nil, cookie)
	check(saved, 200)
	if !strings.Contains(saved.Body.String(), "Private test chart") {
		t.Fatal("charts lost across sessions")
	}
	check(request("DELETE", "/api/me", nil, cookie), 200)
	var n int
	pool.QueryRow(context.Background(), `SELECT count(*) FROM member_chart_store WHERE account_id=$1`, ids[0]).Scan(&n)
	if n != 0 {
		t.Fatal("charts survived account deletion")
	}
	// Recovery-key use also invalidates any outstanding email-reset capability.
	w := request("POST", "/api/me/recovery-key", map[string]string{"password": "initial-password-long"}, cookies[1])
	check(w, 200)
	var recovery struct {
		Key string `json:"recovery_key"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &recovery); err != nil {
		t.Fatal(err)
	}
	check(request("POST", "/api/auth/email/reset-request", map[string]string{"email": emails[1]}, nil), 202)
	last := sender.messages[len(sender.messages)-1]
	oldToken := strings.Split(strings.Split(last, "/#reset/")[1], "\r\n")[0]
	check(request("POST", "/api/auth/recover", map[string]string{"handle": emails[1], "key": recovery.Key, "password": "recovery-key-password-long"}, nil), 200)
	check(request("POST", "/api/auth/email/reset", map[string]string{"token": oldToken, "password": "must-not-be-accepted"}, nil), 400)
	s.mailer = nil
	check(request("POST", "/api/auth/email/reset-request", map[string]string{"email": emails[1]}, nil), 503)
}
