package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Server) membersDB() *pgxpool.Pool { return s.cache.(PostgresCache).Pool }

type memberHandler func(http.ResponseWriter, *http.Request, string)

func (s *Server) memberRoute(pattern string, fn memberHandler) {
	s.mux.Handle(pattern, s.limit(s.accountGuard(func(w http.ResponseWriter, r *http.Request) {
		id, err := s.memberID(r)
		if err != nil {
			problem(w, 401, fmt.Errorf("sign in required"))
			return
		}
		fn(w, r, id)
	}), 120))
}
func (s *Server) communityRoutes() {
	s.memberRoute("GET /api/community/settings", s.communitySettings)
	s.memberRoute("PUT /api/community/settings", s.saveCommunitySettings)
	s.memberRoute("GET /api/community/people", s.communityPeople)
	s.memberRoute("GET /api/community/follows", s.communityFollows)
	s.memberRoute("POST /api/community/follows", s.followAction)
	s.memberRoute("POST /api/community/blocks", s.blockAction)
	s.memberRoute("GET /api/community/blocks", s.listBlocks)
	s.memberRoute("GET /api/community/posts", s.feed)
	s.memberRoute("POST /api/community/posts", s.createPost)
	s.memberRoute("PUT /api/community/posts/{post}", s.editPost)
	s.memberRoute("DELETE /api/community/posts/{post}", s.deletePost)
	s.memberRoute("POST /api/community/media", s.uploadMedia)
	s.memberRoute("GET /api/community/media/{media}", s.getMedia)
	s.memberRoute("POST /api/community/posts/{post}/react", s.reactPost)
	s.memberRoute("GET /api/community/posts/{post}/comments", s.comments)
	s.memberRoute("POST /api/community/posts/{post}/comments", s.addComment)
	s.memberRoute("DELETE /api/community/comments/{comment}", s.deleteComment)
	s.memberRoute("POST /api/community/posts/{post}/report", s.reportPost)
	s.memberRoute("GET /api/community/moderation", s.moderationQueue)
	s.memberRoute("POST /api/community/moderation", s.moderatePost)
	s.familyRoutes()
	s.matrimonyRoutes()
	s.securityRoutes()
	s.notificationRoutes()
}
func (s *Server) memberRows(w http.ResponseWriter, r *http.Request, query string, args ...any) {
	rows, err := s.membersDB().Query(r.Context(), `SELECT row_to_json(q) FROM (`+query+`) q`, args...)
	if err != nil {
		s.logger.Error("member query", "error", err)
		problem(w, 500, fmt.Errorf("request unavailable"))
		return
	}
	defer rows.Close()
	result := []json.RawMessage{}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			problem(w, 500, fmt.Errorf("read failed"))
			return
		}
		result = append(result, json.RawMessage(raw))
	}
	if rows.Err() != nil {
		problem(w, 500, fmt.Errorf("read failed"))
		return
	}
	writeJSON(w, 200, result)
}
func (s *Server) memberExec(w http.ResponseWriter, r *http.Request, query string, args ...any) {
	tag, err := s.membersDB().Exec(r.Context(), query, args...)
	if err != nil {
		s.logger.Error("member write", "error", err)
		problem(w, 400, fmt.Errorf("request could not be saved"))
		return
	}
	if tag.RowsAffected() == 0 {
		problem(w, 404, fmt.Errorf("item unavailable or permission denied"))
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) adultCommunity(w http.ResponseWriter, r *http.Request, id string) bool {
	var ok bool
	err := s.membersDB().QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM member_settings WHERE account_id=$1 AND community AND birth_date <= CURRENT_DATE - INTERVAL '18 years')`, id).Scan(&ok)
	if err != nil || !ok {
		problem(w, 403, fmt.Errorf("enable community in your profile with your adult date of birth first"))
		return false
	}
	return true
}

type communitySettingsInput struct {
	Community   bool              `json:"community"`
	BirthDate   string            `json:"birth_date"`
	Avatar      string            `json:"avatar"`
	Accent      string            `json:"accent"`
	PublicBio   string            `json:"public_bio"`
	Interests   []string          `json:"interests"`
	Links       []string          `json:"links"`
	Preferences map[string]string `json:"preferences"`
}

func (s *Server) communitySettings(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT community,to_char(birth_date,'YYYY-MM-DD') birth_date,avatar,accent,public_bio,interests,links,preferences FROM member_settings WHERE account_id=$1`, id)
}
func (s *Server) saveCommunitySettings(w http.ResponseWriter, r *http.Request, id string) {
	var in communitySettingsInput
	if !memberInput(w, r, &in) {
		return
	}
	dob, err := time.Parse("2006-01-02", in.BirthDate)
	var registeredDOB string
	if e := s.membersDB().QueryRow(r.Context(), `SELECT COALESCE(to_char(birth_date,'YYYY-MM-DD'),'') FROM member_accounts WHERE id=$1`, id).Scan(&registeredDOB); e != nil || (registeredDOB != "" && registeredDOB != in.BirthDate) {
		problem(w, 400, fmt.Errorf("use the date of birth registered with your account"))
		return
	}
	if err != nil || dob.After(time.Now()) || dob.Year() < 1900 || (in.Community && dob.AddDate(18, 0, 0).After(time.Now())) {
		problem(w, 400, fmt.Errorf("enter a valid birth date; community requires age 18 or older"))
		return
	}
	if len(in.PublicBio) > 1000 || len(in.Interests) > 20 || len(in.Links) > 5 || len(in.Preferences) > 12 {
		problem(w, 400, fmt.Errorf("profile exceeds limits"))
		return
	}
	if !strings.Contains("|sun|moon|star|leaf|mountain|", "|"+in.Avatar+"|") || in.Avatar == "" {
		in.Avatar = "sun"
	}
	if len(in.Accent) != 7 || in.Accent[0] != '#' {
		in.Accent = "#d6b467"
	}
	if _, err = strconv.ParseUint(in.Accent[1:], 16, 32); err != nil {
		in.Accent = "#d6b467"
	}
	for i, v := range in.Interests {
		in.Interests[i] = strings.ToLower(strings.TrimSpace(v))
		if len(v) > 40 {
			problem(w, 400, fmt.Errorf("interest too long"))
			return
		}
	}
	for k, v := range in.Preferences {
		if len(k) > 40 || len(v) > 200 {
			problem(w, 400, fmt.Errorf("preference too long"))
			return
		}
	}
	for _, link := range in.Links {
		u, e := url.Parse(link)
		if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || len(link) > 300 {
			problem(w, 400, fmt.Errorf("social links must be valid HTTPS URLs"))
			return
		}
	}
	if in.Interests == nil {
		in.Interests = []string{}
	}
	if in.Links == nil {
		in.Links = []string{}
	}
	if in.Preferences == nil {
		in.Preferences = map[string]string{}
	}
	links, _ := json.Marshal(in.Links)
	prefs, _ := json.Marshal(in.Preferences)
	s.memberExec(w, r, `INSERT INTO member_settings(account_id,community,birth_date,avatar,accent,public_bio,interests,links,preferences) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(account_id) DO UPDATE SET community=$2,birth_date=$3,avatar=$4,accent=$5,public_bio=$6,interests=$7,links=$8,preferences=$9`, id, in.Community, dob, in.Avatar, in.Accent, in.PublicBio, in.Interests, links, prefs)
}
func (s *Server) communityPeople(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultCommunity(w, r, id) {
		return
	}
	s.memberRows(w, r, `SELECT a.id,a.handle,COALESCE(a.profile->>'name','') display_name,c.public_bio,c.avatar,c.accent,c.interests,c.links,
 COALESCE((SELECT CASE WHEN f.accepted THEN 'accepted' ELSE 'pending' END FROM member_follows f WHERE f.follower=$1 AND f.target=a.id),'none') following
 FROM member_accounts a JOIN member_settings c ON c.account_id=a.id WHERE c.community AND a.id<>$1 AND NOT member_blocked($1,a.id) AND (a.handle ILIKE $2 OR a.profile->>'name' ILIKE $2) ORDER BY a.handle LIMIT 50`, id, "%"+r.URL.Query().Get("q")+"%")
}
func (s *Server) communityFollows(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT a.id,a.handle,COALESCE(a.profile->>'name','') display_name,f.accepted FROM member_follows f JOIN member_accounts a ON a.id=f.follower WHERE f.target=$1 AND NOT member_blocked($1,a.id) ORDER BY a.handle`, id)
}
func (s *Server) followAction(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultCommunity(w, r, id) {
		return
	}
	var in struct {
		Target string `json:"target"`
		Action string `json:"action"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	switch in.Action {
	case "request":
		s.memberExec(w, r, `INSERT INTO member_follows(follower,target) SELECT $1,account_id FROM member_settings WHERE account_id=$2 AND community AND account_id<>$1 AND NOT member_blocked($1,$2) ON CONFLICT(follower,target) DO UPDATE SET follower=EXCLUDED.follower`, id, in.Target)
	case "accept":
		s.memberExec(w, r, `UPDATE member_follows SET accepted=true WHERE target=$1 AND follower=$2 AND NOT member_blocked($1,$2)`, id, in.Target)
	case "remove":
		s.memberExec(w, r, `DELETE FROM member_follows WHERE target=$1 AND follower=$2`, id, in.Target)
	case "unfollow":
		s.memberExec(w, r, `DELETE FROM member_follows WHERE follower=$1 AND target=$2`, id, in.Target)
	default:
		problem(w, 400, fmt.Errorf("invalid follow action"))
	}
}
func (s *Server) blockAction(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Target string `json:"target"`
		Block  bool   `json:"block"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if !in.Block {
		s.memberExec(w, r, `DELETE FROM member_blocks WHERE actor=$1 AND target=$2`, id, in.Target)
		return
	}
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	_, err = tx.Exec(r.Context(), `INSERT INTO member_blocks(actor,target) VALUES($1,$2) ON CONFLICT DO NOTHING`, id, in.Target)
	if err == nil {
		_, err = tx.Exec(r.Context(), `DELETE FROM member_follows WHERE (follower=$1 AND target=$2) OR (follower=$2 AND target=$1)`, id, in.Target)
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `DELETE FROM matrimony_interests WHERE (sender=$1 AND recipient=$2) OR (sender=$2 AND recipient=$1)`, id, in.Target)
	}
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 400, fmt.Errorf("block failed"))
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) listBlocks(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT a.id,a.handle FROM member_blocks b JOIN member_accounts a ON a.id=b.target WHERE b.actor=$1`, id)
}
func (s *Server) postAllowed(r *http.Request, id string) bool {
	var allowed bool
	err := s.membersDB().QueryRow(r.Context(), `SELECT post_access($1,$2)`, r.PathValue("post"), id).Scan(&allowed)
	return err == nil && allowed
}
func (s *Server) feed(w http.ResponseWriter, r *http.Request, id string) {
	before := int64(9223372036854775807)
	if v := r.URL.Query().Get("before"); v != "" {
		n, e := strconv.ParseInt(v, 10, 64)
		if e != nil || n < 1 {
			problem(w, 400, fmt.Errorf("invalid cursor"))
			return
		}
		before = n
	}
	mode := r.URL.Query().Get("mode")
	s.memberRows(w, r, `SELECT p.id,p.caption,p.audience,p.family_id,p.hidden,p.created_at,p.owner=$1 mine,a.id author_id,a.handle,COALESCE(a.profile->>'name','') display_name,c.avatar,c.accent,
 COALESCE((SELECT json_agg(json_build_object('id',m.id,'alt',m.alt) ORDER BY m.created_at,m.id) FROM community_media m WHERE m.post_id=p.id),'[]') media,
 (SELECT count(*) FROM community_reactions v WHERE v.post_id=p.id AND v.kind='like') likes,
 EXISTS(SELECT 1 FROM community_reactions v WHERE v.post_id=p.id AND v.account_id=$1 AND v.kind='like') liked,
 EXISTS(SELECT 1 FROM community_reactions v WHERE v.post_id=p.id AND v.account_id=$1 AND v.kind='bookmark') bookmarked
 FROM community_posts p JOIN member_accounts a ON a.id=p.owner LEFT JOIN member_settings c ON c.account_id=p.owner
 WHERE p.id<$2 AND post_access(p.id,$1) AND ($3<>'mine' OR p.owner=$1) AND ($3<>'saved' OR EXISTS(SELECT 1 FROM community_reactions v WHERE v.post_id=p.id AND v.account_id=$1 AND v.kind='bookmark'))
 AND ($3<>'following' OR p.owner=$1 OR EXISTS(SELECT 1 FROM member_follows f WHERE f.follower=$1 AND f.target=p.owner AND f.accepted)) ORDER BY p.id DESC LIMIT 20`, id, before, mode)
}

type postInput struct {
	Caption  string   `json:"caption"`
	Audience string   `json:"audience"`
	FamilyID string   `json:"family_id"`
	Media    []string `json:"media"`
}

func (s *Server) validPost(w http.ResponseWriter, r *http.Request, id string, in postInput) bool {
	if len(in.Caption) > 2200 || len(in.Media) > 10 || (!strings.Contains("|private|followers|family|community|", "|"+in.Audience+"|") || in.Audience == "") {
		problem(w, 400, fmt.Errorf("invalid post; caption limit 2200 and maximum 10 photos"))
		return false
	}
	if strings.TrimSpace(in.Caption) == "" && len(in.Media) == 0 {
		problem(w, 400, fmt.Errorf("add a caption or photo"))
		return false
	}
	if in.Audience == "family" {
		var ok bool
		_ = s.membersDB().QueryRow(r.Context(), `SELECT family_access($1,$2)`, in.FamilyID, id).Scan(&ok)
		if !ok {
			problem(w, 403, fmt.Errorf("family unavailable"))
			return false
		}
	}
	return true
}
func (s *Server) createPost(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultCommunity(w, r, id) {
		return
	}
	var in postInput
	if !memberInput(w, r, &in) || !s.validPost(w, r, id, in) {
		return
	}
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	var post int64
	err = tx.QueryRow(r.Context(), `INSERT INTO community_posts(owner,caption,audience,family_id) VALUES($1,$2,$3,NULLIF($4,'')) RETURNING id`, id, in.Caption, in.Audience, in.FamilyID).Scan(&post)
	seen := map[string]bool{}
	for _, m := range in.Media {
		if err != nil {
			break
		}
		if seen[m] {
			err = fmt.Errorf("duplicate media")
			break
		}
		seen[m] = true
		tag, e := tx.Exec(r.Context(), `UPDATE community_media SET post_id=$1 WHERE id=$2 AND owner=$3 AND post_id IS NULL`, post, m, id)
		err = e
		if e == nil && tag.RowsAffected() != 1 {
			err = fmt.Errorf("media unavailable")
		}
	}
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 400, fmt.Errorf("post could not be created; check your media"))
		return
	}
	writeJSON(w, 201, map[string]any{"id": post})
}
func (s *Server) editPost(w http.ResponseWriter, r *http.Request, id string) {
	var in postInput
	if !memberInput(w, r, &in) || !s.validPost(w, r, id, in) {
		return
	}
	s.memberExec(w, r, `UPDATE community_posts SET caption=$1,audience=$2,family_id=NULLIF($3,''),updated_at=now() WHERE id=$4 AND owner=$5`, in.Caption, in.Audience, in.FamilyID, r.PathValue("post"), id)
}
func (s *Server) deletePost(w http.ResponseWriter, r *http.Request, id string) {
	s.memberExec(w, r, `DELETE FROM community_posts WHERE id=$1 AND owner=$2`, r.PathValue("post"), id)
}
func (s *Server) uploadMedia(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultCommunity(w, r, id) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		problem(w, 400, fmt.Errorf("photo must be under 8 MB"))
		return
	}
	defer r.MultipartForm.RemoveAll()
	f, _, err := r.FormFile("photo")
	if err != nil {
		problem(w, 400, fmt.Errorf("photo required"))
		return
	}
	defer f.Close()
	raw, err := io.ReadAll(f)
	if err != nil {
		problem(w, 400, fmt.Errorf("invalid photo"))
		return
	}
	cfg, kind, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (kind != "jpeg" && kind != "png") || cfg.Width > 6000 || cfg.Height > 6000 || int64(cfg.Width)*int64(cfg.Height) > 16000000 {
		problem(w, 400, fmt.Errorf("use JPEG or PNG up to 16 megapixels"))
		return
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		problem(w, 400, fmt.Errorf("invalid photo"))
		return
	}
	var clean bytes.Buffer
	if jpeg.Encode(&clean, img, &jpeg.Options{Quality: 82}) != nil {
		problem(w, 400, fmt.Errorf("photo processing failed"))
		return
	}
	alt := r.FormValue("alt")
	if len(alt) > 300 {
		problem(w, 400, fmt.Errorf("alt text too long"))
		return
	}
	// Account row lock makes quota checks atomic across concurrent uploads.
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	if _, err = tx.Exec(r.Context(), `SELECT id FROM member_accounts WHERE id=$1 FOR UPDATE`, id); err != nil {
		problem(w, 500, fmt.Errorf("unavailable"))
		return
	}
	_, err = tx.Exec(r.Context(), `DELETE FROM community_media WHERE owner=$1 AND post_id IS NULL AND created_at<now()-interval '1 day'`, id)
	var used int64
	if err == nil {
		err = tx.QueryRow(r.Context(), `SELECT COALESCE(sum(octet_length(data)),0) FROM community_media WHERE owner=$1`, id).Scan(&used)
	}
	if err != nil || used+int64(clean.Len()) > 100<<20 {
		problem(w, 400, fmt.Errorf("photo storage limit reached (100 MB)"))
		return
	}
	mid := randomToken()
	_, err = tx.Exec(r.Context(), `INSERT INTO community_media(id,owner,data,mime,alt) VALUES($1,$2,$3,'image/jpeg',$4)`, mid, id, clean.Bytes(), alt)
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 500, fmt.Errorf("upload failed"))
		return
	}
	writeJSON(w, 201, map[string]string{"id": mid})
}
func (s *Server) getMedia(w http.ResponseWriter, r *http.Request, id string) {
	var raw []byte
	err := s.membersDB().QueryRow(r.Context(), `SELECT data FROM community_media WHERE id=$1 AND (owner=$2 OR post_access(post_id,$2))`, r.PathValue("media"), id).Scan(&raw)
	if err != nil {
		problem(w, 404, fmt.Errorf("photo unavailable"))
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(raw)
}
func (s *Server) reactPost(w http.ResponseWriter, r *http.Request, id string) {
	if !s.postAllowed(r, id) {
		problem(w, 404, fmt.Errorf("post unavailable"))
		return
	}
	var in struct {
		Kind   string `json:"kind"`
		Active bool   `json:"active"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if in.Kind != "like" && in.Kind != "bookmark" {
		problem(w, 400, fmt.Errorf("invalid reaction"))
		return
	}
	if in.Active {
		s.memberExec(w, r, `INSERT INTO community_reactions(account_id,post_id,kind) VALUES($1,$2,$3) ON CONFLICT(account_id,post_id,kind) DO UPDATE SET kind=EXCLUDED.kind`, id, r.PathValue("post"), in.Kind)
	} else {
		s.memberExec(w, r, `DELETE FROM community_reactions WHERE account_id=$1 AND post_id=$2 AND kind=$3`, id, r.PathValue("post"), in.Kind)
	}
}
func (s *Server) comments(w http.ResponseWriter, r *http.Request, id string) {
	if !s.postAllowed(r, id) {
		problem(w, 404, fmt.Errorf("post unavailable"))
		return
	}
	s.memberRows(w, r, `SELECT c.id,c.body,c.created_at,a.handle,COALESCE(a.profile->>'name','') display_name,c.owner=$2 mine FROM community_comments c JOIN member_accounts a ON a.id=c.owner WHERE c.post_id=$1 AND NOT member_blocked(c.owner,$2) ORDER BY c.id DESC LIMIT 100`, r.PathValue("post"), id)
}
func (s *Server) addComment(w http.ResponseWriter, r *http.Request, id string) {
	if !s.postAllowed(r, id) {
		problem(w, 404, fmt.Errorf("post unavailable"))
		return
	}
	var in struct {
		Body string `json:"body"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if len(in.Body) > 1000 || strings.TrimSpace(in.Body) == "" {
		problem(w, 400, fmt.Errorf("comment must be 1–1000 characters"))
		return
	}
	s.memberExec(w, r, `INSERT INTO community_comments(owner,post_id,body) VALUES($1,$2,$3)`, id, r.PathValue("post"), in.Body)
}
func (s *Server) deleteComment(w http.ResponseWriter, r *http.Request, id string) {
	s.memberExec(w, r, `DELETE FROM community_comments c WHERE c.id=$1 AND (c.owner=$2 OR EXISTS(SELECT 1 FROM community_posts p WHERE p.id=c.post_id AND p.owner=$2))`, r.PathValue("comment"), id)
}
func (s *Server) reportPost(w http.ResponseWriter, r *http.Request, id string) {
	if !s.postAllowed(r, id) {
		problem(w, 404, fmt.Errorf("post unavailable"))
		return
	}
	var in struct {
		Reason string `json:"reason"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if len(in.Reason) > 1000 || strings.TrimSpace(in.Reason) == "" {
		problem(w, 400, fmt.Errorf("reason required (maximum 1000 characters)"))
		return
	}
	s.memberExec(w, r, `INSERT INTO community_reports(reporter,post_id,reason) VALUES($1,$2,$3)`, id, r.PathValue("post"), in.Reason)
}
func (s *Server) isModerator(r *http.Request, id string) bool {
	role := s.accountRole(r, id)
	return role == "superadmin" || role == "moderator"
}
func (s *Server) moderationQueue(w http.ResponseWriter, r *http.Request, id string) {
	if !s.isModerator(r, id) {
		problem(w, 403, fmt.Errorf("moderator access required"))
		return
	}
	s.memberRows(w, r, `SELECT r.id,r.post_id,r.reason,r.status,p.caption FROM community_reports r JOIN community_posts p ON p.id=r.post_id WHERE r.status='open' ORDER BY r.id LIMIT 100`)
}
func (s *Server) moderatePost(w http.ResponseWriter, r *http.Request, id string) {
	if !s.isModerator(r, id) {
		problem(w, 403, fmt.Errorf("moderator access required"))
		return
	}
	var in struct {
		Report int64 `json:"report"`
		Hide   bool  `json:"hide"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	s.memberExec(w, r, `WITH report AS (UPDATE community_reports SET status=CASE WHEN $2 THEN 'hidden' ELSE 'dismissed' END WHERE id=$1 RETURNING post_id), changed AS (UPDATE community_posts SET hidden=$2 WHERE id IN (SELECT post_id FROM report) RETURNING id) INSERT INTO member_audit(actor,action,subject) SELECT $3,'moderate',id::text FROM changed`, in.Report, in.Hide, id)
}
