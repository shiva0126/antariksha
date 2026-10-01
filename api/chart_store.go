package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"time"

	"github.com/example/panchang/engine"
)

type savedChart struct {
	engine.ChartInput
	ID    string `json:"id"`
	Name  string `json:"name"`
	Place string `json:"place"`
}
type chartStore struct {
	Profiles  []savedChart `json:"profiles"`
	Revision  int64        `json:"revision"`
	AccountID string       `json:"account_id"`
}

var chartID = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

func validSavedCharts(profiles []savedChart) bool {
	if len(profiles) > 100 {
		return false
	}
	seen := map[string]bool{}
	for _, p := range profiles {
		if !chartID.MatchString(p.ID) || seen[p.ID] || len(p.Name) > 100 || len(p.Place) > 200 || math.IsNaN(p.Lat) || math.IsInf(p.Lat, 0) || math.IsNaN(p.Lon) || math.IsInf(p.Lon, 0) || p.Lat < -90 || p.Lat > 90 || p.Lon < -180 || p.Lon > 180 || p.TZ == "" {
			return false
		}
		z, err := time.LoadLocation(p.TZ)
		if err != nil {
			return false
		}
		d, err := time.ParseInLocation("2006-01-02 15:04", p.Date+" "+p.Time, z)
		if err != nil || d.Year() < 1800 || d.Year() > 2399 || d.Format("2006-01-02 15:04") != p.Date+" "+p.Time {
			return false
		}
		seen[p.ID] = true
	}
	return true
}
func (s *Server) chartStoreRoutes() {
	s.memberRoute("GET /api/me/charts", s.getCharts)
	s.memberRoute("PUT /api/me/charts", s.putCharts)
}
func (s *Server) getCharts(w http.ResponseWriter, r *http.Request, id string) {
	var raw []byte
	var rev int64
	err := s.membersDB().QueryRow(r.Context(), `SELECT COALESCE(c.profiles,'[]'::jsonb),COALESCE(c.revision,0) FROM member_accounts a LEFT JOIN member_chart_store c ON c.account_id=a.id WHERE a.id=$1`, id).Scan(&raw, &rev)
	if err != nil {
		problem(w, 500, fmt.Errorf("saved charts unavailable"))
		return
	}
	writeJSON(w, 200, map[string]any{"profiles": json.RawMessage(raw), "revision": rev, "account_id": id})
}
func (s *Server) putCharts(w http.ResponseWriter, r *http.Request, id string) {
	var in chartStore
	// Bounded batch: 100 small birth-input records; never accepts calculated charts.
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128*1024))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil || in.AccountID != id || in.Profiles == nil || in.Revision < 0 || !validSavedCharts(in.Profiles) {
		problem(w, 400, fmt.Errorf("invalid saved charts or account (maximum 100)"))
		return
	}
	if dec.Decode(new(any)) != io.EOF {
		problem(w, 400, fmt.Errorf("one JSON object required"))
		return
	}
	raw, _ := json.Marshal(in.Profiles)
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("save unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	// Parent row serializes first creation, updates, and deletion for this owner.
	var owner string
	if err = tx.QueryRow(r.Context(), `SELECT id FROM member_accounts WHERE id=$1 FOR UPDATE`, id).Scan(&owner); err != nil {
		problem(w, 401, fmt.Errorf("sign in required"))
		return
	}
	if _, err = tx.Exec(r.Context(), `INSERT INTO member_chart_store(account_id) VALUES($1) ON CONFLICT DO NOTHING`, id); err != nil {
		problem(w, 500, fmt.Errorf("save unavailable"))
		return
	}
	tag, err := tx.Exec(r.Context(), `UPDATE member_chart_store SET profiles=$1,revision=revision+1,updated_at=now() WHERE account_id=$2 AND revision=$3`, raw, id, in.Revision)
	if err != nil {
		problem(w, 500, fmt.Errorf("save unavailable"))
		return
	}
	if tag.RowsAffected() != 1 {
		problem(w, 409, fmt.Errorf("charts changed on another device; reload saved charts before saving again"))
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		problem(w, 500, fmt.Errorf("save unavailable"))
		return
	}
	writeJSON(w, 200, map[string]any{"revision": in.Revision + 1})
}
