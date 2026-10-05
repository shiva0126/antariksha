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

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSuperadminAccessAndLifecycle(t *testing.T) {
	dsn := os.Getenv("ACCOUNT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("isolated migrated database required")
	}
	ctx := context.Background()
	p, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	s := NewServer(nil, PostgresCache{Pool: p}, nil)
	type account struct {
		id, email, handle string
		cookie            *http.Cookie
	}
	request := func(a account, method, path string, body any, origin string) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		r := httptest.NewRequest(method, "http://example.com"+path, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", origin)
		if a.cookie != nil {
			r.AddCookie(a.cookie)
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	check := func(w *httptest.ResponseRecorder, code int) {
		t.Helper()
		if w.Code != code {
			t.Fatalf("want %d, got %d: %s", code, w.Code, w.Body.String())
		}
	}
	password := "admin-integration-password"
	var users []account
	defer func() {
		for _, a := range users {
			p.Exec(ctx, `DELETE FROM member_audit WHERE actor=$1 OR subject=$1`, a.id)
			p.Exec(ctx, `DELETE FROM member_accounts WHERE id=$1`, a.id)
		}
	}()
	for range 3 {
		a := account{handle: "admin_test_" + randomToken()[:10]}
		a.email = a.handle + "@example.com"
		w := request(a, "POST", "/api/auth/register", map[string]any{"handle": a.handle, "email": a.email, "password": password, "birth_date": "1996-01-01", "birth_time": "10:15", "consent": true}, "http://example.com")
		check(w, 201)
		a.cookie = w.Result().Cookies()[0]
		if err = p.QueryRow(ctx, `SELECT id FROM member_accounts WHERE handle=$1`, a.handle).Scan(&a.id); err != nil {
			t.Fatal(err)
		}
		users = append(users, a)
	}
	root, member, moderator := users[0], users[1], users[2]
	if _, err = p.Exec(ctx, `UPDATE member_accounts SET role=CASE WHEN id=$1 THEN 'superadmin' ELSE 'moderator' END WHERE id IN($1,$2)`, root.id, moderator.id); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/admin/users", "/api/admin/summary", "/api/admin/audit"} {
		check(request(account{}, "GET", path, nil, ""), 401)
		check(request(member, "GET", path, nil, ""), 403)
		check(request(moderator, "GET", path, nil, ""), 403)
		check(request(root, "GET", path, nil, ""), 200)
	}
	check(request(member, "PUT", "/api/me", map[string]any{"role": "superadmin"}, "http://example.com"), 400)
	check(request(root, "GET", "/api/admin/users?page=-1", nil, ""), 400)
	check(request(root, "GET", "/api/admin/users?status=typo", nil, ""), 400)
	listed := request(root, "GET", "/api/admin/users?q="+member.handle, nil, "")
	check(listed, 200)
	for _, secret := range []string{"password_hash", "birth_date", "birth_time", "recovery_hash", "ciphertext"} {
		if strings.Contains(listed.Body.String(), secret) {
			t.Fatal("private fields in listing", secret)
		}
	}
	wildcard := request(root, "GET", "/api/admin/users?q=%25", nil, "")
	check(wildcard, 200)
	if strings.TrimSpace(wildcard.Body.String()) != "[]" {
		t.Fatal("wildcard expanded")
	}
	actionPath := "/api/admin/users/" + member.id + "/action"
	body := map[string]any{"action": "suspend", "password": password, "reason": "Integration test"}
	check(request(member, "POST", actionPath, body, "http://example.com"), 403)
	check(request(root, "POST", actionPath, body, "https://elsewhere.example"), 403)
	body["password"] = "wrong"
	check(request(root, "POST", actionPath, body, "http://example.com"), 403)
	body["password"] = password
	check(request(root, "POST", "/api/admin/users/"+root.id+"/action", body, "http://example.com"), 409)
	check(request(root, "DELETE", "/api/me", nil, "http://example.com"), 409)
	p.Exec(ctx, `UPDATE member_settings SET community=true WHERE account_id=$1`, member.id)
	check(request(root, "POST", actionPath, body, "http://example.com"), 200)
	check(request(member, "GET", "/api/me", nil, ""), 401)
	login := func(a account) *httptest.ResponseRecorder {
		return request(account{}, "POST", "/api/auth/login", map[string]string{"email": a.email, "password": password}, "http://example.com")
	}
	check(login(member), 401)
	var paused, blocked bool
	if err = p.QueryRow(ctx, `SELECT NOT community,member_blocked($1,$2) FROM member_settings WHERE account_id=$1`, member.id, root.id).Scan(&paused, &blocked); err != nil || !paused || !blocked {
		t.Fatal("suspension did not pause visibility", err)
	}
	body["action"] = "reactivate"
	check(request(root, "POST", actionPath, body, "http://example.com"), 200)
	check(request(member, "GET", "/api/me", nil, ""), 401) // Old cookie must never revive.
	w := login(member)
	check(w, 200)
	member.cookie = w.Result().Cookies()[0]
	body["action"] = "role"
	body["role"] = "superadmin"
	check(request(root, "POST", actionPath, body, "http://example.com"), 400)
	body["role"] = "moderator"
	check(request(root, "POST", actionPath, body, "http://example.com"), 200)
	check(request(member, "GET", "/api/me", nil, ""), 401)
	w = login(member)
	check(w, 200)
	member.cookie = w.Result().Cookies()[0]
	check(request(member, "GET", "/api/community/moderation", nil, ""), 200)
	check(request(member, "GET", "/api/admin/users", nil, ""), 403)
	body["role"] = "member"
	check(request(root, "POST", actionPath, body, "http://example.com"), 200)
	w = login(member)
	check(w, 200)
	member.cookie = w.Result().Cookies()[0]
	check(request(member, "GET", "/api/community/moderation", nil, ""), 403)
	body["action"] = "revoke_sessions"
	check(request(root, "POST", actionPath, body, "http://example.com"), 200)
	check(request(member, "GET", "/api/me", nil, ""), 401)
	// Exercise cascade deletion and the separately-owned chat cleanup.
	sid := randomToken()
	if _, err = p.Exec(ctx, `INSERT INTO chat_sessions(id,chart_hash,birth) VALUES($1,'test','{}')`, sid); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Exec(ctx, `INSERT INTO chat_owners(session_id,account_id) VALUES($1,$2)`, sid, member.id); err != nil {
		t.Fatal(err)
	}
	body["action"] = "delete"
	body["confirm_handle"] = "wrong"
	check(request(root, "POST", actionPath, body, "http://example.com"), 400)
	body["confirm_handle"] = member.handle
	check(request(root, "POST", actionPath, body, "http://example.com"), 200)
	var count int
	p.QueryRow(ctx, `SELECT count(*) FROM member_accounts WHERE id=$1`, member.id).Scan(&count)
	if count != 0 {
		t.Fatal("account not deleted")
	}
	p.QueryRow(ctx, `SELECT count(*) FROM chat_sessions WHERE id=$1`, sid).Scan(&count)
	if count != 0 {
		t.Fatal("owned chat not deleted")
	}
	p.QueryRow(ctx, `SELECT count(*) FROM member_audit WHERE actor=$1 AND subject=$2 AND action='admin_delete' AND details->>'reason'='Integration test'`, root.id, member.id).Scan(&count)
	if count != 1 {
		t.Fatal("missing deletion audit")
	}
}
