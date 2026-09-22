// Package places searches an embedded index of ~34,000 cities (GeoNames
// cities15000, CC BY 4.0, https://www.geonames.org) with IANA timezones, so
// any birthplace with more than 15,000 people resolves without network access.
package places

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

//go:embed cities.tsv.gz
var raw []byte

type Place struct {
	Name       string  `json:"name"`
	Region     string  `json:"region"`
	Country    string  `json:"country"`
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	TZ         string  `json:"tz"`
	Population int     `json:"population"`
	keys       []string
}

const Attribution = "Place data © GeoNames (geonames.org), CC BY 4.0"

var (
	once  sync.Once
	index []Place
)

// fold lowercases and strips diacritics so "Pondichéry" matches "pondicher".
func fold(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func load() {
	zr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return
	}
	data, err := io.ReadAll(zr)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Split(line, "\t")
		if len(f) < 9 {
			continue
		}
		lat, _ := strconv.ParseFloat(f[5], 64)
		lon, _ := strconv.ParseFloat(f[6], 64)
		pop, _ := strconv.Atoi(f[8])
		p := Place{Name: f[0], Region: f[3], Country: f[4], Lat: lat, Lon: lon, TZ: f[7], Population: pop}
		p.keys = []string{fold(f[0])}
		if f[1] != "" {
			p.keys = append(p.keys, fold(f[1]))
		}
		for _, a := range strings.Split(f[2], ";") {
			if a != "" {
				p.keys = append(p.keys, fold(a))
			}
		}
		index = append(index, p)
	}
}

// Count returns the number of indexed places.
func Count() int { once.Do(load); return len(index) }

// Search returns up to limit places whose name, ASCII name or an alternate
// name starts with the query (or has a word starting with it). Exact matches
// rank first, then larger populations. An optional ", region/country" suffix
// narrows the result ("Aurangabad, Bihar").
func Search(q string, limit int) []Place {
	once.Do(load)
	q = fold(strings.TrimSpace(q))
	if len([]rune(q)) < 2 {
		return []Place{}
	}
	filter := ""
	if i := strings.Index(q, ","); i >= 0 {
		q, filter = strings.TrimSpace(q[:i]), strings.TrimSpace(q[i+1:])
	}
	type hit struct {
		p     Place
		score int
	}
	var hits []hit
	for _, p := range index {
		best := -1
		for i, k := range p.keys {
			s := -1
			switch {
			case k == q:
				s = 3
			case strings.HasPrefix(k, q):
				s = 2
			case strings.Contains(k, " "+q):
				s = 1
			}
			if s >= 0 && i > 0 && s > 0 {
				s-- // alternate names rank just below the primary name
				if s < 0 {
					s = 0
				}
			}
			if s > best {
				best = s
			}
		}
		if best < 0 {
			continue
		}
		if filter != "" && !strings.HasPrefix(fold(p.Region), filter) && !strings.HasPrefix(fold(p.Country), filter) {
			continue
		}
		hits = append(hits, hit{p, best})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return hits[i].p.Population > hits[j].p.Population
	})
	out := []Place{}
	for i := 0; i < len(hits) && i < limit; i++ {
		out = append(out, hits[i].p)
	}
	return out
}
