package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// Alerts tell members about new interests, accepted interests, messages and
// family or follow requests by web push and email. Notifications are already
// recorded in member_notifications by database triggers; this worker delivers
// each one once, after a short delay so that a member who is already online
// and reads it in the app is not emailed. Alerts never contain message text.

const alertDelay = 90 * time.Second

func (s *Server) alertRoutes() {
	s.memberRoute("GET /api/push/key", s.pushKey)
	s.memberRoute("POST /api/push/subscribe", s.pushSubscribe)
	s.memberRoute("DELETE /api/push/subscribe", s.pushUnsubscribe)
}

// RunAlerts delivers pending alerts every minute until ctx ends.
func (s *Server) RunAlerts(ctx context.Context) {
	if _, ok := s.cache.(PostgresCache); !ok {
		return
	}
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		if n, err := s.deliverAlerts(ctx); err != nil {
			s.logger.Error("alerts", "error", err)
		} else if n > 0 {
			s.logger.Info("alerts delivered", "count", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

type pendingAlert struct {
	id                    int64
	recipient, kind, from string
	email                 string
	emailOK               bool
}

func alertText(kind, from string) string {
	switch kind {
	case "interest":
		return from + " is interested in your matrimony profile"
	case "accepted":
		return from + " accepted your interest — you can now chat"
	case "message":
		return "New message from " + from
	case "family":
		return from + " invited you to a family group"
	case "follow":
		return from + " asked to follow you"
	}
	return "New activity from " + from
}

func alertPage(kind string) string {
	if kind == "family" || kind == "follow" {
		return "#community"
	}
	return "#matrimony"
}

func (s *Server) deliverAlerts(ctx context.Context) (int, error) {
	rows, err := s.membersDB().Query(ctx, `SELECT n.id,n.recipient,n.kind,COALESCE(NULLIF(a.profile->>'name',''),'@'||a.handle),
 COALESCE(r.email,''),(r.email_verified_at IS NOT NULL AND r.email_alerts)
 FROM member_notifications n JOIN member_accounts a ON a.id=n.actor JOIN member_accounts r ON r.id=n.recipient
 WHERE n.delivered_at IS NULL AND n.read_at IS NULL AND NOT n.dismissed
 AND n.created_at < now()-$1::interval AND n.created_at > now()-interval '2 days'
 AND notification_visible(n,n.recipient) ORDER BY n.id LIMIT 200`, fmt.Sprintf("%d seconds", int(alertDelay.Seconds())))
	if err != nil {
		return 0, err
	}
	var alerts []pendingAlert
	for rows.Next() {
		var a pendingAlert
		if err = rows.Scan(&a.id, &a.recipient, &a.kind, &a.from, &a.email, &a.emailOK); err != nil {
			rows.Close()
			return 0, err
		}
		alerts = append(alerts, a)
	}
	rows.Close()
	if len(alerts) == 0 {
		return 0, nil
	}
	byRecipient := map[string][]pendingAlert{}
	var order []string
	for _, a := range alerts {
		if _, ok := byRecipient[a.recipient]; !ok {
			order = append(order, a.recipient)
		}
		byRecipient[a.recipient] = append(byRecipient[a.recipient], a)
	}
	origin := publicEmailOrigin()
	if origin == "" {
		origin = strings.TrimRight(os.Getenv("PUBLIC_ORIGIN"), "/")
	}
	ids := make([]int64, 0, len(alerts))
	for _, rcpt := range order {
		list := byRecipient[rcpt]
		lines, seen := []string{}, map[string]bool{}
		for _, a := range list {
			ids = append(ids, a.id)
			if t := alertText(a.kind, a.from); !seen[t] {
				seen[t] = true
				lines = append(lines, t)
			}
		}
		first := list[0]
		s.sendPush(ctx, rcpt, lines[0], len(lines), origin+"/"+alertPage(first.kind))
		if s.mailer != nil && first.emailOK && first.email != "" && origin != "" {
			body := "Hello,\n\nYou have new activity on Astrisk:\n\n- " + strings.Join(lines, "\n- ") +
				"\n\nOpen Astrisk: " + origin + "/" + alertPage(first.kind) +
				"\n\nYou receive these emails because alerts are on in your matrimony profile. Turn them off there at any time.\n"
			mctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			if err := s.mailer.Send(mctx, first.email, "Astrisk: "+lines[0], body); err != nil {
				s.logger.Warn("alert email failed", "error", err)
			}
			cancel()
		}
	}
	_, err = s.membersDB().Exec(ctx, `UPDATE member_notifications SET delivered_at=now() WHERE id=ANY($1)`, ids)
	return len(ids), err
}

// ---- web push ---------------------------------------------------------------

// vapidKeys returns the server's VAPID key pair, creating it once.
func (s *Server) vapidKeys(ctx context.Context) (pub, priv string, err error) {
	db := s.membersDB()
	err = db.QueryRow(ctx, `SELECT (SELECT value FROM app_secrets WHERE name='vapid_public'),(SELECT value FROM app_secrets WHERE name='vapid_private')`).Scan(&pub, &priv)
	if err == nil && pub != "" && priv != "" {
		return pub, priv, nil
	}
	priv, pub, err = webpush.GenerateVAPIDKeys()
	if err != nil {
		return "", "", err
	}
	_, err = db.Exec(ctx, `INSERT INTO app_secrets(name,value) VALUES('vapid_public',$1),('vapid_private',$2) ON CONFLICT(name) DO NOTHING`, pub, priv)
	if err != nil {
		return "", "", err
	}
	// Another process may have won the race: read back what is stored.
	err = db.QueryRow(ctx, `SELECT (SELECT value FROM app_secrets WHERE name='vapid_public'),(SELECT value FROM app_secrets WHERE name='vapid_private')`).Scan(&pub, &priv)
	return pub, priv, err
}

func (s *Server) pushKey(w http.ResponseWriter, r *http.Request, id string) {
	pub, _, err := s.vapidKeys(r.Context())
	if err != nil {
		problem(w, 503, fmt.Errorf("push notifications unavailable"))
		return
	}
	writeJSON(w, 200, map[string]string{"key": pub})
}

// Only the browser vendors' push services are contacted, so a subscription
// cannot make this server send requests to arbitrary hosts.
var pushHosts = []string{"fcm.googleapis.com", "updates.push.services.mozilla.com", ".notify.windows.com", ".push.apple.com", "web.push.apple.com"}

func validPushEndpoint(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || len(raw) > 1000 || u.Port() != "" {
		return false
	}
	h := strings.ToLower(u.Hostname())
	for _, p := range pushHosts {
		if h == strings.TrimPrefix(p, ".") || (strings.HasPrefix(p, ".") && strings.HasSuffix(h, p)) {
			return true
		}
	}
	return false
}

func (s *Server) pushSubscribe(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Endpoint string `json:"endpoint"`
		Keys     struct {
			P256dh string `json:"p256dh"`
			Auth   string `json:"auth"`
		} `json:"keys"`
		ExpirationTime any `json:"expirationTime"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if !validPushEndpoint(in.Endpoint) || len(in.Keys.P256dh) < 40 || len(in.Keys.P256dh) > 200 || len(in.Keys.Auth) < 10 || len(in.Keys.Auth) > 100 {
		problem(w, 400, fmt.Errorf("unsupported push subscription"))
		return
	}
	var n int
	if s.membersDB().QueryRow(r.Context(), `SELECT count(*) FROM member_push_subscriptions WHERE account_id=$1`, id).Scan(&n) == nil && n >= 10 {
		_, _ = s.membersDB().Exec(r.Context(), `DELETE FROM member_push_subscriptions WHERE endpoint IN (SELECT endpoint FROM member_push_subscriptions WHERE account_id=$1 ORDER BY created_at LIMIT 1)`, id)
	}
	s.memberExec(w, r, `INSERT INTO member_push_subscriptions(endpoint,account_id,p256dh,auth) VALUES($1,$2,$3,$4) ON CONFLICT(endpoint) DO UPDATE SET account_id=$2,p256dh=$3,auth=$4,created_at=now()`, in.Endpoint, id, in.Keys.P256dh, in.Keys.Auth)
}

func (s *Server) pushUnsubscribe(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Endpoint string `json:"endpoint"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	s.memberExec(w, r, `DELETE FROM member_push_subscriptions WHERE endpoint=$1 AND account_id=$2`, in.Endpoint, id)
}

// pushSender is replaced in tests.
var pushSender = func(ctx context.Context, payload []byte, sub *webpush.Subscription, opts *webpush.Options) (int, error) {
	resp, err := webpush.SendNotificationWithContext(ctx, payload, sub, opts)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

func (s *Server) sendPush(ctx context.Context, account, text string, count int, link string) {
	rows, err := s.membersDB().Query(ctx, `SELECT endpoint,p256dh,auth FROM member_push_subscriptions WHERE account_id=$1`, account)
	if err != nil {
		return
	}
	var subs []webpush.Subscription
	for rows.Next() {
		var sub webpush.Subscription
		if rows.Scan(&sub.Endpoint, &sub.Keys.P256dh, &sub.Keys.Auth) == nil {
			subs = append(subs, sub)
		}
	}
	rows.Close()
	if len(subs) == 0 {
		return
	}
	pub, priv, err := s.vapidKeys(ctx)
	if err != nil {
		return
	}
	body := text
	if count > 1 {
		body = fmt.Sprintf("%s (and %d more)", text, count-1)
	}
	payload, _ := json.Marshal(map[string]string{"title": "Astrisk", "body": body, "url": link})
	subject := "https://astrisk.space"
	if o := strings.TrimRight(os.Getenv("PUBLIC_ORIGIN"), "/"); strings.HasPrefix(o, "https://") {
		subject = o
	}
	for i := range subs {
		pctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		code, err := pushSender(pctx, payload, &subs[i], &webpush.Options{Subscriber: subject, VAPIDPublicKey: pub, VAPIDPrivateKey: priv, TTL: 86400, Urgency: webpush.UrgencyNormal})
		cancel()
		if err == nil && (code == 404 || code == 410) {
			_, _ = s.membersDB().Exec(ctx, `DELETE FROM member_push_subscriptions WHERE endpoint=$1`, subs[i].Endpoint)
		}
	}
}
