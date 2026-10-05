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
	// Structured biodata (v2). Enumerated values are validated against
	// matrimonyEnums; an empty string means "not shared".
	Region             string `json:"region,omitempty"`
	Religion           string `json:"religion,omitempty"`
	Community          string `json:"community,omitempty"`
	MotherTongue       string `json:"mother_tongue,omitempty"`
	Diet               string `json:"diet,omitempty"`
	HeightCm           int    `json:"height_cm,omitempty"`
	MaritalStatus      string `json:"marital_status,omitempty"`
	EducationLevel     string `json:"education_level,omitempty"`
	OccupationCategory string `json:"occupation_category,omitempty"`
	IncomeBand         string `json:"income_band,omitempty"`
	FamilyType         string `json:"family_type,omitempty"`
}

// matrimonyEnums lists the allowed values of each structured field.
var matrimonyEnums = map[string][]string{
	"religion":            {"hindu", "muslim", "christian", "sikh", "jain", "buddhist", "parsi", "jewish", "spiritual", "none", "other", "prefer_not"},
	"mother_tongue":       {"hindi", "marathi", "kannada", "tamil", "telugu", "malayalam", "gujarati", "bengali", "punjabi", "odia", "urdu", "konkani", "tulu", "assamese", "english", "other"},
	"diet":                {"vegetarian", "eggetarian", "non_vegetarian", "vegan", "jain", "other"},
	"marital_status":      {"never_married", "divorced", "widowed", "separated", "awaiting_divorce"},
	"education_level":     {"high_school", "diploma", "bachelors", "masters", "doctorate", "other"},
	"occupation_category": {"it_software", "engineering", "medicine", "business", "government", "education", "finance", "law", "arts_media", "defence", "other", "not_working"},
	"income_band":         {"under_3l", "3_6l", "6_10l", "10_20l", "20_35l", "35_50l", "over_50l", "prefer_not"},
	"family_type":         {"joint", "nuclear", "other"},
	"timeline":            {"within_6_months", "within_year", "one_to_two_years", "not_sure"},
	"relocation":          {"open", "within_country", "no", "discuss"},
	"children":            {"want", "dont_want", "open", "have_children"},
}

func (d matrimonyDetails) enumValues() map[string]string {
	return map[string]string{"religion": d.Religion, "mother_tongue": d.MotherTongue, "diet": d.Diet, "marital_status": d.MaritalStatus, "education_level": d.EducationLevel, "occupation_category": d.OccupationCategory, "income_band": d.IncomeBand, "family_type": d.FamilyType, "timeline": d.Timeline, "relocation": d.Relocation, "children": d.Children}
}

func validEnum(field, v string) bool {
	if v == "" {
		return true
	}
	for _, x := range matrimonyEnums[field] {
		if x == v {
			return true
		}
	}
	return false
}

func (s *Server) matrimonyRoutes() {
	s.delegateRoutes()
	s.biodataRoutes()
	s.matrimonyV2Routes()
	s.memberRoute("GET /api/matrimony/me", s.matrimonyMe)
	s.memberRoute("PUT /api/matrimony/me", s.saveMatrimony)
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
	s.memberRows(w, r, `SELECT p.active,p.details,p.hidden,p.horoscope_visible,p.verified_at IS NOT NULL verified,p.saved_search,
 (SELECT status FROM matrimony_verifications v WHERE v.account_id=p.account_id) verification,
 a.birth_place IS NOT NULL has_birth_place,a.email_alerts
 FROM matrimony_profiles p JOIN member_accounts a ON a.id=p.account_id WHERE p.account_id=$1`, id)
}
func validMatrimonyDetails(d matrimonyDetails) bool {
	if d.MinAge < 18 || d.MaxAge < d.MinAge || d.MaxAge > 100 || len(d.DisplayName) > 100 || len(d.Introduction) > 1000 || len(d.Values) > 500 || len(d.Hobbies) > 500 || len(d.FamilyAbout) > 1000 || len(d.City) > 100 || len(d.Occupation) > 200 || len(d.Education) > 200 || len(d.Languages) > 200 || len(d.Timeline) > 100 || len(d.Children) > 100 || len(d.Relocation) > 100 || len(d.Lifestyle) > 200 || len(d.SocialLinks) > 5 {
		return false
	}
	if d.ProfileKind != "" && d.ProfileKind != "bride" && d.ProfileKind != "groom" && d.ProfileKind != "person" {
		return false
	}
	if len(d.Region) > 100 || len(d.Community) > 100 || (d.HeightCm != 0 && (d.HeightCm < 120 || d.HeightCm > 230)) {
		return false
	}
	for field, v := range d.enumValues() {
		if !validEnum(field, v) {
			return false
		}
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
		Active    bool             `json:"active"`
		Consent   bool             `json:"consent"`
		Horoscope bool             `json:"horoscope"`
		Details   matrimonyDetails `json:"details"`
		PhotoIDs  []string         `json:"photo_ids"`
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
		_, err = tx.Exec(r.Context(), `INSERT INTO matrimony_profiles(account_id,active,details,horoscope_visible) VALUES($1,$2,$3,$4) ON CONFLICT(account_id) DO UPDATE SET active=$2,details=$3,horoscope_visible=$4,updated_at=now()`, id, in.Active, raw, in.Horoscope)
	}
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 500, fmt.Errorf("profile could not be saved"))
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

const dailyInterestLimit = 20

func (s *Server) matrimonyInterests(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT a.id,a.handle,COALESCE(p.details->>'display_name','') display_name,c.avatar,c.accent,i.sender=$1 outgoing,i.status,i.note,i.created_at,
 (SELECT max(m.created_at) FROM member_messages m WHERE (m.sender=$1 AND m.recipient=a.id) OR (m.sender=a.id AND m.recipient=$1)) last_message
 FROM matrimony_interests i JOIN member_accounts a ON a.id=CASE WHEN i.sender=$1 THEN i.recipient ELSE i.sender END
 LEFT JOIN matrimony_profiles p ON p.account_id=a.id LEFT JOIN member_settings c ON c.account_id=a.id
 WHERE (i.sender=$1 OR i.recipient=$1) AND NOT member_blocked($1,a.id) ORDER BY COALESCE((SELECT max(m.created_at) FROM member_messages m WHERE (m.sender=$1 AND m.recipient=a.id) OR (m.sender=a.id AND m.recipient=$1)),i.created_at) DESC`, id)
}
func (s *Server) matrimonyInterestAction(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Target string `json:"target"`
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	switch in.Action {
	case "send":
		if !s.adultCommunity(w, r, id) {
			return
		}
		in.Note = strings.TrimSpace(in.Note)
		if len([]rune(in.Note)) > 300 {
			problem(w, 400, fmt.Errorf("keep the note under 300 characters"))
			return
		}
		// A daily budget limits mass messaging; real interest is individual.
		var sent int
		if s.membersDB().QueryRow(r.Context(), `SELECT count(*) FROM matrimony_interests WHERE sender=$1 AND created_at>now()-interval '1 day'`, id).Scan(&sent) != nil || sent >= dailyInterestLimit {
			problem(w, 429, fmt.Errorf("you can send %d interests a day; please try again tomorrow", dailyInterestLimit))
			return
		}
		s.memberExec(w, r, `INSERT INTO matrimony_interests(sender,recipient,note) SELECT $1,$2,$3 WHERE matrimony_visible($2,$1) ON CONFLICT(sender,recipient) DO UPDATE SET status=CASE WHEN matrimony_interests.status='declined' THEN 'declined' ELSE matrimony_interests.status END`, id, in.Target, in.Note)
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
