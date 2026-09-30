package api

import (
	"fmt"
	"net/http"
	"net/url"
)

// An existing session ID or matching birth details never establish ownership.
func (s *Server) chatAccess(w http.ResponseWriter, r *http.Request, sid string, create bool) bool {
	// Inner-router unit tests use memory storage; the public auth gate still rejects guests.
	if _, ok := s.cache.(PostgresCache); !ok {
		return true
	}
	if r.Method != "GET" {
		origin, err := url.Parse(r.Header.Get("Origin"))
		if err != nil || origin.Host != r.Host || (origin.Scheme != "http" && origin.Scheme != "https") || r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			problem(w, 403, fmt.Errorf("same-origin request required"))
			return false
		}
	}
	id, err := s.memberID(r)
	if err != nil {
		problem(w, 401, fmt.Errorf("sign in required"))
		return false
	}
	if create {
		_, err = s.membersDB().Exec(r.Context(), `INSERT INTO chat_owners(session_id,account_id) VALUES($1,$2)`, sid, id)
		if err != nil {
			problem(w, 500, fmt.Errorf("chat ownership unavailable"))
			return false
		}
		return true
	}
	var ok bool
	err = s.membersDB().QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM chat_owners WHERE session_id=$1 AND account_id=$2)`, sid, id).Scan(&ok)
	if err == nil && ok {
		return true
	}
	problem(w, 404, fmt.Errorf("chat session unavailable in this account"))
	return false
}
