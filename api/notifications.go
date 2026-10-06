package api

import (
	"fmt"
	"net/http"
	"strconv"
)

func (s *Server) notificationRoutes() {
	s.memberRoute("GET /api/me/notifications", s.notifications)
	s.memberRoute("POST /api/me/notifications/{notification}", s.updateNotification)
}
func (s *Server) notifications(w http.ResponseWriter, r *http.Request, id string) {
	before := int64(9223372036854775807)
	if raw := r.URL.Query().Get("before"); raw != "" {
		var err error
		before, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || before < 1 {
			problem(w, 400, fmt.Errorf("invalid notification cursor"))
			return
		}
	}
	s.memberRows(w, r, `SELECT n.id,n.kind,a.handle,CASE WHEN n.kind='transit' THEN n.subject ELSE '' END AS subject,n.created_at,n.read_at IS NOT NULL AS read FROM member_notifications n JOIN member_accounts a ON a.id=n.actor WHERE n.id<$2 AND notification_visible(n,$1) ORDER BY n.id DESC LIMIT 50`, id, before)
}
func (s *Server) updateNotification(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Action string `json:"action"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if in.Action != "read" && in.Action != "dismiss" {
		problem(w, 400, fmt.Errorf("choose read or dismiss"))
		return
	}
	s.memberExec(w, r, `UPDATE member_notifications n SET read_at=COALESCE(read_at,now()),dismissed=($3='dismiss') WHERE n.id=$2 AND notification_visible(n,$1)`, id, r.PathValue("notification"), in.Action)
}
