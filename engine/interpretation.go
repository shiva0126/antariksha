package engine

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Dignity is deliberately a small, explicit data structure. The interpreter
// is never allowed to infer a dignity from prose or from a model response.
type Dignity struct {
	State        string `json:"state"`
	Sign         string `json:"sign"`
	NeechaBhanga bool   `json:"neecha_bhanga,omitempty"`
}
type Combustion map[string]bool
type Yoga struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Planets  []string       `json:"planets"`
	Houses   []string       `json:"houses"`
	Strength string         `json:"strength"`
	Geometry map[string]any `json:"geometry"`
}
type DashaPeriod struct {
	Lord   string `json:"lord"`
	Maha   string `json:"maha,omitempty"`
	Antara string `json:"antara,omitempty"`
	From   string `json:"from"`
	To     string `json:"to"`
	Level  string `json:"level"`
}
type Dasha struct {
	BirthBalance struct {
		Lord           string  `json:"lord"`
		YearsRemaining float64 `json:"years_remaining"`
	} `json:"birth_balance"`
	Current  DashaPeriod   `json:"current"`
	Upcoming DashaPeriod   `json:"upcoming"`
	Sequence []DashaPeriod `json:"sequence"`
}
type ChartFacts struct {
	Chart       Chart              `json:"chart"`
	Dignities   map[string]Dignity `json:"dignities"`
	Combustion  Combustion         `json:"combustion"`
	Retrograde  []string           `json:"retrograde"`
	Vimshottari Dasha              `json:"vimshottari"`
	Yogas       []Yoga             `json:"yogas"`
	AsOf        string             `json:"as_of"`
}

var dashaOrder = []string{"ketu", "venus", "sun", "moon", "mars", "rahu", "jupiter", "saturn", "mercury"}
var dashaYears = map[string]float64{"ketu": 7, "venus": 20, "sun": 6, "moon": 10, "mars": 7, "rahu": 18, "jupiter": 16, "saturn": 19, "mercury": 17}
var nakLords = dashaOrder
var exalted = map[string]int{"sun": 0, "moon": 1, "mars": 9, "mercury": 5, "jupiter": 3, "venus": 11, "saturn": 6}
var owned = map[string]map[int]bool{"sun": {4: true}, "moon": {3: true}, "mars": {0: true, 7: true}, "mercury": {2: true, 5: true}, "jupiter": {8: true, 11: true}, "venus": {1: true, 6: true}, "saturn": {9: true, 10: true}}
var friends = map[string]map[string]bool{"sun": {"moon": true, "mars": true, "jupiter": true}, "moon": {"sun": true, "mercury": true}, "mars": {"sun": true, "moon": true, "jupiter": true}, "mercury": {"sun": true, "venus": true}, "jupiter": {"sun": true, "moon": true, "mars": true}, "venus": {"mercury": true, "saturn": true}, "saturn": {"mercury": true, "venus": true}}
var planetNames = map[string]string{"sun": "Surya", "moon": "Chandra", "mars": "Mangala", "mercury": "Budha", "jupiter": "Guru", "venus": "Shukra", "saturn": "Shani", "rahu": "Rahu", "ketu": "Ketu"}

func signOf(g Graha) int { return int(g.Longitude / 30) }
func chartMap(c Chart) map[string]Graha {
	out := make(map[string]Graha, len(c.Grahas))
	for _, g := range c.Grahas {
		out[g.ID] = g
	}
	return out
}
func houseOf(lon, asc float64) int { return (int(lon/30)-int(asc/30)+12)%12 + 1 }
func inSet(v int, xs ...int) bool {
	for _, x := range xs {
		if v == x {
			return true
		}
	}
	return false
}
func angularDistance(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}

func Dignities(c Chart) map[string]Dignity {
	gs := chartMap(c)
	out := map[string]Dignity{}
	for id, g := range gs {
		if _, ok := exalted[id]; !ok {
			continue
		}
		sign := signOf(g)
		state := "neutral"
		if sign == exalted[id] {
			state = "exalted"
		} else if sign == (exalted[id]+6)%12 {
			state = "debilitated"
		} else if owned[id][sign] {
			state = "own"
		} else {
			lord := rashiLord[sign]
			if friends[id][lord] {
				state = "friendly"
			} else if friends[lord] != nil && friends[lord][id] {
				state = "enemy"
			}
		}
		out[id] = Dignity{State: state, Sign: g.Rashi}
	}
	for id, d := range out {
		if d.State == "debilitated" {
			out[id] = d
			out[id] = neechaBhanga(c, id, out[id], gs)
		}
	}
	return out
}
func neechaBhanga(c Chart, id string, d Dignity, gs map[string]Graha) Dignity {
	g := gs[id]
	dispositor := rashiLord[(signOf(g)+6)%12]
	if dg, ok := gs[dispositor]; ok && inSet(houseOf(dg.Longitude, c.Ascendant.Longitude), 1, 4, 7, 10) {
		d.NeechaBhanga = true
	}
	return d
}

var rashiLord = []string{"mars", "venus", "mercury", "moon", "sun", "mercury", "venus", "mars", "jupiter", "saturn", "saturn", "jupiter"}

func CombustionFlags(c Chart) Combustion {
	gs := chartMap(c)
	out := Combustion{}
	sun, ok := gs["sun"]
	if !ok {
		return out
	}
	threshold := map[string]float64{"moon": 12, "mars": 17, "mercury": 14, "jupiter": 11, "venus": 10, "saturn": 15}
	for id, limit := range threshold {
		g, ok := gs[id]
		if ok {
			if id == "mercury" && g.Retrograde {
				limit = 12
			}
			out[id] = angularDistance(g.Longitude, sun.Longitude) <= limit
		}
	}
	return out
}
func Facts(c Chart, asOf time.Time) (ChartFacts, error) {
	if c.Ayanamsa != "lahiri" {
		return ChartFacts{}, fmt.Errorf("facts require Lahiri chart")
	}
	d := Dignities(c)
	return ChartFacts{Chart: c, Dignities: d, Combustion: CombustionFlags(c), Retrograde: retrograde(c), Vimshottari: Vimshottari(c, asOf), Yogas: DetectYogas(c, d), AsOf: asOf.UTC().Format(time.RFC3339)}, nil
}
func retrograde(c Chart) []string {
	out := []string{}
	for _, g := range c.Grahas {
		if g.Retrograde && g.ID != "rahu" && g.ID != "ketu" {
			out = append(out, g.ID)
		}
	}
	sort.Strings(out)
	return out
}

func Vimshottari(c Chart, asOf time.Time) Dasha {
	gs := chartMap(c)
	moon := gs["moon"]
	nakSize := 360.0 / 27
	idx := int(moon.Longitude / nakSize)
	lord := nakLords[idx%9]
	elapsed := math.Mod(moon.Longitude, nakSize) / nakSize
	remaining := dashaYears[lord] * (1 - elapsed)
	birth := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	if t, e := time.ParseInLocation("2006-01-02 15:04", c.Input.Date+" "+c.Input.Time, time.UTC); e == nil {
		z, _ := time.LoadLocation(c.Input.TZ)
		birth = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, z).UTC()
	}
	periods := []DashaPeriod{}
	cur := birth
	currentIndex := -1
	for n := 0; n < 18; n++ {
		i := (idx + n) % 9
		years := dashaYears[dashaOrder[i]]
		if n == 0 {
			years = remaining
		}
		end := cur.Add(time.Duration(years*365.2425*24) * time.Hour)
		p := DashaPeriod{Lord: dashaOrder[i], From: cur.Format("2006-01-02"), To: end.Format("2006-01-02"), Level: "maha"}
		periods = append(periods, p)
		if !asOf.Before(cur) && asOf.Before(end) {
			currentIndex = n
		}
		cur = end
		if n == 0 {
			idx = (idx + 1) % 9
		}
	}
	var current, upcoming DashaPeriod
	if currentIndex >= 0 {
		current = periods[currentIndex]
		current.Maha = current.Lord
		current.Antara, current.From, current.To = antaraFor(current, asOf)
		if currentIndex+1 < len(periods) {
			upcoming = periods[currentIndex+1]
		}
	}
	return Dasha{BirthBalance: struct {
		Lord           string  `json:"lord"`
		YearsRemaining float64 `json:"years_remaining"`
	}{lord, remaining}, Current: current, Upcoming: upcoming, Sequence: periods}
}

func antaraFor(maha DashaPeriod, at time.Time) (string, string, string) {
	from, _ := time.Parse("2006-01-02", maha.From)
	cycleStart := 0
	for i, id := range dashaOrder {
		if id == maha.Lord {
			cycleStart = i
			break
		}
	}
	cur := from
	for n := 0; n < 9; n++ {
		lord := dashaOrder[(cycleStart+n)%9]
		days := dashaYears[maha.Lord] * dashaYears[lord] / 120 * 365.2425
		end := cur.Add(time.Duration(days*24) * time.Hour)
		if !at.Before(cur) && at.Before(end) {
			return lord, cur.Format("2006-01-02"), end.Format("2006-01-02")
		}
		cur = end
	}
	return dashaOrder[cycleStart], maha.From, maha.To
}
