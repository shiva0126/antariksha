package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func TestCommunityAccessMatrix(t *testing.T) {
	dsn := os.Getenv("ACCOUNT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("ACCOUNT_TEST_DATABASE_URL required")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	s := NewServer(nil, PostgresCache{Pool: pool}, nil)
	type user struct {
		id, handle string
		cookie     *http.Cookie
	}
	var users []user
	defer func() {
		for _, u := range users {
			_, _ = pool.Exec(context.Background(), `DELETE FROM member_accounts WHERE id=$1`, u.id)
		}
	}()
	request := func(u user, method, path string, body any) *httptest.ResponseRecorder {
		raw, _ := json.Marshal(body)
		r := httptest.NewRequest(method, "http://example.com"+path, bytes.NewReader(raw))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", "http://example.com")
		if u.cookie != nil {
			r.AddCookie(u.cookie)
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	expect := func(w *httptest.ResponseRecorder, code int) {
		t.Helper()
		if w.Code != code {
			t.Fatalf("want %d got %d: %s", code, w.Code, w.Body.String())
		}
	}
	for range 3 {
		u := user{handle: "socialtest_" + randomToken()[:12]}
		w := request(u, "POST", "/api/auth/register", map[string]any{"handle": u.handle, "email": u.handle + "@example.com", "birth_date": "1996-01-01", "birth_time": "10:15", "password": "community-test-password", "consent": true})
		expect(w, 201)
		u.cookie = w.Result().Cookies()[0]
		if err = pool.QueryRow(context.Background(), `SELECT id FROM member_accounts WHERE handle=$1`, u.handle).Scan(&u.id); err != nil {
			t.Fatal(err)
		}
		users = append(users, u)
		expect(request(u, "PUT", "/api/community/settings", map[string]any{"community": true, "birth_date": "1996-01-01", "avatar": "sun", "accent": "#d6b467", "public_bio": "Hello", "interests": []string{"hiking"}}), 200)
	}
	a, b, c := users[0], users[1], users[2]
	expect(request(a, "PUT", "/api/community/settings", map[string]any{"birth_date": "1990-01-01", "community": true}), 400)
	t.Setenv("MODERATOR_HANDLES", c.handle)
	post := func(audience string, media []string) int64 {
		w := request(a, "POST", "/api/community/posts", map[string]any{"caption": "test " + audience, "audience": audience, "media": media})
		expect(w, 201)
		var x struct {
			ID int64 `json:"id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &x)
		return x.ID
	}
	private := post("private", nil)
	follower := post("followers", nil)
	public := post("community", nil)
	expect(request(b, "GET", fmt.Sprintf("/api/community/posts/%d/comments", private), nil), 404)
	expect(request(b, "DELETE", fmt.Sprintf("/api/community/posts/%d", public), nil), 404)
	expect(request(b, "GET", fmt.Sprintf("/api/community/posts/%d/comments", public), nil), 200)
	expect(request(b, "GET", fmt.Sprintf("/api/community/posts/%d/comments", follower), nil), 404)
	expect(request(b, "POST", "/api/community/follows", map[string]string{"target": a.id, "action": "request"}), 200)
	expect(request(b, "GET", fmt.Sprintf("/api/community/posts/%d/comments", follower), nil), 404)
	expect(request(a, "POST", "/api/community/follows", map[string]string{"target": b.id, "action": "accept"}), 200)
	expect(request(b, "GET", fmt.Sprintf("/api/community/posts/%d/comments", follower), nil), 200)

	// Media endpoints use the same ACL and accept only decodable images.
	var data bytes.Buffer
	writer := multipart.NewWriter(&data)
	f, _ := writer.CreateFormFile("photo", "test.png")
	_ = png.Encode(f, image.NewRGBA(image.Rect(0, 0, 3, 3)))
	_ = writer.Close()
	r := httptest.NewRequest("POST", "http://example.com/api/community/media", &data)
	r.AddCookie(a.cookie)
	r.Header.Set("Origin", "http://example.com")
	r.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	expect(w, 201)
	var media struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &media)
	expect(request(b, "GET", "/api/community/media/"+media.ID, nil), 404)
	photoPost := post("followers", []string{media.ID})
	expect(request(b, "GET", "/api/community/media/"+media.ID, nil), 200)
	expect(request(a, "POST", "/api/community/follows", map[string]string{"target": b.id, "action": "remove"}), 200)
	expect(request(b, "GET", "/api/community/media/"+media.ID, nil), 404)
	expect(request(a, "DELETE", fmt.Sprintf("/api/community/posts/%d", photoPost), nil), 200)
	expect(request(a, "GET", "/api/community/media/"+media.ID, nil), 404)

	// Family invitations grant no access until accepted, and viewers cannot edit.
	expect(request(a, "POST", "/api/families", map[string]string{"name": "Test family"}), 200)
	var group string
	_ = pool.QueryRow(context.Background(), `SELECT id FROM family_groups WHERE owner=$1`, a.id).Scan(&group)
	path := "/api/families/" + group
	expect(request(a, "POST", path+"/members", map[string]string{"action": "invite", "handle": b.handle, "role": "viewer"}), 200)
	noticeResponse := request(b, "GET", "/api/me/notifications", nil)
	expect(noticeResponse, 200)
	if !strings.Contains(noticeResponse.Body.String(), `"kind":"family"`) {
		t.Fatal("family notification missing")
	}
	var noticeID int64
	if err = pool.QueryRow(context.Background(), `SELECT id FROM member_notifications WHERE recipient=$1 AND kind='family'`, b.id).Scan(&noticeID); err != nil {
		t.Fatal(err)
	}
	noticePath := fmt.Sprintf("/api/me/notifications/%d", noticeID)
	expect(request(c, "POST", noticePath, map[string]string{"action": "read"}), 404)
	expect(request(b, "POST", noticePath, map[string]string{"action": "read"}), 200)
	if !strings.Contains(request(b, "GET", "/api/me/notifications", nil).Body.String(), `"read":true`) {
		t.Fatal("read state missing")
	}
	expect(request(b, "GET", path, nil), 404)
	expect(request(b, "POST", path+"/members", map[string]string{"action": "accept"}), 200)
	expect(request(b, "GET", path, nil), 200)
	if strings.Contains(request(b, "GET", "/api/me/notifications", nil).Body.String(), `"kind":"family"`) {
		t.Fatal("accepted invitation still actionable")
	}
	expect(request(c, "GET", path, nil), 404)
	expect(request(b, "POST", path+"/people", map[string]any{"name": "Person", "consent": true}), 403)
	expect(request(a, "POST", path+"/people", map[string]any{"name": "Person", "consent": true}), 200)
	expect(request(a, "POST", path+"/members", map[string]string{"action": "remove", "handle": b.handle}), 200)
	expect(request(b, "GET", path, nil), 404)

	// Matched-only messaging and revocation, followed by block propagation.
	for _, u := range []user{a, b} {
		expect(request(u, "PUT", "/api/matrimony/me", map[string]any{"active": true, "consent": true, "details": matrimonyDetails{Introduction: "Hello", MinAge: 18, MaxAge: 60}}), 200)
	}
	expect(request(a, "POST", "/api/matrimony/messages/"+b.id, map[string]string{"body": "hello"}), 403)
	expect(request(a, "POST", "/api/matrimony/interests", map[string]string{"target": b.id, "action": "send"}), 200)
	expect(request(a, "POST", "/api/matrimony/messages/"+b.id, map[string]string{"body": "hello"}), 403)
	expect(request(b, "POST", "/api/matrimony/interests", map[string]string{"target": a.id, "action": "accept"}), 200)
	expect(request(a, "POST", "/api/matrimony/messages/"+b.id, map[string]string{"body": "hello"}), 200)
	notices := request(b, "GET", "/api/me/notifications", nil)
	expect(notices, 200)
	if !strings.Contains(notices.Body.String(), `"kind":"message"`) || strings.Contains(notices.Body.String(), "hello") {
		t.Fatal("notification content invalid")
	}
	if err = pool.QueryRow(context.Background(), `SELECT id FROM member_notifications WHERE recipient=$1 AND kind='message'`, b.id).Scan(&noticeID); err != nil {
		t.Fatal(err)
	}
	expect(request(b, "POST", fmt.Sprintf("/api/me/notifications/%d", noticeID), map[string]string{"action": "dismiss"}), 200)
	if strings.Contains(request(b, "GET", "/api/me/notifications", nil).Body.String(), `"kind":"message"`) {
		t.Fatal("dismissed notification leaked")
	}
	expect(request(a, "POST", "/api/matrimony/messages/"+b.id, map[string]string{"body": "another"}), 200)
	expect(request(c, "GET", "/api/matrimony/messages/"+b.id, nil), 403)
	expect(request(b, "POST", "/api/community/blocks", map[string]any{"target": a.id, "block": true}), 200)
	if strings.Contains(request(b, "GET", "/api/me/notifications", nil).Body.String(), a.handle) {
		t.Fatal("blocked actor leaked into notifications")
	}
	expect(request(a, "POST", "/api/matrimony/messages/"+b.id, map[string]string{"body": "hello"}), 403)
	expect(request(b, "GET", fmt.Sprintf("/api/community/posts/%d/comments", public), nil), 404)

	// Moderator access is explicit, report decisions are recorded and hide content.
	expect(request(a, "GET", "/api/community/moderation", nil), 403)
	expect(request(c, "POST", fmt.Sprintf("/api/community/posts/%d/report", public), map[string]string{"reason": "test report"}), 200)
	var report int64
	_ = pool.QueryRow(context.Background(), `SELECT id FROM community_reports WHERE post_id=$1`, public).Scan(&report)
	expect(request(c, "POST", "/api/community/moderation", map[string]any{"report": report, "hide": true}), 200)
	expect(request(c, "GET", fmt.Sprintf("/api/community/posts/%d/comments", public), nil), 404)
	expect(request(a, "GET", fmt.Sprintf("/api/community/posts/%d/comments", public), nil), 200)

	// Private exports never contain session/password material.
	w = request(a, "GET", "/api/me/export", nil)
	expect(w, 200)
	if strings.Contains(w.Body.String(), "password_hash") || strings.Contains(w.Body.String(), a.cookie.Value) {
		t.Fatal("secret in export")
	}
	// Recovery is single use and invalidates every old session.
	w = request(a, "POST", "/api/me/recovery-key", map[string]string{"password": "community-test-password"})
	expect(w, 200)
	var key struct {
		Key string `json:"recovery_key"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &key)
	recovery := map[string]string{"handle": a.handle + "@example.com", "key": key.Key, "password": "replacement-password"}
	expect(request(user{}, "POST", "/api/auth/recover", recovery), 200)
	expect(request(a, "GET", "/api/me", nil), 401)
	expect(request(user{}, "POST", "/api/auth/recover", recovery), 401)
}

func TestPhoneEncryption(t *testing.T) {
	t.Setenv("CONTACT_ENCRYPTION_KEY", strings.Repeat("a", 64))
	raw, err := encryptPhone("+919876543210")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("9876543210")) {
		t.Fatal("plaintext stored")
	}
	phone, err := decryptPhone(raw)
	if err != nil || phone != "+919876543210" {
		t.Fatal("roundtrip failed")
	}
	raw[len(raw)-1] ^= 1
	if _, err = decryptPhone(raw); err == nil {
		t.Fatal("tampered ciphertext accepted")
	}
}
