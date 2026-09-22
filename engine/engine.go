package engine

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/example/panchang/engine/swe"
)

type Engine struct {
	mu       sync.Mutex
	ephePath string
}

func New(ephePath string) *Engine {
	return &Engine{ephePath: ephePath}
}

// Swiss Ephemeris stores configuration in C thread-local storage. Keep every
// calculation on its configured OS thread, including all boundary searches.
func (e *Engine) begin() func() {
	e.mu.Lock()
	runtime.LockOSThread()
	swe.SetEphemerisPath(e.ephePath)
	swe.SetLahiri()
	return func() { runtime.UnlockOSThread(); e.mu.Unlock() }
}

func normalize(v float64) float64 {
	v = math.Mod(v, 360)
	if v < 0 {
		v += 360
	}
	return v
}

type angles struct{ sun, moon, diff, sum float64 }

func positions(jd float64) (angles, error) {
	sun, _, err := swe.Position(jd, swe.Sun)
	if err != nil {
		return angles{}, err
	}
	moon, _, err := swe.Position(jd, swe.Moon)
	if err != nil {
		return angles{}, err
	}
	return angles{sun, moon, normalize(moon - sun), normalize(moon + sun)}, nil
}

func localMidnightJD(date time.Time, zone *time.Location) float64 {
	m := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, zone).UTC()
	h := float64(m.Hour()) + float64(m.Minute())/60 + float64(m.Second())/3600
	return swe.JulianDay(m.Year(), int(m.Month()), m.Day(), h)
}

func jdTime(jd float64) time.Time {
	y, m, d, h := swe.ReverseJulian(jd)
	hour := int(h)
	minf := (h - float64(hour)) * 60
	minute := int(minf)
	sec := int(math.Round((minf - float64(minute)) * 60))
	return time.Date(y, time.Month(m), d, hour, minute, 0, 0, time.UTC).Add(time.Duration(sec) * time.Second)
}

func clock(t time.Time, zone *time.Location) string {
	return t.Round(time.Minute).In(zone).Format("15:04")
}

func dayClock(jd float64, base time.Time, zone *time.Location) string {
	t := jdTime(jd).In(zone)
	if t.Format("2006-01-02") != base.Format("2006-01-02") {
		return t.Format("15:04 · 02 Jan")
	}
	return t.Format("15:04")
}

func (e *Engine) Calculate(date time.Time, loc Location) (Computed, error) {
	defer e.begin()()
	zone, err := time.LoadLocation(loc.TZ)
	if err != nil {
		return Computed{}, fmt.Errorf("timezone: %w", err)
	}
	base := localMidnightJD(date, zone)
	sunriseJD, err := swe.RiseSet(base, loc.Lat, loc.Lon, swe.Sun, swe.Rise)
	if err != nil {
		return Computed{}, err
	}
	sunsetJD, err := swe.RiseSet(sunriseJD, loc.Lat, loc.Lon, swe.Sun, swe.Set)
	if err != nil {
		return Computed{}, err
	}
	nextRise, err := swe.RiseSet(sunsetJD, loc.Lat, loc.Lon, swe.Sun, swe.Rise)
	if err != nil {
		return Computed{}, err
	}
	moonriseJD, mrErr := swe.RiseSet(sunriseJD, loc.Lat, loc.Lon, swe.Moon, swe.Rise)
	moonsetJD, msErr := swe.RiseSet(sunriseJD, loc.Lat, loc.Lon, swe.Moon, swe.Set)
	a, err := positions(sunriseJD)
	if err != nil {
		return Computed{}, err
	}
	tithi := int(a.diff / 12)
	nak := int(a.moon / (360.0 / 27))
	yoga := int(a.sum / (360.0 / 27))
	half := int(a.diff / 6)
	tithiEnd, err := findBoundary(sunriseJD, a.diff, 12, func(x angles) float64 { return x.diff })
	if err != nil {
		return Computed{}, err
	}
	nakEnd, err := findBoundary(sunriseJD, a.moon, 360.0/27, func(x angles) float64 { return x.moon })
	if err != nil {
		return Computed{}, err
	}
	yogaEnd, err := findBoundary(sunriseJD, a.sum, 360.0/27, func(x angles) float64 { return x.sum })
	if err != nil {
		return Computed{}, err
	}
	karanaEnd, err := findBoundary(sunriseJD, a.diff, 6, func(x angles) float64 { return x.diff })
	if err != nil {
		return Computed{}, err
	}
	sunrise, sunset := jdTime(sunriseJD), jdTime(sunsetJD)
	if sunrise.In(zone).Format("2006-01-02") != date.Format("2006-01-02") {
		return Computed{}, fmt.Errorf("no sunrise on this local date")
	}
	month, err := lunarMonth(sunriseJD)
	if err != nil {
		return Computed{}, err
	}
	day := Day{SchemaVersion: SchemaVersion, Date: date.Format("2006-01-02"), Location: loc,
		Sunrise: clock(sunrise, zone), Sunset: clock(sunset, zone), Vaara: vaaraNames[int(date.Weekday())],
		Paksha: map[bool]string{true: "Shukla", false: "Krishna"}[a.diff < 180], LunarMonth: month,
		Tithi: Limb{tithiNames[tithi], tithi + 1, dayClock(tithiEnd, date, zone)}, Nakshatra: Limb{nakshatraNames[nak], nak + 1, dayClock(nakEnd, date, zone)},
		Yoga: Limb{yogaNames[yoga], yoga + 1, dayClock(yogaEnd, date, zone)}, Karana: Limb{karanaName(half), half + 1, dayClock(karanaEnd, date, zone)}, Festivals: []string{},
	}
	if mrErr == nil && moonriseJD < nextRise {
		day.Moonrise = dayClock(moonriseJD, date, zone)
	}
	if msErr == nil && moonsetJD < nextRise {
		day.Moonset = dayClock(moonsetJD, date, zone)
	}
	applyWindows(&day, sunrise, sunset, zone, date.Weekday())
	day.Choghadiya = append(choghadiya(sunrise, sunset, date.Weekday(), true, zone), choghadiya(sunset, jdTime(nextRise), date.Weekday(), false, zone)...)
	return Computed{day, sunrise, sunset}, nil
}

// Amanta month: name from the solar sign at the ending new moon. Equal
// signs at consecutive new moons indicate an intercalary (Adhika) month.
func lunarMonth(jd float64) (string, error) {
	next, e := newMoonAfter(jd)
	if e != nil {
		return "", e
	}
	prev, e := newMoonAfter(next - 32)
	if e != nil {
		return "", e
	}
	a, e := positions(next)
	if e != nil {
		return "", e
	}
	b, e := positions(prev)
	if e != nil {
		return "", e
	}
	sign := int(a.sun / 30)
	name := monthNames[sign]
	if int(b.sun/30) == sign {
		name = "Adhika " + name
	}
	return name, nil
}
func newMoonAfter(jd float64) (float64, error) {
	a, e := positions(jd)
	if e != nil {
		return 0, e
	}
	for hi := jd + 0.5; hi <= jd+35; hi += 0.5 {
		b, e := positions(hi)
		if e != nil {
			return 0, e
		}
		if a.diff > 180 && b.diff < 180 {
			lo := hi - 0.5
			for i := 0; i < 45; i++ {
				mid := (lo + hi) / 2
				c, e := positions(mid)
				if e != nil {
					return 0, e
				}
				if c.diff > 180 {
					lo = mid
				} else {
					hi = mid
				}
			}
			return (lo + hi) / 2, nil
		}
		a = b
	}
	return 0, fmt.Errorf("could not bracket new moon")
}

func findBoundary(start, initial, segment float64, value func(angles) float64) (float64, error) {
	target := (math.Floor(initial/segment) + 1) * segment
	unwrapped := func(jd float64) (float64, error) {
		a, e := positions(jd)
		if e != nil {
			return 0, e
		}
		v := value(a)
		for v+1e-9 < initial {
			v += 360
		}
		return v, nil
	}
	lo, hi := start, start+2
	for i := 0; i < 48; i++ {
		mid := (lo + hi) / 2
		v, e := unwrapped(mid)
		if e != nil {
			return 0, e
		}
		if v >= target {
			hi = mid
		} else {
			lo = mid
		}
	}
	return (lo + hi) / 2, nil
}
