package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strings"
)

func (s *Server) biodataRoutes() {
	s.memberRoute("GET /api/matrimony/photos", s.myMatrimonyPhotos)
	s.memberRoute("POST /api/matrimony/photos", s.uploadMatrimonyPhoto)
	s.memberRoute("GET /api/matrimony/photos/{photo}", s.matrimonyPhoto)
	s.memberRoute("DELETE /api/matrimony/photos/{photo}", s.deleteMatrimonyPhoto)
	s.memberRoute("GET /api/matrimony/drafts", s.biodataDrafts)
	s.memberRoute("POST /api/matrimony/drafts", s.createBiodataDraft)
	s.memberRoute("POST /api/matrimony/drafts/{draft}", s.biodataAction)
	s.memberRoute("DELETE /api/matrimony/drafts/{draft}", s.deleteBiodataDraft)
	s.memberRoute("POST /api/matrimony/report/{subject}", s.reportMatrimony)
	s.memberRoute("GET /api/matrimony/moderation", s.matrimonyReports)
	s.memberRoute("POST /api/matrimony/moderation", s.moderateMatrimony)
}

func (s *Server) adultMember(w http.ResponseWriter, r *http.Request, id string) bool {
	var adult bool
	err := s.membersDB().QueryRow(r.Context(), `SELECT birth_date<=CURRENT_DATE-INTERVAL '18 years' FROM member_accounts WHERE id=$1`, id).Scan(&adult)
	if err != nil || !adult {
		problem(w, 403, fmt.Errorf("an adult account is required"))
		return false
	}
	return true
}

func (s *Server) myMatrimonyPhotos(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT id,alt,published FROM matrimony_photos WHERE owner=$1 AND draft_id IS NULL ORDER BY created_at,id`, id)
}

// Decoding and re-encoding removes EXIF/location metadata and rejects active files.
func readMatrimonyPhoto(w http.ResponseWriter, r *http.Request) ([]byte, string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		return nil, "", fmt.Errorf("photo must be under 8 MB")
	}
	defer r.MultipartForm.RemoveAll()
	f, _, err := r.FormFile("photo")
	if err != nil {
		return nil, "", fmt.Errorf("photo required")
	}
	defer f.Close()
	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, "", fmt.Errorf("invalid photo")
	}
	cfg, kind, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (kind != "jpeg" && kind != "png") || cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 6000 || cfg.Height > 6000 || int64(cfg.Width)*int64(cfg.Height) > 16000000 {
		return nil, "", fmt.Errorf("use JPEG or PNG up to 16 megapixels")
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, "", fmt.Errorf("invalid photo")
	}
	var clean bytes.Buffer
	if err = jpeg.Encode(&clean, img, &jpeg.Options{Quality: 82}); err != nil {
		return nil, "", fmt.Errorf("photo processing failed")
	}
	alt := strings.TrimSpace(r.FormValue("alt"))
	if len(alt) > 300 {
		return nil, "", fmt.Errorf("photo description must be under 300 characters")
	}
	return clean.Bytes(), alt, nil
}

func (s *Server) uploadMatrimonyPhoto(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultMember(w, r, id) {
		return
	}
	data, alt, err := readMatrimonyPhoto(w, r)
	if err != nil {
		problem(w, 400, err)
		return
	}
	if r.FormValue("consent") != "true" {
		problem(w, 400, fmt.Errorf("permission to upload this photo is required"))
		return
	}
	draft := r.FormValue("draft_id")
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("upload unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	if _, err = tx.Exec(r.Context(), `SELECT id FROM member_accounts WHERE id=$1 FOR UPDATE`, id); err != nil {
		problem(w, 500, fmt.Errorf("upload unavailable"))
		return
	}
	if draft != "" {
		var found string
		if err = tx.QueryRow(r.Context(), `SELECT id FROM matrimony_drafts WHERE id=$1 AND creator=$2 AND status='draft' FOR UPDATE`, draft, id).Scan(&found); err != nil {
			problem(w, 404, fmt.Errorf("editable draft unavailable"))
			return
		}
	}
	var count int
	var used int64
	err = tx.QueryRow(r.Context(), `SELECT count(*) FROM matrimony_photos WHERE owner=$1 AND draft_id IS NOT DISTINCT FROM NULLIF($2,'')`, id, draft).Scan(&count)
	if err == nil {
		err = tx.QueryRow(r.Context(), `SELECT COALESCE(sum(octet_length(data)),0) FROM matrimony_photos WHERE owner=$1`, id).Scan(&used)
	}
	if err != nil || count >= 6 || used+int64(len(data)) > 100<<20 {
		problem(w, 400, fmt.Errorf("limit: 6 photos per profile/draft and 100 MB total matrimony photos"))
		return
	}
	pid := randomToken()
	_, err = tx.Exec(r.Context(), `INSERT INTO matrimony_photos(id,owner,draft_id,data,alt) VALUES($1,$2,NULLIF($3,''),$4,$5)`, pid, id, draft, data, alt)
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 500, fmt.Errorf("upload failed"))
		return
	}
	writeJSON(w, 201, map[string]string{"id": pid})
}

func (s *Server) matrimonyPhoto(w http.ResponseWriter, r *http.Request, id string) {
	forOwner := r.URL.Query().Get("for")
	if forOwner == "" {
		forOwner = id
	}
	var data []byte
	err := s.membersDB().QueryRow(r.Context(), `SELECT ph.data FROM matrimony_photos ph WHERE ph.id=$1 AND (
 ph.owner=$2 OR
 (ph.draft_id IS NOT NULL AND EXISTS(SELECT 1 FROM matrimony_drafts d WHERE d.id=ph.draft_id AND d.recipient=$2 AND d.status='pending' AND NOT member_blocked(d.creator,$2))) OR
 (ph.draft_id IS NULL AND ph.published AND matrimony_visible(ph.owner,$3) AND NOT member_blocked(ph.owner,$2) AND ($2=$3 OR delegate_allowed($3,$2)))
 )`, r.PathValue("photo"), id, forOwner).Scan(&data)
	if err != nil {
		problem(w, 404, fmt.Errorf("photo unavailable"))
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(data)
}

func (s *Server) deleteMatrimonyPhoto(w http.ResponseWriter, r *http.Request, id string) {
	s.memberExec(w, r, `DELETE FROM matrimony_photos ph WHERE ph.id=$1 AND ph.owner=$2 AND (ph.draft_id IS NULL OR EXISTS(SELECT 1 FROM matrimony_drafts d WHERE d.id=ph.draft_id AND d.status='draft'))`, r.PathValue("photo"), id)
}

func (s *Server) biodataDrafts(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT d.id,d.relationship,d.details,d.status,d.creator=$1 mine,a.handle creator_handle,b.handle recipient_handle,
 COALESCE((SELECT json_agg(json_build_object('id',p.id,'alt',p.alt) ORDER BY p.created_at,p.id) FROM matrimony_photos p WHERE p.draft_id=d.id),'[]') photos
 FROM matrimony_drafts d JOIN member_accounts a ON a.id=d.creator LEFT JOIN member_accounts b ON b.id=d.recipient
 WHERE (d.creator=$1 OR d.recipient=$1) AND (d.recipient IS NULL OR NOT member_blocked(d.creator,d.recipient)) ORDER BY d.created_at DESC`, id)
}

func (s *Server) createBiodataDraft(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultMember(w, r, id) {
		return
	}
	var in struct {
		Relationship string           `json:"relationship"`
		Details      matrimonyDetails `json:"details"`
		Consent      bool             `json:"consent"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if !in.Consent || !validMatrimonyDetails(in.Details) || strings.TrimSpace(in.Details.DisplayName) == "" || !strings.Contains("|parent|sibling|relative|friend|", "|"+in.Relationship+"|") || in.Relationship == "" {
		problem(w, 400, fmt.Errorf("adult person's name, valid biodata, relationship and permission are required"))
		return
	}
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("draft unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	if _, err = tx.Exec(r.Context(), `SELECT id FROM member_accounts WHERE id=$1 FOR UPDATE`, id); err != nil {
		problem(w, 500, fmt.Errorf("draft unavailable"))
		return
	}
	var n int
	if err = tx.QueryRow(r.Context(), `SELECT count(*) FROM matrimony_drafts WHERE creator=$1`, id).Scan(&n); err != nil || n >= 10 {
		problem(w, 400, fmt.Errorf("keep at most 10 pending biodata drafts"))
		return
	}
	raw, _ := json.Marshal(in.Details)
	draft := randomToken()
	_, err = tx.Exec(r.Context(), `INSERT INTO matrimony_drafts(id,creator,relationship,details) VALUES($1,$2,$3,$4)`, draft, id, in.Relationship, raw)
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 500, fmt.Errorf("draft could not be saved"))
		return
	}
	writeJSON(w, 201, map[string]string{"id": draft})
}

func (s *Server) biodataAction(w http.ResponseWriter, r *http.Request, id string) {
	if !s.adultMember(w, r, id) {
		return
	}
	var in struct {
		Action  string `json:"action"`
		Handle  string `json:"handle"`
		Consent bool   `json:"consent"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	draft := r.PathValue("draft")
	if in.Action == "send" {
		s.memberExec(w, r, `UPDATE matrimony_drafts d SET recipient=a.id,status='pending' FROM member_accounts a WHERE d.id=$1 AND d.creator=$2 AND d.status='draft' AND a.handle=$3 AND a.id<>$2 AND a.birth_date<=CURRENT_DATE-INTERVAL '18 years' AND NOT member_blocked(a.id,$2)`, draft, id, strings.ToLower(strings.TrimSpace(in.Handle)))
		return
	}
	if in.Action != "accept" || !in.Consent {
		problem(w, 400, fmt.Errorf("review and consent to using this biodata first"))
		return
	}
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("draft unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	var raw []byte
	err = tx.QueryRow(r.Context(), `SELECT details FROM matrimony_drafts WHERE id=$1 AND recipient=$2 AND status='pending' AND NOT member_blocked(creator,$2) FOR UPDATE`, draft, id).Scan(&raw)
	if err != nil {
		problem(w, 404, fmt.Errorf("draft unavailable"))
		return
	}
	if _, err = tx.Exec(r.Context(), `SELECT id FROM member_accounts WHERE id=$1 FOR UPDATE`, id); err != nil {
		problem(w, 500, fmt.Errorf("draft unavailable"))
		return
	}
	var n int
	var used int64
	err = tx.QueryRow(r.Context(), `SELECT count(*),COALESCE(sum(octet_length(data)),0) FROM matrimony_photos WHERE (owner=$1 AND draft_id IS NULL) OR draft_id=$2`, id, draft).Scan(&n, &used)
	if err != nil || n > 6 {
		problem(w, 400, fmt.Errorf("remove some existing profile photos first; maximum 6 after import"))
		return
	}
	err = tx.QueryRow(r.Context(), `SELECT COALESCE(sum(octet_length(data)),0) FROM matrimony_photos WHERE owner=$1 OR draft_id=$2`, id, draft).Scan(&used)
	if err != nil || used > 100<<20 {
		problem(w, 400, fmt.Errorf("matrimony photo storage limit reached"))
		return
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO matrimony_profiles(account_id,active,details) VALUES($1,false,$2) ON CONFLICT(account_id) DO UPDATE SET active=false,details=$2,updated_at=now()`, id, raw)
	if err == nil {
		_, err = tx.Exec(r.Context(), `UPDATE matrimony_photos SET owner=$1,draft_id=NULL,published=false WHERE draft_id=$2`, id, draft)
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `DELETE FROM matrimony_drafts WHERE id=$1`, draft)
	}
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 500, fmt.Errorf("biodata could not be imported"))
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (s *Server) deleteBiodataDraft(w http.ResponseWriter, r *http.Request, id string) {
	s.memberExec(w, r, `DELETE FROM matrimony_drafts WHERE id=$1 AND (creator=$2 OR recipient=$2)`, r.PathValue("draft"), id)
}

func (s *Server) reportMatrimony(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Reason string `json:"reason"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Reason) == "" || len(in.Reason) > 1000 {
		problem(w, 400, fmt.Errorf("report reason required, up to 1000 characters"))
		return
	}
	s.memberExec(w, r, `INSERT INTO matrimony_reports(reporter,subject,reason) SELECT $1,$2,$3 WHERE matrimony_visible($2,$1) AND NOT EXISTS(SELECT 1 FROM matrimony_reports WHERE reporter=$1 AND subject=$2 AND status='open')`, id, r.PathValue("subject"), in.Reason)
}

func (s *Server) matrimonyReports(w http.ResponseWriter, r *http.Request, id string) {
	if !s.isModerator(r, id) {
		problem(w, 403, fmt.Errorf("moderator access required"))
		return
	}
	s.memberRows(w, r, `SELECT r.id,a.handle,r.reason,p.details,p.hidden FROM matrimony_reports r JOIN member_accounts a ON a.id=r.subject JOIN matrimony_profiles p ON p.account_id=r.subject WHERE r.status='open' ORDER BY r.id LIMIT 100`)
}

func (s *Server) moderateMatrimony(w http.ResponseWriter, r *http.Request, id string) {
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
	s.memberExec(w, r, `WITH reported AS (UPDATE matrimony_reports SET status=CASE WHEN $2 THEN 'hidden' ELSE 'dismissed' END WHERE id=$1 AND status='open' RETURNING subject), changed AS (UPDATE matrimony_profiles SET hidden=$2 WHERE account_id IN(SELECT subject FROM reported) RETURNING account_id) INSERT INTO member_audit(actor,action,subject) SELECT $3,'matrimony_moderation',account_id FROM changed`, in.Report, in.Hide, id)
}
