package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type matrimonyDetails struct {
	Introduction string `json:"introduction"`
	City         string `json:"city"`
	Occupation   string `json:"occupation"`
	Education    string `json:"education"`
	Languages    string `json:"languages"`
	Timeline     string `json:"timeline"`
	Children     string `json:"children"`
	Relocation   string `json:"relocation"`
	Lifestyle    string `json:"lifestyle"`
	Values       string `json:"values"`
	MinAge       int    `json:"min_age"`
	MaxAge       int    `json:"max_age"`
}

func (s *Server) matrimonyRoutes() {
	s.delegateRoutes()
	s.memberRoute("GET /api/matrimony/me", s.matrimonyMe)
	s.memberRoute("PUT /api/matrimony/me", s.saveMatrimony)
	s.memberRoute("GET /api/matrimony/discover", s.discoverMatrimony)
	s.memberRoute("GET /api/matrimony/interests", s.matrimonyInterests)
	s.memberRoute("POST /api/matrimony/interests", s.matrimonyInterestAction)
	s.memberRoute("GET /api/matrimony/messages/{peer}", s.memberMessages)
	s.memberRoute("POST /api/matrimony/messages/{peer}", s.sendMemberMessage)
}
func (s *Server) matrimonyMe(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT active,details FROM matrimony_profiles WHERE account_id=$1`, id)
}
func (s *Server) saveMatrimony(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Active  bool             `json:"active"`
		Consent bool             `json:"consent"`
		Details matrimonyDetails `json:"details"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if in.Active && (!in.Consent || !s.adultCommunity(w, r, id)) {
		if !in.Consent {
			problem(w, 400, fmt.Errorf("explicit publishing consent required"))
		}
		return
	}
	d := in.Details
	if d.MinAge < 18 || d.MaxAge < d.MinAge || d.MaxAge > 100 || len(d.Introduction) > 1000 || len(d.Values) > 500 || len(d.City) > 100 || len(d.Occupation) > 200 || len(d.Education) > 200 || len(d.Languages) > 200 || len(d.Timeline) > 100 || len(d.Children) > 100 || len(d.Relocation) > 100 || len(d.Lifestyle) > 200 {
		problem(w, 400, fmt.Errorf("check profile lengths and age preferences (18–100)"))
		return
	}
	raw, _ := json.Marshal(d)
	s.memberExec(w, r, `INSERT INTO matrimony_profiles(account_id,active,details) VALUES($1,$2,$3) ON CONFLICT(account_id) DO UPDATE SET active=$2,details=$3,updated_at=now()`, id, in.Active, raw)
}
func (s *Server) discoverMatrimony(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultCommunity(w, r, id) {
		return
	}
	s.matrimonyCandidates(w, r, id, id)
}
func (s *Server) matrimonyCandidates(w http.ResponseWriter, r *http.Request, id, viewer string) {
	s.memberRows(w, r, `SELECT a.id,a.handle,p.details,c.avatar,c.accent,c.interests,date_part('year',age(c.birth_date))::int age,
 ARRAY(SELECT unnest(c.interests) INTERSECT SELECT unnest(me.interests)) shared_interests,
 (p.details->>'city'=mine.details->>'city' AND p.details->>'city'<>'') same_city,
 (p.details->>'timeline'=mine.details->>'timeline' AND p.details->>'timeline'<>'') same_timeline,
 (p.details->>'relocation'<>mine.details->>'relocation' AND p.details->>'relocation'<>'' AND mine.details->>'relocation'<>'') discuss_relocation
 FROM matrimony_profiles p JOIN member_accounts a ON a.id=p.account_id JOIN member_settings c ON c.account_id=a.id
 JOIN member_settings me ON me.account_id=$1 JOIN matrimony_profiles mine ON mine.account_id=$1
 WHERE p.active AND mine.active AND c.community AND a.id<>$1 AND a.id<>$2 AND NOT member_blocked(a.id,$1) AND NOT member_blocked(a.id,$2)
 AND ($1=$2 OR delegate_allowed($1,$2))
 AND date_part('year',age(c.birth_date)) BETWEEN (mine.details->>'min_age')::int AND (mine.details->>'max_age')::int
 AND date_part('year',age(me.birth_date)) BETWEEN (p.details->>'min_age')::int AND (p.details->>'max_age')::int
 ORDER BY p.updated_at DESC LIMIT 50`, id, viewer)
}
func (s *Server) matrimonyInterests(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT a.id,a.handle,i.sender=$1 outgoing,i.status FROM matrimony_interests i JOIN member_accounts a ON a.id=CASE WHEN i.sender=$1 THEN i.recipient ELSE i.sender END WHERE (i.sender=$1 OR i.recipient=$1) AND NOT member_blocked($1,a.id) ORDER BY i.created_at DESC`, id)
}
func (s *Server) matrimonyInterestAction(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Target string `json:"target"`
		Action string `json:"action"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	switch in.Action {
	case "send":
		if !s.adultCommunity(w, r, id) {
			return
		}
		s.memberExec(w, r, `INSERT INTO matrimony_interests(sender,recipient) SELECT $1,p.account_id FROM matrimony_profiles p JOIN member_settings c ON c.account_id=p.account_id JOIN matrimony_profiles m ON m.account_id=$1 JOIN member_settings me ON me.account_id=$1 WHERE p.account_id=$2 AND p.active AND m.active AND c.community AND p.account_id<>$1 AND NOT member_blocked($1,$2) AND date_part('year',age(c.birth_date)) BETWEEN (m.details->>'min_age')::int AND (m.details->>'max_age')::int AND date_part('year',age(me.birth_date)) BETWEEN (p.details->>'min_age')::int AND (p.details->>'max_age')::int ON CONFLICT(sender,recipient) DO UPDATE SET status=CASE WHEN matrimony_interests.status='declined' THEN 'declined' ELSE matrimony_interests.status END`, id, in.Target)
	case "accept":
		s.memberExec(w, r, `UPDATE matrimony_interests SET status='accepted' WHERE recipient=$1 AND sender=$2 AND status='pending' AND NOT member_blocked($1,$2) AND EXISTS(SELECT 1 FROM matrimony_profiles WHERE account_id=$1 AND active) AND EXISTS(SELECT 1 FROM matrimony_profiles WHERE account_id=$2 AND active)`, id, in.Target)
	case "decline":
		s.memberExec(w, r, `UPDATE matrimony_interests SET status='declined' WHERE recipient=$1 AND sender=$2`, id, in.Target)
	case "unmatch":
		s.memberExec(w, r, `UPDATE matrimony_interests SET status='declined' WHERE (sender=$1 AND recipient=$2) OR (sender=$2 AND recipient=$1)`, id, in.Target)
	default:
		problem(w, 400, fmt.Errorf("invalid interest action"))
	}
}
func (s *Server) canMessage(r *http.Request, id, peer string) bool {
	var ok bool
	err := s.membersDB().QueryRow(r.Context(), `SELECT NOT member_blocked($1,$2) AND (SELECT count(*)=2 FROM member_settings WHERE account_id IN ($1,$2) AND community AND birth_date <= CURRENT_DATE - INTERVAL '18 years') AND EXISTS(SELECT 1 FROM matrimony_interests WHERE status='accepted' AND ((sender=$1 AND recipient=$2) OR (sender=$2 AND recipient=$1)))`, id, peer).Scan(&ok)
	return err == nil && ok
}
func (s *Server) memberMessages(w http.ResponseWriter, r *http.Request, id string) {
	peer := r.PathValue("peer")
	if !s.canMessage(r, id, peer) {
		problem(w, 403, fmt.Errorf("mutual acceptance required"))
		return
	}
	s.memberRows(w, r, `SELECT * FROM (SELECT id,body,created_at,sender=$1 mine FROM member_messages WHERE (sender=$1 AND recipient=$2) OR (sender=$2 AND recipient=$1) ORDER BY id DESC LIMIT 100) recent ORDER BY id`, id, peer)
}
func (s *Server) sendMemberMessage(w http.ResponseWriter, r *http.Request, id string) {
	peer := r.PathValue("peer")
	if !s.canMessage(r, id, peer) {
		problem(w, 403, fmt.Errorf("mutual acceptance required"))
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Body) == "" || len(in.Body) > 2000 {
		problem(w, 400, fmt.Errorf("message must be 1–2000 characters"))
		return
	}
	s.memberExec(w, r, `INSERT INTO member_messages(sender,recipient,body) VALUES($1,$2,$3)`, id, peer, in.Body)
}
