package api

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBiodataConsentAndPhotoAccess(t *testing.T) {
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
	type member struct {
		id, handle string
		cookie     *http.Cookie
	}
	request := func(u member, method, path string, body any) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		r := httptest.NewRequest(method, "http://example.com"+path, bytes.NewReader(raw))
		r.Header.Set("Origin", "http://example.com")
		r.Header.Set("Content-Type", "application/json")
		if u.cookie != nil {
			r.AddCookie(u.cookie)
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
	readID := func(w *httptest.ResponseRecorder) string {
		var v struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		return v.ID
	}
	users := []member{}
	for range 3 {
		h := "biodata_" + randomToken()[:10]
		w := request(member{}, "POST", "/api/auth/register", map[string]any{"handle": h, "email": h + "@example.com", "password": "biodata-test-password", "birth_date": "1996-01-01", "birth_time": "10:15", "consent": true})
		check(w, 201)
		u := member{handle: h, cookie: w.Result().Cookies()[0]}
		u.id = readID(request(u, "GET", "/api/me", nil))
		users = append(users, u)
		defer pool.Exec(context.Background(), `DELETE FROM member_accounts WHERE id=$1`, u.id)
		check(request(u, "PUT", "/api/community/settings", map[string]any{"community": true, "birth_date": "1996-01-01", "avatar": "sun"}), 200)
		check(request(u, "PUT", "/api/matrimony/me", map[string]any{"active": true, "consent": true, "details": matrimonyDetails{MinAge: 18, MaxAge: 70}}), 200)
	}
	a, b, c := users[0], users[1], users[2]
	explanationPath := "/api/matrimony/explanation/" + b.id
	check(request(member{}, "POST", explanationPath, map[string]bool{"use_ai": false}), 401)
	check(request(a, "POST", explanationPath, map[string]bool{"use_ai": false}), 200)
	check(request(a, "POST", "/api/matrimony/explanation/"+a.id, map[string]bool{}), 404)
	spy := &matchPrivacyLLM{after: func() {
		pool.Exec(context.Background(), `UPDATE matrimony_profiles SET hidden=true WHERE account_id=$1`, b.id)
	}}
	s.reading.LLM = spy
	check(request(a, "POST", explanationPath, map[string]bool{"use_ai": true}), 404)
	if spy.calls != 1 || strings.Contains(spy.prompt, a.handle) || strings.Contains(spy.prompt, b.handle) || strings.Contains(spy.prompt, "1996-01-01") {
		t.Fatal("model call privacy failure")
	}
	pool.Exec(context.Background(), `UPDATE matrimony_profiles SET hidden=false WHERE account_id=$1`, b.id)
	s.reading.LLM = nil
	t.Setenv("MODERATOR_HANDLES", c.handle)
	details := matrimonyDetails{DisplayName: "Family prepared bride", ProfileKind: "bride", Education: "Engineering", MinAge: 18, MaxAge: 70, SocialLinks: []string{"https://www.linkedin.com/in/example"}}
	check(request(a, "POST", "/api/matrimony/drafts", map[string]any{"details": details, "relationship": "parent", "consent": false}), 400)
	w := request(a, "POST", "/api/matrimony/drafts", map[string]any{"details": details, "relationship": "parent", "consent": true})
	check(w, 201)
	draft := readID(w)
	upload := func(u member, d string) *httptest.ResponseRecorder {
		var buf bytes.Buffer
		mp := multipart.NewWriter(&buf)
		f, _ := mp.CreateFormFile("photo", "portrait.png")
		_ = png.Encode(f, image.NewRGBA(image.Rect(0, 0, 3, 3)))
		_ = mp.WriteField("draft_id", d)
		_ = mp.WriteField("consent", "true")
		_ = mp.WriteField("alt", "Test portrait")
		_ = mp.Close()
		r := httptest.NewRequest("POST", "http://example.com/api/matrimony/photos", &buf)
		r.AddCookie(u.cookie)
		r.Header.Set("Origin", "http://example.com")
		r.Header.Set("Content-Type", mp.FormDataContentType())
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	w = upload(a, draft)
	check(w, 201)
	photo := readID(w)
	photoPath := "/api/matrimony/photos/" + photo
	draftPath := "/api/matrimony/drafts/" + draft
	check(upload(c, draft), 404)
	check(request(a, "GET", photoPath, nil), 200)
	check(request(b, "GET", photoPath, nil), 404)
	check(request(c, "GET", photoPath, nil), 404)
	check(request(a, "POST", draftPath, map[string]any{"action": "send", "handle": b.handle}), 200)
	check(request(b, "GET", photoPath, nil), 200)
	check(upload(a, draft), 404)
	check(request(a, "DELETE", photoPath, nil), 404)
	check(request(c, "POST", draftPath, map[string]any{"action": "accept", "consent": true}), 404)
	check(request(b, "POST", draftPath, map[string]any{"action": "accept", "consent": false}), 400)
	check(request(b, "POST", draftPath, map[string]any{"action": "accept", "consent": true}), 200)
	check(request(b, "POST", draftPath, map[string]any{"action": "accept", "consent": true}), 404)
	check(request(a, "GET", photoPath, nil), 404)
	check(request(b, "GET", photoPath, nil), 200)
	if !strings.Contains(request(b, "GET", "/api/matrimony/me", nil).Body.String(), `"active":false`) {
		t.Fatal("accept published profile without review")
	}
	save := func(active bool, ids []string) {
		check(request(b, "PUT", "/api/matrimony/me", map[string]any{"active": active, "consent": true, "details": details, "photo_ids": ids}), 200)
	}
	save(true, []string{photo})
	check(request(a, "GET", photoPath, nil), 200)
	check(request(c, "PUT", "/api/matrimony/me", map[string]any{"active": true, "consent": true, "details": details, "photo_ids": []string{photo}}), 400)
	save(true, []string{})
	check(request(a, "GET", photoPath, nil), 404)
	save(true, []string{photo})
	save(false, []string{photo})
	check(request(a, "GET", photoPath, nil), 404)
	check(request(a, "POST", explanationPath, map[string]bool{}), 404)
	save(true, []string{photo})
	check(request(a, "GET", photoPath+"?for="+c.id, nil), 404)
	check(request(a, "POST", "/api/matrimony/report/"+b.id, map[string]string{"reason": "Photo concern"}), 200)
	check(request(a, "GET", "/api/matrimony/moderation", nil), 403)
	var report int64
	pool.QueryRow(context.Background(), `SELECT id FROM matrimony_reports WHERE subject=$1`, b.id).Scan(&report)
	check(request(c, "POST", "/api/matrimony/moderation", map[string]any{"report": report, "hide": true}), 200)
	check(request(a, "POST", explanationPath, map[string]bool{}), 404)
	check(request(a, "GET", photoPath, nil), 404)
	save(true, []string{photo})
	check(request(a, "GET", photoPath, nil), 404)
	pool.Exec(context.Background(), `UPDATE matrimony_profiles SET hidden=false WHERE account_id=$1`, b.id)
	check(request(a, "POST", "/api/community/blocks", map[string]any{"target": b.id, "block": true}), 200)
	check(request(a, "GET", photoPath, nil), 404)
	check(request(b, "GET", photoPath, nil), 200)
	if !strings.Contains(request(b, "GET", "/api/me/export", nil).Body.String(), "Test portrait") {
		t.Fatal("photo metadata missing from export")
	}
	check(request(b, "DELETE", "/api/me", nil), 200)
	check(request(c, "GET", photoPath, nil), 404)
}
