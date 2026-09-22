package engine

import (
	"fmt"
	"strings"
)

// MuhurtaEvent describes the classical day-level rules for one kind of event
// (Muhurta Chintamani conventions as commonly applied). A day qualifies when
// its sunrise nakshatra, tithi and weekday are all favourable for the event and
// it carries no general exclusion (Rikta tithi, Amavasya, Bhadra at sunrise).
type MuhurtaEvent struct {
	ID, Name    string
	Nakshatras  []string
	Tithis      []int // 1–15 within the paksha
	Weekdays    []string
	ShuklaOnly  bool
	AvoidAdhika bool
	Description string
}

var MuhurtaEvents = []MuhurtaEvent{
	{"marriage", "Marriage (Vivaha)", []string{"Rohini", "Mrigashira", "Magha", "Uttara Phalguni", "Hasta", "Swati", "Anuradha", "Mula", "Uttara Ashadha", "Uttara Bhadrapada", "Revati"},
		[]int{2, 3, 5, 7, 10, 11, 12, 13}, []string{"Somavara", "Budhavara", "Guruvara", "Shukravara"}, false, true,
		"Traditional vivaha nakshatras on benefic weekdays; month-level rules (Kharmas, Chaturmas, combust Jupiter/Venus) also apply and should be checked with an astrologer."},
	{"griha_pravesh", "Griha Pravesh (house warming)", []string{"Rohini", "Mrigashira", "Uttara Phalguni", "Chitra", "Anuradha", "Uttara Ashadha", "Uttara Bhadrapada", "Revati"},
		[]int{2, 3, 5, 7, 10, 11, 13}, []string{"Somavara", "Budhavara", "Guruvara", "Shukravara"}, true, true,
		"Fixed (dhruva) and gentle nakshatras in the bright fortnight."},
	{"vehicle", "Vehicle purchase", []string{"Ashwini", "Mrigashira", "Punarvasu", "Pushya", "Hasta", "Chitra", "Swati", "Anuradha", "Shravana", "Dhanishtha", "Shatabhisha", "Revati"},
		[]int{1, 2, 3, 5, 6, 7, 10, 11, 12, 13}, []string{"Somavara", "Budhavara", "Guruvara", "Shukravara"}, false, false,
		"Movable and swift nakshatras suit vehicles."},
	{"business", "New business / venture", []string{"Ashwini", "Rohini", "Pushya", "Uttara Phalguni", "Hasta", "Chitra", "Anuradha", "Uttara Ashadha", "Shravana", "Uttara Bhadrapada", "Revati"},
		[]int{2, 3, 5, 7, 10, 11, 13}, []string{"Ravivara", "Budhavara", "Guruvara", "Shukravara"}, false, false,
		"Light and fixed nakshatras for lasting beginnings."},
	{"property", "Property purchase", []string{"Mrigashira", "Punarvasu", "Ashlesha", "Magha", "Purva Phalguni", "Vishakha", "Anuradha", "Mula", "Purva Ashadha", "Purva Bhadrapada", "Revati"},
		[]int{2, 3, 5, 6, 10, 11, 13}, []string{"Somavara", "Guruvara", "Shukravara"}, false, false,
		"Traditional bhoomi-kraya nakshatras."},
	{"naming", "Naming ceremony (Namakarana)", []string{"Ashwini", "Rohini", "Mrigashira", "Punarvasu", "Pushya", "Uttara Phalguni", "Hasta", "Chitra", "Swati", "Anuradha", "Uttara Ashadha", "Shravana", "Dhanishtha", "Shatabhisha", "Uttara Bhadrapada", "Revati"},
		[]int{1, 2, 3, 5, 7, 10, 11, 12, 13}, []string{"Somavara", "Budhavara", "Guruvara", "Shukravara"}, false, false,
		"Gentle and light nakshatras on benefic weekdays."},
	{"travel", "Travel (Yatra)", []string{"Ashwini", "Mrigashira", "Punarvasu", "Pushya", "Hasta", "Anuradha", "Shravana", "Dhanishtha", "Revati"},
		[]int{2, 3, 5, 7, 10, 11, 13}, []string{"Somavara", "Budhavara", "Guruvara", "Shukravara"}, false, false,
		"Classical yatra nakshatras; direction-specific rules (disha shula) are not applied."},
}

func MuhurtaEventByID(id string) (MuhurtaEvent, bool) {
	for _, e := range MuhurtaEvents {
		if e.ID == id {
			return e, true
		}
	}
	return MuhurtaEvent{}, false
}

var taraNames = []string{"Janma", "Sampat", "Vipat", "Kshema", "Pratyak", "Sadhaka", "Vadha", "Mitra", "Parama Mitra"}

// TaraBala returns the tara (1–9) of a day's nakshatra counted from the birth
// nakshatra, its name, and whether it is favourable (2, 4, 6, 8, 9).
func TaraBala(janma, day int) (int, string, bool) {
	t := ((day-janma+27)%27)%9 + 1
	return t, taraNames[t-1], inSet(t, 2, 4, 6, 8, 9)
}

// ChandraBala: the transiting Moon's sign counted from the natal Moon sign is
// favourable in the 1st, 3rd, 6th, 7th, 10th and 11th.
func ChandraBala(natalSign, transitSign int) (int, bool) {
	h := (transitSign-natalSign+12)%12 + 1
	return h, inSet(h, 1, 3, 6, 7, 10, 11)
}

type MuhurtaDay struct {
	Date      string   `json:"date"`
	Vaara     string   `json:"vaara"`
	Tithi     string   `json:"tithi"`
	Paksha    string   `json:"paksha"`
	Nakshatra string   `json:"nakshatra"`
	Yoga      string   `json:"yoga"`
	Good      bool     `json:"good"`
	Score     int      `json:"score"`
	Reasons   []string `json:"reasons"`
	Cautions  []string `json:"cautions"`
	Windows   []Window `json:"windows"`
	Avoid     []Window `json:"avoid"`
}

// Inauspicious nitya yogas that spoil a muhurta.
var badYogas = map[string]bool{"Vishkambha": true, "Atiganda": true, "Shula": true, "Ganda": true, "Vyaghata": true, "Vajra": true, "Vyatipata": true, "Parigha": true, "Vaidhriti": true}

func indexOf(xs []string, x string) int {
	for i, v := range xs {
		if v == x {
			return i
		}
	}
	return -1
}

// EvaluateMuhurta scores one Panchang day for an event. janmaNak and moonSign
// are the person's birth nakshatra index (0–26) and Moon sign (0–11), or -1.
func EvaluateMuhurta(ev MuhurtaEvent, d Day, janmaNak, moonSign int, transitMoonSign int) MuhurtaDay {
	out := MuhurtaDay{Date: d.Date, Vaara: d.Vaara, Tithi: d.Tithi.Name, Paksha: d.Paksha, Nakshatra: d.Nakshatra.Name, Yoga: d.Yoga.Name, Reasons: []string{}, Cautions: []string{}, Windows: []Window{}, Avoid: []Window{d.RahuKaal, d.Yamaganda, d.Gulika}}
	ok := true
	tn := d.Tithi.Number
	if tn > 15 {
		tn -= 15
	}
	fail := func(s string) { ok = false; out.Cautions = append(out.Cautions, s) }
	if indexOf(ev.Nakshatras, d.Nakshatra.Name) >= 0 {
		out.Score += 3
		out.Reasons = append(out.Reasons, d.Nakshatra.Name+" is a favoured nakshatra (until "+d.Nakshatra.EndsAt+")")
	} else {
		fail(d.Nakshatra.Name + " nakshatra is not recommended for this event")
	}
	if inSet(tn, ev.Tithis...) {
		out.Score += 2
		out.Reasons = append(out.Reasons, d.Tithi.Name+" is a suitable tithi")
	} else {
		fail(d.Tithi.Name + " tithi is not recommended")
	}
	if indexOf(ev.Weekdays, d.Vaara) >= 0 {
		out.Score += 1
		out.Reasons = append(out.Reasons, d.Vaara+" is a favourable weekday")
	} else {
		fail(d.Vaara + " is not a favoured weekday")
	}
	if ev.ShuklaOnly && d.Paksha != "Shukla" {
		fail("the bright fortnight (Shukla paksha) is preferred")
	}
	if inSet(tn, 4, 9, 14) {
		fail("Rikta tithi (4th, 9th or 14th)")
	}
	if d.Tithi.Number == 30 {
		fail("Amavasya")
	}
	if d.Karana.Name == "Vishti" {
		fail("Bhadra (Vishti karana) at sunrise, until " + d.Karana.EndsAt)
	}
	if badYogas[d.Yoga.Name] {
		out.Cautions = append(out.Cautions, d.Yoga.Name+" yoga is inauspicious until "+d.Yoga.EndsAt)
		out.Score--
	}
	if ev.AvoidAdhika && strings.HasPrefix(d.LunarMonth, "Adhika") {
		fail("Adhika (intercalary) month")
	}
	if janmaNak >= 0 {
		dn := indexOf(nakshatraNames, d.Nakshatra.Name)
		t, name, good := TaraBala(janmaNak, dn)
		if good {
			out.Score++
			out.Reasons = append(out.Reasons, fmt.Sprintf("Tara bala: %s (%d) from your birth star", name, t))
		} else {
			fail(fmt.Sprintf("Tara bala: %s (%d) from your birth star is unfavourable", name, t))
		}
	}
	if moonSign >= 0 && transitMoonSign >= 0 {
		h, good := ChandraBala(moonSign, transitMoonSign)
		if good {
			out.Score++
			out.Reasons = append(out.Reasons, fmt.Sprintf("Chandra bala: Moon transits your %s from the natal Moon", ordinalEn(h)))
		} else {
			fail(fmt.Sprintf("Chandra bala: Moon transits your %s from the natal Moon", ordinalEn(h)))
		}
	}
	out.Good = ok
	if ok {
		// Good windows: Abhijit (not used on Wednesdays) and the Amrit, Shubh
		// and Labh day choghadiyas, minus anything touching Rahu kaal,
		// Yamaganda or Gulika.
		clash := func(w Window) bool {
			for _, a := range out.Avoid {
				if w.Start < a.End && a.Start < w.End {
					return true
				}
			}
			return false
		}
		if d.Vaara != "Budhavara" && !clash(d.Abhijit) {
			out.Windows = append(out.Windows, d.Abhijit)
		}
		for _, c := range d.Choghadiya[:8] {
			w := Window{c.Start, c.End}
			if (c.Name == "Amrit" || c.Name == "Shubh" || c.Name == "Labh") && !clash(w) {
				out.Windows = append(out.Windows, w)
			}
		}
	}
	return out
}

func ordinalEn(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	}
	return fmt.Sprintf("%dth", n)
}

// NakshatraIndex and RashiIndex expose name lookups for callers.
func NakshatraIndex(name string) int { return indexOf(nakshatraNames, name) }
func RashiIndex(name string) int     { return indexOf(rashiNames, name) }
