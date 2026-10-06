package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) accountRole(r *http.Request, id string) string {
	var role string
	if err := s.membersDB().QueryRow(r.Context(), `SELECT role FROM member_accounts WHERE id=$1 AND NOT suspended`, id).Scan(&role); err != nil {
		return ""
	}
	return role
}

func (s *Server) adminRoutes() {
	for _, route := range []struct {
		pattern string
		fn      memberHandler
	}{
		{"GET /api/admin/summary", s.adminSummary}, {"GET /api/admin/users", s.adminUsers},
		{"GET /api/admin/operations", s.adminOperations},
		{"GET /api/admin/audit", s.adminAudit}, {"POST /api/admin/users/{user}/action", s.adminUserAction},
	} {
		fn := route.fn
		s.memberRoute(route.pattern, func(w http.ResponseWriter, r *http.Request, id string) {
			if s.accountRole(r, id) != "superadmin" {
				problem(w, 403, fmt.Errorf("superadmin access required"))
				return
			}
			fn(w, r, id)
		})
	}
}

func (s *Server) adminSummary(w http.ResponseWriter, r *http.Request, _ string) {
	var raw []byte
	err := s.membersDB().QueryRow(r.Context(), `SELECT json_build_object(
 'users',count(*),'suspended',count(*) FILTER(WHERE suspended),
 'moderators',count(*) FILTER(WHERE role='moderator'),'superadmins',count(*) FILTER(WHERE role='superadmin'),
 'active_sessions',(SELECT count(*) FROM member_sessions WHERE expires_at>now()),
 'open_post_reports',(SELECT count(*) FROM community_reports WHERE status='open'),
 'open_profile_reports',(SELECT count(*) FROM matrimony_reports WHERE status='open')) FROM member_accounts`).Scan(&raw)
	if err != nil {
		problem(w, 503, fmt.Errorf("admin summary unavailable"))
		return
	}
	writeJSON(w, 200, json.RawMessage(raw))
}

func adminPage(w http.ResponseWriter, r *http.Request) (int, bool) {
	value := r.URL.Query().Get("page")
	if value == "" {
		return 0, true
	}
	page, err := strconv.Atoi(value)
	if err != nil || page < 1 || page > 100000 {
		problem(w, 400, fmt.Errorf("invalid page"))
		return 0, false
	}
	return (page - 1) * 25, true
}

func (s *Server) adminUsers(w http.ResponseWriter, r *http.Request, _ string) {
	offset, ok := adminPage(w, r)
	if !ok {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	status := r.URL.Query().Get("status")
	if len(q) > 254 || (status != "" && status != "active" && status != "suspended") {
		problem(w, 400, fmt.Errorf("invalid user filter"))
		return
	}
	// strpos gives literal case-insensitive search, without LIKE wildcard input.
	s.memberRows(w, r, `SELECT a.id,a.handle,COALESCE(a.email,'') email,a.role,a.suspended,a.created_at,
 a.email_verified_at IS NOT NULL email_verified,
 (SELECT count(*) FROM member_sessions WHERE account_id=a.id AND expires_at>now()) sessions,
 COALESCE(c.community,false) community,COALESCE(m.active,false) matrimony
 FROM member_accounts a LEFT JOIN member_settings c ON c.account_id=a.id LEFT JOIN matrimony_profiles m ON m.account_id=a.id
 WHERE ($1='' OR strpos(lower(a.handle),lower($1))>0 OR strpos(lower(COALESCE(a.email,'')),lower($1))>0)
 AND ($2='' OR ($2='suspended' AND a.suspended) OR ($2='active' AND NOT a.suspended))
 ORDER BY a.created_at DESC,a.id DESC LIMIT 26 OFFSET $3`, q, status, offset)
}

func (s *Server) adminAudit(w http.ResponseWriter, r *http.Request, _ string) {
	offset, ok := adminPage(w, r)
	if !ok {
		return
	}
	s.memberRows(w, r, `SELECT v.id,COALESCE(a.handle,'server operator / deleted account') actor,v.action,v.subject,v.details,v.created_at
 FROM member_audit v LEFT JOIN member_accounts a ON a.id=v.actor ORDER BY v.id DESC LIMIT 26 OFFSET $1`, offset)
}

func (s *Server) adminUserAction(w http.ResponseWriter, r *http.Request, actor string) {
	var in struct {
		Action        string `json:"action"`
		Role          string `json:"role"`
		Reason        string `json:"reason"`
		Password      string `json:"password"`
		ConfirmHandle string `json:"confirm_handle"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if len(in.Reason) < 3 || len(in.Reason) > 500 || len(in.Password) > 72 {
		problem(w, 400, fmt.Errorf("provide your password and a reason (3–500 characters)"))
		return
	}
	switch in.Action {
	case "suspend", "reactivate", "revoke_sessions", "delete":
	case "role":
		if in.Role != "member" && in.Role != "moderator" {
			problem(w, 400, fmt.Errorf("choose member or moderator"))
			return
		}
	default:
		problem(w, 400, fmt.Errorf("invalid admin action"))
		return
	}
	// Per-account budget also bounds password checks and all mutation attempts.
	if !s.adminBudget.allow(actor, time.Now()) {
		problem(w, 429, fmt.Errorf("too many admin actions; try again shortly"))
		return
	}
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 503, fmt.Errorf("admin action unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	var hash string
	if err = tx.QueryRow(r.Context(), `SELECT password_hash FROM member_accounts WHERE id=$1 AND role='superadmin' AND NOT suspended FOR UPDATE`, actor).Scan(&hash); err != nil {
		problem(w, 403, fmt.Errorf("superadmin access required"))
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		problem(w, 403, fmt.Errorf("current superadmin password is incorrect"))
		return
	}
	target := r.PathValue("user")
	var handle, role string
	var suspended bool
	if err = tx.QueryRow(r.Context(), `SELECT handle,role,suspended FROM member_accounts WHERE id=$1 FOR UPDATE`, target).Scan(&handle, &role, &suspended); err != nil {
		problem(w, 404, fmt.Errorf("account unavailable"))
		return
	}
	if target == actor || role == "superadmin" {
		problem(w, 409, fmt.Errorf("superadmin accounts are protected; use your own Security page to manage sessions"))
		return
	}
	if in.Action == "delete" && in.ConfirmHandle != handle {
		problem(w, 400, fmt.Errorf("type the exact account handle to confirm deletion"))
		return
	}
	switch in.Action {
	case "suspend":
		_, err = tx.Exec(r.Context(), `UPDATE member_accounts SET suspended=true WHERE id=$1`, target)
		if err == nil {
			_, err = tx.Exec(r.Context(), `UPDATE member_settings SET community=false WHERE account_id=$1`, target)
		}
		if err == nil {
			_, err = tx.Exec(r.Context(), `UPDATE matrimony_profiles SET active=false WHERE account_id=$1`, target)
		}
	case "reactivate":
		_, err = tx.Exec(r.Context(), `UPDATE member_accounts SET suspended=false WHERE id=$1`, target)
	case "role":
		_, err = tx.Exec(r.Context(), `UPDATE member_accounts SET role=$2 WHERE id=$1`, target, in.Role)
	case "delete":
		_, err = tx.Exec(r.Context(), `WITH removed_chats AS (DELETE FROM chat_sessions WHERE id IN(SELECT session_id FROM chat_owners WHERE account_id=$1)) DELETE FROM member_accounts WHERE id=$1`, target)
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `DELETE FROM member_sessions WHERE account_id=$1`, target)
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `DELETE FROM member_email_tokens WHERE account_id=$1`, target)
	}
	details, _ := json.Marshal(map[string]any{"handle": handle, "reason": in.Reason, "previous_role": role, "previous_suspended": suspended, "requested_role": in.Role})
	if err == nil {
		_, err = tx.Exec(r.Context(), `INSERT INTO member_audit(actor,action,subject,details) VALUES($1,$2,$3,$4)`, actor, "admin_"+in.Action, target, details)
	}
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 503, fmt.Errorf("admin action could not be saved"))
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
