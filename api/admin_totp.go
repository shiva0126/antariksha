package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func totpCipher() (cipher.AEAD, error) {
	path := os.Getenv("ADMIN_TOTP_KEY_FILE")
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("admin authenticator key unavailable")
	}
	key, err := os.ReadFile(path)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("admin authenticator key unavailable")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// RFC 6238 / HOTP dynamic truncation; SHA-1 is the authenticator interoperability default.
func totpCode(secret []byte, counter int64, digits int) string {
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(counter))
	mac := hmac.New(sha1.New, secret)
	mac.Write(msg[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 15
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	mod := uint32(1000000)
	if digits == 8 {
		mod = 100000000
	}
	return fmt.Sprintf("%0*d", digits, value%mod)
}

func matchTOTP(secret []byte, code string, now time.Time, last int64) (int64, bool) {
	if len(code) != 6 {
		return 0, false
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	counter := now.Unix() / 30
	for _, n := range []int64{counter, counter - 1, counter + 1} {
		if n > last && subtle.ConstantTimeCompare([]byte(totpCode(secret, n, 6)), []byte(code)) == 1 {
			return n, true
		}
	}
	return 0, false
}

func (s *Server) totpStatus(w http.ResponseWriter, r *http.Request, id string) {
	if s.accountRole(r, id) != "superadmin" {
		problem(w, 403, fmt.Errorf("superadmin access required"))
		return
	}
	var enabled bool
	err := s.membersDB().QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM member_totp WHERE account_id=$1 AND enabled)`, id).Scan(&enabled)
	if err != nil {
		problem(w, 503, fmt.Errorf("authenticator status unavailable"))
		return
	}
	_, keyErr := totpCipher()
	writeJSON(w, 200, map[string]bool{"enabled": enabled, "available": keyErr == nil})
}

func (s *Server) manageTOTP(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Action   string `json:"action"`
		Password string `json:"password"`
		Code     string `json:"code"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if in.Action != "begin" && in.Action != "confirm" && in.Action != "disable" {
		problem(w, 400, fmt.Errorf("invalid authenticator action"))
		return
	}
	tx, ok := s.lockSecurity(w, r, id, in.Password)
	if !ok {
		return
	}
	defer tx.Rollback(r.Context())
	var email, role string
	if err := tx.QueryRow(r.Context(), `SELECT COALESCE(email,handle),role FROM member_accounts WHERE id=$1`, id).Scan(&email, &role); err != nil || role != "superadmin" {
		problem(w, 403, fmt.Errorf("superadmin access required"))
		return
	}
	box, err := totpCipher()
	if err != nil {
		problem(w, 503, fmt.Errorf("server authenticator encryption is not configured"))
		return
	}
	var encrypted []byte
	var enabled bool
	var until time.Time
	var last int64
	err = tx.QueryRow(r.Context(), `SELECT secret,enabled,pending_until,last_counter FROM member_totp WHERE account_id=$1 FOR UPDATE`, id).Scan(&encrypted, &enabled, &until, &last)
	if err != nil && err != pgx.ErrNoRows {
		problem(w, 503, fmt.Errorf("authenticator unavailable"))
		return
	}
	result := map[string]any{"ok": true}
	if in.Action == "begin" {
		if enabled {
			problem(w, 409, fmt.Errorf("authenticator is already enabled"))
			return
		}
		secret := make([]byte, 20)
		nonce := make([]byte, box.NonceSize())
		if _, err = rand.Read(secret); err != nil {
			problem(w, 503, fmt.Errorf("authenticator unavailable"))
			return
		}
		if _, err = rand.Read(nonce); err != nil {
			problem(w, 503, fmt.Errorf("authenticator unavailable"))
			return
		}
		encrypted = box.Seal(nonce, nonce, secret, []byte(id))
		_, err = tx.Exec(r.Context(), `INSERT INTO member_totp(account_id,secret) VALUES($1,$2) ON CONFLICT(account_id) DO UPDATE SET secret=$2,pending_until=now()+interval '10 minutes',last_counter=-1`, id, encrypted)
		encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
		result["secret"] = encoded
		result["uri"] = "otpauth://totp/" + url.PathEscape("Astrisk:"+email) + "?" + url.Values{"secret": {encoded}, "issuer": {"Astrisk"}, "algorithm": {"SHA1"}, "digits": {"6"}, "period": {"30"}}.Encode()
	} else {
		if err == pgx.ErrNoRows || len(encrypted) < box.NonceSize() || (in.Action == "confirm" && (enabled || time.Now().After(until))) || (in.Action == "disable" && !enabled) {
			problem(w, 409, fmt.Errorf("authenticator setup missing, expired or already completed"))
			return
		}
		secret, openErr := box.Open(nil, encrypted[:box.NonceSize()], encrypted[box.NonceSize():], []byte(id))
		if openErr != nil {
			problem(w, 503, fmt.Errorf("authenticator key unavailable"))
			return
		}
		counter, valid := matchTOTP(secret, in.Code, time.Now(), last)
		if !valid {
			problem(w, 403, fmt.Errorf("invalid or already used authenticator code; try the next code"))
			return
		}
		if in.Action == "confirm" {
			_, err = tx.Exec(r.Context(), `UPDATE member_totp SET enabled=true,last_counter=$2 WHERE account_id=$1`, id, counter)
		} else {
			_, err = tx.Exec(r.Context(), `DELETE FROM member_totp WHERE account_id=$1`, id)
		}
		if err == nil {
			_, err = tx.Exec(r.Context(), `DELETE FROM member_sessions WHERE account_id=$1`, id)
		}
		result["signed_out"] = true
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `INSERT INTO member_audit(actor,action,subject,details) VALUES($1,$2,$1,'{}')`, id, "admin_totp_"+in.Action)
	}
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 503, fmt.Errorf("authenticator change could not be saved"))
		return
	}
	if in.Action != "begin" {
		clearSessionCookie(w, r)
	}
	writeJSON(w, 200, result)
}

// Invoked under the account lock as part of the session-creation transaction.
// Password resets do not remove this factor; there is no public MFA bypass.
func checkLoginTOTP(r *http.Request, tx pgx.Tx, id, code string) error {
	var encrypted []byte
	var last int64
	err := tx.QueryRow(r.Context(), `SELECT secret,last_counter FROM member_totp WHERE account_id=$1 AND enabled FOR UPDATE`, id).Scan(&encrypted, &last)
	if err == pgx.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	box, err := totpCipher()
	if err != nil {
		return err
	}
	if len(encrypted) < box.NonceSize() {
		return fmt.Errorf("invalid authenticator storage")
	}
	secret, err := box.Open(nil, encrypted[:box.NonceSize()], encrypted[box.NonceSize():], []byte(id))
	if err != nil {
		return err
	}
	counter, ok := matchTOTP(secret, code, time.Now(), last)
	if !ok {
		return fmt.Errorf("invalid authenticator code")
	}
	_, err = tx.Exec(r.Context(), `UPDATE member_totp SET last_counter=$2 WHERE account_id=$1`, id, counter)
	return err
}
