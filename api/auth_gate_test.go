package api

import (
	"net/http/httptest"
	"testing"
)

func TestAuthenticationBoundary(t *testing.T) {
	s := NewServer(fakeCalc{}, NoCache{}, nil)
	for _, path := range []string{"/api/chart", "/api/panchang", "/api/month", "/api/chat/history", "/api/me", "/api/community/feed"} {
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 401 {
			t.Errorf("%s: got %d", path, w.Code)
		}
	}
}
