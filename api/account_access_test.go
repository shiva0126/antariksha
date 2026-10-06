package api

import (
	"bytes"
	"context"
	"encoding/base32"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type accessFixture struct {
	s                   *Server
	p                   *pgxpool.Pool
	id, email, password string
	cookie              *http.Cookie
	t                   *testing.T
}

func newAccessFixture(t *testing.T) *accessFixture {
	t.Helper()
	dsn := os.Getenv("ACCOUNT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable migrated database")
	}
	p, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	f := &accessFixture{s: NewServer(nil, PostgresCache{Pool: p}, nil), p: p, email: "access_" + randomToken()[:12] + "@example.com", password: "original-secure-password", t: t}
	w := f.req(nil, "POST", "/api/auth/register", map[string]any{"email": f.email, "password": f.password, "birth_date": "1996-01-01", "birth_time": "10:15", "consent": true})
	f.expect(w, 201)
	f.cookie = w.Result().Cookies()[0]
	if err = p.QueryRow(context.Background(), `SELECT id FROM member_accounts WHERE email=$1`, f.email).Scan(&f.id); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		p.Exec(context.Background(), `DELETE FROM member_audit WHERE actor=$1 OR subject=$1`, f.id)
		p.Exec(context.Background(), `DELETE FROM member_accounts WHERE id=$1`, f.id)
		p.Close()
	})
	return f
}
func (f *accessFixture) req(cookie *http.Cookie, method, path string, body any) *httptest.ResponseRecorder {
	f.t.Helper()
	raw, _ := json.Marshal(body)
	r := httptest.NewRequest(method, "http://example.com"+path, bytes.NewReader(raw))
	r.Header.Set("Origin", "http://example.com")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", "Mozilla/5.0 Chrome/120 Linux")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	f.s.Handler().ServeHTTP(w, r)
	return w
}
func (f *accessFixture) expect(w *httptest.ResponseRecorder, status int) {
	f.t.Helper()
	if w.Code != status {
		f.t.Fatalf("got %d want %d: %s", w.Code, status, w.Body.String())
	}
}
func (f *accessFixture) login(password, code string) *httptest.ResponseRecorder {
	return f.req(nil, "POST", "/api/auth/login", map[string]string{"email": f.email, "password": password, "code": code})
}

func TestPasswordAndSessionLifecycle(t *testing.T) {
	f := newAccessFixture(t)
	other := newAccessFixture(t)
	login := f.login(f.password, "")
	f.expect(login, 200)
	second := login.Result().Cookies()[0]
	w := f.req(f.cookie, "GET", "/api/me/sessions", nil)
	f.expect(w, 200)
	if strings.Contains(w.Body.String(), "token_hash") || strings.Contains(w.Body.String(), f.cookie.Value) {
		t.Fatal("session secret leaked")
	}
	var sessions []struct {
		ID      string `json:"id"`
		Current bool   `json:"current"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Fatal("expected two sessions")
	}
	var target string
	for _, s := range sessions {
		if !s.Current {
			target = s.ID
		}
	}
	f.expect(f.req(other.cookie, "POST", "/api/me/sessions/"+target+"/revoke", map[string]string{"password": other.password}), 404)
	f.expect(f.req(f.cookie, "POST", "/api/me/sessions/"+target+"/revoke", map[string]string{"password": "wrong"}), 403)
	f.expect(f.req(f.cookie, "POST", "/api/me/sessions/"+target+"/revoke", map[string]string{"password": f.password}), 200)
	f.expect(f.req(second, "GET", "/api/me", nil), 401)
	f.expect(f.req(f.cookie, "POST", "/api/me/password", map[string]string{"password": "wrong", "new_password": "new-secure-password"}), 403)
	f.expect(f.req(f.cookie, "POST", "/api/me/password", map[string]string{"password": f.password, "new_password": "short"}), 400)
	f.expect(f.req(f.cookie, "POST", "/api/me/recovery-key", map[string]string{"password": f.password}), 200)
	_, err := f.p.Exec(context.Background(), `INSERT INTO member_email_tokens(account_id,purpose,token_hash,email,expires_at) VALUES($1,'reset',$2,$3,now()+interval '30 minutes')`, f.id, sessionHash(randomToken()), f.email)
	if err != nil {
		t.Fatal(err)
	}
	f.expect(f.req(f.cookie, "POST", "/api/me/password", map[string]string{"password": f.password, "new_password": "new-secure-password"}), 200)
	f.expect(f.req(f.cookie, "GET", "/api/me", nil), 401)
	f.expect(f.login(f.password, ""), 401)
	f.expect(f.login("new-secure-password", ""), 200)
	var clean bool
	err = f.p.QueryRow(context.Background(), `SELECT recovery_hash IS NULL AND NOT EXISTS(SELECT 1 FROM member_email_tokens WHERE account_id=$1) FROM member_accounts WHERE id=$1`, f.id).Scan(&clean)
	if err != nil || !clean {
		t.Fatal("old recovery credentials survived")
	}
	w = httptest.NewRecorder()
	r := httptest.NewRequest("POST", "http://example.com/api/auth/login", nil)
	if err = f.s.issuePasswordSession(w, r, f.id, "stale-password-hash", ""); err == nil {
		t.Fatal("stale password validation created a session")
	}
}

func TestAdminAuthenticatorLifecycle(t *testing.T) {
	f := newAccessFixture(t)
	path := filepath.Join(t.TempDir(), "key")
	if err := os.WriteFile(path, []byte(strings.Repeat("k", 32)), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADMIN_TOTP_KEY_FILE", path)
	f.expect(f.req(f.cookie, "GET", "/api/me/admin-totp", nil), 403)
	if _, err := f.p.Exec(context.Background(), `UPDATE member_accounts SET role='superadmin' WHERE id=$1`, f.id); err != nil {
		t.Fatal(err)
	}
	action := func(a, code string) *httptest.ResponseRecorder {
		return f.req(f.cookie, "POST", "/api/me/admin-totp", map[string]string{"action": a, "password": f.password, "code": code})
	}
	w := action("begin", "")
	f.expect(w, 200)
	var setup struct {
		Secret string `json:"secret"`
	}
	json.Unmarshal(w.Body.Bytes(), &setup)
	secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(setup.Secret)
	if err != nil || len(secret) != 20 {
		t.Fatal("invalid setup secret")
	}
	var encrypted []byte
	f.p.QueryRow(context.Background(), `SELECT secret FROM member_totp WHERE account_id=$1`, f.id).Scan(&encrypted)
	if bytes.Equal(encrypted, secret) || bytes.Contains(encrypted, []byte(setup.Secret)) {
		t.Fatal("stored plaintext secret")
	}
	// Use previous time step for enrollment so login can test the next unused code without sleeping.
	f.expect(action("confirm", totpCode(secret, time.Now().Unix()/30-1, 6)), 200)
	f.expect(f.req(f.cookie, "GET", "/api/me", nil), 401)
	f.expect(f.login(f.password, ""), 401)
	code := totpCode(secret, time.Now().Unix()/30, 6)
	w = f.login(f.password, code)
	f.expect(w, 200)
	f.cookie = w.Result().Cookies()[0]
	f.expect(f.login(f.password, code), 401)
	f.expect(action("begin", ""), 409)
	f.expect(action("disable", code), 403)
	// Recovery changes a password but MUST NOT remove an enabled second factor.
	w = f.req(f.cookie, "POST", "/api/me/recovery-key", map[string]string{"password": f.password})
	f.expect(w, 200)
	var recovery struct {
		Key string `json:"recovery_key"`
	}
	json.Unmarshal(w.Body.Bytes(), &recovery)
	f.expect(f.req(nil, "POST", "/api/auth/recover", map[string]string{"handle": f.email, "key": recovery.Key, "password": "recovered-secure-password"}), 200)
	f.expect(f.login("recovered-secure-password", ""), 401)
	w = f.login("recovered-secure-password", totpCode(secret, time.Now().Unix()/30+1, 6))
	f.expect(w, 200)
	f.cookie = w.Result().Cookies()[0]
	f.password = "recovered-secure-password"
	// Simulate the next time step deterministically for disable (never sleep in a test).
	f.p.Exec(context.Background(), `UPDATE member_totp SET last_counter=$2 WHERE account_id=$1`, f.id, time.Now().Unix()/30-1)
	f.expect(action("disable", totpCode(secret, time.Now().Unix()/30, 6)), 200)
	f.expect(f.login(f.password, ""), 200)
}

func TestMatrimonyRequiresConfirmedBirthplace(t *testing.T) {
	s := NewServer(realEngine(t), nil, nil)
	b := memberBirth{Date: "1996-05-14", Time: "10:15"}
	if _, ok := s.chartOf("test-no-place", b); ok {
		t.Fatal("missing birthplace was guessed")
	}
	b.Place = &birthPlace{Name: "New York", Lat: 40.71, Lon: -74.01, TZ: "invalid"}
	if _, ok := s.chartOf("test-invalid-place", b); ok {
		t.Fatal("invalid timezone accepted")
	}
	b.Place.TZ = "America/New_York"
	a, ok := s.chartOf("test-new-york", b)
	if !ok || a.Chart.Input.TZ != "America/New_York" {
		t.Fatal("confirmed timezone not used")
	}
}

func TestPhotoReviewTracksPublishedSet(t *testing.T) {
	f := newAccessFixture(t)
	ctx := context.Background()
	photo := "photo_" + randomToken()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := f.p.Exec(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO matrimony_profiles(account_id,details) VALUES($1,'{}')`, f.id)
	exec(`INSERT INTO matrimony_photos(id,owner,data,published) VALUES($1,$2,$3,true)`, photo, f.id, []byte("photo"))
	exec(`UPDATE matrimony_profiles SET verified_at=now() WHERE account_id=$1`, f.id)
	exec(`INSERT INTO matrimony_verifications(account_id,status) VALUES($1,'approved')`, f.id)
	// Captions and private photos do not change what was reviewed.
	exec(`UPDATE matrimony_photos SET alt='New caption' WHERE id=$1`, photo)
	var reviewed bool
	f.p.QueryRow(ctx, `SELECT verified_at IS NOT NULL FROM matrimony_profiles WHERE account_id=$1`, f.id).Scan(&reviewed)
	if !reviewed {
		t.Fatal("caption removed review")
	}
	exec(`UPDATE matrimony_photos SET published=false WHERE id=$1`, photo)
	f.p.QueryRow(ctx, `SELECT verified_at IS NOT NULL FROM matrimony_profiles WHERE account_id=$1`, f.id).Scan(&reviewed)
	if reviewed {
		t.Fatal("changed photo set retained review")
	}
	var status string
	f.p.QueryRow(ctx, `SELECT status FROM matrimony_verifications WHERE account_id=$1`, f.id).Scan(&status)
	if status != "rejected" {
		t.Fatal("review status not invalidated")
	}
}
