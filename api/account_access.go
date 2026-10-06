package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// Coarse, self-reported labels only: never store full user agents or IP addresses.
func sessionClientLabel(ua string) string {
	browser, platform := "Browser", "unknown device"
	switch {
	case strings.Contains(ua, "Edg/"):
		browser = "Edge"
	case strings.Contains(ua, "Firefox/"):
		browser = "Firefox"
	case strings.Contains(ua, "Chrome/"):
		browser = "Chrome"
	case strings.Contains(ua, "Safari/"):
		browser = "Safari"
	}
	switch {
	case strings.Contains(ua, "Android"):
		platform = "Android"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		platform = "iPhone / iPad"
	case strings.Contains(ua, "Windows"):
		platform = "Windows"
	case strings.Contains(ua, "Macintosh"):
		platform = "Mac"
	case strings.Contains(ua, "Linux"):
		platform = "Linux"
	}
	return browser + " on " + platform
}

func (s *Server) listSessions(w http.ResponseWriter, r *http.Request, id string) {
	cookie, _ := r.Cookie("antariksha_session")
	s.memberRows(w, r, `SELECT id::text id,client_label,created_at,expires_at,token_hash=$2 current FROM member_sessions WHERE account_id=$1 AND expires_at>now() ORDER BY created_at DESC NULLS LAST,id DESC`, id, sessionHash(cookie.Value))
}

// Serialize credential mutations against sign-in and recheck the requesting
// session under the account lock, so a just-revoked session cannot change access.
func (s *Server) lockSecurity(w http.ResponseWriter, r *http.Request, id, password string) (pgx.Tx, bool) {
	if len(password) > 72 || !s.securityBudget.allow(id, time.Now()) {
		problem(w, 429, fmt.Errorf("too many security attempts; try again shortly"))
		return nil, false
	}
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 503, fmt.Errorf("security unavailable"))
		return nil, false
	}
	var hash string
	err = tx.QueryRow(r.Context(), `SELECT password_hash FROM member_accounts WHERE id=$1 AND NOT suspended FOR UPDATE`, id).Scan(&hash)
	cookie, cookieErr := r.Cookie("antariksha_session")
	var active bool
	if err == nil && cookieErr == nil {
		err = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM member_sessions WHERE account_id=$1 AND token_hash=$2 AND expires_at>now())`, id, sessionHash(cookie.Value)).Scan(&active)
	}
	if err != nil || cookieErr != nil || !active || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		tx.Rollback(r.Context())
		problem(w, 403, fmt.Errorf("current password and active session required"))
		return nil, false
	}
	return tx, true
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Password    string `json:"password"`
		NewPassword string `json:"new_password"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if len(in.NewPassword) < 12 || len(in.NewPassword) > 72 || in.NewPassword == in.Password {
		problem(w, 400, fmt.Errorf("choose a different password of 12–72 bytes"))
		return
	}
	tx, ok := s.lockSecurity(w, r, id, in.Password)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	hash, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err == nil {
		_, err = tx.Exec(r.Context(), `UPDATE member_accounts SET password_hash=$2,recovery_hash=NULL WHERE id=$1`, id, string(hash))
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `DELETE FROM member_sessions WHERE account_id=$1`, id)
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `DELETE FROM member_email_tokens WHERE account_id=$1`, id)
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `INSERT INTO member_audit(actor,action,subject,details) VALUES($1,'password_changed',$1,'{}')`, id)
	}
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 503, fmt.Errorf("password could not be changed"))
		return
	}
	clearSessionCookie(w, r)
	writeJSON(w, 200, map[string]string{"message": "Password changed. Sign in again on each device and generate a new recovery key."})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "antariksha_session", Value: "", Path: "/api", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: secureMemberCookie(r)})
}

func (s *Server) revokeSession(w http.ResponseWriter, r *http.Request, id string) {
	target, err := strconv.ParseInt(r.PathValue("session"), 10, 64)
	if err != nil || target < 1 {
		problem(w, 400, fmt.Errorf("invalid session"))
		return
	}
	var in struct {
		Password string `json:"password"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	tx, ok := s.lockSecurity(w, r, id, in.Password)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	var hash string
	err = tx.QueryRow(r.Context(), `DELETE FROM member_sessions WHERE account_id=$1 AND id=$2 RETURNING token_hash`, id, target).Scan(&hash)
	if err != nil {
		problem(w, 404, fmt.Errorf("session unavailable"))
		return
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO member_audit(actor,action,subject,details) VALUES($1,'session_revoked',$1,'{}')`, id)
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 503, fmt.Errorf("session could not be revoked"))
		return
	}
	cookie, _ := r.Cookie("antariksha_session")
	current := hash == sessionHash(cookie.Value)
	if current {
		clearSessionCookie(w, r)
	}
	writeJSON(w, 200, map[string]bool{"ok": true, "signed_out": current})
}
