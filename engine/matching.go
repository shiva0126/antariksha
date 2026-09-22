package engine

import (
	"fmt"
	"math"
)

// Ashtakoota (36-point) Guna Milan from the Moon's nakshatra and rashi of both
// charts, following the widely used North Indian scheme. Where schools differ
// (Vashya, Gana scoring) the most commonly published tables are used and the
// koota's description says so.

type Koota struct {
	Name        string  `json:"name"`
	Max         float64 `json:"max"`
	Score       float64 `json:"score"`
	Boy         string  `json:"boy"`
	Girl        string  `json:"girl"`
	Description string  `json:"description"`
}

type Match struct {
	Kootas     []Koota  `json:"kootas"`
	Total      float64  `json:"total"`
	Max        float64  `json:"max"`
	Verdict    string   `json:"verdict"`
	Doshas     []string `json:"doshas"`
	Exceptions []string `json:"exceptions"`
	BoyMangal  bool     `json:"boy_mangal_dosha"`
	GirlMangal bool     `json:"girl_mangal_dosha"`
	MangalNote string   `json:"mangal_note"`
	BoyMoon    string   `json:"boy_moon"`
	GirlMoon   string   `json:"girl_moon"`
}

var varnaNames = []string{"Shudra", "Vaishya", "Kshatriya", "Brahmin"}

// Varna by rashi (0 = Shudra … 3 = Brahmin).
var rashiVarna = []int{2, 1, 0, 3, 2, 1, 0, 3, 2, 1, 0, 3}

const (
	vChatushpada = iota
	vManava
	vJalachara
	vVanachara
	vKeeta
)

var vashyaNames = []string{"Chatushpada (quadruped)", "Manava (human)", "Jalachara (water)", "Vanachara (wild)", "Keeta (insect)"}

// vashyaScore[boy][girl], the common 5-group table.
var vashyaScore = [5][5]float64{
	{2, 1, 1, 0.5, 1},
	{1, 2, 0.5, 0, 1},
	{1, 0.5, 2, 1, 1},
	{0.5, 0, 1, 2, 0},
	{1, 1, 1, 0, 2},
}

// vashyaOf groups a Moon longitude; Dhanu and Makara are split at 15°.
func vashyaOf(lon float64) int {
	sign, deg := int(lon/30)%12, math.Mod(lon, 30)
	switch sign {
	case 0, 1:
		return vChatushpada
	case 2, 5, 6, 10:
		return vManava
	case 3, 11:
		return vJalachara
	case 4:
		return vVanachara
	case 7:
		return vKeeta
	case 8:
		if deg < 15 {
			return vManava
		}
		return vChatushpada
	case 9:
		if deg < 15 {
			return vChatushpada
		}
		return vJalachara
	}
	return vManava
}

var yoniNames = []string{"Horse", "Elephant", "Sheep", "Serpent", "Dog", "Cat", "Rat", "Cow", "Buffalo", "Tiger", "Deer", "Monkey", "Mongoose", "Lion"}

// Yoni animal of each nakshatra.
var nakYoni = []int{0, 1, 2, 3, 3, 4, 5, 2, 5, 6, 6, 7, 8, 9, 8, 9, 10, 10, 4, 11, 12, 11, 13, 0, 13, 7, 1}

var yoniScore = [14][14]float64{
	{4, 2, 2, 3, 2, 2, 2, 1, 0, 1, 3, 3, 2, 1},
	{2, 4, 3, 3, 2, 2, 2, 2, 3, 1, 2, 3, 2, 0},
	{2, 3, 4, 2, 1, 2, 1, 3, 3, 1, 2, 0, 3, 1},
	{3, 3, 2, 4, 2, 1, 1, 1, 1, 2, 2, 2, 0, 2},
	{2, 2, 1, 2, 4, 2, 1, 2, 2, 1, 0, 2, 1, 1},
	{2, 2, 2, 1, 2, 4, 0, 2, 2, 1, 3, 3, 2, 1},
	{2, 2, 1, 1, 1, 0, 4, 2, 2, 2, 2, 2, 1, 2},
	{1, 2, 3, 1, 2, 2, 2, 4, 3, 0, 3, 2, 2, 1},
	{0, 3, 3, 1, 2, 2, 2, 3, 4, 1, 2, 2, 2, 1},
	{1, 1, 1, 2, 1, 1, 2, 0, 1, 4, 1, 1, 2, 1},
	{3, 2, 2, 2, 0, 3, 2, 3, 2, 1, 4, 2, 2, 1},
	{3, 3, 0, 2, 2, 3, 2, 2, 2, 1, 2, 4, 3, 2},
	{2, 2, 3, 0, 1, 2, 1, 2, 2, 2, 2, 3, 4, 2},
	{1, 0, 1, 2, 1, 1, 2, 1, 1, 1, 1, 2, 2, 4},
}

var ganaNames = []string{"Deva", "Manushya", "Rakshasa"}

// Gana of each nakshatra (0 Deva, 1 Manushya, 2 Rakshasa).
var nakGana = []int{0, 1, 2, 1, 0, 1, 0, 0, 2, 2, 1, 1, 0, 2, 0, 2, 0, 2, 2, 1, 1, 0, 2, 2, 1, 1, 0}

// ganaScore[boy][girl].
var ganaScore = [3][3]float64{{6, 6, 1}, {5, 6, 0}, {1, 0, 6}}

var nadiNames = []string{"Adi (Vata)", "Madhya (Pitta)", "Antya (Kapha)"}

func nadiOf(nak int) int { return []int{0, 1, 2, 2, 1, 0}[nak%6] }

// relation returns 2 friend, 1 neutral, 0 enemy (naisargika) of a towards b.
func relation(a, b string) int {
	if a == b || friends[a][b] {
		return 2
	}
	if enemies[a][b] {
		return 0
	}
	return 1
}

func maitriScore(a, b string) float64 {
	if a == b {
		return 5
	}
	switch ra, rb := relation(a, b), relation(b, a); {
	case ra == 2 && rb == 2:
		return 5
	case ra+rb == 3 && ra != 0 && rb != 0: // friend + neutral
		return 4
	case ra == 1 && rb == 1:
		return 3
	case ra+rb == 2: // friend + enemy
		return 1
	case ra+rb == 1: // neutral + enemy
		return 0.5
	}
	return 0
}

func moonOf(c Chart) (Graha, error) {
	for _, g := range c.Grahas {
		if g.ID == "moon" {
			return g, nil
		}
	}
	return Graha{}, fmt.Errorf("chart has no Moon")
}

func nakIndex(lon float64) int { return int(normalize(lon)/(360.0/27)) % 27 }

// MatchCharts computes Ashtakoota Guna Milan for a boy's and a girl's chart.
func MatchCharts(boy, girl Chart) (Match, error) {
	bm, err := moonOf(boy)
	if err != nil {
		return Match{}, err
	}
	gm, err := moonOf(girl)
	if err != nil {
		return Match{}, err
	}
	bn, gn := nakIndex(bm.Longitude), nakIndex(gm.Longitude)
	bs, gs := int(bm.Longitude/30)%12, int(gm.Longitude/30)%12
	m := Match{Max: 36, BoyMoon: fmt.Sprintf("%s, %s pada %d", rashiNames[bs], nakshatraNames[bn], bm.NakshatraPada), GirlMoon: fmt.Sprintf("%s, %s pada %d", rashiNames[gs], nakshatraNames[gn], gm.NakshatraPada)}
	add := func(k Koota) { m.Kootas = append(m.Kootas, k); m.Total += k.Score }

	vb, vg := rashiVarna[bs], rashiVarna[gs]
	add(Koota{"Varna", 1, b2f(vb >= vg), varnaNames[vb], varnaNames[vg], "Spiritual and temperamental compatibility; 1 point when the groom's varna is equal to or higher than the bride's."})

	wb, wg := vashyaOf(bm.Longitude), vashyaOf(gm.Longitude)
	add(Koota{"Vashya", 2, vashyaScore[wb][wg], vashyaNames[wb], vashyaNames[wg], "Mutual attraction and influence, by rashi group (common 5-group table; schools differ in details)."})

	taraGood := func(from, to int) bool { r := ((to-from+27)%27 + 1) % 9; return r != 3 && r != 5 && r != 7 }
	tara := 0.0
	if taraGood(gn, bn) {
		tara += 1.5
	}
	if taraGood(bn, gn) {
		tara += 1.5
	}
	add(Koota{"Tara", 3, tara, nakshatraNames[bn], nakshatraNames[gn], "Birth-star harmony counted both ways; the 3rd, 5th and 7th taras (Vipat, Pratyak, Vadha) are inauspicious."})

	yb, yg := nakYoni[bn], nakYoni[gn]
	add(Koota{"Yoni", 4, yoniScore[yb][yg], yoniNames[yb], yoniNames[yg], "Physical and instinctive compatibility by nakshatra animal; sworn-enemy pairs score 0."})

	lb, lg := rashiLord[bs], rashiLord[gs]
	add(Koota{"Graha Maitri", 5, maitriScore(lb, lg), GrahaEnglish(lb), GrahaEnglish(lg), "Mental compatibility and friendship between the lords of the two Moon signs."})

	gb, gg := nakGana[bn], nakGana[gn]
	add(Koota{"Gana", 6, ganaScore[gb][gg], ganaNames[gb], ganaNames[gg], "Temperament: Deva, Manushya or Rakshasa nature of the birth stars."})

	d := (bs-gs+12)%12 + 1
	bhakoot := 7.0
	if inSet(d, 2, 12, 5, 9, 6, 8) {
		bhakoot = 0
	}
	add(Koota{"Bhakoot", 7, bhakoot, rashiNames[bs], rashiNames[gs], "Emotional and family well-being from the distance between Moon signs; 2/12, 5/9 and 6/8 placements form Bhakoot dosha."})

	nb, ng := nadiOf(bn), nadiOf(gn)
	nadi := 8.0
	if nb == ng {
		nadi = 0
	}
	add(Koota{"Nadi", 8, nadi, nadiNames[nb], nadiNames[ng], "Physiological constitution; the same nadi for both forms Nadi dosha, the most weighted koota."})

	if bhakoot == 0 {
		m.Doshas = append(m.Doshas, "Bhakoot dosha")
		if lb == lg || (relation(lb, lg) == 2 && relation(lg, lb) == 2) {
			m.Exceptions = append(m.Exceptions, "Bhakoot dosha is traditionally cancelled because the Moon-sign lords are the same or mutual friends.")
		}
	}
	if nadi == 0 {
		m.Doshas = append(m.Doshas, "Nadi dosha")
		if bn == gn && bs != gs {
			m.Exceptions = append(m.Exceptions, "Nadi dosha is traditionally cancelled: same nakshatra but different Moon signs.")
		} else if bs == gs && bn != gn {
			m.Exceptions = append(m.Exceptions, "Nadi dosha is traditionally cancelled: same Moon sign but different nakshatras.")
		} else if bn == gn && bm.NakshatraPada != gm.NakshatraPada {
			m.Exceptions = append(m.Exceptions, "Some traditions cancel Nadi dosha when the nakshatra is shared but the padas differ.")
		}
	}
	if ganaScore[gb][gg] <= 1 && gb != gg {
		m.Doshas = append(m.Doshas, "Gana dosha")
	}

	m.BoyMangal, _ = MangalDosha(boy)
	m.GirlMangal, _ = MangalDosha(girl)
	switch {
	case m.BoyMangal && m.GirlMangal:
		m.MangalNote = "Both charts have lagna-based Mangal dosha, which traditions treat as mutually cancelling."
	case m.BoyMangal || m.GirlMangal:
		m.MangalNote = "Only one chart has lagna-based Mangal dosha; traditions also check it from the Moon and Venus and for cancellations before concluding."
		m.Doshas = append(m.Doshas, "Mangal dosha (one chart)")
	default:
		m.MangalNote = "Neither chart has lagna-based Mangal dosha."
	}

	switch {
	case m.Total < 18:
		m.Verdict = "Below 18 points: traditionally not recommended without a detailed chart comparison."
	case m.Total <= 24:
		m.Verdict = "18–24 points: an average, acceptable match."
	case m.Total <= 32:
		m.Verdict = "25–32 points: a good match."
	default:
		m.Verdict = "33–36 points: an excellent match."
	}
	if m.Doshas == nil {
		m.Doshas = []string{}
	}
	if m.Exceptions == nil {
		m.Exceptions = []string{}
	}
	return m, nil
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
