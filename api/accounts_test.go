package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMemberIsolationAndSessionRevocation(t *testing.T) {
	dsn := os.Getenv("ACCOUNT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set ACCOUNT_TEST_DATABASE_URL to a migrated database")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	s := NewServer(nil, PostgresCache{Pool: pool}, nil)
	request := func(method, path, body, origin string, cookie *http.Cookie) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://example.com"+path, strings.NewReader(body))
		r.Header.Set("Origin", origin)
		r.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			r.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	handles := []string{"test_" + randomToken()[:16], "test_" + randomToken()[:16]}
	defer func() {
		for _, h := range handles {
			_, _ = pool.Exec(context.Background(), `DELETE FROM member_accounts WHERE handle=$1`, h)
		}
	}()
	var cookies []*http.Cookie
	for _, h := range handles {
		w := request("POST", "/api/auth/register", `{"handle":"`+h+`","email":"`+h+`@example.com","birth_date":"1996-01-01","birth_time":"10:15","password":"test-password-long","consent":true}`, "http://example.com", nil)
		if w.Code != 201 {
			t.Fatalf("register: %d %s", w.Code, w.Body.String())
		}
		cookies = append(cookies, w.Result().Cookies()[0])
	}
	if w := request("PUT", "/api/me", `{"bio":"private-secret"}`, "https://evil.example", cookies[0]); w.Code != 403 {
		t.Fatal("cross origin accepted")
	}
	if w := request("PUT", "/api/me", `{"bio":"private-secret"}`, "http://example.com", cookies[0]); w.Code != 200 {
		t.Fatalf("save: %s", w.Body.String())
	}
	if w := request("GET", "/api/me", "", "", cookies[1]); w.Code != 200 || strings.Contains(w.Body.String(), "private-secret") {
		t.Fatal("profile isolation failed")
	}
	if w := request("GET", "/api/me", "", "", nil); w.Code != 401 {
		t.Fatal("anonymous access")
	}
	if w := request("POST", "/api/auth/logout", `{}`, "http://example.com", cookies[0]); w.Code != 200 {
		t.Fatal("logout failed")
	}
	if w := request("GET", "/api/me", "", "", cookies[0]); w.Code != 401 {
		t.Fatal("revoked session accepted")
	}
	if w := request("DELETE", "/api/me", "", "http://example.com", cookies[1]); w.Code != 200 {
		t.Fatal("delete failed")
	}
	if w := request("GET", "/api/me", "", "", cookies[1]); w.Code != 401 {
		t.Fatal("deleted account session accepted")
	}
}
