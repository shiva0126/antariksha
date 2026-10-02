package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCharacterPrivacyAndPreview(t *testing.T) {
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
	call := func(cookie *http.Cookie, method, path string, body any) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		r := httptest.NewRequest(method, "http://example.com"+path, bytes.NewReader(raw))
		r.Header.Set("Origin", "http://example.com")
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
			t.Fatalf("want %d got %d: %s", status, w.Code, w.Body.String())
		}
	}
	var ids []string
	var cookies []*http.Cookie
	for range 2 {
		w := call(nil, "POST", "/api/auth/register", map[string]any{"email": "character_" + randomToken()[:10] + "@example.com", "password": "character-test-password", "birth_date": "1996-01-01", "birth_time": "10:15", "consent": true})
		check(w, 201)
		cookie := w.Result().Cookies()[0]
		cookies = append(cookies, cookie)
		var me struct {
			ID string `json:"id"`
		}
		json.Unmarshal(call(cookie, "GET", "/api/me", nil).Body.Bytes(), &me)
		ids = append(ids, me.ID)
		defer pool.Exec(context.Background(), `DELETE FROM member_accounts WHERE id=$1`, me.ID)
	}
	preview := call(cookies[0], "POST", "/api/me/character/preview", map[string]string{"text": "I like reading. My friend likes hiking; I do not."})
	check(preview, 200)
	if !strings.Contains(preview.Body.String(), `"stored":false`) {
		t.Fatal("preview does not report storage status")
	}
	var count int
	pool.QueryRow(context.Background(), `SELECT count(*) FROM member_characters WHERE account_id=$1`, ids[0]).Scan(&count)
	if count != 0 {
		t.Fatal("preview persisted private source")
	}
	p := characterDefaults()
	p.Name = "Private constellation"
	p.Interests = []string{"reading"}
	p.Sources = []characterSource{{ID: "source-1", Platform: "instagram", Text: "I like reading. My friend likes hiking; I do not."}}
	body := map[string]any{"account_id": ids[0], "revision": 0, "profile": p, "consent": true}
	check(call(cookies[1], "PUT", "/api/me/character", body), 400)
	w := call(cookies[0], "PUT", "/api/me/character", body)
	check(w, 200)
	if strings.Contains(w.Body.String(), "hiking") {
		t.Fatal("unselected source topic became a recommendation")
	}
	check(call(cookies[0], "PUT", "/api/me/character", body), 409)
	other := call(cookies[1], "GET", "/api/me/character?account_id="+ids[0], nil)
	check(other, 200)
	if strings.Contains(other.Body.String(), "Private constellation") {
		t.Fatal("cross-account character leak")
	}
	if !strings.Contains(call(cookies[0], "GET", "/api/me/export", nil).Body.String(), "Private constellation") {
		t.Fatal("character missing from export")
	}
	p.Sources[0].URL = "javascript:alert(1)"
	body["profile"] = p
	body["revision"] = 1
	check(call(cookies[0], "PUT", "/api/me/character", body), 400)
	p.Sources = []characterSource{}
	body["profile"] = p
	check(call(cookies[0], "PUT", "/api/me/character", body), 200)
	if strings.Contains(call(cookies[0], "GET", "/api/me/character", nil).Body.String(), "My friend") {
		t.Fatal("removed source retained")
	}
	check(call(cookies[0], "DELETE", "/api/me", nil), 200)
	pool.QueryRow(context.Background(), `SELECT count(*) FROM member_characters WHERE account_id=$1`, ids[0]).Scan(&count)
	if count != 0 {
		t.Fatal("character survived account deletion")
	}
}
