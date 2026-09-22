package engine

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/example/panchang/engine/swe"
)

// Shadbala (BPHS ch. 27): the six-fold strength of the seven grahas, in
// virupas (shashtiamsas); 60 virupas = 1 rupa. Documented approximations:
//   - Dig bala uses equal houses from the ascendant for the 4th, 7th and 10th
//     cusps (the engine uses whole-sign houses and has no separate MC).
//   - Nathonnata bala uses local mean time from longitude (no equation of time).
//   - Chesta bala for Mars–Saturn uses the eight-motion (ashta chesta) scheme
//     from the planet's apparent speed, not the full chesta-kendra computation.
//   - Year and month lords come from the Kali ahargana (360- and 30-day cycles).

var shadbalaGrahas = []string{"sun", "moon", "mars", "mercury", "jupiter", "venus", "saturn"}

// Required minimum strength in rupas (BPHS).
var requiredRupas = map[string]float64{"sun": 5, "moon": 6, "mars": 5, "mercury": 7, "jupiter": 6.5, "venus": 5.5, "saturn": 5}

// Deep debilitation points (sidereal longitude).
var debilPoint = map[string]float64{"sun": 190, "moon": 213, "mars": 118, "mercury": 345, "jupiter": 275, "venus": 177, "saturn": 20}

// Moolatrikona sign and degree range.
var moolatrikona = map[string][3]float64{"sun": {4, 0, 20}, "moon": {1, 3, 30}, "mars": {0, 0, 12}, "mercury": {5, 15, 20}, "jupiter": {8, 0, 10}, "venus": {6, 0, 15}, "saturn": {10, 0, 20}}

var naisargika = map[string]float64{"sun": 60, "moon": 51.43, "venus": 42.86, "jupiter": 34.29, "mercury": 25.71, "mars": 17.14, "saturn": 8.57}

// Weekday lords, Sunday first.
var weekdayLord = []string{"sun", "moon", "mars", "mercury", "jupiter", "venus", "saturn"}

// Mean daily geocentric motion (degrees/day) for the ashta chesta scheme.
var meanMotion = map[string]float64{"mars": 0.524, "mercury": 0.9856, "jupiter": 0.0831, "venus": 0.9856, "saturn": 0.0335}

type ShadbalaRow struct {
	Graha        string             `json:"graha"`
	Sthana       float64            `json:"sthana"`
	Dig          float64            `json:"dig"`
	Kala         float64            `json:"kala"`
	Chesta       float64            `json:"chesta"`
	Naisargika   float64            `json:"naisargika"`
	Drik         float64            `json:"drik"`
	Total        float64            `json:"total"` // virupas
	Rupas        float64            `json:"rupas"`
	Required     float64            `json:"required_rupas"`
	Ratio        float64            `json:"ratio"`
	Rank         int                `json:"rank"`
	Ishta        float64            `json:"ishta_phala"`
	Kashta       float64            `json:"kashta_phala"`
	Details      map[string]float64 `json:"details"`
	ChestaMotion string             `json:"chesta_motion,omitempty"`
}

type Shadbala struct {
	Rows  []ShadbalaRow `json:"rows"`
	Notes []string      `json:"notes"`
}

// compound returns the Saptavargaja points for graha a in a sign ruled by b,
// from the compound (panchadha) relationship: natural friendship combined
// with temporary friendship (b in the 2nd, 3rd, 4th, 10th, 11th or 12th from a).
func compound(a, b string, gs map[string]Graha) float64 {
	if a == b {
		return 30
	}
	d := (signOf(gs[b])-signOf(gs[a])+12)%12 + 1
	tf := inSet(d, 2, 3, 4, 10, 11, 12)
	switch natural := relation(a, b); {
	case natural == 2 && tf:
		return 22.5 // adhi mitra
	case natural == 1 && tf:
		return 15 // mitra
	case natural == 2 && !tf, natural == 0 && tf:
		return 7.5 // sama
	case natural == 1 && !tf:
		return 3.75 // shatru
	}
	return 1.875 // adhi shatru
}

func arcFrom(a, b float64) float64 { return angularDistance(a, b) }

// drishti is the BPHS sphuta drishti (0–60) of a graha at `from` on a point
// at `to`, with the special aspects of Mars, Jupiter and Saturn.
func drishti(id string, from, to float64) float64 {
	d := normalize(to - from)
	var v float64
	switch {
	case d < 30:
		v = 0
	case d < 60:
		v = (d - 30) / 2
	case d < 90:
		v = d - 60 + 15
	case d < 120:
		v = (120-d)/2 + 30
	case d < 150:
		v = 150 - d
	case d < 180:
		v = (d - 150) * 2
	case d < 300:
		v = (300 - d) / 2
	default:
		v = 0
	}
	switch id {
	case "mars":
		if (d >= 90 && d < 120) || (d >= 210 && d < 240) {
			v += 15
		}
	case "jupiter":
		if (d >= 120 && d < 150) || (d >= 240 && d < 270) {
			v += 30
		}
	case "saturn":
		if (d >= 60 && d < 90) || (d >= 270 && d < 300) {
			v += 45
		}
	}
	return math.Min(v, 60)
}

// Shadbala computes the six-fold strength for a birth chart.
func (e *Engine) Shadbala(c Chart) (Shadbala, error) {
	birth := birthTime(c)
	if birth.Year() < 1 {
		return Shadbala{}, fmt.Errorf("invalid birth time")
	}
	h := float64(birth.Hour()) + float64(birth.Minute())/60 + float64(birth.Second())/3600
	jd := swe.JulianDay(birth.Year(), int(birth.Month()), birth.Day(), h)

	var rise, set, nextRise, ayan float64
	var err error
	func() {
		defer e.begin()()
		ayan = swe.Ayanamsa(jd)
		// Sunrise on or before the birth instant, and the following sunset.
		rise, err = swe.RiseSet(jd-1, c.Input.Lat, c.Input.Lon, swe.Sun, swe.Rise)
		if err != nil {
			return
		}
		for {
			next, e2 := swe.RiseSet(rise+0.01, c.Input.Lat, c.Input.Lon, swe.Sun, swe.Rise)
			if e2 != nil || next > jd {
				nextRise = next
				break
			}
			rise = next
		}
		set, err = swe.RiseSet(rise, c.Input.Lat, c.Input.Lon, swe.Sun, swe.Set)
	}()
	if err != nil {
		return Shadbala{}, err
	}
	isDay := jd < set

	gs := chartMap(c)
	asc := c.Ascendant.Longitude
	sun, moon := gs["sun"], gs["moon"]
	elong := arcFrom(moon.Longitude, sun.Longitude) // 0–180
	waxing := normalize(moon.Longitude-sun.Longitude) < 180

	rows := map[string]*ShadbalaRow{}
	for _, id := range shadbalaGrahas {
		rows[id] = &ShadbalaRow{Graha: id, Details: map[string]float64{}, Required: requiredRupas[id]}
	}

	// ---- Sthana bala ----
	vargaList := []int{1, 2, 3, 7, 9, 12, 30}
	for _, id := range shadbalaGrahas {
		g := gs[id]
		r := rows[id]
		uchcha := arcFrom(g.Longitude, debilPoint[id]) / 3
		sv := 0.0
		for _, n := range vargaList {
			var sign int
			if n == 30 {
				sign = trimsamsaSign(g.Longitude)
			} else {
				sign, _ = vargaSign(g.Longitude, n)
			}
			lord := rashiLord[sign]
			mt := moolatrikona[id]
			deg := math.Mod(g.Longitude, 30)
			switch {
			case n == 1 && int(mt[0]) == sign && deg >= mt[1] && deg < mt[2]:
				sv += 45
			case lord == id:
				sv += 30
			default:
				sv += compound(id, lord, gs)
			}
		}
		oj := 0.0
		signOdd := signOf(g)%2 == 0
		nav, _ := vargaSign(g.Longitude, 9)
		navOdd := nav%2 == 0
		femalePref := id == "moon" || id == "venus"
		if signOdd != femalePref {
			oj += 15
		}
		if navOdd != femalePref {
			oj += 15
		}
		house := houseOf(g.Longitude, asc)
		kendradi := 15.0
		if inSet(house, 1, 4, 7, 10) {
			kendradi = 60
		} else if inSet(house, 2, 5, 8, 11) {
			kendradi = 30
		}
		dk := 0.0
		dec := int(math.Mod(g.Longitude, 30) / 10)
		switch id {
		case "sun", "mars", "jupiter":
			if dec == 0 {
				dk = 15
			}
		case "mercury", "saturn":
			if dec == 1 {
				dk = 15
			}
		default:
			if dec == 2 {
				dk = 15
			}
		}
		r.Details["uchcha"], r.Details["saptavargaja"], r.Details["ojhayugma"], r.Details["kendradi"], r.Details["drekkana"] = uchcha, sv, oj, kendradi, dk
		r.Sthana = uchcha + sv + oj + kendradi + dk
	}

	// ---- Dig bala: strongest at the cusp shown, zero at its opposite ----
	cusp := map[string]float64{"jupiter": asc, "mercury": asc, "sun": asc + 270, "mars": asc + 270, "saturn": asc + 180, "moon": asc + 90, "venus": asc + 90}
	for _, id := range shadbalaGrahas {
		weak := normalize(cusp[id] + 180)
		rows[id].Dig = arcFrom(gs[id].Longitude, weak) / 3
	}

	// ---- Kala bala ----
	lmtHours := math.Mod(float64(birth.Hour())+float64(birth.Minute())/60+c.Input.Lon/15+48, 24)
	fromNoon := math.Abs(lmtHours - 12) // 0 at midday, 12 at midnight
	// The Vedic day runs from sunrise, so it is the local date of `rise`.
	// Ahargana counts days from the Kali epoch (JDN 588466, a Friday).
	zone, zerr := time.LoadLocation(c.Input.TZ)
	if zerr != nil {
		zone = time.UTC
	}
	ry, rm, rd, rh := swe.ReverseJulian(rise)
	riseLocal := time.Date(ry, time.Month(rm), rd, 0, 0, 0, 0, time.UTC).Add(time.Duration(rh * float64(time.Hour))).In(zone)
	jdn := math.Floor(swe.JulianDay(riseLocal.Year(), int(riseLocal.Month()), riseLocal.Day(), 12) + 0.5)
	day := jdn - 588466
	wd := func(n float64) int { return (int(math.Mod(5+n, 7)) + 7) % 7 } // Sunday = 0
	varaLord := weekdayLord[wd(day)]
	abdaLord := weekdayLord[wd(math.Floor(day/360)*360)]
	masaLord := weekdayLord[wd(math.Floor(day/30)*30)]
	// Hora: 24 planetary hours from sunrise in Chaldean order, first hora = vara lord.
	chaldean := []string{"saturn", "jupiter", "mars", "sun", "venus", "mercury", "moon"}
	start := 0
	for i, id := range chaldean {
		if id == varaLord {
			start = i
		}
	}
	horaIdx := 0
	if isDay {
		horaIdx = int((jd - rise) / ((set - rise) / 12))
	} else {
		horaIdx = 12 + int((jd-set)/((nextRise-set)/12))
	}
	if horaIdx > 23 {
		horaIdx = 23
	}
	horaLord := chaldean[(start+horaIdx)%7]

	// Tribhaga: thirds of the day or night.
	third := 0
	if isDay {
		third = int(3 * (jd - rise) / (set - rise))
	} else {
		third = int(3 * (jd - set) / (nextRise - set))
	}
	if third > 2 {
		third = 2
	}
	var triLord string
	if isDay {
		triLord = []string{"mercury", "sun", "saturn"}[third]
	} else {
		triLord = []string{"moon", "venus", "mars"}[third]
	}

	const obliquity = 23.44
	for _, id := range shadbalaGrahas {
		g := gs[id]
		r := rows[id]
		var nat float64
		switch id {
		case "sun", "jupiter", "venus":
			nat = 60 * (1 - fromNoon/12)
		case "moon", "mars", "saturn":
			nat = 60 * fromNoon / 12
		default:
			nat = 60
		}
		benefic := id == "jupiter" || id == "venus" || id == "mercury" || id == "moon"
		paksha := elong / 3
		if !benefic {
			paksha = 60 - paksha
		}
		if id == "moon" {
			paksha *= 2 // BPHS doubles the Moon's paksha bala
		}
		tri := 0.0
		if id == "jupiter" || id == triLord {
			tri = 60
		}
		vmdh := 0.0
		if id == abdaLord {
			vmdh += 15
		}
		if id == masaLord {
			vmdh += 30
		}
		if id == varaLord {
			vmdh += 45
		}
		if id == horaLord {
			vmdh += 60
		}
		// Ayana bala from tropical declination.
		lt := normalize(g.Longitude+ayan) * math.Pi / 180
		b := g.Latitude * math.Pi / 180
		eps := obliquity * math.Pi / 180
		decl := math.Asin(math.Sin(b)*math.Cos(eps)+math.Cos(b)*math.Sin(eps)*math.Sin(lt)) * 180 / math.Pi
		var ayana float64
		switch id {
		case "moon", "saturn":
			ayana = (24 - decl) / 48 * 60
		case "mercury":
			ayana = (24 + math.Abs(decl)) / 48 * 60
		default:
			ayana = (24 + decl) / 48 * 60
		}
		if id == "sun" {
			ayana *= 2
		}
		r.Details["nathonnata"], r.Details["paksha"], r.Details["tribhaga"], r.Details["abda_masa_vara_hora"], r.Details["ayana"] = nat, paksha, tri, vmdh, ayana
		r.Kala = nat + paksha + tri + vmdh + ayana
	}

	// ---- Chesta bala ----
	for _, id := range shadbalaGrahas {
		r := rows[id]
		switch id {
		case "sun":
			r.Chesta = r.Details["ayana"] / 2 // BPHS: the Sun's chesta is its ayana bala
			r.ChestaMotion = "ayana"
		case "moon":
			r.Chesta = r.Details["paksha"] / 2
			r.ChestaMotion = "paksha"
		default:
			sp, mean := gs[id].Speed, meanMotion[id]
			ratio := sp / mean
			switch {
			case sp < 0 && ratio < -0.5:
				r.Chesta, r.ChestaMotion = 60, "vakra (retrograde)"
			case sp < 0:
				r.Chesta, r.ChestaMotion = 30, "anuvakra (turning retrograde)"
			case ratio < 0.1:
				r.Chesta, r.ChestaMotion = 15, "vikala (stationary)"
			case ratio < 0.5:
				r.Chesta, r.ChestaMotion = 15, "mandatara (very slow)"
			case ratio < 0.9:
				r.Chesta, r.ChestaMotion = 30, "manda (slow)"
			case ratio <= 1.1:
				r.Chesta, r.ChestaMotion = 7.5, "sama (mean)"
			case ratio <= 1.5:
				r.Chesta, r.ChestaMotion = 45, "chara (fast)"
			default:
				r.Chesta, r.ChestaMotion = 30, "atichara (very fast)"
			}
		}
	}

	// ---- Naisargika ----
	for _, id := range shadbalaGrahas {
		rows[id].Naisargika = naisargika[id]
	}

	// ---- Drik bala: benefic aspects add, malefic aspects subtract (÷4) ----
	isBenefic := func(id string) bool {
		switch id {
		case "jupiter", "venus", "mercury":
			return true
		case "moon":
			return waxing
		}
		return false
	}
	for _, id := range shadbalaGrahas {
		sum := 0.0
		for _, other := range shadbalaGrahas {
			if other == id {
				continue
			}
			v := drishti(other, gs[other].Longitude, gs[id].Longitude)
			if isBenefic(other) {
				sum += v
			} else {
				sum -= v
			}
		}
		rows[id].Drik = sum / 4
	}

	// ---- Yuddha (planetary war) among Mars–Saturn within 1° ----
	notes := []string{}
	war := []string{"mars", "mercury", "jupiter", "venus", "saturn"}
	for i := 0; i < len(war); i++ {
		for j := i + 1; j < len(war); j++ {
			a, b := war[i], war[j]
			if arcFrom(gs[a].Longitude, gs[b].Longitude) >= 1 {
				continue
			}
			win, lose := a, b
			if gs[b].Latitude > gs[a].Latitude {
				win, lose = b, a
			}
			ra, rb := rows[win], rows[lose]
			diff := math.Abs((ra.Sthana + ra.Dig + ra.Kala) - (rb.Sthana + rb.Dig + rb.Kala))
			ra.Kala += diff
			rb.Kala -= diff
			ra.Details["yuddha"], rb.Details["yuddha"] = diff, -diff
			notes = append(notes, fmt.Sprintf("Planetary war: %s defeats %s (within 1°).", GrahaEnglish(win), GrahaEnglish(lose)))
		}
	}

	out := Shadbala{Notes: notes}
	for _, id := range shadbalaGrahas {
		r := rows[id]
		r.Total = r.Sthana + r.Dig + r.Kala + r.Chesta + r.Naisargika + r.Drik
		r.Rupas = r.Total / 60
		r.Ratio = r.Rupas / r.Required
		u := r.Details["uchcha"]
		r.Ishta = math.Sqrt(math.Max(0, u*r.Chesta))
		r.Kashta = math.Sqrt(math.Max(0, (60-u)*(60-r.Chesta)))
		out.Rows = append(out.Rows, *r)
	}
	idx := make([]int, len(out.Rows))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool { return out.Rows[idx[a]].Ratio > out.Rows[idx[b]].Ratio })
	for rank, i := range idx {
		out.Rows[i].Rank = rank + 1
	}
	out.Notes = append(out.Notes,
		fmt.Sprintf("Weekday lord %s, hora lord %s, year lord %s, month lord %s; birth in the %s.", GrahaEnglish(varaLord), GrahaEnglish(horaLord), GrahaEnglish(abdaLord), GrahaEnglish(masaLord), map[bool]string{true: "day", false: "night"}[isDay]),
		"Dig bala uses equal-house cusps from the ascendant; chesta bala for Mars–Saturn uses the eight-motion scheme; nathonnata uses local mean time.")
	return out, nil
}

// trimsamsaSign returns the D30 sign (BPHS): odd signs Mars 5°, Saturn 5°,
// Jupiter 8°, Mercury 7°, Venus 5°; even signs the reverse.
func trimsamsaSign(lon float64) int {
	sign, deg := int(normalize(lon)/30), math.Mod(normalize(lon), 30)
	if sign%2 == 0 {
		switch {
		case deg < 5:
			return 0 // Mesha (Mars)
		case deg < 10:
			return 10 // Kumbha (Saturn)
		case deg < 18:
			return 8 // Dhanu (Jupiter)
		case deg < 25:
			return 2 // Mithuna (Mercury)
		}
		return 6 // Tula (Venus)
	}
	switch {
	case deg < 5:
		return 1 // Vrishabha (Venus)
	case deg < 12:
		return 5 // Kanya (Mercury)
	case deg < 20:
		return 11 // Meena (Jupiter)
	case deg < 25:
		return 9 // Makara (Saturn)
	}
	return 7 // Vrishchika (Mars)
}
