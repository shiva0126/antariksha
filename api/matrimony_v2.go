package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/example/panchang/divination"
	"github.com/example/panchang/engine"
	"github.com/example/panchang/reading"
)

// Matrimony v2: discovery with filters and reasons, horoscope matching between
// members who both opted in, shortlists, interest notes, contact sharing,
// selfie verification and the full comparison with Ask Astrisk. Birth date,
// time and place never leave the server; only derived results are returned.

func (s *Server) matrimonyV2Routes() {
	s.memberRoute("PUT /api/me/birth-place", s.saveBirthPlace)
	s.memberRoute("PUT /api/me/alerts", s.saveAlertPreference)
	s.memberRoute("GET /api/matrimony/discover", s.discoverV2)
	s.memberRoute("PUT /api/matrimony/search", s.saveMatrimonySearch)
	s.memberRoute("GET /api/matrimony/saved", s.matrimonySavedList)
	s.memberRoute("POST /api/matrimony/saved", s.matrimonySave)
	s.memberRoute("GET /api/matrimony/contact/{peer}", s.matrimonyContacts)
	s.memberRoute("PUT /api/matrimony/contact/{peer}", s.shareMatrimonyContact)
	s.memberRoute("DELETE /api/matrimony/contact/{peer}", s.unshareMatrimonyContact)
	s.memberRoute("GET /api/matrimony/compare/{peer}", s.matrimonyCompare)
	s.memberRoute("POST /api/matrimony/compare/{peer}/chat", s.matrimonyCompareChat)
	s.memberRoute("GET /api/matrimony/verification", s.myVerification)
	s.memberRoute("POST /api/matrimony/verification", s.submitVerification)
	s.memberRoute("GET /api/matrimony/verifications", s.verificationQueue)
	s.memberRoute("GET /api/matrimony/verifications/{account}/photo", s.verificationPhoto)
	s.memberRoute("POST /api/matrimony/verifications/{account}", s.reviewVerification)
}

// ---- birth place and alert preferences -------------------------------------

type birthPlace struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	TZ   string  `json:"tz"`
}

func (p birthPlace) valid() bool {
	if len(p.Name) == 0 || len(p.Name) > 160 || math.IsNaN(p.Lat) || math.IsNaN(p.Lon) || p.Lat < -90 || p.Lat > 90 || p.Lon < -180 || p.Lon > 180 {
		return false
	}
	_, err := time.LoadLocation(p.TZ)
	return err == nil && p.TZ != "" && p.TZ != "Local"
}

func (s *Server) saveBirthPlace(w http.ResponseWriter, r *http.Request, id string) {
	var p birthPlace
	if !memberInput(w, r, &p) {
		return
	}
	if !p.valid() {
		problem(w, 400, fmt.Errorf("choose a birthplace from the list"))
		return
	}
	raw, _ := json.Marshal(p)
	s.memberExec(w, r, `UPDATE member_accounts SET birth_place=$2 WHERE id=$1`, id, raw)
	s.forgetChart(id)
}

func (s *Server) saveAlertPreference(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Email bool `json:"email"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	s.memberExec(w, r, `UPDATE member_accounts SET email_alerts=$2 WHERE id=$1`, id, in.Email)
}

// ---- birth charts of members ---------------------------------------------

// India's centre, used only for the Moon when no birthplace is saved: the Moon
// needs the moment of birth, not the place; lagna-based rules are skipped.
const fallbackLat, fallbackLon = 22.97, 78.66

type memberBirth struct {
	Date, Time string
	Place      *birthPlace
}

type memberChart struct {
	Chart    engine.Chart
	HasPlace bool
}

type chartCache struct {
	mu sync.Mutex
	m  map[string]memberChart
}

var memberCharts = chartCache{m: map[string]memberChart{}}

func (s *Server) forgetChart(id string) {
	memberCharts.mu.Lock()
	defer memberCharts.mu.Unlock()
	for k := range memberCharts.m {
		if strings.HasPrefix(k, id+"|") {
			delete(memberCharts.m, k)
		}
	}
}

func (s *Server) chartOf(id string, b memberBirth) (memberChart, bool) {
	if b.Date == "" || b.Time == "" || s.engine == nil {
		return memberChart{}, false
	}
	in := engine.ChartInput{Date: b.Date, Time: b.Time, Lat: fallbackLat, Lon: fallbackLon, TZ: "Asia/Kolkata"}
	if b.Place != nil {
		in.Lat, in.Lon, in.TZ = b.Place.Lat, b.Place.Lon, b.Place.TZ
	}
	key := fmt.Sprintf("%s|%s|%s|%.4f|%.4f|%s", id, in.Date, in.Time, in.Lat, in.Lon, in.TZ)
	memberCharts.mu.Lock()
	if c, ok := memberCharts.m[key]; ok {
		memberCharts.mu.Unlock()
		return c, true
	}
	memberCharts.mu.Unlock()
	c, err := s.engine.BirthChart(in)
	if err != nil {
		return memberChart{}, false
	}
	mc := memberChart{Chart: c, HasPlace: b.Place != nil}
	memberCharts.mu.Lock()
	if len(memberCharts.m) > 4096 {
		memberCharts.m = map[string]memberChart{}
	}
	memberCharts.m[key] = mc
	memberCharts.mu.Unlock()
	return mc, true
}

func (s *Server) birthOf(ctx context.Context, id string) (memberBirth, error) {
	var b memberBirth
	var place []byte
	err := s.membersDB().QueryRow(ctx, `SELECT COALESCE(to_char(birth_date,'YYYY-MM-DD'),''),COALESCE(to_char(birth_time,'HH24:MI'),''),birth_place FROM member_accounts WHERE id=$1`, id).Scan(&b.Date, &b.Time, &place)
	if err == nil && len(place) > 0 {
		var p birthPlace
		if json.Unmarshal(place, &p) == nil && p.valid() {
			b.Place = &p
		}
	}
	return b, err
}

// ---- horoscope comparison between two members -------------------------------

type horoscopeMatch struct {
	Guna        float64  `json:"guna"`
	Max         float64  `json:"max"`
	Kootas      []kScore `json:"kootas"`
	Doshas      []string `json:"doshas"`
	Exceptions  []string `json:"exceptions"`
	TheirMoon   string   `json:"their_moon"`
	YourMoon    string   `json:"your_moon"`
	TheirMangal string   `json:"their_mangal"` // yes | no | unknown (no birthplace)
	YourMangal  string   `json:"your_mangal"`
	NumberRel   string   `json:"number_relation"`
	TheirNumber int      `json:"their_root_number"`
	TheirGraha  string   `json:"their_root_graha"`
	YourNumber  int      `json:"your_root_number"`
	YourGraha   string   `json:"your_root_graha"`
	ViewerIsBoy bool     `json:"viewer_is_groom_side"`
	engineMatch engine.Match
	boyC, girlC memberChart
	boyB, girlB memberBirth
}
type kScore struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
	Max   float64 `json:"max"`
}

func mangalWord(mc memberChart) string {
	if !mc.HasPlace {
		return "unknown"
	}
	if on, _ := engine.MangalDosha(mc.Chart); on {
		return "yes"
	}
	return "no"
}

// viewerIsBoy orients the Ashtakoota tables, which are defined for a groom
// and a bride: profile kinds decide, otherwise the viewer is the first chart.
func viewerIsBoy(viewerKind, theirKind string) bool {
	return !(viewerKind == "bride" || theirKind == "groom")
}

func (s *Server) compareMembers(viewer, them string, vb, tb memberBirth, vKind, tKind string) (*horoscopeMatch, bool) {
	vc, ok1 := s.chartOf(viewer, vb)
	tc, ok2 := s.chartOf(them, tb)
	if !ok1 || !ok2 {
		return nil, false
	}
	boy, girl, bb, gb := vc, tc, vb, tb
	isBoy := viewerIsBoy(vKind, tKind)
	if !isBoy {
		boy, girl, bb, gb = tc, vc, tb, vb
	}
	m, err := engine.MatchCharts(boy.Chart, girl.Chart)
	if err != nil {
		return nil, false
	}
	h := &horoscopeMatch{Guna: m.Total, Max: m.Max, Doshas: m.Doshas, Exceptions: m.Exceptions, TheirMangal: mangalWord(tc), YourMangal: mangalWord(vc), ViewerIsBoy: isBoy, engineMatch: m, boyC: boy, girlC: girl, boyB: bb, girlB: gb}
	if h.Doshas == nil {
		h.Doshas = []string{}
	}
	if h.Exceptions == nil {
		h.Exceptions = []string{}
	}
	for _, k := range m.Kootas {
		h.Kootas = append(h.Kootas, kScore{k.Name, k.Score, k.Max})
	}
	h.TheirMoon, h.YourMoon = m.GirlMoon, m.BoyMoon
	if !isBoy {
		h.TheirMoon, h.YourMoon = m.BoyMoon, m.GirlMoon
	}
	vn, e1 := divination.ReadNumerology(vb.Date, "", time.Now().Year())
	tn, e2 := divination.ReadNumerology(tb.Date, "", time.Now().Year())
	if e1 == nil && e2 == nil {
		h.YourNumber, h.YourGraha, h.TheirNumber, h.TheirGraha = vn.Mulank, vn.MulankGraha, tn.Mulank, tn.MulankGraha
		h.NumberRel, _ = reading.NumberGrahaRelation(vn.MulankGraha, tn.MulankGraha)
	}
	return h, true
}

// ---- discovery ---------------------------------------------------------------

type candidate struct {
	ID         string            `json:"id"`
	Handle     string            `json:"handle"`
	Details    matrimonyDetails  `json:"details"`
	Avatar     string            `json:"avatar"`
	Accent     string            `json:"accent"`
	Age        int               `json:"age"`
	Photos     []json.RawMessage `json:"photos"`
	Shared     []string          `json:"shared_interests"`
	Verified   bool              `json:"verified"`
	Saved      string            `json:"saved"`    // "", saved, skipped
	Interest   string            `json:"interest"` // "", sent, received, accepted, declined
	Horoscope  *horoscopeMatch   `json:"horoscope"`
	HoroReason string            `json:"horoscope_note,omitempty"`
	Reasons    []string          `json:"reasons"`
	Updated    time.Time         `json:"updated_at"`
	score      float64
}

type discoverFilters struct {
	Kind, City, Religion, MotherTongue, Diet, MaritalStatus, EducationLevel, Mangal, Sort string
	AgeMin, AgeMax, Page                                                                  int
	MinGuna                                                                               float64
	Verified, Saved, ShowSkipped                                                          bool
}

func parseDiscoverFilters(q map[string][]string) discoverFilters {
	get := func(k string) string {
		if v := q[k]; len(v) > 0 {
			return strings.TrimSpace(v[0])
		}
		return ""
	}
	num := func(k string) int { n, _ := strconv.Atoi(get(k)); return n }
	f := discoverFilters{Kind: get("kind"), City: strings.ToLower(get("city")), Religion: get("religion"), MotherTongue: get("mother_tongue"), Diet: get("diet"), MaritalStatus: get("marital_status"), EducationLevel: get("education_level"), Mangal: get("mangal"), Sort: get("sort"), AgeMin: num("age_min"), AgeMax: num("age_max"), Page: num("page"), Verified: get("verified") == "1", Saved: get("saved") == "1", ShowSkipped: get("skipped") == "1"}
	f.MinGuna, _ = strconv.ParseFloat(get("min_guna"), 64)
	if f.Page < 0 {
		f.Page = 0
	}
	return f
}

var enumLabel = map[string]string{"vegetarian": "vegetarian", "eggetarian": "eggetarian", "non_vegetarian": "non-vegetarian", "vegan": "vegan", "jain": "Jain diet", "within_6_months": "within 6 months", "within_year": "within a year", "one_to_two_years": "in 1–2 years"}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (c *candidate) addReasons(mine matrimonyDetails, h *horoscopeMatch) {
	add := func(s string, w float64) { c.Reasons = append(c.Reasons, s); c.score += w }
	if h != nil {
		add(fmt.Sprintf("%g/36 gunas", h.Guna), h.Guna/36*3)
	}
	if c.Details.City != "" && strings.EqualFold(c.Details.City, mine.City) {
		add("Same city", 1)
	}
	if c.Details.MotherTongue != "" && c.Details.MotherTongue == mine.MotherTongue && c.Details.MotherTongue != "other" {
		add("Both speak "+title(c.Details.MotherTongue), 1)
	}
	if r := c.Details.Religion; r != "" && r == mine.Religion && r != "prefer_not" && r != "other" && r != "none" {
		add("Same religion", 1)
	}
	if d := c.Details.Diet; d != "" && d == mine.Diet && d != "other" {
		add("Both "+enumLabel[d], 0.5)
	}
	if t := c.Details.Timeline; t != "" && t == mine.Timeline && t != "not_sure" {
		add("Both hope to marry "+enumLabel[t], 0.5)
	}
	if len(c.Shared) > 0 {
		add("Shared interests: "+strings.Join(c.Shared, ", "), 0.5*float64(len(c.Shared)))
	}
	if c.Verified {
		add("Photo verified", 0.5)
	}
	if c.Details.Relocation != "" && mine.Relocation != "" && c.Details.Relocation != mine.Relocation {
		c.Reasons = append(c.Reasons, "Discuss: relocation preferences differ")
	}
}

func (s *Server) discoverV2(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultCommunity(w, r, id) {
		return
	}
	s.discoverFor(w, r, id, id)
}

// discoverFor lists eligible profiles for owner, as seen by viewer (the owner,
// or a family helper with an accepted grant).
func (s *Server) discoverFor(w http.ResponseWriter, r *http.Request, owner, viewer string) {
	f := parseDiscoverFilters(r.URL.Query())
	ctx := r.Context()
	var mineRaw []byte
	var myHoro bool
	if err := s.membersDB().QueryRow(ctx, `SELECT details,horoscope_visible FROM matrimony_profiles WHERE account_id=$1`, owner).Scan(&mineRaw, &myHoro); err != nil {
		writeJSON(w, 200, map[string]any{"items": []candidate{}, "total": 0, "page": 0, "pages": 0, "needs_profile": true})
		return
	}
	var mine matrimonyDetails
	_ = json.Unmarshal(mineRaw, &mine)
	ownerBirth, _ := s.birthOf(ctx, owner)
	rows, err := s.membersDB().Query(ctx, `SELECT a.id,a.handle,p.details,c.avatar,c.accent,date_part('year',age(c.birth_date))::int,
 COALESCE((SELECT json_agg(json_build_object('id',ph.id,'alt',ph.alt) ORDER BY ph.created_at,ph.id) FROM matrimony_photos ph WHERE ph.owner=a.id AND ph.draft_id IS NULL AND ph.published),'[]'),
 ARRAY(SELECT unnest(c.interests) INTERSECT SELECT unnest(me.interests)),
 p.verified_at IS NOT NULL, p.horoscope_visible, p.updated_at,
 COALESCE((SELECT kind FROM matrimony_saved sv WHERE sv.owner=$1 AND sv.target=a.id),''),
 COALESCE((SELECT CASE WHEN i.status='accepted' THEN 'accepted' WHEN i.status='declined' THEN 'declined' WHEN i.sender=$1 THEN 'sent' ELSE 'received' END FROM matrimony_interests i WHERE (i.sender=$1 AND i.recipient=a.id) OR (i.sender=a.id AND i.recipient=$1) ORDER BY i.created_at DESC LIMIT 1),''),
 COALESCE(to_char(a.birth_date,'YYYY-MM-DD'),''),COALESCE(to_char(a.birth_time,'HH24:MI'),''),a.birth_place
 FROM matrimony_profiles p JOIN member_accounts a ON a.id=p.account_id JOIN member_settings c ON c.account_id=a.id
 JOIN member_settings me ON me.account_id=$1 JOIN matrimony_profiles mine ON mine.account_id=$1
 WHERE matrimony_visible(a.id,$1) AND a.id<>$2 AND NOT member_blocked(a.id,$2)
 AND ($1=$2 OR delegate_allowed($1,$2))
 ORDER BY p.updated_at DESC LIMIT 500`, owner, viewer)
	if err != nil {
		s.logger.Error("discover", "error", err)
		problem(w, 500, fmt.Errorf("discovery unavailable"))
		return
	}
	var all []candidate
	for rows.Next() {
		var c candidate
		var details, photos, place []byte
		var theirHoro bool
		var b memberBirth
		if err = rows.Scan(&c.ID, &c.Handle, &details, &c.Avatar, &c.Accent, &c.Age, &photos, &c.Shared, &c.Verified, &theirHoro, &c.Updated, &c.Saved, &c.Interest, &b.Date, &b.Time, &place); err != nil {
			rows.Close()
			problem(w, 500, fmt.Errorf("discovery unavailable"))
			return
		}
		_ = json.Unmarshal(details, &c.Details)
		_ = json.Unmarshal(photos, &c.Photos)
		if len(place) > 0 {
			var p birthPlace
			if json.Unmarshal(place, &p) == nil && p.valid() {
				b.Place = &p
			}
		}
		if c.Shared == nil {
			c.Shared = []string{}
		}
		switch {
		case !myHoro:
			c.HoroReason = "Turn on horoscope matching in your profile to see guna scores."
		case !theirHoro:
			c.HoroReason = "This member has not shared horoscope matching."
		default:
			if h, ok := s.compareMembers(owner, c.ID, ownerBirth, b, mine.ProfileKind, c.Details.ProfileKind); ok {
				c.Horoscope = h
			} else {
				c.HoroReason = "Birth details are incomplete."
			}
		}
		c.addReasons(mine, c.Horoscope)
		all = append(all, c)
	}
	rows.Close()
	items := []candidate{}
	for _, c := range all {
		d := c.Details
		switch {
		case f.Saved && c.Saved != "saved",
			!f.ShowSkipped && !f.Saved && c.Saved == "skipped",
			f.Kind != "" && d.ProfileKind != f.Kind,
			f.City != "" && strings.ToLower(d.City) != f.City,
			f.Religion != "" && d.Religion != f.Religion,
			f.MotherTongue != "" && d.MotherTongue != f.MotherTongue,
			f.Diet != "" && d.Diet != f.Diet,
			f.MaritalStatus != "" && d.MaritalStatus != f.MaritalStatus,
			f.EducationLevel != "" && d.EducationLevel != f.EducationLevel,
			f.AgeMin > 0 && c.Age < f.AgeMin,
			f.AgeMax > 0 && c.Age > f.AgeMax,
			f.Verified && !c.Verified,
			f.MinGuna > 0 && (c.Horoscope == nil || c.Horoscope.Guna < f.MinGuna),
			f.Mangal == "no" && (c.Horoscope == nil || c.Horoscope.TheirMangal != "no"),
			f.Mangal == "yes" && (c.Horoscope == nil || c.Horoscope.TheirMangal != "yes"):
			continue
		}
		items = append(items, c)
	}
	switch f.Sort {
	case "recent":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Updated.After(items[j].Updated) })
	case "guna":
		sort.SliceStable(items, func(i, j int) bool {
			gi, gj := -1.0, -1.0
			if items[i].Horoscope != nil {
				gi = items[i].Horoscope.Guna
			}
			if items[j].Horoscope != nil {
				gj = items[j].Horoscope.Guna
			}
			return gi > gj
		})
	default:
		sort.SliceStable(items, func(i, j int) bool { return items[i].score > items[j].score })
	}
	const pageSize = 12
	total := len(items)
	pages := (total + pageSize - 1) / pageSize
	start := min(f.Page*pageSize, total)
	writeJSON(w, 200, map[string]any{"items": items[start:min(start+pageSize, total)], "total": total, "page": f.Page, "pages": pages, "horoscope_enabled": myHoro})
}

func (s *Server) saveMatrimonySearch(w http.ResponseWriter, r *http.Request, id string) {
	var in map[string]string
	if !memberInput(w, r, &in) {
		return
	}
	allowed := map[string]bool{"kind": true, "city": true, "religion": true, "mother_tongue": true, "diet": true, "marital_status": true, "education_level": true, "mangal": true, "sort": true, "age_min": true, "age_max": true, "min_guna": true, "verified": true}
	for k, v := range in {
		if !allowed[k] || len(v) > 100 {
			problem(w, 400, fmt.Errorf("invalid search field %q", k))
			return
		}
	}
	raw, _ := json.Marshal(in)
	s.memberExec(w, r, `UPDATE matrimony_profiles SET saved_search=$2 WHERE account_id=$1`, id, raw)
}

// ---- shortlist ("saved") and "not now" ------------------------------------

func (s *Server) matrimonySave(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Target string `json:"target"`
		Kind   string `json:"kind"` // saved | skipped | clear
	}
	if !memberInput(w, r, &in) {
		return
	}
	switch in.Kind {
	case "saved", "skipped":
		s.memberExec(w, r, `INSERT INTO matrimony_saved(owner,target,kind) SELECT $1,$2,$3 WHERE matrimony_visible($2,$1) ON CONFLICT(owner,target) DO UPDATE SET kind=$3,created_at=now()`, id, in.Target, in.Kind)
	case "clear":
		s.memberExec(w, r, `DELETE FROM matrimony_saved WHERE owner=$1 AND target=$2`, id, in.Target)
	default:
		problem(w, 400, fmt.Errorf("choose saved, skipped or clear"))
	}
}

func (s *Server) matrimonySavedList(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT sv.target id,a.handle,p.details->>'display_name' display_name,sv.kind,sv.created_at FROM matrimony_saved sv JOIN member_accounts a ON a.id=sv.target JOIN matrimony_profiles p ON p.account_id=sv.target WHERE sv.owner=$1 AND matrimony_visible(sv.target,$1) ORDER BY sv.created_at DESC`, id)
}

// ---- contact sharing --------------------------------------------------------

func (s *Server) matrimonyContacts(w http.ResponseWriter, r *http.Request, id string) {
	peer := r.PathValue("peer")
	if !s.canMessage(r, id, peer) {
		problem(w, 403, fmt.Errorf("mutual acceptance required"))
		return
	}
	var mine, theirs string
	_ = s.membersDB().QueryRow(r.Context(), `SELECT COALESCE((SELECT contact FROM matrimony_contacts WHERE owner=$1 AND peer=$2),''),COALESCE((SELECT contact FROM matrimony_contacts WHERE owner=$2 AND peer=$1),'')`, id, peer).Scan(&mine, &theirs)
	writeJSON(w, 200, map[string]string{"mine": mine, "theirs": theirs})
}

func (s *Server) shareMatrimonyContact(w http.ResponseWriter, r *http.Request, id string) {
	peer := r.PathValue("peer")
	if !s.canMessage(r, id, peer) {
		problem(w, 403, fmt.Errorf("mutual acceptance required"))
		return
	}
	var in struct {
		Contact string `json:"contact"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	in.Contact = strings.TrimSpace(in.Contact)
	if len(in.Contact) < 3 || len(in.Contact) > 200 || strings.ContainsAny(in.Contact, "<>") {
		problem(w, 400, fmt.Errorf("enter a phone number or email (3–200 characters)"))
		return
	}
	s.memberExec(w, r, `INSERT INTO matrimony_contacts(owner,peer,contact) VALUES($1,$2,$3) ON CONFLICT(owner,peer) DO UPDATE SET contact=$3,created_at=now()`, id, peer, in.Contact)
}

func (s *Server) unshareMatrimonyContact(w http.ResponseWriter, r *http.Request, id string) {
	s.memberExec(w, r, `DELETE FROM matrimony_contacts WHERE owner=$1 AND peer=$2`, id, r.PathValue("peer"))
}

// ---- full comparison and Ask Astrisk about a profile -------------------------

func (s *Server) comparablePair(w http.ResponseWriter, r *http.Request, id, peer string) (*horoscopeMatch, bool) {
	ctx := r.Context()
	var ok bool
	var myKind, theirKind string
	err := s.membersDB().QueryRow(ctx, `SELECT matrimony_visible($2,$1) AND m.horoscope_visible AND p.horoscope_visible, COALESCE(m.details->>'profile_kind',''), COALESCE(p.details->>'profile_kind','')
 FROM matrimony_profiles m, matrimony_profiles p WHERE m.account_id=$1 AND p.account_id=$2`, id, peer).Scan(&ok, &myKind, &theirKind)
	if err != nil || !ok {
		problem(w, 404, fmt.Errorf("horoscope comparison needs both members to share horoscope matching"))
		return nil, false
	}
	mb, err1 := s.birthOf(ctx, id)
	pb, err2 := s.birthOf(ctx, peer)
	if err1 != nil || err2 != nil {
		problem(w, 404, fmt.Errorf("profile unavailable"))
		return nil, false
	}
	h, ok := s.compareMembers(id, peer, mb, pb, myKind, theirKind)
	if !ok {
		problem(w, 409, fmt.Errorf("birth details are incomplete for this comparison"))
		return nil, false
	}
	return h, true
}

func (s *Server) matrimonyCompare(w http.ResponseWriter, r *http.Request, id string) {
	peer := r.PathValue("peer")
	h, ok := s.comparablePair(w, r, id, peer)
	if !ok {
		return
	}
	m := h.engineMatch
	// Lagna-based Mangal dosha is only meaningful with a birthplace.
	if !h.boyC.HasPlace {
		m.BoyMangal = false
	}
	if !h.girlC.HasPlace {
		m.GirlMangal = false
	}
	_, bh := engine.MangalDosha(h.boyC.Chart)
	_, gh := engine.MangalDosha(h.girlC.Chart)
	bs, gs := summary(h.boyC.Chart, bh), summary(h.girlC.Chart, gh)
	writeJSON(w, 200, map[string]any{"match": m, "boy": bs, "girl": gs, "explanation": reading.ChartMatchExplanation(m), "horoscope": h, "groom_has_place": h.boyC.HasPlace, "bride_has_place": h.girlC.HasPlace})
}

func (s *Server) matrimonyCompareChat(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Question string             `json:"question"`
		History  []reading.ChatTurn `json:"history"`
		Lang     string             `json:"lang"`
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 24<<10))
	if d.Decode(&req) != nil {
		problem(w, 400, fmt.Errorf("invalid request"))
		return
	}
	if len(req.History) > 6 {
		req.History = req.History[len(req.History)-6:]
	}
	for _, t := range req.History {
		if (t.Role != "user" && t.Role != "assistant") || len(t.Content) > 4000 {
			problem(w, 400, fmt.Errorf("invalid conversation history"))
			return
		}
	}
	h, ok := s.comparablePair(w, r, id, r.PathValue("peer"))
	if !ok {
		return
	}
	now := time.Now()
	person := func(c memberChart, b memberBirth) (reading.MatchPerson, error) {
		facts, rules, err := s.reading.BuildFacts(r.Context(), c.Chart, now)
		if err != nil {
			return reading.MatchPerson{}, err
		}
		p := reading.MatchPerson{Facts: facts, Rules: rules}
		if n, err := divination.ReadNumerology(b.Date, "", now.Year()); err == nil {
			p.Numerology = &n
		}
		return p, nil
	}
	bp, err1 := person(h.boyC, h.boyB)
	gp, err2 := person(h.girlC, h.girlB)
	if err1 != nil || err2 != nil {
		problem(w, 500, fmt.Errorf("comparison unavailable"))
		return
	}
	ans, err := s.reading.AnswerMatch(r.Context(), reading.MatchChatContext{Match: h.engineMatch, Boy: bp, Girl: gp, Lang: req.Lang}, req.Question, req.History)
	if err != nil {
		problem(w, 400, err)
		return
	}
	writeJSON(w, 200, ans)
}

// ---- selfie verification ---------------------------------------------------

func (s *Server) myVerification(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT status,note,created_at,reviewed_at FROM matrimony_verifications WHERE account_id=$1`, id)
}

func (s *Server) submitVerification(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultMember(w, r, id) {
		return
	}
	data, _, err := readMatrimonyPhoto(w, r)
	if err != nil {
		problem(w, 400, err)
		return
	}
	if r.FormValue("consent") != "true" {
		problem(w, 400, fmt.Errorf("consent to a moderator viewing this selfie is required"))
		return
	}
	var published int
	if s.membersDB().QueryRow(r.Context(), `SELECT count(*) FROM matrimony_photos WHERE owner=$1 AND draft_id IS NULL AND published`, id).Scan(&published) != nil || published == 0 {
		problem(w, 400, fmt.Errorf("publish at least one profile photo first, so a moderator can compare"))
		return
	}
	s.memberExec(w, r, `INSERT INTO matrimony_verifications(account_id,photo,status) VALUES($1,$2,'pending') ON CONFLICT(account_id) DO UPDATE SET photo=$2,status='pending',note='',created_at=now(),reviewed_at=NULL,reviewer=NULL`, id, data)
}

func (s *Server) verificationQueue(w http.ResponseWriter, r *http.Request, id string) {
	if !s.isModerator(r, id) {
		problem(w, 403, fmt.Errorf("moderator access required"))
		return
	}
	s.memberRows(w, r, `SELECT v.account_id,a.handle,p.details->>'display_name' display_name,v.created_at,
 COALESCE((SELECT json_agg(ph.id ORDER BY ph.created_at) FROM matrimony_photos ph WHERE ph.owner=v.account_id AND ph.draft_id IS NULL AND ph.published),'[]') photos
 FROM matrimony_verifications v JOIN member_accounts a ON a.id=v.account_id LEFT JOIN matrimony_profiles p ON p.account_id=v.account_id
 WHERE v.status='pending' ORDER BY v.created_at LIMIT 100`)
}

func (s *Server) verificationPhoto(w http.ResponseWriter, r *http.Request, id string) {
	if !s.isModerator(r, id) {
		problem(w, 403, fmt.Errorf("moderator access required"))
		return
	}
	var data []byte
	if err := s.membersDB().QueryRow(r.Context(), `SELECT photo FROM matrimony_verifications WHERE account_id=$1 AND status='pending' AND photo IS NOT NULL`, r.PathValue("account")).Scan(&data); err != nil {
		problem(w, 404, fmt.Errorf("selfie unavailable"))
		return
	}
	w.Header().Set("Content-Type", http.DetectContentType(data))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

// reviewVerification approves or rejects a selfie; the selfie is deleted
// either way, and only the outcome is kept.
func (s *Server) reviewVerification(w http.ResponseWriter, r *http.Request, id string) {
	if !s.isModerator(r, id) {
		problem(w, 403, fmt.Errorf("moderator access required"))
		return
	}
	var in struct {
		Approve bool   `json:"approve"`
		Note    string `json:"note"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if len(in.Note) > 300 {
		problem(w, 400, fmt.Errorf("note must be under 300 characters"))
		return
	}
	account := r.PathValue("account")
	s.memberExec(w, r, `WITH v AS (UPDATE matrimony_verifications SET photo=NULL,status=CASE WHEN $2 THEN 'approved' ELSE 'rejected' END,note=$3,reviewed_at=now(),reviewer=$4 WHERE account_id=$1 AND status='pending' RETURNING account_id),
 p AS (UPDATE matrimony_profiles SET verified_at=CASE WHEN $2 THEN now() ELSE NULL END WHERE account_id IN (SELECT account_id FROM v) RETURNING account_id)
 INSERT INTO member_audit(actor,action,subject) SELECT $4,CASE WHEN $2 THEN 'verification_approved' ELSE 'verification_rejected' END,account_id FROM v`, account, in.Approve, in.Note, id)
}
