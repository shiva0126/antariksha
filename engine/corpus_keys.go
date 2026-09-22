package engine

import (
	"fmt"
	"sort"
	"strings"
)

// Corpus document types. The detection vocabulary below is the corpus schema:
// every token the engine can emit for a chart has a (DocType, Key) pair, and
// retrieval is an exact fetch on that pair rather than a fuzzy search.
const (
	DocYoga         = "yoga"
	DocGrahaInHouse = "graha_in_house"
	DocGrahaInSign  = "graha_in_sign"
	DocNakshatra    = "nakshatra"
	DocDasha        = "dasha"
	DocDignity      = "dignity"
	DocBhava        = "bhava"
	DocAspect       = "aspect"
)

// SystemParashari is the only school the engine detects; Jaimini entries live
// in their own namespace and must never answer a Parashari detection.
const SystemParashari = "parashari"

type CorpusKey struct {
	DocType string `json:"doc_type"`
	Key     string `json:"key"`
}

func (k CorpusKey) String() string { return k.DocType + ":" + k.Key }

// VocabEntry is one token the engine can emit. Label is a plain-English
// description used as the semantic query in retrieval smoke tests. Required
// tokens must resolve to at least one corpus entry for coverage to be green.
type VocabEntry struct {
	CorpusKey
	Label    string `json:"label"`
	Required bool   `json:"required"`
}

// YogaCatalog lists every yoga name DetectYogas can emit.
var YogaCatalog = []string{"Gajakesari", "Budha-Aditya", "Chandra-Mangala", "Ruchaka", "Bhadra", "Hamsa", "Malavya", "Sasa", "Raja Yoga", "Dhana Yoga", "Vipreet Raja Yoga", "Adhi Yoga", "Sunapha", "Anapha", "Durudhara", "Kemadruma", "Neecha Bhanga Raja Yoga", "Kala Sarpa"}

// GrahaIDs is the engine's graha order.
var GrahaIDs = []string{"sun", "moon", "mars", "mercury", "jupiter", "venus", "saturn", "rahu", "ketu"}

var DignityStates = []string{"exalted", "own", "friendly", "neutral", "enemy", "debilitated"}

var grahaEnglish = map[string]string{"sun": "Sun", "moon": "Moon", "mars": "Mars", "mercury": "Mercury", "jupiter": "Jupiter", "venus": "Venus", "saturn": "Saturn", "rahu": "Rahu", "ketu": "Ketu"}

// Slug converts a display name to a corpus key token: "Budha-Aditya" →
// "budha_aditya", "Purva Phalguni" → "purva_phalguni".
func Slug(s string) string {
	var b strings.Builder
	under := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			under = false
		} else if !under && b.Len() > 0 {
			b.WriteByte('_')
			under = true
		}
	}
	return strings.TrimSuffix(b.String(), "_")
}

func ordinal(n int) string {
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

// CorpusVocabulary returns every token the engine can emit, sorted.
func CorpusVocabulary() []VocabEntry {
	out := []VocabEntry{}
	add := func(doc, key, label string, required bool) {
		out = append(out, VocabEntry{CorpusKey{doc, key}, label, required})
	}
	for _, y := range YogaCatalog {
		add(DocYoga, Slug(y), y+" yoga", true)
	}
	for _, g := range GrahaIDs {
		for h := 1; h <= 12; h++ {
			add(DocGrahaInHouse, fmt.Sprintf("%s_in_%d", g, h), fmt.Sprintf("%s in the %s house from the ascendant", grahaEnglish[g], ordinal(h)), true)
		}
		for _, s := range rashiNames {
			add(DocGrahaInSign, g+"_in_"+Slug(s), fmt.Sprintf("%s in the sign %s", grahaEnglish[g], s), true)
		}
	}
	for _, n := range nakshatraNames {
		add(DocNakshatra, Slug(n), "Moon in the nakshatra "+n+" at birth", true)
		for p := 1; p <= 4; p++ {
			add(DocNakshatra, fmt.Sprintf("%s_pada%d", Slug(n), p), fmt.Sprintf("Moon in %s pada %d", n, p), false)
		}
	}
	for _, g := range GrahaIDs {
		add(DocDasha, "dasha_"+g, grahaEnglish[g]+" Vimshottari dasha period", true)
	}
	for _, st := range DignityStates {
		add(DocDignity, "dignity_"+st, "a graha in its "+st+" dignity", true)
	}
	for id := range exalted {
		for _, st := range DignityStates {
			req := st == "exalted" || st == "debilitated" || st == "own"
			add(DocDignity, st+"_"+id, fmt.Sprintf("%s %s", grahaEnglish[id], dignityPhrase(st)), req)
		}
	}
	for _, id := range []string{"moon", "mars", "mercury", "jupiter", "venus", "saturn"} {
		add(DocDignity, "combust_"+id, grahaEnglish[id]+" combust, too close to the Sun", true)
	}
	for _, id := range []string{"mars", "mercury", "jupiter", "venus", "saturn"} {
		add(DocDignity, "retrograde_"+id, grahaEnglish[id]+" retrograde", true)
	}
	for h := 1; h <= 12; h++ {
		add(DocBhava, fmt.Sprintf("bhava_%d", h), fmt.Sprintf("significations of the %s house", ordinal(h)), true)
	}
	for _, s := range rashiNames {
		add(DocBhava, "lagna_"+Slug(s), s+" rising as the ascendant", true)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}

func dignityPhrase(st string) string {
	switch st {
	case "exalted":
		return "exalted"
	case "debilitated":
		return "debilitated, in its sign of fall"
	case "own":
		return "in its own sign"
	}
	return "in a " + st + " sign"
}

// ThemeHouses are always retrieved: the reading schema has explicit lagna,
// career and relationship sections.
var ThemeHouses = []int{1, 7, 10}

// CorpusKeys returns the exact tokens detected in a chart, deduplicated and
// sorted. It is the only place the reading layer learns what to retrieve.
func CorpusKeys(f ChartFacts) []CorpusKey {
	seen := map[CorpusKey]bool{}
	add := func(doc, key string) {
		if key != "" {
			seen[CorpusKey{doc, key}] = true
		}
	}
	for _, y := range f.Yogas {
		add(DocYoga, Slug(y.Name))
	}
	asc := f.Chart.Ascendant.Longitude
	for _, g := range f.Chart.Grahas {
		add(DocGrahaInHouse, fmt.Sprintf("%s_in_%d", g.ID, houseOf(g.Longitude, asc)))
		add(DocGrahaInSign, g.ID+"_in_"+Slug(rashiNames[signOf(g)]))
		if g.ID == "moon" {
			add(DocNakshatra, Slug(g.Nakshatra))
			if g.NakshatraPada > 0 {
				add(DocNakshatra, fmt.Sprintf("%s_pada%d", Slug(g.Nakshatra), g.NakshatraPada))
			}
		}
	}
	for id, d := range f.Dignities {
		add(DocDignity, d.State+"_"+id)
		add(DocDignity, "dignity_"+d.State)
	}
	for id, c := range f.Combustion {
		if c {
			add(DocDignity, "combust_"+id)
		}
	}
	for _, id := range f.Retrograde {
		add(DocDignity, "retrograde_"+id)
	}
	for _, lord := range []string{f.Vimshottari.Current.Maha, f.Vimshottari.Current.Antara, f.Vimshottari.Upcoming.Lord} {
		if lord != "" {
			add(DocDasha, "dasha_"+lord)
		}
	}
	if len(f.Chart.Grahas) > 0 {
		add(DocBhava, "lagna_"+Slug(rashiNames[int(asc/30)%12]))
		for _, h := range ThemeHouses {
			add(DocBhava, fmt.Sprintf("bhava_%d", h))
		}
	}
	out := make([]CorpusKey, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}
