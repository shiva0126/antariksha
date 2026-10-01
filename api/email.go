package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// EmailSender is injected in tests. Production only enables delivery with a
// complete, validated SMTP configuration; no tokens are returned or logged.
type EmailSender interface {
	Send(context.Context, string, string, string) error
}
type smtpSender struct{ host, port, user, password, from string }

func configuredSMTP() EmailSender {
	s := smtpSender{os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"), os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASSWORD"), os.Getenv("SMTP_FROM")}
	if s.host == "" || s.port == "" || s.user == "" || s.password == "" || s.from == "" || strings.ContainsAny(s.host+s.port+s.from, "\r\n") {
		return nil
	}
	if _, err := mail.ParseAddress(s.from); err != nil {
		return nil
	}
	if publicEmailOrigin() == "" {
		return nil
	}
	return s
}
func publicEmailOrigin() string {
	raw := os.Getenv("PUBLIC_ORIGIN")
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return ""
	}
	return strings.TrimRight(raw, "/")
}
func (s smtpSender) Send(ctx context.Context, to, subject, body string) error {
	if strings.ContainsAny(to+subject, "\r\n") {
		return fmt.Errorf("invalid mail header")
	}
	from, err := mail.ParseAddress(s.from)
	if err != nil {
		return err
	}
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp", net.JoinHostPort(s.host, s.port))
	if err != nil {
		return err
	}
	defer conn.Close()
	deadline := time.Now().Add(20 * time.Second)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	_ = conn.SetDeadline(deadline)
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("SMTP server must support STARTTLS")
	}
	if err = client.StartTLS(&tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12}); err != nil {
		return err
	}
	if err = client.Auth(smtp.PlainAuth("", s.user, s.password, s.host)); err != nil {
		return err
	}
	if err = client.Mail(from.Address); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	_, err = io.WriteString(writer, "From: "+s.from+"\r\nTo: "+to+"\r\nSubject: "+subject+"\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n"+body)
	if err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}
func (s *Server) emailRoutes() {
	s.mux.HandleFunc("GET /api/auth/options", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]bool{"email_delivery_available": s.mailer != nil})
	})
	s.mux.Handle("POST /api/auth/email/reset-request", s.limit(s.accountGuard(s.requestPasswordEmail), 5))
	s.mux.Handle("POST /api/auth/email/reset", s.limit(s.accountGuard(s.resetPasswordEmail), 10))
	s.mux.Handle("POST /api/auth/email/verify", s.limit(s.accountGuard(s.verifyEmail), 10))
	s.memberRoute("POST /api/me/email/verification", s.requestVerificationEmail)
}
func (s *Server) requestPasswordEmail(w http.ResponseWriter, r *http.Request) {
	if s.mailer == nil {
		problem(w, 503, fmt.Errorf("email delivery is not configured; use your recovery key"))
		return
	}
	var in struct {
		Email string `json:"email"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if len(email) > 254 {
		problem(w, 400, fmt.Errorf("invalid email"))
		return
	}
	var id string
	// Same response for unknown accounts, throttled requests, or provider failures.
	if err := s.membersDB().QueryRow(r.Context(), `SELECT id FROM member_accounts WHERE lower(email)=$1`, email).Scan(&id); err == nil {
		s.sendAccountEmail(r, id, email, "reset")
	}
	writeJSON(w, 202, map[string]string{"message": "If an eligible account exists, a reset link will be sent. Please check your inbox."})
}
func (s *Server) requestVerificationEmail(w http.ResponseWriter, r *http.Request, id string) {
	if s.mailer == nil {
		problem(w, 503, fmt.Errorf("email delivery is not configured"))
		return
	}
	var email string
	if err := s.membersDB().QueryRow(r.Context(), `SELECT email FROM member_accounts WHERE id=$1 AND email_verified_at IS NULL`, id).Scan(&email); err == nil {
		s.sendAccountEmail(r, id, email, "verify")
	}
	writeJSON(w, 202, map[string]string{"message": "If verification is needed, a link will be sent. Please check your inbox."})
}
func (s *Server) sendAccountEmail(r *http.Request, id, email, purpose string) {
	token := randomToken()
	hash := sessionHash(token)
	// Persistent per-account cooldown survives process restarts and concurrent requests.
	tag, err := s.membersDB().Exec(r.Context(), `INSERT INTO member_email_tokens(account_id,purpose,token_hash,email,expires_at) VALUES($1,$2,$3,$4,now()+interval '30 minutes') ON CONFLICT(account_id,purpose) DO UPDATE SET token_hash=$3,email=$4,created_at=now(),expires_at=now()+interval '30 minutes' WHERE member_email_tokens.created_at < now()-interval '5 minutes'`, id, purpose, hash, email)
	if err != nil || tag.RowsAffected() != 1 {
		return
	}
	subject := "Verify your Astrisk email"
	action := "verify"
	if purpose == "reset" {
		subject = "Reset your Astrisk password"
		action = "reset"
	}
	body := subject + "\r\n\r\nOpen this link within 30 minutes:\r\n" + publicEmailOrigin() + "/#" + action + "/" + token + "\r\n\r\nIf you did not request this, ignore this email."
	if err = s.mailer.Send(r.Context(), email, subject, body); err != nil {
		s.logger.Warn("account email delivery failed")
		// Keep cooldown but invalidate undelivered tokens.
		_, _ = s.membersDB().Exec(r.Context(), `UPDATE member_email_tokens SET expires_at=now() WHERE token_hash=$1`, hash)
	}
}

type emailTokenInput struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (s *Server) verifyEmail(w http.ResponseWriter, r *http.Request) {
	s.consumeEmailToken(w, r, "verify")
}
func (s *Server) resetPasswordEmail(w http.ResponseWriter, r *http.Request) {
	s.consumeEmailToken(w, r, "reset")
}
func (s *Server) consumeEmailToken(w http.ResponseWriter, r *http.Request, purpose string) {
	var in emailTokenInput
	if !memberInput(w, r, &in) {
		return
	}
	if len(in.Token) != 64 || (purpose == "reset" && (len(in.Password) < 12 || len(in.Password) > 72)) {
		problem(w, 400, fmt.Errorf("valid link and a 12–72 byte password required"))
		return
	}
	var passwordHash []byte
	var err error
	if purpose == "reset" {
		passwordHash, err = bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			problem(w, 500, fmt.Errorf("password reset unavailable"))
			return
		}
	}
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("request unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	var id, email string
	err = tx.QueryRow(r.Context(), `SELECT t.account_id,t.email FROM member_email_tokens t JOIN member_accounts a ON a.id=t.account_id WHERE t.token_hash=$1 AND t.purpose=$2 AND t.expires_at>now() AND t.email=a.email FOR UPDATE OF a,t`, sessionHash(in.Token), purpose).Scan(&id, &email)
	if err != nil {
		problem(w, 400, fmt.Errorf("link is invalid, expired or already used"))
		return
	}
	if purpose == "reset" {
		_, err = tx.Exec(r.Context(), `UPDATE member_accounts SET password_hash=$1,email_verified_at=now(),recovery_hash=NULL WHERE id=$2`, string(passwordHash), id)
		if err == nil {
			_, err = tx.Exec(r.Context(), `DELETE FROM member_sessions WHERE account_id=$1`, id)
		}
	} else {
		_, err = tx.Exec(r.Context(), `UPDATE member_accounts SET email_verified_at=now() WHERE id=$1`, id)
	}
	if err == nil {
		_, err = tx.Exec(r.Context(), `DELETE FROM member_email_tokens WHERE account_id=$1 AND (purpose=$2 OR $2='reset')`, id, purpose)
	}
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 500, fmt.Errorf("request unavailable"))
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}
