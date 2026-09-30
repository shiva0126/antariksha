package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func secureMemberCookie(r *http.Request) bool {
	return r.TLS != nil || os.Getenv("COOKIE_SECURE") == "true"
}
func (s *Server) securityRoutes() {
	s.memberRoute("GET /api/me/security", s.memberSecurity)
	s.memberRoute("POST /api/me/recovery-key", s.recoveryKey)
	s.memberRoute("POST /api/me/logout-all", s.logoutAll)
	s.memberRoute("GET /api/me/export", s.exportMember)
	s.memberRoute("POST /api/me/phone/start", s.phoneStart)
	s.memberRoute("POST /api/me/phone/check", s.phoneCheck)
	s.mux.Handle("POST /api/auth/recover", s.limit(s.accountGuard(s.recoverMember), 5))
}
func (s *Server) memberSecurity(w http.ResponseWriter, r *http.Request, id string) {
	var sessions int
	var recovery, verified bool
	_ = s.membersDB().QueryRow(r.Context(), `SELECT count(*) FROM member_sessions WHERE account_id=$1 AND expires_at>now()`, id).Scan(&sessions)
	_ = s.membersDB().QueryRow(r.Context(), `SELECT recovery_hash IS NOT NULL FROM member_accounts WHERE id=$1`, id).Scan(&recovery)
	_ = s.membersDB().QueryRow(r.Context(), `SELECT verified FROM member_phone WHERE account_id=$1`, id).Scan(&verified)
	writeJSON(w, 200, map[string]any{"sessions": sessions, "recovery_key_created": recovery, "phone_verified": verified, "phone_available": phoneReady(), "moderator": s.isModerator(r, id)})
}
func (s *Server) passwordOK(r *http.Request, id, password string) bool {
	var hash string
	err := s.membersDB().QueryRow(r.Context(), `SELECT password_hash FROM member_accounts WHERE id=$1`, id).Scan(&hash)
	return err == nil && bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
func (s *Server) recoveryKey(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Password string `json:"password"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if !s.passwordOK(r, id, in.Password) {
		problem(w, 401, fmt.Errorf("password required"))
		return
	}
	token := randomToken()
	_, err := s.membersDB().Exec(r.Context(), `UPDATE member_accounts SET recovery_hash=$1 WHERE id=$2`, sessionHash(token), id)
	if err != nil {
		problem(w, 500, fmt.Errorf("recovery unavailable"))
		return
	}
	writeJSON(w, 200, map[string]string{"recovery_key": token})
}
func (s *Server) recoverMember(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Handle   string `json:"handle"`
		Key      string `json:"key"`
		Password string `json:"password"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if len(in.Key) != 64 || len(in.Password) < 12 || len(in.Password) > 72 {
		problem(w, 400, fmt.Errorf("recovery key and 12–72 byte password required"))
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		problem(w, 500, fmt.Errorf("recovery unavailable"))
		return
	}
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("recovery unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	var id string
	err = tx.QueryRow(r.Context(), `UPDATE member_accounts SET password_hash=$1,recovery_hash=NULL WHERE (lower(email)=$2 OR (email IS NULL AND handle=$2)) AND recovery_hash=$3 RETURNING id`, string(hash), strings.ToLower(strings.TrimSpace(in.Handle)), sessionHash(in.Key)).Scan(&id)
	if err != nil {
		problem(w, 401, fmt.Errorf("invalid recovery details"))
		return
	}
	_, err = tx.Exec(r.Context(), `DELETE FROM member_sessions WHERE account_id=$1`, id)
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 500, fmt.Errorf("recovery unavailable"))
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
func (s *Server) logoutAll(w http.ResponseWriter, r *http.Request, id string) {
	s.memberExec(w, r, `DELETE FROM member_sessions WHERE account_id=$1`, id)
}
func (s *Server) exportMember(w http.ResponseWriter, r *http.Request, id string) {
	// Explicit allowlists exclude password hashes, tokens and other members' private messages.
	s.memberRows(w, r, `SELECT a.handle,a.email,a.birth_date,a.birth_time,a.profile,a.created_at,
 COALESCE((SELECT json_agg(json_build_object('kind',n.kind,'created_at',n.created_at,'read_at',n.read_at,'dismissed',n.dismissed)) FROM member_notifications n WHERE n.recipient=a.id),'[]') notifications,
 (SELECT to_jsonb(s)-'account_id' FROM member_settings s WHERE s.account_id=a.id) settings,
 COALESCE((SELECT json_agg(json_build_object('id',p.id,'caption',p.caption,'audience',p.audience,'created_at',p.created_at)) FROM community_posts p WHERE p.owner=a.id),'[]') posts,
 COALESCE((SELECT json_agg(json_build_object('id',c.id,'body',c.body,'post_id',c.post_id)) FROM community_comments c WHERE c.owner=a.id),'[]') comments,
 COALESCE((SELECT json_agg(json_build_object('body',m.body,'created_at',m.created_at)) FROM member_messages m WHERE m.sender=a.id),'[]') sent_messages,
 (SELECT details FROM matrimony_profiles WHERE account_id=a.id) matrimony,
 COALESCE((SELECT json_agg(json_build_object('name',f.name,'people',(SELECT json_agg(json_build_object('name',p.name,'note',p.note)) FROM family_people p WHERE p.group_id=f.id))) FROM family_groups f WHERE f.owner=a.id),'[]') owned_families
 FROM member_accounts a WHERE a.id=$1`, id)
}

var phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

func phoneKey() []byte {
	key, _ := hex.DecodeString(os.Getenv("CONTACT_ENCRYPTION_KEY"))
	if len(key) != 32 {
		return nil
	}
	return key
}
func phoneReady() bool {
	return phoneKey() != nil && os.Getenv("TWILIO_ACCOUNT_SID") != "" && os.Getenv("TWILIO_AUTH_TOKEN") != "" && regexp.MustCompile(`^VA[0-9a-fA-F]{32}$`).MatchString(os.Getenv("TWILIO_VERIFY_SERVICE_SID"))
}
func phoneDigest(phone string) string {
	h := hmac.New(sha256.New, phoneKey())
	h.Write([]byte("phone-lookup:" + phone))
	return hex.EncodeToString(h.Sum(nil))
}
func encryptPhone(phone string) ([]byte, error) {
	b, err := aes.NewCipher(phoneKey())
	if err != nil {
		return nil, err
	}
	g, err := cipher.NewGCM(b)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, g.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return g.Seal(nonce, nonce, []byte(phone), []byte("member-phone-v1")), nil
}
func decryptPhone(raw []byte) (string, error) {
	b, err := aes.NewCipher(phoneKey())
	if err != nil {
		return "", err
	}
	g, err := cipher.NewGCM(b)
	if err != nil || len(raw) < g.NonceSize() {
		return "", fmt.Errorf("invalid phone ciphertext")
	}
	plain, err := g.Open(nil, raw[:g.NonceSize()], raw[g.NonceSize():], []byte("member-phone-v1"))
	return string(plain), err
}
func verifyPhoneProvider(r *http.Request, path string, values url.Values) (string, error) {
	req, err := http.NewRequestWithContext(r.Context(), "POST", "https://verify.twilio.com/v2/Services/"+os.Getenv("TWILIO_VERIFY_SERVICE_SID")+"/"+path, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(os.Getenv("TWILIO_ACCOUNT_SID"), os.Getenv("TWILIO_AUTH_TOKEN"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("verification provider unavailable")
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return "", fmt.Errorf("verification provider rejected request")
	}
	var out struct {
		Status string `json:"status"`
	}
	err = json.NewDecoder(res.Body).Decode(&out)
	return out.Status, err
}
func (s *Server) phoneStart(w http.ResponseWriter, r *http.Request, id string) {
	if !phoneReady() {
		problem(w, 503, fmt.Errorf("SMS verification requires configured provider credentials and an encryption key"))
		return
	}
	var in struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Consent  bool   `json:"consent"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if !in.Consent || !phonePattern.MatchString(in.Phone) || !s.passwordOK(r, id, in.Password) {
		problem(w, 400, fmt.Errorf("international phone number, password and SMS consent required"))
		return
	}
	encrypted, err := encryptPhone(in.Phone)
	if err != nil {
		problem(w, 500, fmt.Errorf("phone encryption unavailable"))
		return
	}
	// Persist cooldown before contacting the paid provider, including failed attempts.
	budget, err := s.membersDB().Exec(r.Context(), `INSERT INTO member_sms_budget(account_id,day,sends) VALUES($1,CURRENT_DATE,1) ON CONFLICT(account_id,day) DO UPDATE SET sends=member_sms_budget.sends+1 WHERE member_sms_budget.sends<5`, id)
	if err != nil || budget.RowsAffected() != 1 {
		problem(w, 429, fmt.Errorf("daily SMS limit reached"))
		return
	}
	tag, err := s.membersDB().Exec(r.Context(), `INSERT INTO member_phone(account_id,phone_hash,ciphertext,sent_at) VALUES($1,$2,$3,now()) ON CONFLICT(account_id) DO UPDATE SET phone_hash=$2,ciphertext=$3,verified=false,attempts=0,sent_at=now() WHERE member_phone.sent_at<now()-interval '10 minutes'`, id, phoneDigest(in.Phone), encrypted)
	if err != nil || tag.RowsAffected() != 1 {
		problem(w, 429, fmt.Errorf("verification unavailable; wait 10 minutes before retrying"))
		return
	}
	status, err := verifyPhoneProvider(r, "Verifications", url.Values{"To": {in.Phone}, "Channel": {"sms"}})
	if err != nil || status != "pending" {
		problem(w, 502, fmt.Errorf("SMS could not be sent; check provider configuration"))
		return
	}
	writeJSON(w, 200, map[string]bool{"sent": true})
}
func (s *Server) phoneCheck(w http.ResponseWriter, r *http.Request, id string) {
	if !phoneReady() {
		problem(w, 503, fmt.Errorf("SMS verification unavailable"))
		return
	}
	var in struct {
		Code string `json:"code"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if !regexp.MustCompile(`^[0-9]{4,10}$`).MatchString(in.Code) {
		problem(w, 400, fmt.Errorf("invalid code"))
		return
	}
	var raw []byte
	var digest string
	err := s.membersDB().QueryRow(r.Context(), `UPDATE member_phone SET attempts=attempts+1 WHERE account_id=$1 AND NOT verified AND attempts<5 AND sent_at>now()-interval '10 minutes' RETURNING ciphertext,phone_hash`, id).Scan(&raw, &digest)
	if err != nil {
		problem(w, 400, fmt.Errorf("verification expired or attempts exhausted"))
		return
	}
	phone, err := decryptPhone(raw)
	if err != nil {
		problem(w, 500, fmt.Errorf("phone unavailable"))
		return
	}
	status, err := verifyPhoneProvider(r, "VerificationCheck", url.Values{"To": {phone}, "Code": {in.Code}})
	if err != nil || status != "approved" {
		problem(w, 400, fmt.Errorf("code incorrect or expired"))
		return
	}
	s.memberExec(w, r, `UPDATE member_phone SET verified=true WHERE account_id=$1 AND phone_hash=$2 AND NOT verified`, id, digest)
}
