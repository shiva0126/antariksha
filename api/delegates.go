package api

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) delegateRoutes() {
	s.memberRoute("GET /api/matrimony/delegates", s.delegates)
	s.memberRoute("POST /api/matrimony/delegates", s.delegateAction)
	s.memberRoute("GET /api/matrimony/assistance/{owner}", s.delegatedDiscovery)
	s.memberRoute("GET /api/matrimony/shortlist", s.shortlist)
	s.memberRoute("POST /api/matrimony/shortlist", s.saveShortlist)
}
func (s *Server) delegates(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT g.owner,g.delegate,g.owner=$1 mine,a.handle,g.accepted,g.expires_at,delegate_allowed(g.owner,g.delegate) available FROM matrimony_delegates g JOIN member_accounts a ON a.id=CASE WHEN g.owner=$1 THEN g.delegate ELSE g.owner END WHERE (g.owner=$1 OR g.delegate=$1) AND NOT member_blocked(g.owner,g.delegate) AND family_access(g.family_id,$1) ORDER BY g.expires_at DESC`, id)
}
func (s *Server) delegateAction(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Action   string `json:"action"`
		Handle   string `json:"handle"`
		Family   string `json:"family_id"`
		Owner    string `json:"owner"`
		Delegate string `json:"delegate"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	switch in.Action {
	case "grant":
		if !s.adultCommunity(w, r, id) {
			return
		}
		s.memberExec(w, r, `INSERT INTO matrimony_delegates(owner,delegate,family_id) SELECT $1,a.id,$3 FROM member_accounts a JOIN member_settings m ON m.account_id=a.id JOIN matrimony_profiles p ON p.account_id=$1 WHERE a.handle=$2 AND a.id<>$1 AND p.active AND m.community AND m.birth_date<=CURRENT_DATE-INTERVAL '18 years' AND family_access($3,$1) AND family_access($3,a.id) AND NOT member_blocked($1,a.id) ON CONFLICT(owner,delegate) DO UPDATE SET family_id=EXCLUDED.family_id,accepted=false,expires_at=now()+INTERVAL '30 days'`, id, strings.ToLower(strings.TrimSpace(in.Handle)), in.Family)
	case "accept":
		if !s.adultCommunity(w, r, id) {
			return
		}
		s.memberExec(w, r, `UPDATE matrimony_delegates g SET accepted=true WHERE owner=$1 AND delegate=$2 AND expires_at>now() AND family_access(family_id,$1) AND family_access(family_id,$2) AND NOT member_blocked($1,$2)`, in.Owner, id)
	case "revoke":
		s.memberExec(w, r, `DELETE FROM matrimony_delegates WHERE owner=$1 AND delegate=$2`, id, in.Delegate)
	case "leave":
		s.memberExec(w, r, `DELETE FROM matrimony_delegates WHERE owner=$1 AND delegate=$2`, in.Owner, id)
	default:
		problem(w, 400, fmt.Errorf("invalid assistance action"))
	}
}
func (s *Server) delegatedDiscovery(w http.ResponseWriter, r *http.Request, id string) {
	owner := r.PathValue("owner")
	var allowed bool
	if err := s.membersDB().QueryRow(r.Context(), `SELECT delegate_allowed($1,$2)`, owner, id).Scan(&allowed); err != nil || !allowed {
		problem(w, 403, fmt.Errorf("an accepted, active family grant is required"))
		return
	}
	s.matrimonyCandidates(w, r, owner, id)
}
func (s *Server) shortlist(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT l.owner,l.delegate,l.candidate,l.note,l.created_at,a.handle candidate_handle,d.handle delegate_handle,l.owner=$1 mine FROM matrimony_shortlist l JOIN member_accounts a ON a.id=l.candidate JOIN member_accounts d ON d.id=l.delegate WHERE (l.owner=$1 OR l.delegate=$1) AND delegate_allowed(l.owner,l.delegate) AND NOT member_blocked(l.candidate,l.owner) AND NOT member_blocked(l.candidate,l.delegate) AND EXISTS(
 SELECT 1 FROM matrimony_profiles p JOIN member_settings s ON s.account_id=p.account_id
 JOIN matrimony_profiles mine ON mine.account_id=l.owner JOIN member_settings me ON me.account_id=l.owner
 WHERE p.account_id=l.candidate AND p.active AND s.community
 AND date_part('year',age(s.birth_date)) BETWEEN (mine.details->>'min_age')::int AND (mine.details->>'max_age')::int
 AND date_part('year',age(me.birth_date)) BETWEEN (p.details->>'min_age')::int AND (p.details->>'max_age')::int
 ) ORDER BY l.created_at DESC LIMIT 100`, id)
}
func (s *Server) saveShortlist(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Owner     string `json:"owner"`
		Candidate string `json:"candidate"`
		Delegate  string `json:"delegate"`
		Note      string `json:"note"`
		Remove    bool   `json:"remove"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if in.Remove {
		s.memberExec(w, r, `DELETE FROM matrimony_shortlist WHERE owner=$1 AND delegate=$2 AND candidate=$3 AND ($4=owner OR $4=delegate)`, in.Owner, in.Delegate, in.Candidate, id)
		return
	}
	if len(in.Note) > 500 {
		problem(w, 400, fmt.Errorf("keep the shortlist note under 500 characters"))
		return
	}
	s.memberExec(w, r, `INSERT INTO matrimony_shortlist(owner,delegate,candidate,note)
 SELECT $1,$2,p.account_id,$4 FROM matrimony_profiles p JOIN member_settings c ON c.account_id=p.account_id JOIN matrimony_profiles mine ON mine.account_id=$1 JOIN member_settings me ON me.account_id=$1
 WHERE p.account_id=$3 AND p.active AND c.community AND p.account_id NOT IN($1,$2) AND delegate_allowed($1,$2) AND NOT member_blocked($3,$1) AND NOT member_blocked($3,$2)
 AND date_part('year',age(c.birth_date)) BETWEEN (mine.details->>'min_age')::int AND (mine.details->>'max_age')::int
 AND date_part('year',age(me.birth_date)) BETWEEN (p.details->>'min_age')::int AND (p.details->>'max_age')::int
 ON CONFLICT(owner,delegate,candidate) DO UPDATE SET note=EXCLUDED.note,created_at=now()`, in.Owner, id, in.Candidate, in.Note)
}
