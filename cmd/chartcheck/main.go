// Command chartcheck compares the engine's planet positions with NASA JPL's
// Horizons ephemeris, an independent source, at random moments between 1900
// and 2050. It reports the largest difference per planet in arc-seconds and
// fails if any exceeds the limit.
//
//	go run ./cmd/chartcheck -n 60 -limit 5
//
// Positions compared are geocentric, apparent, tropical ecliptic longitudes
// of date (Horizons quantity 31). The engine's sidereal longitudes are these
// minus the Lahiri ayanamsa, which is checked separately against the value
// the engine itself uses for charts. Rahu and Ketu are not in Horizons.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/example/panchang/engine/swe"
)

var bodies = []struct {
	id      string
	swe     int
	horizon string
}{
	{"sun", swe.Sun, "10"}, {"moon", swe.Moon, "301"}, {"mercury", swe.Mercury, "199"}, {"venus", swe.Venus, "299"},
	{"mars", swe.Mars, "499"}, {"jupiter", swe.Jupiter, "599"}, {"saturn", swe.Saturn, "699"},
}

// horizons returns apparent ecliptic longitudes (degrees) for one body at
// the given UT Julian days, in order.
func horizons(body string, jds []float64) ([]float64, error) {
	tl := make([]string, len(jds))
	for i, jd := range jds {
		tl[i] = "'" + strconv.FormatFloat(jd, 'f', 6, 64) + "'"
	}
	q := url.Values{
		"format": {"json"}, "COMMAND": {"'" + body + "'"}, "OBJ_DATA": {"NO"}, "MAKE_EPHEM": {"YES"},
		"EPHEM_TYPE": {"OBSERVER"}, "CENTER": {"'500@399'"}, "QUANTITIES": {"'31'"}, "TIME_TYPE": {"UT"},
		"TLIST_TYPE": {"JD"}, "TLIST": {strings.Join(tl, " ")}, "CSV_FORMAT": {"YES"}, "ANG_FORMAT": {"DEG"},
		"EXTRA_PREC": {"YES"},
	}
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Get("https://ssd.jpl.nasa.gov/api/horizons.api?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var v struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if err = json.Unmarshal(b, &v); err != nil || v.Error != "" {
		return nil, fmt.Errorf("horizons %s: %v %s", body, err, v.Error)
	}
	start, end := strings.Index(v.Result, "$$SOE"), strings.Index(v.Result, "$$EOE")
	if start < 0 || end < 0 {
		return nil, fmt.Errorf("horizons %s: no ephemeris in reply", body)
	}
	var out []float64
	for _, line := range strings.Split(strings.TrimSpace(v.Result[start+5:end]), "\n") {
		cols := strings.Split(line, ",")
		// Date, solar-presence flag, lunar-presence flag, ObsEcLon, ObsEcLat
		if len(cols) < 5 {
			continue
		}
		lon, err := strconv.ParseFloat(strings.TrimSpace(cols[3]), 64)
		if err != nil {
			return nil, fmt.Errorf("horizons %s: %q: %v", body, line, err)
		}
		out = append(out, lon)
	}
	if len(out) != len(jds) {
		return nil, fmt.Errorf("horizons %s: %d rows for %d times", body, len(out), len(jds))
	}
	return out, nil
}

// arcsec is the smallest angle between two longitudes, in arc-seconds.
func arcsec(a, b float64) float64 {
	d := math.Mod(math.Abs(a-b), 360)
	if d > 180 {
		d = 360 - d
	}
	return d * 3600
}

func main() {
	n := flag.Int("n", 60, "number of random moments")
	limit := flag.Float64("limit", 5, "largest allowed difference in arc-seconds")
	ephe := flag.String("ephe", "ephe", "Swiss Ephemeris data directory")
	seed := flag.Int64("seed", 42, "random seed")
	flag.Parse()
	// Swiss Ephemeris keeps its settings per OS thread, as the engine does.
	runtime.LockOSThread()
	swe.SetEphemerisPath(*ephe)
	swe.SetLahiri()

	r := rand.New(rand.NewSource(*seed))
	start := swe.JulianDay(1900, 1, 1, 0)
	end := swe.JulianDay(2050, 12, 31, 0)
	jds := make([]float64, *n)
	for i := range jds {
		jds[i] = start + r.Float64()*(end-start)
	}
	sort.Float64s(jds)

	failed := false
	fmt.Printf("%-8s %10s %10s   (arc-seconds; %d moments 1900-2050, limit %.1f\")\n", "planet", "max", "mean", *n, *limit)
	for _, b := range bodies {
		ref, err := horizons(b.horizon, jds)
		if err != nil {
			log.Fatal(err)
		}
		worst, sum, worstAt := 0.0, 0.0, 0.0
		for i, jd := range jds {
			trop, _, _, _, err := swe.TropicalPosition3D(jd, b.swe)
			if err != nil {
				log.Fatal(err)
			}
			sid, _, _, _, err := swe.Position3D(jd, b.swe)
			if err != nil {
				log.Fatal(err)
			}
			// The sidereal chart must be the tropical position minus Lahiri.
			aya, err := swe.TrueAyanamsa(jd)
			if err != nil {
				log.Fatal(err)
			}
			if d := arcsec(sid, trop-aya); d > 0.01 {
				fmt.Printf("  %s sidereal differs from tropical minus ayanamsa by %.3f\" at JD %.4f\n", b.id, d, jd)
				failed = true
			}
			d := arcsec(trop, ref[i])
			sum += d
			if d > worst {
				worst, worstAt = d, jd
			}
		}
		status := "ok"
		if worst > *limit {
			status, failed = "FAIL", true
		}
		y, m, d, _ := swe.ReverseJulian(worstAt)
		fmt.Printf("%-8s %10.3f %10.3f   %s (worst on %04d-%02d-%02d)\n", b.id, worst, sum/float64(len(jds)), status, y, m, d)
	}
	if failed {
		os.Exit(1)
	}
}
