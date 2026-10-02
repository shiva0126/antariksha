package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type memberProfile struct {
	Name    string `json:"name"`
	City    string `json:"city"`
	Bio     string `json:"bio"`
	Hobbies string `json:"hobbies"`
	Goals   string `json:"goals"`
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func sessionHash(token string) string {
	v := sha256.Sum256([]byte(token))
	return hex.EncodeToString(v[:])
}

var accountHandle = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)

func (s *Server) accountRoutes() {
	for _, route := range []struct {
		method, path string
		fn           http.HandlerFunc
	}{
		{"POST", "/api/auth/register", s.registerMember}, {"POST", "/api/auth/login", s.loginMember},
		{"POST", "/api/auth/logout", s.logoutMember}, {"GET", "/api/me", s.getMember},
		{"PUT", "/api/me", s.updateMember}, {"DELETE", "/api/me", s.deleteMember},
	} {
		limit := 10
		if route.method == "GET" {
			limit = 120
		}
		s.mux.Handle(route.method+" "+route.path, s.limit(s.accountGuard(route.fn), limit))
	}
}

func (s *Server) accountGuard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if _, ok := s.cache.(PostgresCache); !ok {
			problem(w, 503, fmt.Errorf("accounts require database configuration"))
			return
		}
		if r.Method != "GET" {
			origin, err := url.Parse(r.Header.Get("Origin"))
			if err != nil || origin.Host != r.Host || (origin.Scheme != "http" && origin.Scheme != "https") || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				problem(w, 403, fmt.Errorf("same-origin request required"))
				return
			}
			mediaType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			photoUpload := r.URL.Path == "/api/community/media" || r.URL.Path == "/api/matrimony/photos"
			if r.Method != "DELETE" && mediaType != "application/json" && !(photoUpload && mediaType == "multipart/form-data") {
				problem(w, 415, fmt.Errorf("JSON required"))
				return
			}
		}
		next(w, r)
	}
}

func memberInput(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		problem(w, 400, fmt.Errorf("invalid request"))
		return false
	}
	if d.Decode(new(any)) != io.EOF {
		problem(w, 400, fmt.Errorf("one JSON object required"))
		return false
	}
	return true
}

type memberCredentials struct {
	Handle    string `json:"handle"`
	Email     string `json:"email"`
	BirthDate string `json:"birth_date"`
	BirthTime string `json:"birth_time"`
	Password  string `json:"password"`
	Consent   bool   `json:"consent"`
}

func (s *Server) issueSession(w http.ResponseWriter, r *http.Request, id string) error {
	token := randomToken()
	expires := time.Now().Add(7 * 24 * time.Hour)
	_, err := s.cache.(PostgresCache).Pool.Exec(r.Context(), `INSERT INTO member_sessions(token_hash,account_id,expires_at) VALUES($1,$2,$3)`, sessionHash(token), id, expires)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: "antariksha_session", Value: token, Path: "/api", HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: secureMemberCookie(r), Expires: expires})
	return nil
}
func (s *Server) memberID(r *http.Request) (string, error) {
	if _, ok := s.cache.(PostgresCache); !ok {
		return "", fmt.Errorf("accounts require a database")
	}
	c, err := r.Cookie("antariksha_session")
	if err != nil || len(c.Value) != 64 {
		return "", fmt.Errorf("sign in required")
	}
	var id string
	err = s.cache.(PostgresCache).Pool.QueryRow(r.Context(), `SELECT account_id FROM member_sessions WHERE token_hash=$1 AND expires_at>now()`, sessionHash(c.Value)).Scan(&id)
	return id, err
}
func (s *Server) registerMember(w http.ResponseWriter, r *http.Request) {
	var in memberCredentials
	if !memberInput(w, r, &in) {
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	address, addressErr := mail.ParseAddress(in.Email)
	dob, dateErr := time.Parse("2006-01-02", in.BirthDate)
	bt, timeErr := time.Parse("15:04", in.BirthTime)
	if !in.Consent || addressErr != nil || address.Address != in.Email || len(in.Email) > 254 || dateErr != nil || timeErr != nil || dob.Year() < 1900 || dob.After(time.Now()) || len(in.Password) < 12 || len(in.Password) > 72 {
		problem(w, 400, fmt.Errorf("enter a valid email, date of birth, birth time and a 12–72 byte password"))
		return
	}
	in.Handle = strings.ToLower(strings.TrimSpace(in.Handle))
	if in.Handle == "" {
		in.Handle = "member_" + randomToken()[:12]
	}
	if !accountHandle.MatchString(in.Handle) {
		problem(w, 400, fmt.Errorf("invalid handle"))
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		problem(w, 500, fmt.Errorf("registration unavailable"))
		return
	}
	id := randomToken()
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("registration unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	_, err = tx.Exec(r.Context(), `INSERT INTO member_accounts(id,handle,password_hash,consent_version,email,birth_date,birth_time) VALUES($1,$2,$3,'private-profile-v2',$4,$5,$6)`, id, in.Handle, string(hash), in.Email, dob, bt.Format("15:04"))
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			problem(w, 409, fmt.Errorf("account already exists; sign in or recover your account"))
		} else {
			problem(w, 500, fmt.Errorf("registration unavailable"))
		}
		return
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO member_settings(account_id,birth_date) VALUES($1,$2)`, id, dob)
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 500, fmt.Errorf("registration unavailable"))
		return
	}
	if err = s.issueSession(w, r, id); err != nil {
		problem(w, 500, fmt.Errorf("account created; please sign in"))
		return
	}
	writeJSON(w, 201, map[string]bool{"ok": true})
}

var dummyPassword, _ = bcrypt.GenerateFromPassword([]byte("unused-comparison-password"), bcrypt.DefaultCost)

func (s *Server) loginMember(w http.ResponseWriter, r *http.Request) {
	var in memberCredentials
	if !memberInput(w, r, &in) {
		return
	}
	var id, hash string
	identifier := in.Email
	if identifier == "" {
		identifier = in.Handle
	}
	err := s.cache.(PostgresCache).Pool.QueryRow(r.Context(), `SELECT id,password_hash FROM member_accounts WHERE lower(email)=$1 OR (email IS NULL AND handle=$1)`, strings.ToLower(strings.TrimSpace(identifier))).Scan(&id, &hash)
	if err != nil {
		hash = string(dummyPassword)
	}
	valid := bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) == nil
	if err != nil || !valid {
		problem(w, 401, fmt.Errorf("invalid credentials"))
		return
	}
	if s.issueSession(w, r, id) != nil {
		problem(w, 500, fmt.Errorf("sign-in unavailable"))
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) logoutMember(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("antariksha_session"); err == nil {
		if _, err = s.cache.(PostgresCache).Pool.Exec(r.Context(), `DELETE FROM member_sessions WHERE token_hash=$1`, sessionHash(c.Value)); err != nil {
			problem(w, 500, fmt.Errorf("logout failed"))
			return
		}
	}
	http.SetCookie(w, &http.Cookie{Name: "antariksha_session", Value: "", Path: "/api", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: secureMemberCookie(r)})
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) getMember(w http.ResponseWriter, r *http.Request) {
	id, err := s.memberID(r)
	if err != nil {
		problem(w, 401, fmt.Errorf("sign in required"))
		return
	}
	var handle string
	var email, date, birthTime string
	var raw []byte
	var verified bool
	err = s.cache.(PostgresCache).Pool.QueryRow(r.Context(), `SELECT handle,profile,COALESCE(email,''),COALESCE(to_char(birth_date,'YYYY-MM-DD'),''),COALESCE(to_char(birth_time,'HH24:MI'),'') FROM member_accounts WHERE id=$1`, id).Scan(&handle, &raw, &email, &date, &birthTime)
	if err != nil {
		problem(w, 500, fmt.Errorf("profile unavailable"))
		return
	}
	if err = s.membersDB().QueryRow(r.Context(), `SELECT email_verified_at IS NOT NULL FROM member_accounts WHERE id=$1`, id).Scan(&verified); err != nil {
		problem(w, 500, fmt.Errorf("profile unavailable"))
		return
	}
	writeJSON(w, 200, map[string]any{"id": id, "handle": handle, "email": email, "birth_date": date, "birth_time": birthTime, "profile": json.RawMessage(raw), "visibility": "private", "phone_verification_available": phoneReady(), "email_verified": verified, "email_delivery_available": s.mailer != nil})
}
func (s *Server) updateMember(w http.ResponseWriter, r *http.Request) {
	id, err := s.memberID(r)
	if err != nil {
		problem(w, 401, fmt.Errorf("sign in required"))
		return
	}
	var p memberProfile
	if !memberInput(w, r, &p) {
		return
	}
	if len(p.Name) > 100 || len(p.City) > 100 || len(p.Bio) > 2000 || len(p.Hobbies) > 1000 || len(p.Goals) > 1000 {
		problem(w, 400, fmt.Errorf("profile field too long"))
		return
	}
	raw, _ := json.Marshal(p)
	_, err = s.cache.(PostgresCache).Pool.Exec(r.Context(), `UPDATE member_accounts SET profile=$1 WHERE id=$2`, raw, id)
	if err != nil {
		problem(w, 500, fmt.Errorf("save failed"))
		return
	}
	s.getMember(w, r)
}
func (s *Server) deleteMember(w http.ResponseWriter, r *http.Request) {
	id, err := s.memberID(r)
	if err != nil {
		problem(w, 401, fmt.Errorf("sign in required"))
		return
	}
	tag, err := s.cache.(PostgresCache).Pool.Exec(r.Context(), `WITH removed_chats AS (DELETE FROM chat_sessions WHERE id IN (SELECT session_id FROM chat_owners WHERE account_id=$1)) DELETE FROM member_accounts WHERE id=$1`, id)
	if err != nil || tag.RowsAffected() != 1 {
		problem(w, 500, fmt.Errorf("deletion failed"))
		return
	}
	s.logoutMember(w, r)
}
