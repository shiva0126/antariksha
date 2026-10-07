package engine

import (
	"fmt"
	"math"
	"time"

	"github.com/example/panchang/engine/swe"
)

// Varshaphal (Tajika annual horoscopy): the chart for the moment the
// sidereal Sun returns to its natal longitude, read for the year that
// follows. The engine computes:
//   - the return moment, to the second, and the year chart (Varsha
//     kundali) for it at the birth place (many astrologers cast it for the
//     place of residence instead; the rest of the method is the same);
//   - Muntha: the natal lagna sign advanced one sign per completed year;
//   - the five office-holders (Pancha-adhikari) and the year lord
//     (Varshesha): the strongest office-holder that aspects the year lagna
//     by Tajika aspect.
// Documented approximation: classical texts rank the office-holders by
// Pancha-vargiya bala; the engine ranks them by Shadbala of the year chart
// (the ratio to the required minimum), which it already computes.

// Tri-rashi lords by lagna sign, for a day and a night year chart.
var triRashiDay = []string{"sun", "venus", "saturn", "venus", "jupiter", "moon", "mercury", "mars", "saturn", "mars", "jupiter", "moon"}
var triRashiNight = []string{"jupiter", "moon", "mercury", "mars", "sun", "venus", "saturn", "venus", "saturn", "mars", "jupiter", "moon"}

var varshaHouse = []string{"", "self and health", "money and family", "courage, siblings and short travel", "home, mother and comforts", "children, study and creativity", "health troubles, debts and rivals", "partnership and marriage", "obstacles and sudden change", "fortune, father and dharma", "work and standing", "gains and friends", "expenses, losses and faraway places"}

type YearOffice struct {
	Role     string  `json:"role"`
	Graha    string  `json:"graha"`
	House    int     `json:"house"`            // in the year chart
	Aspect   string  `json:"aspect,omitempty"` // to the year lagna: conjunct, friendly, inimical or ""
	Strength float64 `json:"strength"`         // Shadbala ratio in the year chart
}

type Muntha struct {
	Rashi  string `json:"rashi"`
	House  int    `json:"house"` // from the year lagna
	Lord   string `json:"lord"`
	Tone   string `json:"tone"` // good | mixed | hard
	Detail string `json:"detail"`
}

type VarshaReport struct {
	Year       int          `json:"year"`
	Age        int          `json:"age"` // completed years at the return
	ReturnAt   time.Time    `json:"return_at"`
	Chart      Chart        `json:"chart"`
	Day        bool         `json:"day_chart"`
	Muntha     Muntha       `json:"muntha"`
	Offices    []YearOffice `json:"offices"`
	YearLord   string       `json:"year_lord"`
	LordReason string       `json:"year_lord_reason"`
	Strengths  []string     `json:"strengths"`
	Cautions   []string     `json:"cautions"`
	Method     string       `json:"method"`
}

// SolarReturn finds the moment in the given calendar year when the
// sidereal Sun is back at its natal longitude.
func (e *Engine) SolarReturn(natal Chart, year int) (time.Time, error) {
	var sun float64
	for _, g := range natal.Grahas {
		if g.ID == "sun" {
			sun = g.Longitude
		}
	}
	birth := birthTime(natal)
	if birth.Year() < 1 {
		return time.Time{}, fmt.Errorf("invalid birth time")
	}
	guess := time.Date(year, birth.Month(), birth.Day(), birth.Hour(), birth.Minute(), 0, 0, time.UTC)
	defer e.begin()()
	jd := jdOf(guess)
	for i := 0; i < 8; i++ {
		lon, speed, err := swe.Position(jd, swe.Sun)
		if err != nil {
			return time.Time{}, err
		}
		d := math.Mod(normalize(lon)-sun+540, 360) - 180
		if math.Abs(d) < 1e-7 {
			break
		}
		jd -= d / speed
	}
	return timeOfJD(jd), nil
}

// tajikaAspect is the Tajika aspect between two signs: the same sign is a
// conjunction, 5/9 and 3/11 are friendly, 7 and 4/10 inimical, and 2/12
// and 6/8 no aspect.
func tajikaAspect(a, b int) string {
	switch (b-a+12)%12 + 1 {
	case 1:
		return "conjunct"
	case 3, 5, 9, 11:
		return "friendly"
	case 4, 7, 10:
		return "inimical"
	}
	return ""
}

// Varshaphal casts and reads the year chart for the return in the given
// calendar year.
func (e *Engine) Varshaphal(natal Chart, year int) (VarshaReport, error) {
	birth := birthTime(natal)
	age := year - birth.Year()
	if age < 1 || age > 120 {
		return VarshaReport{}, fmt.Errorf("choose a year after the birth year")
	}
	at, err := e.SolarReturn(natal, year)
	if err != nil {
		return VarshaReport{}, err
	}
	zone, err := time.LoadLocation(natal.Input.TZ)
	if err != nil {
		return VarshaReport{}, err
	}
	local := at.In(zone).Round(time.Minute)
	vc, err := e.BirthChart(ChartInput{Date: local.Format("2006-01-02"), Time: local.Format("15:04"), Lat: natal.Input.Lat, Lon: natal.Input.Lon, TZ: natal.Input.TZ})
	if err != nil {
		return VarshaReport{}, err
	}
	sb, err := e.Shadbala(vc)
	if err != nil {
		return VarshaReport{}, err
	}
	return readVarsha(natal, vc, year, age, at, sb), nil
}

func readVarsha(natal, vc Chart, year, age int, at time.Time, sb Shadbala) VarshaReport {
	g := chartMap(vc)
	asc := vc.Ascendant.Longitude
	lagna := int(asc/30) % 12
	natalLagna := int(natal.Ascendant.Longitude/30) % 12
	day := inSet(houseOf(g["sun"].Longitude, asc), 7, 8, 9, 10, 11, 12)
	strength := map[string]float64{}
	for _, r := range sb.Rows {
		strength[r.Graha] = r.Ratio
	}
	rep := VarshaReport{Year: year, Age: age, ReturnAt: at, Chart: vc, Day: day, Strengths: []string{}, Cautions: []string{},
		Method: "Tajika: the year lord is the strongest of the five office-holders that aspects the year lagna. Strength here is Shadbala of the year chart, in place of the classical Pancha-vargiya bala. The chart is cast for the birth place."}

	// Muntha.
	ms := (natalLagna + age) % 12
	mh := (ms-lagna+12)%12 + 1
	m := Muntha{Rashi: rashiNames[ms], House: mh, Lord: SignLord(ms)}
	switch {
	case inSet(mh, 9, 10, 11):
		m.Tone, m.Detail = "good", fmt.Sprintf("Muntha falls in the %s house of the year chart (%s), one of the best places for it: a year of progress in these areas.", ordinal(mh), varshaHouse[mh])
	case inSet(mh, 1, 2, 3, 5):
		m.Tone, m.Detail = "mixed", fmt.Sprintf("Muntha falls in the %s house of the year chart (%s), a fair place for it: effort brings results here.", ordinal(mh), varshaHouse[mh])
	default:
		m.Tone, m.Detail = "hard", fmt.Sprintf("Muntha falls in the %s house of the year chart (%s), traditionally a harder place: go carefully in these areas.", ordinal(mh), varshaHouse[mh])
	}
	rep.Muntha = m

	// The five office-holders.
	dinaRatri := SignLord(signOf(g["moon"]))
	tri := triRashiNight[lagna]
	if day {
		dinaRatri, tri = SignLord(signOf(g["sun"])), triRashiDay[lagna]
	}
	roles := []struct{ role, graha string }{
		{"Muntha lord", m.Lord},
		{"Birth lagna lord", SignLord(natalLagna)},
		{"Year lagna lord", SignLord(lagna)},
		{"Tri-rashi lord", tri},
		{map[bool]string{true: "Day lord (Sun's sign)", false: "Night lord (Moon's sign)"}[day], dinaRatri},
	}
	for _, r := range roles {
		rep.Offices = append(rep.Offices, YearOffice{Role: r.role, Graha: r.graha, House: houseOf(g[r.graha].Longitude, asc),
			Aspect: tajikaAspect(signOf(g[r.graha]), lagna), Strength: math.Round(strength[r.graha]*100) / 100})
	}
	best := -1
	for i, o := range rep.Offices {
		if o.Aspect != "" && (best < 0 || o.Strength > rep.Offices[best].Strength) {
			best = i
		}
	}
	if best >= 0 {
		rep.LordReason = fmt.Sprintf("%s (%s) is the strongest office-holder that aspects the year lagna.", GrahaEnglish(rep.Offices[best].Graha), rep.Offices[best].Role)
	} else {
		for i, o := range rep.Offices {
			if best < 0 || o.Strength > rep.Offices[best].Strength {
				best = i
			}
		}
		rep.LordReason = fmt.Sprintf("No office-holder aspects the year lagna, so the strongest, %s (%s), rules the year.", GrahaEnglish(rep.Offices[best].Graha), rep.Offices[best].Role)
	}
	lord := rep.Offices[best].Graha
	rep.YearLord = lord

	// The year lord's condition in the year chart.
	dig := Dignities(vc)
	lh := houseOf(g[lord].Longitude, asc)
	switch {
	case dig[lord].State == "exalted" || dig[lord].State == "own":
		rep.Strengths = append(rep.Strengths, fmt.Sprintf("The year lord, %s, is in %s sign: a well-supported year.", GrahaEnglish(lord), dignityWord(dig[lord].State)))
	case dig[lord].State == "debilitated" && !dig[lord].NeechaBhanga:
		rep.Cautions = append(rep.Cautions, fmt.Sprintf("The year lord, %s, is in its weakest sign: results come with more effort.", GrahaEnglish(lord)))
	}
	if inSet(lh, 6, 8, 12) {
		rep.Cautions = append(rep.Cautions, fmt.Sprintf("The year lord sits in the %s house (%s), a harder house.", ordinal(lh), varshaHouse[lh]))
	} else if inSet(lh, 1, 4, 5, 7, 9, 10, 11) {
		rep.Strengths = append(rep.Strengths, fmt.Sprintf("The year lord sits in the %s house (%s), so the year centres on this area.", ordinal(lh), varshaHouse[lh]))
	}
	if rep.Offices[best].Aspect == "inimical" {
		rep.Cautions = append(rep.Cautions, "The year lord's aspect to the year lagna is the inimical kind (Tajika): progress comes through effort and friction.")
	}
	for _, id := range []string{"jupiter", "venus"} {
		if h := houseOf(g[id].Longitude, asc); inSet(h, 1, 4, 7, 10, 5, 9) {
			rep.Strengths = append(rep.Strengths, fmt.Sprintf("%s, a natural benefic, sits in the %s house of the year chart (%s).", GrahaEnglish(id), ordinal(h), varshaHouse[h]))
		}
	}
	for _, id := range []string{"saturn", "mars"} {
		if h := houseOf(g[id].Longitude, asc); inSet(h, 1, 7, 8) {
			rep.Cautions = append(rep.Cautions, fmt.Sprintf("%s sits in the %s house of the year chart (%s): go carefully there.", GrahaEnglish(id), ordinal(h), varshaHouse[h]))
		}
	}
	return rep
}
