package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/example/panchang/reading"
)

type matrimonyDetails struct {
	DisplayName  string   `json:"display_name"`
	ProfileKind  string   `json:"profile_kind"`
	Introduction string   `json:"introduction"`
	City         string   `json:"city"`
	Occupation   string   `json:"occupation"`
	Education    string   `json:"education"`
	Languages    string   `json:"languages"`
	Timeline     string   `json:"timeline"`
	Children     string   `json:"children"`
	Relocation   string   `json:"relocation"`
	Lifestyle    string   `json:"lifestyle"`
	Values       string   `json:"values"`
	Hobbies      string   `json:"hobbies"`
	FamilyAbout  string   `json:"family_about"`
	SocialLinks  []string `json:"social_links"`
	MinAge       int      `json:"min_age"`
	MaxAge       int      `json:"max_age"`
}

func (s *Server) matrimonyRoutes() {
	s.delegateRoutes()
	s.biodataRoutes()
	s.memberRoute("GET /api/matrimony/me", s.matrimonyMe)
	s.memberRoute("PUT /api/matrimony/me", s.saveMatrimony)
	s.memberRoute("GET /api/matrimony/discover", s.discoverMatrimony)
	s.memberRoute("POST /api/matrimony/explanation/{peer}", s.matrimonyExplanation)
	s.memberRoute("GET /api/matrimony/interests", s.matrimonyInterests)
	s.memberRoute("POST /api/matrimony/interests", s.matrimonyInterestAction)
	s.memberRoute("GET /api/matrimony/messages/{peer}", s.memberMessages)
	s.memberRoute("POST /api/matrimony/messages/{peer}", s.sendMemberMessage)
}

func (s *Server) matrimonyExplanation(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		UseAI bool `json:"use_ai"`
	}
	if !memberInput(w, r, &req) {
		return
	}
	peer := r.PathValue("peer")
	if peer == id {
		problem(w, 404, fmt.Errorf("profile unavailable"))
		return
	}
	var mineRaw, theirsRaw []byte
	var shared int
	err := s.membersDB().QueryRow(r.Context(), `SELECT m.details,p.details,
 cardinality(ARRAY(SELECT unnest(ms.interests) INTERSECT SELECT unnest(ps.interests)))
 FROM matrimony_profiles m JOIN matrimony_profiles p ON p.account_id=$2
 JOIN member_settings ms ON ms.account_id=m.account_id JOIN member_settings ps ON ps.account_id=p.account_id
 WHERE m.account_id=$1 AND matrimony_visible($2,$1)`, id, peer).Scan(&mineRaw, &theirsRaw, &shared)
	if err != nil {
		problem(w, 404, fmt.Errorf("profile unavailable"))
		return
	}
	var mine, theirs map[string]any
	if json.Unmarshal(mineRaw, &mine) != nil || json.Unmarshal(theirsRaw, &theirs) != nil {
		problem(w, 500, fmt.Errorf("comparison unavailable"))
		return
	}
	comparisons := map[string]string{}
	for _, key := range []string{"city", "timeline", "children", "relocation", "lifestyle", "values", "hobbies"} {
		a, _ := mine[key].(string)
		b, _ := theirs[key].(string)
		a = strings.ToLower(strings.Join(strings.Fields(a), " "))
		b = strings.ToLower(strings.Join(strings.Fields(b), " "))
		comparisons[key] = "missing"
		if a != "" && b != "" {
			comparisons[key] = "different"
			if a == b {
				comparisons[key] = "same"
			}
		}
	}
	if req.UseAI && !s.matchBudget.allow(id, time.Now()) {
		problem(w, 429, fmt.Errorf("please wait before requesting another AI explanation"))
		return
	}
	out := s.reading.ExplainMatch(r.Context(), reading.ProfileMatchExplanation(comparisons, shared), req.UseAI)
	// Recheck after a potentially slow provider call: a block, pause, age-filter
	// change or moderation action must revoke access before a result is returned.
	var visible bool
	if s.membersDB().QueryRow(r.Context(), `SELECT matrimony_visible($1,$2)`, peer, id).Scan(&visible) != nil || !visible {
		problem(w, 404, fmt.Errorf("profile unavailable"))
		return
	}
	writeJSON(w, 200, out)
}
func (s *Server) matrimonyMe(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT active,details,hidden FROM matrimony_profiles WHERE account_id=$1`, id)
}
func validMatrimonyDetails(d matrimonyDetails) bool {
	if d.MinAge < 18 || d.MaxAge < d.MinAge || d.MaxAge > 100 || len(d.DisplayName) > 100 || len(d.Introduction) > 1000 || len(d.Values) > 500 || len(d.Hobbies) > 500 || len(d.FamilyAbout) > 1000 || len(d.City) > 100 || len(d.Occupation) > 200 || len(d.Education) > 200 || len(d.Languages) > 200 || len(d.Timeline) > 100 || len(d.Children) > 100 || len(d.Relocation) > 100 || len(d.Lifestyle) > 200 || len(d.SocialLinks) > 5 {
		return false
	}
	if d.ProfileKind != "" && d.ProfileKind != "bride" && d.ProfileKind != "groom" && d.ProfileKind != "person" {
		return false
	}
	for _, link := range d.SocialLinks {
		u, err := url.Parse(link)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || len(link) > 300 {
			return false
		}
	}
	return true
}
func (s *Server) saveMatrimony(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Active   bool             `json:"active"`
		Consent  bool             `json:"consent"`
		Details  matrimonyDetails `json:"details"`
		PhotoIDs []string         `json:"photo_ids"`
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
	if !validMatrimonyDetails(d) || len(in.PhotoIDs) > 6 {
		problem(w, 400, fmt.Errorf("check profile lengths and age preferences (18–100)"))
		return
	}
	raw, _ := json.Marshal(d)
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("profile unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	if _, err = tx.Exec(r.Context(), `SELECT id FROM member_accounts WHERE id=$1 FOR UPDATE`, id); err != nil {
		problem(w, 500, fmt.Errorf("profile unavailable"))
		return
	}
	// Explicit selection prevents a newly uploaded photo being published implicitly.
	if in.PhotoIDs != nil {
		var count int
		err = tx.QueryRow(r.Context(), `SELECT count(*) FROM matrimony_photos WHERE id=ANY($1) AND owner=$2 AND draft_id IS NULL`, in.PhotoIDs, id).Scan(&count)
		if err != nil || count != len(in.PhotoIDs) {
			problem(w, 400, fmt.Errorf("select only your own profile photos"))
			return
		}
		_, err = tx.Exec(r.Context(), `UPDATE matrimony_photos SET published=(id=ANY($2)) WHERE owner=$1 AND draft_id IS NULL`, id, in.PhotoIDs)
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `INSERT INTO matrimony_profiles(account_id,active,details) VALUES($1,$2,$3) ON CONFLICT(account_id) DO UPDATE SET active=$2,details=$3,updated_at=now()`, id, in.Active, raw)
	}
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 500, fmt.Errorf("profile could not be saved"))
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) discoverMatrimony(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultCommunity(w, r, id) {
		return
	}
	s.matrimonyCandidates(w, r, id, id)
}
func (s *Server) matrimonyCandidates(w http.ResponseWriter, r *http.Request, id, viewer string) {
	s.memberRows(w, r, `SELECT a.id,a.handle,p.details,c.avatar,c.accent,c.interests,date_part('year',age(c.birth_date))::int age,
 COALESCE((SELECT json_agg(json_build_object('id',ph.id,'alt',ph.alt) ORDER BY ph.created_at,ph.id) FROM matrimony_photos ph WHERE ph.owner=a.id AND ph.draft_id IS NULL AND ph.published),'[]') photos,
 ARRAY(SELECT unnest(c.interests) INTERSECT SELECT unnest(me.interests)) shared_interests,
 (p.details->>'city'=mine.details->>'city' AND p.details->>'city'<>'') same_city,
 (p.details->>'timeline'=mine.details->>'timeline' AND p.details->>'timeline'<>'') same_timeline,
 (p.details->>'relocation'<>mine.details->>'relocation' AND p.details->>'relocation'<>'' AND mine.details->>'relocation'<>'') discuss_relocation
 FROM matrimony_profiles p JOIN member_accounts a ON a.id=p.account_id JOIN member_settings c ON c.account_id=a.id
 JOIN member_settings me ON me.account_id=$1 JOIN matrimony_profiles mine ON mine.account_id=$1
 WHERE matrimony_visible(a.id,$1) AND a.id<>$2 AND NOT member_blocked(a.id,$2)
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
		s.memberExec(w, r, `INSERT INTO matrimony_interests(sender,recipient) SELECT $1,$2 WHERE matrimony_visible($2,$1) ON CONFLICT(sender,recipient) DO UPDATE SET status=CASE WHEN matrimony_interests.status='declined' THEN 'declined' ELSE matrimony_interests.status END`, id, in.Target)
	case "accept":
		s.memberExec(w, r, `UPDATE matrimony_interests SET status='accepted' WHERE recipient=$1 AND sender=$2 AND status='pending' AND matrimony_visible($2,$1)`, id, in.Target)
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
