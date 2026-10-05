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
	"sync"
	"testing"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fakeMailer struct {
	mu   sync.Mutex
	sent []string
}

func (f *fakeMailer) Send(_ context.Context, to, subject, body string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, to+"|"+subject+"|"+body)
	return nil
}

func TestMatrimonyV2Journey(t *testing.T) {
	dsn := os.Getenv("ACCOUNT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires isolated migrated database")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	s := NewServer(realEngine(t), PostgresCache{Pool: pool}, nil)
	mail := &fakeMailer{}
	s.mailer = mail
	t.Setenv("PUBLIC_ORIGIN", "https://astrisk.example")
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
	bengaluru := map[string]any{"name": "Bengaluru, Karnataka", "lat": 12.97, "lon": 77.59, "tz": "Asia/Kolkata"}
	newMember := func(prefix, dob, tob, kind, city, diet string, place map[string]any, horoscope bool) member {
		h := prefix + "_" + randomToken()[:8]
		body := map[string]any{"handle": h, "name": strings.ToUpper(prefix[:1]) + prefix[1:], "email": h + "@example.com", "password": "matrimony-v2-password", "birth_date": dob, "birth_time": tob, "consent": true}
		if place != nil {
			body["birth_place"] = place
		}
		w := request(member{}, "POST", "/api/auth/register", body)
		check(w, 201)
		u := member{handle: h, cookie: w.Result().Cookies()[0]}
		var me struct {
			ID         string          `json:"id"`
			BirthPlace json.RawMessage `json:"birth_place"`
		}
		_ = json.Unmarshal(request(u, "GET", "/api/me", nil).Body.Bytes(), &me)
		u.id = me.ID
		t.Cleanup(func() { pool.Exec(context.Background(), `DELETE FROM member_accounts WHERE id=$1`, u.id) })
		if place != nil && !strings.Contains(string(me.BirthPlace), "Bengaluru") {
			t.Fatalf("birth place not stored: %s", me.BirthPlace)
		}
		check(request(u, "PUT", "/api/community/settings", map[string]any{"community": true, "birth_date": dob, "avatar": "sun", "interests": []string{"music"}}), 200)
		check(request(u, "PUT", "/api/matrimony/me", map[string]any{"active": true, "consent": true, "horoscope": horoscope, "details": map[string]any{
			"display_name": prefix, "profile_kind": kind, "city": city, "religion": "hindu", "mother_tongue": "kannada", "diet": diet,
			"marital_status": "never_married", "timeline": "within_year", "height_cm": 170, "min_age": 18, "max_age": 60}}), 200)
		return u
	}
	groom := newMember("rohan", "1994-11-02", "06:40", "groom", "Bengaluru", "vegetarian", bengaluru, true)
	bride := newMember("ananya", "1996-05-14", "10:15", "bride", "Bengaluru", "vegetarian", bengaluru, true)
	private := newMember("meera", "1997-03-03", "08:00", "bride", "Mysuru", "non_vegetarian", nil, false)

	// Invalid structured values are rejected.
	check(request(groom, "PUT", "/api/matrimony/me", map[string]any{"active": true, "consent": true, "details": map[string]any{"diet": "fish-only", "min_age": 18, "max_age": 60}}), 400)
	check(request(groom, "PUT", "/api/me/birth-place", map[string]any{"name": "Nowhere", "lat": 12, "lon": 77, "tz": "Mars/Base"}), 400)

	type disc struct {
		Items []candidate `json:"items"`
		Total int         `json:"total"`
		Pages int         `json:"pages"`
	}
	discover := func(u member, q string) disc {
		t.Helper()
		w := request(u, "GET", "/api/matrimony/discover"+q, nil)
		check(w, 200)
		var d disc
		if err := json.Unmarshal(w.Body.Bytes(), &d); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(w.Body.String(), "1996-05-14") || strings.Contains(w.Body.String(), "10:15") || strings.Contains(w.Body.String(), "77.59") {
			t.Fatal("discovery leaked birth details: " + w.Body.String())
		}
		return d
	}
	d := discover(groom, "")
	var seenBride, seenPrivate *candidate
	for i := range d.Items {
		switch d.Items[i].ID {
		case bride.id:
			seenBride = &d.Items[i]
		case private.id:
			seenPrivate = &d.Items[i]
		}
	}
	if seenBride == nil || seenPrivate == nil {
		t.Fatalf("both profiles should be listed: %+v", d)
	}
	// 1994-11-02 06:40 and 1996-05-14 10:15 Bengaluru score 19.5 (see TestMatchChat).
	if seenBride.Horoscope == nil || seenBride.Horoscope.Guna != 19.5 || seenBride.Horoscope.TheirMangal == "unknown" || seenBride.Horoscope.NumberRel == "" {
		t.Fatalf("horoscope match: %+v", seenBride.Horoscope)
	}
	if !strings.Contains(strings.Join(seenBride.Reasons, ";"), "19.5/36 gunas") || !strings.Contains(strings.Join(seenBride.Reasons, ";"), "Same city") || !strings.Contains(strings.Join(seenBride.Reasons, ";"), "Both vegetarian") {
		t.Fatalf("reasons: %v", seenBride.Reasons)
	}
	if seenPrivate.Horoscope != nil || seenPrivate.HoroReason == "" {
		t.Fatal("a member who did not opt in must not show a horoscope")
	}
	if d = discover(groom, "?diet=vegetarian"); d.Total != 1 || d.Items[0].ID != bride.id {
		t.Fatalf("diet filter: %+v", d)
	}
	if d = discover(groom, "?min_guna=20"); d.Total != 0 {
		t.Fatalf("min guna filter: %+v", d)
	}
	if d = discover(groom, "?city=mysuru"); d.Total != 1 || d.Items[0].ID != private.id {
		t.Fatalf("city filter: %+v", d)
	}
	// Not now hides a profile until "show skipped"; saved filters to the shortlist.
	check(request(groom, "POST", "/api/matrimony/saved", map[string]string{"target": private.id, "kind": "skipped"}), 200)
	check(request(groom, "POST", "/api/matrimony/saved", map[string]string{"target": bride.id, "kind": "saved"}), 200)
	if d = discover(groom, ""); d.Total != 1 {
		t.Fatalf("skipped profile still listed: %+v", d)
	}
	if d = discover(groom, "?saved=1"); d.Total != 1 || d.Items[0].Saved != "saved" {
		t.Fatalf("saved filter: %+v", d)
	}
	check(request(groom, "PUT", "/api/matrimony/search", map[string]string{"diet": "vegetarian", "min_guna": "18"}), 200)
	check(request(groom, "PUT", "/api/matrimony/search", map[string]string{"password": "x"}), 400)

	// Full comparison and Ask Astrisk work only when both opted in.
	w := request(groom, "GET", "/api/matrimony/compare/"+bride.id, nil)
	check(w, 200)
	if !strings.Contains(w.Body.String(), `"total":19.5`) || strings.Contains(w.Body.String(), "1996-05-14") {
		t.Fatalf("compare: %s", w.Body.String())
	}
	check(request(groom, "GET", "/api/matrimony/compare/"+private.id, nil), 404)
	w = request(groom, "POST", "/api/matrimony/compare/"+bride.id+"/chat", map[string]any{"question": "Explain our guna score", "history": []any{}})
	check(w, 200)
	if !strings.Contains(w.Body.String(), "19.5 out of 36 gunas") {
		t.Fatalf("compare chat: %s", w.Body.String())
	}

	// Interest with a note, acceptance notice, alerts and contact sharing.
	check(request(groom, "POST", "/api/matrimony/interests", map[string]string{"target": bride.id, "action": "send", "note": strings.Repeat("x", 301)}), 400)
	check(request(groom, "POST", "/api/matrimony/interests", map[string]string{"target": bride.id, "action": "send", "note": "We both love music. Would you like to talk?"}), 200)
	w = request(bride, "GET", "/api/matrimony/interests", nil)
	if !strings.Contains(w.Body.String(), "We both love music") || !strings.Contains(w.Body.String(), `"display_name":"rohan"`) {
		t.Fatalf("interest note: %s", w.Body.String())
	}
	check(request(bride, "GET", "/api/matrimony/contact/"+groom.id, nil), 403)
	check(request(bride, "POST", "/api/matrimony/interests", map[string]string{"target": groom.id, "action": "accept"}), 200)
	w = request(groom, "GET", "/api/me/notifications", nil)
	if !strings.Contains(w.Body.String(), `"kind":"accepted"`) {
		t.Fatalf("acceptance notice: %s", w.Body.String())
	}
	check(request(bride, "PUT", "/api/matrimony/contact/"+groom.id, map[string]string{"contact": "+91 98765 43210"}), 200)
	w = request(groom, "GET", "/api/matrimony/contact/"+bride.id, nil)
	if !strings.Contains(w.Body.String(), `"theirs":"+91 98765 43210"`) {
		t.Fatalf("contact: %s", w.Body.String())
	}
	check(request(groom, "POST", "/api/matrimony/messages/"+bride.id, map[string]string{"body": "Hello!"}), 200)

	// Alerts go out after the delay: verified email only, never message text.
	pool.Exec(context.Background(), `UPDATE member_accounts SET email_verified_at=now() WHERE id=$1`, bride.id)
	pool.Exec(context.Background(), `UPDATE member_notifications SET created_at=now()-interval '5 minutes' WHERE recipient IN ($1,$2)`, groom.id, bride.id)
	var pushed []string
	pushSender = func(_ context.Context, payload []byte, sub *webpush.Subscription, _ *webpush.Options) (int, error) {
		pushed = append(pushed, sub.Endpoint+" "+string(payload))
		return 201, nil
	}
	check(request(groom, "POST", "/api/push/subscribe", map[string]any{"endpoint": "https://evil.example/push", "keys": map[string]string{"p256dh": strings.Repeat("a", 87), "auth": strings.Repeat("b", 22)}}), 400)
	check(request(groom, "POST", "/api/push/subscribe", map[string]any{"endpoint": "https://fcm.googleapis.com/fcm/send/abc", "keys": map[string]string{"p256dh": strings.Repeat("a", 87), "auth": strings.Repeat("b", 22)}}), 200)
	check(request(groom, "GET", "/api/push/key", nil), 200)
	n, err := s.deliverAlerts(context.Background())
	if err != nil || n < 2 {
		t.Fatalf("alerts delivered %d: %v", n, err)
	}
	if len(mail.sent) != 1 || !strings.HasPrefix(mail.sent[0], bride.handle+"@example.com|Astrisk: New message from Rohan") || strings.Contains(mail.sent[0], "Hello!") {
		t.Fatalf("email alerts: %q", mail.sent)
	}
	if len(pushed) != 1 || !strings.Contains(pushed[0], "accepted your interest") {
		t.Fatalf("push alerts: %q", pushed)
	}
	if n, _ = s.deliverAlerts(context.Background()); n != 0 {
		t.Fatal("alerts must be delivered once")
	}

	// The daily interest budget.
	for i := 0; i < dailyInterestLimit; i++ {
		pool.Exec(context.Background(), `INSERT INTO member_accounts(id,handle,password_hash,consent_version) VALUES($1,$1,'x','test')`, "filler"+randomToken()[:10])
	}
	pool.Exec(context.Background(), `INSERT INTO matrimony_interests(sender,recipient) SELECT $1,id FROM member_accounts WHERE id LIKE 'filler%' LIMIT $2`, private.id, dailyInterestLimit)
	defer pool.Exec(context.Background(), `DELETE FROM member_accounts WHERE id LIKE 'filler%'`)
	check(request(private, "POST", "/api/matrimony/interests", map[string]string{"target": groom.id, "action": "send"}), 429)

	// Selfie verification: needs a published photo, reviewed by a moderator.
	upload := func(u member, path string) *httptest.ResponseRecorder {
		var buf bytes.Buffer
		mp := multipart.NewWriter(&buf)
		f, _ := mp.CreateFormFile("photo", "selfie.png")
		_ = png.Encode(f, image.NewRGBA(image.Rect(0, 0, 3, 3)))
		_ = mp.WriteField("consent", "true")
		_ = mp.Close()
		r := httptest.NewRequest("POST", "http://example.com"+path, &buf)
		r.AddCookie(u.cookie)
		r.Header.Set("Origin", "http://example.com")
		r.Header.Set("Content-Type", mp.FormDataContentType())
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		return w
	}
	check(upload(bride, "/api/matrimony/verification"), 400)
	w = upload(bride, "/api/matrimony/photos")
	check(w, 201)
	var photo struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &photo)
	check(request(bride, "PUT", "/api/matrimony/me", map[string]any{"active": true, "consent": true, "horoscope": true, "photo_ids": []string{photo.ID}, "details": map[string]any{"display_name": "ananya", "profile_kind": "bride", "city": "Bengaluru", "diet": "vegetarian", "min_age": 18, "max_age": 60}}), 200)
	check(upload(bride, "/api/matrimony/verification"), 200)
	check(request(groom, "GET", "/api/matrimony/verifications", nil), 403)
	pool.Exec(context.Background(), `UPDATE member_accounts SET role='moderator' WHERE id=$1`, private.id)
	w = request(private, "GET", "/api/matrimony/verifications", nil)
	if !strings.Contains(w.Body.String(), bride.id) {
		t.Fatalf("verification queue: %s", w.Body.String())
	}
	check(request(private, "GET", "/api/matrimony/verifications/"+bride.id+"/photo", nil), 200)
	check(request(private, "POST", "/api/matrimony/verifications/"+bride.id, map[string]any{"approve": true}), 200)
	check(request(private, "GET", "/api/matrimony/verifications/"+bride.id+"/photo", nil), 404)
	if d = discover(groom, "?verified=1"); d.Total != 1 || !d.Items[0].Verified {
		t.Fatalf("verified filter: %+v", d)
	}

	// Blocking removes saved entries and shared contacts for the pair.
	check(request(groom, "POST", "/api/community/blocks", map[string]any{"target": bride.id, "block": true}), 200)
	var left int
	pool.QueryRow(context.Background(), `SELECT (SELECT count(*) FROM matrimony_saved WHERE owner=$1 AND target=$2)+(SELECT count(*) FROM matrimony_contacts WHERE owner=$2 AND peer=$1)`, groom.id, bride.id).Scan(&left)
	if left != 0 {
		t.Fatal("block must remove shortlist and contacts")
	}
}
