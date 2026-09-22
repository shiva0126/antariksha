package engine

import "strings"

// Kaal is the part of the day at which a festival's tithi must prevail.
// Festivals are not decided by the sunrise tithi alone: Ganesh Chaturthi needs
// Chaturthi at midday, Janmashtami Ashtami at midnight, Diwali Amavasya at
// pradosh.
type Kaal int

const (
	Sunrise   Kaal = iota
	Madhyahna      // midday, the middle fifth of daylight
	Aparahna       // afternoon, the fourth fifth of daylight
	Pradosh        // the first three muhurtas after sunset
	Nishita        // the midnight muhurta, midpoint of the night
)

// FestivalRule describes a lunar observance in the Amanta month system.
// Tithi is 1–15 within the paksha (15 = Purnima or Amavasya).
type FestivalRule struct {
	Name   string
	Month  string // Amanta month; "" = every month
	Shukla bool
	Tithi  int
	Kaal   Kaal
}

// Festivals lists the major pan-Indian observances. Regional variants and the
// full vrat list (named Ekadashis etc.) are out of scope.
var Festivals = []FestivalRule{
	{"Ugadi / Gudi Padwa", "Chaitra", true, 1, Sunrise},
	{"Rama Navami", "Chaitra", true, 9, Madhyahna},
	{"Hanuman Jayanti", "Chaitra", true, 15, Sunrise},
	{"Akshaya Tritiya", "Vaishakha", true, 3, Sunrise},
	{"Buddha Purnima", "Vaishakha", true, 15, Sunrise},
	{"Devshayani Ekadashi", "Ashadha", true, 11, Sunrise},
	{"Guru Purnima", "Ashadha", true, 15, Sunrise},
	{"Nag Panchami", "Shravana", true, 5, Sunrise},
	{"Raksha Bandhan", "Shravana", true, 15, Sunrise},
	{"Krishna Janmashtami", "Shravana", false, 8, Nishita},
	{"Ganesh Chaturthi", "Bhadrapada", true, 4, Madhyahna},
	{"Anant Chaturdashi", "Bhadrapada", true, 14, Sunrise},
	{"Sharad Navaratri begins", "Ashwina", true, 1, Sunrise},
	{"Durga Ashtami", "Ashwina", true, 8, Sunrise},
	{"Vijayadashami (Dussehra)", "Ashwina", true, 10, Aparahna},
	{"Sharad Purnima", "Ashwina", true, 15, Pradosh},
	{"Dhanteras", "Ashwina", false, 13, Pradosh},
	{"Naraka Chaturdashi", "Ashwina", false, 14, Sunrise},
	{"Diwali (Lakshmi Puja)", "Ashwina", false, 15, Pradosh},
	{"Govardhan Puja", "Kartika", true, 1, Sunrise},
	{"Bhai Dooj", "Kartika", true, 2, Aparahna},
	{"Dev Uthani Ekadashi", "Kartika", true, 11, Sunrise},
	{"Kartika Purnima", "Kartika", true, 15, Sunrise},
	{"Vasant Panchami", "Magha", true, 5, Sunrise},
	{"Maha Shivaratri", "Magha", false, 14, Nishita},
	{"Holika Dahan", "Phalguna", true, 15, Pradosh},
	{"Holi", "Phalguna", false, 1, Sunrise},
	// Recurring observances.
	{"Ekadashi", "", true, 11, Sunrise},
	{"Ekadashi", "", false, 11, Sunrise},
	{"Pradosh Vrat", "", true, 13, Pradosh},
	{"Pradosh Vrat", "", false, 13, Pradosh},
	{"Sankashti Chaturthi", "", false, 4, Pradosh},
	{"Purnima", "", true, 15, Sunrise},
	{"Amavasya", "", false, 15, Sunrise},
}

// kaalJD returns the Julian day of a kaal for a Panchang day.
func kaalJD(k Kaal, rise, set, nextRise float64) float64 {
	day, night := set-rise, nextRise-set
	switch k {
	case Madhyahna:
		return rise + day*0.5
	case Aparahna:
		return rise + day*0.7
	case Pradosh:
		return set + night*0.1 // middle of the 3-muhurta window
	case Nishita:
		return set + night*0.5
	}
	return rise
}

// tithiAt returns the 0–29 tithi index at jd.
func tithiAt(jd float64) (int, error) {
	a, err := positions(jd)
	if err != nil {
		return 0, err
	}
	return int(a.diff / 12), nil
}

// festivalsFor evaluates every rule for one Panchang day. A festival falls on
// the first day whose kaal has the target tithi; if the tithi is skipped at
// kaal on both days (kshaya), it falls on the day the tithi was running into.
func festivalsFor(rise, set, nextRise float64, sunriseMonth string, sunriseTithi int) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	type sample struct {
		yesterday, today, tomorrow int
		jd                         float64
	}
	cache := map[Kaal]sample{}
	months := map[float64]string{}
	monthAt := func(jd float64) (string, error) {
		if m, ok := months[jd]; ok {
			return m, nil
		}
		m := sunriseMonth
		// The Amanta month changes at new moon; only look it up again when a
		// new moon may lie between sunrise and jd.
		if t, err := tithiAt(jd); err != nil {
			return "", err
		} else if t < sunriseTithi {
			if m, err = lunarMonth(jd); err != nil {
				return "", err
			}
		}
		months[jd] = m
		return m, nil
	}
	for _, r := range Festivals {
		sm, ok := cache[r.Kaal]
		if !ok {
			sm.jd = kaalJD(r.Kaal, rise, set, nextRise)
			var err error
			if sm.yesterday, err = tithiAt(sm.jd - 1); err != nil {
				return nil, err
			}
			if sm.today, err = tithiAt(sm.jd); err != nil {
				return nil, err
			}
			if sm.tomorrow, err = tithiAt(sm.jd + 1); err != nil {
				return nil, err
			}
			cache[r.Kaal] = sm
		}
		target := r.Tithi - 1
		if !r.Shukla {
			target += 15
		}
		monthJD := sm.jd
		hit := sm.today == target && sm.yesterday != target
		// Kshaya: the tithi falls between two consecutive kaals and is never
		// current at either. It belongs to the Panchang day holding its middle,
		// about half a day after the first kaal: the same day for daytime
		// kaals, the next day for pradosh and nishita.
		if !hit {
			if r.Kaal == Pradosh || r.Kaal == Nishita {
				hit = sm.yesterday == (target+29)%30 && sm.today == (target+1)%30
				monthJD = sm.jd - 0.5
			} else {
				hit = sm.today == (target+29)%30 && sm.tomorrow == (target+1)%30
				monthJD = sm.jd + 0.5
			}
		}
		if !hit {
			continue
		}
		month, err := monthAt(monthJD)
		if err != nil {
			return nil, err
		}
		if r.Month != "" && strings.HasPrefix(month, "Adhika") {
			continue // named festivals are not observed in an intercalary month
		}
		if r.Month != "" && r.Month != month {
			continue
		}
		if !seen[r.Name] {
			seen[r.Name] = true
			out = append(out, r.Name)
		}
	}
	// Named festivals first; recurring observances after them.
	named, recurring := []string{}, []string{}
	for _, n := range out {
		if isRecurring(n) {
			recurring = append(recurring, n)
		} else {
			named = append(named, n)
		}
	}
	return append(named, recurring...), nil
}

func isRecurring(name string) bool {
	switch name {
	case "Ekadashi", "Pradosh Vrat", "Sankashti Chaturthi", "Purnima", "Amavasya":
		return true
	}
	return false
}

// solarFestivals marks sankranti days. By the usual convention an ingress
// after sunset is observed the next day, so the observance day is the one
// whose span from the previous sunset to its own sunset contains the ingress.
func solarFestivals(rise, set, nextRise float64) ([]string, error) {
	prevSet := rise - (nextRise - set)
	a, err := positions(prevSet)
	if err != nil {
		return nil, err
	}
	b, err := positions(set)
	if err != nil {
		return nil, err
	}
	sa, sb := int(a.sun/30), int(b.sun/30)
	if sa == sb {
		return nil, nil
	}
	switch sb {
	case 9:
		return []string{"Makar Sankranti"}, nil
	case 0:
		return []string{"Mesha Sankranti"}, nil
	}
	return []string{rashiNames[sb] + " Sankranti"}, nil
}
