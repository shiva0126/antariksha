package engine

import (
	"fmt"
	"github.com/example/panchang/engine/swe"
	"math"
	"time"
)

var westernSigns = []string{"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo", "Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces"}
var westernThemes = []string{
	"initiative and a willingness to begin; consider leaving room to listen before acting",
	"steadiness and comfort; consider balancing familiar routines with useful change",
	"curiosity and communication; consider giving one idea enough time to develop",
	"care and belonging; consider expressing your needs clearly",
	"creativity and self-expression; consider sharing attention and encouragement",
	"careful observation and useful work; consider allowing room for imperfection",
	"cooperation and fairness; consider stating your own preferences as well as listening",
	"depth and commitment; consider how trust develops through clear boundaries",
	"exploration and a wider perspective; consider grounding big ideas in practical steps",
	"responsibility and long-term effort; consider making space for rest and enjoyment",
	"independent ideas and community; consider how your plans affect the people around you",
	"imagination and empathy; consider pairing compassion with workable limits",
}

type WesternPoint struct {
	Longitude float64 `json:"longitude"`
	Sign      string  `json:"sign"`
	Degree    float64 `json:"degree"`
}
type WesternPlanet struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	WesternPoint
	Latitude   float64 `json:"latitude"`
	DistanceAU float64 `json:"distance_au"`
	Speed      float64 `json:"speed"`
	Retrograde bool    `json:"retrograde"`
	House      int     `json:"house"`
	Meaning    string  `json:"meaning"`
}
type WesternAspect struct {
	First   string  `json:"first"`
	Second  string  `json:"second"`
	Name    string  `json:"name"`
	Orb     float64 `json:"orb"`
	Meaning string  `json:"meaning"`
}
type WesternReading struct {
	Zodiac    string          `json:"zodiac"`
	Input     ChartInput      `json:"input"`
	Ascendant WesternPoint    `json:"ascendant"`
	Houses    Houses          `json:"houses"`
	Planets   []WesternPlanet `json:"planets"`
	Aspects   []WesternAspect `json:"aspects"`
	Summary   []string        `json:"summary"`
	Note      string          `json:"note"`
}

func westernPoint(lon float64) WesternPoint {
	lon = normalize(lon)
	return WesternPoint{lon, westernSigns[int(lon/30)], math.Mod(lon, 30)}
}
func westernHouse(lon float64, cusps []float64) int {
	for i, start := range cusps {
		end := cusps[(i+1)%12]
		if normalize(lon-start) < normalize(end-start) {
			return i + 1
		}
	}
	return 1
}
func (e *Engine) WesternChart(in ChartInput, system string) (WesternReading, error) {
	if system == "" {
		system = "placidus"
	}
	if system != "placidus" && system != "whole_sign" {
		return WesternReading{}, fmt.Errorf("choose placidus or whole_sign houses")
	}
	if math.IsNaN(in.Lat) || math.IsInf(in.Lat, 0) || math.IsNaN(in.Lon) || math.IsInf(in.Lon, 0) || in.Lat <= -90 || in.Lat >= 90 || in.Lon < -180 || in.Lon > 180 || in.TZ == "" {
		return WesternReading{}, fmt.Errorf("valid coordinates and timezone are required")
	}
	z, err := time.LoadLocation(in.TZ)
	if err != nil {
		return WesternReading{}, fmt.Errorf("invalid IANA timezone")
	}
	birth, err := time.ParseInLocation("2006-01-02 15:04", in.Date+" "+in.Time, z)
	if err != nil || birth.Format("2006-01-02 15:04") != in.Date+" "+in.Time || birth.Year() < 1800 || birth.Year() > 2399 {
		return WesternReading{}, fmt.Errorf("enter an existing local birth time between 1800 and 2399")
	}
	defer e.begin()()
	u := birth.UTC()
	jd := swe.JulianDay(u.Year(), int(u.Month()), u.Day(), float64(u.Hour())+float64(u.Minute())/60)
	asc, cusps, err := swe.TropicalHouses(jd, in.Lat, in.Lon, system == "whole_sign")
	if err != nil {
		return WesternReading{}, err
	}
	out := WesternReading{Zodiac: "tropical", Input: in, Ascendant: westernPoint(asc), Houses: Houses{system, cusps}, Planets: []WesternPlanet{}, Aspects: []WesternAspect{}, Summary: []string{}, Note: "Positions are geocentric tropical coordinates from Swiss Ephemeris. Houses use longitude sectors between the returned cusps. Meanings are symbolic reflection, not personality measurements or predictions."}
	defs := []struct {
		id, name, topic string
		body            int
	}{{"sun", "Sun", "sense of identity and direction", swe.Sun}, {"moon", "Moon", "emotional needs and familiar comforts", swe.Moon}, {"mercury", "Mercury", "thinking and communication", swe.Mercury}, {"venus", "Venus", "affection and appreciation", swe.Venus}, {"mars", "Mars", "action and assertiveness", swe.Mars}, {"jupiter", "Jupiter", "learning and broadening your outlook", swe.Jupiter}, {"saturn", "Saturn", "responsibility and boundaries", swe.Saturn}, {"uranus", "Uranus", "change and independence", swe.Uranus}, {"neptune", "Neptune", "imagination and ideals", swe.Neptune}, {"pluto", "Pluto", "deep change and the use of power", swe.Pluto}}
	for _, d := range defs {
		lon, lat, dist, speed, err := swe.TropicalPosition3D(jd, d.body)
		if err != nil {
			return WesternReading{}, err
		}
		point := westernPoint(lon)
		meaning := fmt.Sprintf("In Western astrology, %s is used to reflect on %s. Its sign, %s, is associated with %s.", d.name, d.topic, point.Sign, westernThemes[int(point.Longitude/30)])
		out.Planets = append(out.Planets, WesternPlanet{d.id, d.name, point, lat, dist, speed, speed < 0, westernHouse(point.Longitude, cusps), meaning})
		if d.id == "sun" || d.id == "moon" {
			out.Summary = append(out.Summary, meaning)
		}
	}
	out.Summary = append(out.Summary, fmt.Sprintf("Your rising sign is %s: the sign rising on the eastern horizon at birth. This tradition uses it as a prompt about how you approach new situations, with themes of %s.", out.Ascendant.Sign, westernThemes[int(out.Ascendant.Longitude/30)]))
	angles := []struct {
		angle, orb    float64
		name, meaning string
	}{{0, 8, "conjunction", "These two themes are interpreted together."}, {60, 4, "sextile", "A prompt to notice opportunities to combine these themes."}, {90, 6, "square", "A prompt to work with competing needs rather than assume a fixed conflict."}, {120, 6, "trine", "A prompt to develop a combination that may feel familiar or easy."}, {180, 8, "opposition", "A prompt to find balance between these themes."}}
	for i, a := range out.Planets {
		for _, b := range out.Planets[i+1:] {
			gap := math.Abs(a.Longitude - b.Longitude)
			if gap > 180 {
				gap = 360 - gap
			}
			for _, v := range angles {
				orb := math.Abs(gap - v.angle)
				if orb <= v.orb {
					out.Aspects = append(out.Aspects, WesternAspect{a.Name, b.Name, v.name, orb, v.meaning})
					break
				}
			}
		}
	}
	return out, nil
}
