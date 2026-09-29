package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
		s.mux.Handle(route.method+" "+route.path, s.limit(s.accountGuard(route.fn), 10))
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
			if r.Method != "DELETE" && r.Header.Get("Content-Type") != "application/json" {
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
	return true
}

type memberCredentials struct {
	Handle   string `json:"handle"`
	Password string `json:"password"`
	Consent  bool   `json:"consent"`
}

func (s *Server) issueSession(w http.ResponseWriter, r *http.Request, id string) error {
	token := randomToken()
	expires := time.Now().Add(7 * 24 * time.Hour)
	_, err := s.cache.(PostgresCache).Pool.Exec(r.Context(), `INSERT INTO member_sessions(token_hash,account_id,expires_at) VALUES($1,$2,$3)`, sessionHash(token), id, expires)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: "antariksha_session", Value: token, Path: "/api", HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: r.TLS != nil, Expires: expires})
	return nil
}
func (s *Server) memberID(r *http.Request) (string, error) {
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
	in.Handle = strings.ToLower(strings.TrimSpace(in.Handle))
	if !in.Consent || !accountHandle.MatchString(in.Handle) || len(in.Password) < 12 || len(in.Password) > 72 {
		problem(w, 400, fmt.Errorf("use a 3–32 character handle, a 12–72 byte password, and accept private account storage"))
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		problem(w, 500, fmt.Errorf("registration unavailable"))
		return
	}
	id := randomToken()
	_, err = s.cache.(PostgresCache).Pool.Exec(r.Context(), `INSERT INTO member_accounts(id,handle,password_hash,consent_version) VALUES($1,$2,$3,'private-profile-v1')`, id, in.Handle, string(hash))
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			problem(w, 409, fmt.Errorf("handle unavailable"))
		} else {
			problem(w, 500, fmt.Errorf("registration unavailable"))
		}
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
	err := s.cache.(PostgresCache).Pool.QueryRow(r.Context(), `SELECT id,password_hash FROM member_accounts WHERE handle=$1`, strings.ToLower(strings.TrimSpace(in.Handle))).Scan(&id, &hash)
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
	http.SetCookie(w, &http.Cookie{Name: "antariksha_session", Value: "", Path: "/api", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: r.TLS != nil})
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) getMember(w http.ResponseWriter, r *http.Request) {
	id, err := s.memberID(r)
	if err != nil {
		problem(w, 401, fmt.Errorf("sign in required"))
		return
	}
	var handle string
	var raw []byte
	err = s.cache.(PostgresCache).Pool.QueryRow(r.Context(), `SELECT handle,profile FROM member_accounts WHERE id=$1`, id).Scan(&handle, &raw)
	if err != nil {
		problem(w, 500, fmt.Errorf("profile unavailable"))
		return
	}
	writeJSON(w, 200, map[string]any{"handle": handle, "profile": json.RawMessage(raw), "visibility": "private", "phone_verification_available": false})
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
	tag, err := s.cache.(PostgresCache).Pool.Exec(r.Context(), `DELETE FROM member_accounts WHERE id=$1`, id)
	if err != nil || tag.RowsAffected() != 1 {
		problem(w, 500, fmt.Errorf("deletion failed"))
		return
	}
	s.logoutMember(w, r)
}
