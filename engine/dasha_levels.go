package engine

import (
	"math"
	"time"
)

// subPeriods splits a period of `lord` running from start to end into the nine
// Vimshottari sub-periods, each proportional to its lord's years, starting
// with the period's own lord.
func subPeriods(lord string, start, end time.Time, level string) []DashaPeriod {
	cycle := 0
	for i, id := range dashaOrder {
		if id == lord {
			cycle = i
		}
	}
	total := end.Sub(start)
	cur := start
	out := make([]DashaPeriod, 0, 9)
	for n := 0; n < 9; n++ {
		l := dashaOrder[(cycle+n)%9]
		next := cur.Add(time.Duration(float64(total) * dashaYears[l] / 120))
		if n == 8 {
			next = end
		}
		out = append(out, DashaPeriod{Lord: l, From: cur.Format("2006-01-02"), To: next.Format("2006-01-02"), Level: level})
		cur = next
	}
	return out
}

func running(ps []DashaPeriod, at string) int {
	for i, p := range ps {
		if p.From <= at && at < p.To {
			return i
		}
	}
	return -1
}

// addLevels fills antardashas of the running mahadasha and pratyantardashas of
// the running antardasha.
func addLevels(d *Dasha, mahaStart time.Time, asOf time.Time) {
	if d.Current.Maha == "" {
		return
	}
	mahaEnd := mahaStart.Add(yearsDuration(dashaYears[d.Current.Maha]))
	d.Antaras = subPeriods(d.Current.Maha, mahaStart, mahaEnd, "antara")
	at := asOf.Format("2006-01-02")
	if i := running(d.Antaras, at); i >= 0 {
		a := d.Antaras[i]
		s, _ := time.Parse("2006-01-02", a.From)
		e, _ := time.Parse("2006-01-02", a.To)
		d.Pratyantaras = subPeriods(a.Lord, s, e, "pratyantara")
		if j := running(d.Pratyantaras, at); j >= 0 {
			d.Current.Pratyantara = d.Pratyantaras[j].Lord
		}
	}
}

// Yogini dasha: eight yoginis of 1–8 years (36-year cycle), starting from the
// yogini of the Moon's nakshatra.
var yoginiNames = []string{"Mangala", "Pingala", "Dhanya", "Bhramari", "Bhadrika", "Ulka", "Siddha", "Sankata"}
var yoginiLords = []string{"moon", "sun", "jupiter", "mars", "mercury", "saturn", "venus", "rahu"}

type YoginiPeriod struct {
	Yogini string `json:"yogini"`
	Lord   string `json:"lord"`
	From   string `json:"from"`
	To     string `json:"to"`
}

type YoginiDasha struct {
	Current  YoginiPeriod   `json:"current"`
	Sequence []YoginiPeriod `json:"sequence"`
}

func Yogini(c Chart, birth, asOf time.Time) YoginiDasha {
	moon := chartMap(c)["moon"]
	nakSize := 360.0 / 27
	nak := int(moon.Longitude / nakSize)
	// (nakshatra number + 3) mod 8, where 1 = Mangala and 0 = Sankata.
	idx := ((nak+1+3)%8 + 7) % 8
	elapsed := math.Mod(moon.Longitude, nakSize) / nakSize
	start := birth.Add(-yearsDuration(float64(idx+1) * elapsed))
	var out YoginiDasha
	cur := start
	for n := 0; n < 16; n++ {
		i := (idx + n) % 8
		end := cur.Add(yearsDuration(float64(i + 1)))
		from := cur
		if n == 0 {
			from = birth
		}
		p := YoginiPeriod{yoginiNames[i], yoginiLords[i], from.Format("2006-01-02"), end.Format("2006-01-02")}
		out.Sequence = append(out.Sequence, p)
		if !asOf.Before(from) && asOf.Before(end) {
			out.Current = p
		}
		cur = end
	}
	return out
}
