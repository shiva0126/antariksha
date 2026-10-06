package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/panchang/engine"
)

// Transit alerts: members who opt in hear when Jupiter, Saturn or Rahu and
// Ketu change sign, read for their own chart (house from the natal Moon and
// the Brihat Samhita 104.4 verdict). Events are found a week ahead; the
// notification's subject encodes the event, so each member gets each event
// once (member_notifications is unique on recipient, actor, kind, subject).

const transitAlertHorizon = 7 * 24 * time.Hour

func (s *Server) transitAlertRoutes() {
	s.memberRoute("GET /api/me/transit-alerts", s.transitAlertsSetting)
	s.memberRoute("PUT /api/me/transit-alerts", s.setTransitAlerts)
}

func (s *Server) transitAlertsSetting(w http.ResponseWriter, r *http.Request, id string) {
	var on, ready bool
	err := s.membersDB().QueryRow(r.Context(), `SELECT transit_alerts, birth_date IS NOT NULL AND birth_time IS NOT NULL AND birth_place IS NOT NULL FROM member_accounts WHERE id=$1`, id).Scan(&on, &ready)
	if err != nil {
		problem(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"enabled": on, "birth_details": ready})
}

func (s *Server) setTransitAlerts(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	s.memberExec(w, r, `UPDATE member_accounts SET transit_alerts=$2 WHERE id=$1`, id, in.Enabled)
}

// transitSubject encodes one event for one member.
func transitSubject(ev engine.TransitEvent, fromMoon int, ketu *engine.TransitEvent, ketuHouse int) string {
	fav := "-"
	if f, known := engine.GocharaFavourable(ev.Graha, fromMoon); known {
		fav = map[bool]string{true: "1", false: "0"}[f]
	}
	s := fmt.Sprintf("%s|%s|%s|%d|%s", ev.Graha, ev.Rashi, ev.At.Format("2006-01-02"), fromMoon, fav)
	if ketu != nil {
		s += fmt.Sprintf("|%s|%d", ketu.Rashi, ketuHouse)
	}
	return s
}

// transitAlertText turns a subject back into one plain sentence.
func transitAlertText(subject string) string {
	p := strings.Split(subject, "|")
	if len(p) < 5 {
		return "A slow planet changes sign for your chart"
	}
	day, err := time.Parse("2006-01-02", p[2])
	when := p[2]
	if err == nil {
		when = day.Format("2 January")
	}
	h, _ := strconv.Atoi(p[3])
	name := engine.GrahaEnglish(p[0])
	if p[0] == "rahu" && len(p) >= 7 {
		kh, _ := strconv.Atoi(p[6])
		return fmt.Sprintf("On %s Rahu moves into %s and Ketu into %s: your %s and %s houses from the Moon", when, p[1], p[5], ordinalNum(h), ordinalNum(kh))
	}
	verdict := map[string]string{"1": ", a good house for it in the classical books", "0": ", a harder house for it in the classical books"}[p[4]]
	return fmt.Sprintf("On %s %s moves into %s: your %s house from the Moon%s", when, name, p[1], ordinalNum(h), verdict)
}

func ordinalNum(n int) string {
	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", n, suffix)
}

// queueTransitAlerts records a notification for every opted-in member for
// each slow sign change in the coming week. It returns how many were new.
func (s *Server) queueTransitAlerts(ctx context.Context, now time.Time) (int, error) {
	tc, ok := s.engine.(TransitCalculator)
	if !ok {
		return 0, nil
	}
	all, err := tc.TransitEvents(now, now.Add(transitAlertHorizon))
	if err != nil {
		return 0, err
	}
	var events []engine.TransitEvent
	var ketuFor = map[time.Time]*engine.TransitEvent{}
	for i, ev := range all {
		if ev.Kind != "ingress" {
			continue
		}
		switch ev.Graha {
		case "jupiter", "saturn", "rahu":
			events = append(events, ev)
		case "ketu":
			ketuFor[ev.At] = &all[i]
		}
	}
	if len(events) == 0 {
		return 0, nil
	}
	db := s.membersDB()
	rows, err := db.Query(ctx, `SELECT id, to_char(birth_date,'YYYY-MM-DD'), to_char(birth_time,'HH24:MI'), birth_place FROM member_accounts
 WHERE transit_alerts AND birth_date IS NOT NULL AND birth_time IS NOT NULL AND birth_place IS NOT NULL`)
	if err != nil {
		return 0, err
	}
	type member struct {
		id    string
		birth engine.ChartInput
	}
	var members []member
	for rows.Next() {
		var m member
		var place []byte
		if rows.Scan(&m.id, &m.birth.Date, &m.birth.Time, &place) != nil {
			continue
		}
		var p struct {
			Lat, Lon float64
			TZ       string `json:"tz"`
		}
		if json.Unmarshal(place, &p) != nil || p.TZ == "" {
			continue
		}
		m.birth.Lat, m.birth.Lon, m.birth.TZ = p.Lat, p.Lon, p.TZ
		members = append(members, m)
	}
	rows.Close()
	added := 0
	for _, m := range members {
		natal, err := s.engine.BirthChart(m.birth)
		if err != nil {
			continue
		}
		var moon int
		for _, g := range natal.Grahas {
			if g.ID == "moon" {
				moon = engine.RashiIndex(g.Rashi)
			}
		}
		house := func(rashi string) int { return (engine.RashiIndex(rashi)-moon+12)%12 + 1 }
		for _, ev := range events {
			var subject string
			if ev.Graha == "rahu" {
				k := ketuFor[ev.At]
				kh := 0
				if k != nil {
					kh = house(k.Rashi)
				}
				subject = transitSubject(ev, house(ev.Rashi), k, kh)
			} else {
				subject = transitSubject(ev, house(ev.Rashi), nil, 0)
			}
			tag, err := db.Exec(ctx, `INSERT INTO member_notifications(recipient,actor,kind,subject) VALUES($1,$1,'transit',$2) ON CONFLICT(recipient,actor,kind,subject) DO NOTHING`, m.id, subject)
			if err != nil {
				return added, err
			}
			added += int(tag.RowsAffected())
		}
	}
	return added, nil
}
