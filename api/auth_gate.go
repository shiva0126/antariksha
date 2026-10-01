package api

import (
	"fmt"
	"net/http"
	"strings"
)

// Authentication is enforced at the API boundary, not just by hiding the UI.
func (s *Server) requireAccount(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
			public := r.URL.Path == "/api/auth/register" || r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/recover"
			public = public || r.URL.Path == "/api/auth/options" || r.URL.Path == "/api/auth/email/reset-request" || r.URL.Path == "/api/auth/email/reset" || r.URL.Path == "/api/auth/email/verify"
			if !public {
				if _, err := s.memberID(r); err != nil {
					problem(w, 401, fmt.Errorf("please sign in to use Astrisk"))
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
