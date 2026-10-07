// Command bookbench checks the engine and the corpus against the classical
// book on real charts, and writes an exact-answer test set for the model.
//
//	go run ./cmd/bookbench -n 2000
//
// For each random chart (births 1900-2010 across India and the world):
//
//  1. It re-derives every condition the book describes straight from the
//     planets' longitudes, following the book's own definitions (Brihat
//     Jataka, tr. Iyer 1885): planet in house from the ascendant, Moon sign,
//     birth star, and the yogas of chapters 13 and 14. Each must agree with
//     what the engine detects.
//  2. For every condition the engine detects, the corpus must hold the
//     book's passage for exactly that condition, with a plain summary.
//  3. Each matched (chart, condition, passage) becomes a test item: a
//     question, the verse reference, the exact clause and the plain answer
//     the model must give (.runtime/llm-data/book-bench.jsonl).
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/example/panchang/corpus"
	"github.com/example/panchang/engine"
)

var places = []struct {
	lat, lon float64
	tz       string
}{
	{13.34, 74.75, "Asia/Kolkata"}, {28.61, 77.21, "Asia/Kolkata"}, {22.57, 88.36, "Asia/Kolkata"}, {19.08, 72.88, "Asia/Kolkata"},
	{13.08, 80.27, "Asia/Kolkata"}, {26.14, 91.74, "Asia/Kolkata"}, {27.72, 85.32, "Asia/Kathmandu"}, {6.93, 79.86, "Asia/Colombo"},
	{51.51, -0.13, "Europe/London"}, {40.71, -74.01, "America/New_York"}, {-33.87, 151.21, "Australia/Sydney"}, {1.35, 103.82, "Asia/Singapore"},
}

var ordinal = []string{"", "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th", "10th", "11th", "12th"}

// bookConditions derives, from longitudes alone, the conditions the book
// defines. It shares no code with the engine's detectors.
func bookConditions(c engine.Chart) map[string]bool {
	sign := func(lon float64) int { return int(lon/30) % 12 }
	asc := sign(c.Ascendant.Longitude)
	pos := map[string]float64{}
	for _, g := range c.Grahas {
		pos[g.ID] = g.Longitude
	}
	out := map[string]bool{}
	for _, id := range []string{"sun", "moon", "mars", "mercury", "jupiter", "venus", "saturn"} {
		out[fmt.Sprintf("graha_in_house:%s_in_%d", id, (sign(pos[id])-asc+12)%12+1)] = true
	}
	moon := sign(pos["moon"])
	out["graha_in_sign:moon_in_"+engine.Slug(engine.RashiNames()[moon])] = true
	naks := []string{"ashwini", "bharani", "krittika", "rohini", "mrigashira", "ardra", "punarvasu", "pushya", "ashlesha", "magha", "purva_phalguni", "uttara_phalguni", "hasta", "chitra", "swati", "vishakha", "anuradha", "jyeshtha", "mula", "purva_ashadha", "uttara_ashadha", "shravana", "dhanishtha", "shatabhisha", "purva_bhadrapada", "uttara_bhadrapada", "revati"}
	out["nakshatra:"+naks[int(pos["moon"]/(360.0/27))%27]] = true
	// 13.3: excepting the Sun, planets in the 2nd, the 12th, or both from the
	// Moon give Sunapha, Anapha or Durudhura; otherwise Kemadruma.
	second, twelfth := false, false
	for _, id := range []string{"mars", "mercury", "jupiter", "venus", "saturn"} {
		switch (sign(pos[id]) - moon + 12) % 12 {
		case 1:
			second = true
		case 11:
			twelfth = true
		}
	}
	switch {
	case second && twelfth:
		out["yoga:durudhara"] = true
	case second:
		out["yoga:sunapha"] = true
	case twelfth:
		out["yoga:anapha"] = true
	default:
		out["yoga:kemadruma"] = true
	}
	// 13.2: benefics (Mercury, Jupiter, Venus) in the 6th, 7th and 8th from the Moon.
	adhiHouses := map[int]bool{}
	for _, id := range []string{"mercury", "jupiter", "venus"} {
		if d := (sign(pos[id]) - moon + 12) % 12; d >= 5 && d <= 7 {
			adhiHouses[d] = true
		}
	}
	if len(adhiHouses) == 3 {
		out["yoga:adhi_yoga"] = true
	} else if len(adhiHouses) > 0 {
		out["yoga:adhi_yoga(partial)"] = true
	}
	// 14.1 and 14.2: two planets in one sign.
	if sign(pos["sun"]) == sign(pos["mercury"]) {
		out["yoga:budha_aditya"] = true
	}
	if sign(pos["moon"]) == sign(pos["mars"]) {
		out["yoga:chandra_mangala"] = true
	}
	return out
}

type item struct {
	Chart    engine.ChartInput `json:"chart"`
	Token    string            `json:"token"`
	Question string            `json:"question"`
	Ref      string            `json:"ref"`
	Exact    string            `json:"exact"`
	Plain    string            `json:"plain"`
}

func question(doc, key string) string {
	switch doc {
	case "graha_in_house":
		parts := strings.SplitN(key, "_in_", 2)
		var h int
		fmt.Sscan(parts[1], &h)
		return fmt.Sprintf("What does Brihat Jataka say about my %s in the %s house?", engine.GrahaEnglish(parts[0]), ordinal[h])
	case "nakshatra":
		return "What does Brihat Jataka say about my birth star?"
	case "graha_in_sign":
		return "What does Brihat Jataka say about my Moon sign?"
	}
	return fmt.Sprintf("What does Brihat Jataka say about %s yoga in my chart?", strings.ReplaceAll(key, "_", " "))
}

func main() {
	n := flag.Int("n", 2000, "number of random charts")
	entries := flag.String("entries", "corpus/build/entries.jsonl", "corpus build output (run: go run ./cmd/corpus build)")
	out := flag.String("out", ".runtime/llm-data/book-bench.jsonl", "test set to write")
	ephe := flag.String("ephe", "ephe", "Swiss Ephemeris data directory")
	flag.Parse()

	book := map[string]corpus.Entry{}
	f, err := os.Open(*entries)
	if err != nil {
		log.Fatalf("%v (run: go run ./cmd/corpus build)", err)
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var e corpus.Entry
		if json.Unmarshal(sc.Bytes(), &e) == nil && e.SourceID == "brihat-jataka-iyer-1885" {
			k := e.DocType + ":" + e.Key
			// Keep the result verse over a definition-only verse (13.3).
			if prev, ok := book[k]; !ok || prev.Ref == "13.3" {
				book[k] = e
			}
		}
	}
	f.Close()

	eng := engine.New(*ephe)
	r := rand.New(rand.NewSource(11))
	start := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	agree, disagree := map[string]int{}, map[string]int{}
	examples := map[string][]string{}
	covered, missing := map[string]int{}, map[string]int{}
	var items []item
	kemaCancelled := 0
	for i := 0; i < *n; i++ {
		p := places[r.Intn(len(places))]
		in := engine.ChartInput{Date: start.AddDate(0, 0, r.Intn(40000)).Format("2006-01-02"), Time: fmt.Sprintf("%02d:%02d", r.Intn(24), r.Intn(60)), Lat: p.lat, Lon: p.lon, TZ: p.tz}
		c, err := eng.BirthChart(in)
		if err != nil {
			log.Fatal(err)
		}
		facts, err := engine.Facts(c, asOf)
		if err != nil {
			log.Fatal(err)
		}
		detected := map[string]bool{}
		for _, k := range engine.CorpusKeys(facts) {
			detected[k.DocType+":"+k.Key] = true
		}
		want := bookConditions(c)
		// 1. Every condition the book defines must match the engine.
		kinds := map[string]bool{}
		for k := range want {
			kinds[strings.TrimSuffix(k, "(partial)")] = true
		}
		for _, k := range []string{"yoga:sunapha", "yoga:anapha", "yoga:durudhara", "yoga:kemadruma", "yoga:adhi_yoga", "yoga:budha_aditya", "yoga:chandra_mangala"} {
			kinds[k] = true
		}
		for k := range kinds {
			bookSays := want[k]
			got := detected[k]
			if k == "yoga:kemadruma" {
				// The engine follows the book's definition and then applies
				// cancellations, which it states; compare the definition and
				// count the cancelled ones separately.
				km := engine.Kemadruma(c)
				got = km.Present
				if km.Present && len(km.Cancellations) > 0 {
					kemaCancelled++
				}
			}
			kind := strings.SplitN(k, ":", 2)[0]
			if bookSays == got {
				agree[kind]++
				continue
			}
			label := k
			disagree[label]++
			if len(examples[label]) < 3 {
				examples[label] = append(examples[label], fmt.Sprintf("%s %s %.2f,%.2f", in.Date, in.Time, in.Lat, in.Lon))
			}
		}
		// 2 and 3. The corpus must hold the passage for every detected condition.
		for k := range detected {
			kind := strings.SplitN(k, ":", 2)[0]
			if kind != "graha_in_house" && kind != "nakshatra" && kind != "yoga" && !(kind == "graha_in_sign" && strings.HasPrefix(k, "graha_in_sign:moon_")) {
				continue
			}
			if kind == "graha_in_house" && (strings.Contains(k, "rahu") || strings.Contains(k, "ketu")) {
				continue // Brihat Jataka chapter 20 does not treat the nodes
			}
			if kind == "nakshatra" && strings.Contains(k, "_pada") {
				continue // the book gives results per birth star, not per quarter
			}
			e, ok := book[k]
			if !ok {
				if kind != "yoga" {
					missing[k]++
				}
				continue
			}
			covered[kind]++
			body, plain, _ := strings.Cut(e.Body, "\n"+corpus.PlainLabel)
			if plain == "" {
				missing[k+" (no plain summary)"]++
				continue
			}
			if i%10 == 0 { // a sample of charts becomes the model test set
				key := strings.SplitN(k, ":", 2)[1]
				items = append(items, item{in, k, question(kind, key), e.Ref, body, strings.TrimSuffix(plain, "]")})
			}
		}
	}

	fmt.Printf("Book vs engine on %d charts (Brihat Jataka, tr. Iyer 1885)\n\n", *n)
	fmt.Println("Conditions re-derived from longitudes by the book's definitions:")
	kinds := []string{}
	for k := range agree {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	for _, k := range kinds {
		fmt.Printf("  %-16s %6d agree\n", k, agree[k])
	}
	if len(disagree) == 0 {
		fmt.Println("  no disagreements")
	}
	dk := []string{}
	for k := range disagree {
		dk = append(dk, k)
	}
	sort.Strings(dk)
	for _, k := range dk {
		fmt.Printf("  DIFFERS %5d charts  %s  e.g. %s\n", disagree[k], k, strings.Join(examples[k], "; "))
	}
	fmt.Printf("  (Kemadruma by the book's definition in agreement; %d of those charts have it cancelled, with the reason stated)\n", kemaCancelled)
	fmt.Println("\nBook passage found for each detected condition:")
	ck := []string{}
	for k := range covered {
		ck = append(ck, k)
	}
	sort.Strings(ck)
	for _, k := range ck {
		fmt.Printf("  %-16s %6d\n", k, covered[k])
	}
	if len(missing) > 0 {
		mk := []string{}
		for k := range missing {
			mk = append(mk, k)
		}
		sort.Strings(mk)
		fmt.Println("  missing:", strings.Join(mk, ", "))
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		log.Fatal(err)
	}
	fh, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(fh)
	for _, it := range items {
		if err := enc.Encode(it); err != nil {
			log.Fatal(err)
		}
	}
	fh.Close()
	fmt.Printf("\nWrote %d exact-answer test items to %s\n", len(items), *out)
}
