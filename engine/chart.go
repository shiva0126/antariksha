package engine

import (
	"fmt"
	"math"
	"time"

	"github.com/example/panchang/engine/swe"
)

var rashiNames = []string{"Mesha", "Vrishabha", "Mithuna", "Karka", "Simha", "Kanya", "Tula", "Vrishchika", "Dhanu", "Makara", "Kumbha", "Meena"}

type ChartInput struct {
	Date string  `json:"date"`
	Time string  `json:"time"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	TZ   string  `json:"tz"`
}
type Point struct {
	Longitude float64 `json:"longitude"`
	Rashi     string  `json:"rashi"`
	Degree    float64 `json:"degree"`
}
type Graha struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	DistanceAU    float64 `json:"distance_au"`
	Rashi         string  `json:"rashi"`
	RashiDegree   float64 `json:"rashi_degree"`
	Nakshatra     string  `json:"nakshatra"`
	NakshatraPada int     `json:"nakshatra_pada"`
	Retrograde    bool    `json:"retrograde"`
	Speed         float64 `json:"speed"`
}
type Houses struct {
	System string    `json:"system"`
	Cusps  []float64 `json:"cusps"`
}
type Chart struct {
	SchemaVersion int        `json:"schema_version"`
	Input         ChartInput `json:"input"`
	Ayanamsa      string     `json:"ayanamsa"`
	Ascendant     Point      `json:"ascendant"`
	Grahas        []Graha    `json:"grahas"`
	Houses        Houses     `json:"houses"`
}

func (e *Engine) BirthChart(in ChartInput) (Chart, error) {
	defer e.begin()()
	if math.IsNaN(in.Lat) || math.IsInf(in.Lat, 0) || math.IsNaN(in.Lon) || math.IsInf(in.Lon, 0) || in.Lat < -90 || in.Lat > 90 || in.Lon < -180 || in.Lon > 180 || in.TZ == "" {
		return Chart{}, fmt.Errorf("valid latitude, longitude and IANA timezone are required")
	}
	z, err := time.LoadLocation(in.TZ)
	if err != nil {
		return Chart{}, err
	}
	local, err := time.ParseInLocation("2006-01-02 15:04", in.Date+" "+in.Time, z)
	if err != nil {
		return Chart{}, fmt.Errorf("birth date/time: %w", err)
	}
	if local.Format("2006-01-02 15:04") != in.Date+" "+in.Time {
		return Chart{}, fmt.Errorf("time does not exist in this timezone")
	}
	u := local.UTC()
	jd := swe.JulianDay(u.Year(), int(u.Month()), u.Day(), float64(u.Hour())+float64(u.Minute())/60+float64(u.Second())/3600)
	asc, err := swe.Ascendant(jd, in.Lat, in.Lon)
	if err != nil {
		return Chart{}, err
	}
	asc = normalize(asc)
	defs := []struct {
		id, name string
		body     int
	}{{"sun", "Surya", swe.Sun}, {"moon", "Chandra", swe.Moon}, {"mars", "Mangala", swe.Mars}, {"mercury", "Budha", swe.Mercury}, {"jupiter", "Guru", swe.Jupiter}, {"venus", "Shukra", swe.Venus}, {"saturn", "Shani", swe.Saturn}, {"rahu", "Rahu", swe.TrueNode}}
	gs := make([]Graha, 0, 9)
	for _, d := range defs {
		lon, lat, dist, speed, er := swe.Position3D(jd, d.body)
		if er != nil {
			return Chart{}, er
		}
		gs = append(gs, makeGraha(d.id, d.name, lon, lat, dist, speed))
	}
	r := gs[len(gs)-1]
	gs = append(gs, makeGraha("ketu", "Ketu", normalize(r.Longitude+180), -r.Latitude, r.DistanceAU, r.Speed))
	cusps := make([]float64, 12)
	for i := range cusps {
		cusps[i] = normalize(math.Floor(asc/30)*30 + float64(i)*30)
	}
	return Chart{SchemaVersion: SchemaVersion, Input: in, Ayanamsa: "lahiri", Ascendant: Point{asc, rashiNames[int(asc/30)], math.Mod(asc, 30)}, Grahas: gs, Houses: Houses{"whole_sign", cusps}}, nil
}
func makeGraha(id, name string, lon, lat, dist, speed float64) Graha {
	lon = normalize(lon)
	n := int(lon / (360.0 / 27))
	within := math.Mod(lon, 360.0/27)
	return Graha{id, name, lon, lat, dist, rashiNames[int(lon/30)], math.Mod(lon, 30), nakshatraNames[n], int(within/(360.0/108)) + 1, speed < 0, speed}
}
