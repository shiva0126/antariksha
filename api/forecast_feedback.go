package api

import (
	"fmt"
	"net/http"
	"time"
)

// Forecast feedback: members may say whether a past period felt as the
// forecast described. Superadmins see agreement by tone and area, the
// measure used to tune the forecast weights.

var feedbackAreas = map[string]bool{"career": true, "money": true, "partnership": true, "home": true, "learning": true, "effort": true, "wellbeing": true, "change": true, "travel": true}

func (s *Server) forecastFeedbackRoutes() {
	s.memberRoute("GET /api/me/forecast-feedback", s.myForecastFeedback)
	s.memberRoute("PUT /api/me/forecast-feedback", s.putForecastFeedback)
	s.memberRoute("GET /api/admin/forecast-accuracy", func(w http.ResponseWriter, r *http.Request, id string) {
		if s.accountRole(r, id) != "superadmin" {
			problem(w, 403, fmt.Errorf("superadmin access required"))
			return
		}
		s.forecastAccuracy(w, r, id)
	})
}

func (s *Server) myForecastFeedback(w http.ResponseWriter, r *http.Request, id string) {
	s.memberRows(w, r, `SELECT to_char(period_from,'YYYY-MM-DD') period_from, area, response FROM forecast_feedback WHERE account_id=$1 ORDER BY period_from DESC LIMIT 200`, id)
}

func (s *Server) putForecastFeedback(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		From     string `json:"period_from"`
		To       string `json:"period_to"`
		Area     string `json:"area"`
		Tone     string `json:"tone"`
		Response string `json:"response"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	from, e1 := time.Parse("2006-01-02", in.From)
	to, e2 := time.Parse("2006-01-02", in.To)
	switch {
	case e1 != nil || e2 != nil || !to.After(from):
		problem(w, 400, fmt.Errorf("invalid period"))
		return
	case from.After(time.Now()):
		problem(w, 400, fmt.Errorf("only past periods can be answered"))
		return
	case !feedbackAreas[in.Area] || (in.Tone != "supportive" && in.Tone != "mixed" && in.Tone != "challenging") || (in.Response != "yes" && in.Response != "partly" && in.Response != "no"):
		problem(w, 400, fmt.Errorf("invalid answer"))
		return
	}
	s.memberExec(w, r, `INSERT INTO forecast_feedback(account_id,period_from,period_to,area,tone,response) VALUES($1,$2,$3,$4,$5,$6)
 ON CONFLICT(account_id,period_from,area) DO UPDATE SET response=EXCLUDED.response,tone=EXCLUDED.tone,period_to=EXCLUDED.period_to,created_at=now()`, id, in.From, in.To, in.Area, in.Tone, in.Response)
}

// forecastAccuracy reports answers by tone and by area: "yes" counts as
// agreement, "partly" as half.
func (s *Server) forecastAccuracy(w http.ResponseWriter, r *http.Request, _ string) {
	s.memberRows(w, r, `SELECT tone, area, count(*) answers,
 round(avg(CASE response WHEN 'yes' THEN 1 WHEN 'partly' THEN 0.5 ELSE 0 END)::numeric, 3) agreement
 FROM forecast_feedback GROUP BY ROLLUP(tone, area) ORDER BY tone NULLS LAST, area NULLS FIRST`)
}
