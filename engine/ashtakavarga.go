package engine

// Ashtakavarga (BPHS): for each of the seven grahas, the houses counted from
// each contributor (the seven grahas and the lagna) where it receives a bindu.
// Totals per graha are 48, 49, 39, 54, 56, 52, 39 (Sarvashtakavarga 337).
var avContributors = []string{"sun", "moon", "mars", "mercury", "jupiter", "venus", "saturn", "lagna"}

var avTable = map[string][8][]int{
	"sun":     {{1, 2, 4, 7, 8, 9, 10, 11}, {3, 6, 10, 11}, {1, 2, 4, 7, 8, 9, 10, 11}, {3, 5, 6, 9, 10, 11, 12}, {5, 6, 9, 11}, {6, 7, 12}, {1, 2, 4, 7, 8, 9, 10, 11}, {3, 4, 6, 10, 11, 12}},
	"moon":    {{3, 6, 7, 8, 10, 11}, {1, 3, 6, 7, 10, 11}, {2, 3, 5, 6, 9, 10, 11}, {1, 3, 4, 5, 7, 8, 10, 11}, {1, 4, 7, 8, 10, 11, 12}, {3, 4, 5, 7, 9, 10, 11}, {3, 5, 6, 11}, {3, 6, 10, 11}},
	"mars":    {{3, 5, 6, 10, 11}, {3, 6, 11}, {1, 2, 4, 7, 8, 10, 11}, {3, 5, 6, 11}, {6, 10, 11, 12}, {6, 8, 11, 12}, {1, 4, 7, 8, 9, 10, 11}, {1, 3, 6, 10, 11}},
	"mercury": {{5, 6, 9, 11, 12}, {2, 4, 6, 8, 10, 11}, {1, 2, 4, 7, 8, 9, 10, 11}, {1, 3, 5, 6, 9, 10, 11, 12}, {6, 8, 11, 12}, {1, 2, 3, 4, 5, 8, 9, 11}, {1, 2, 4, 7, 8, 9, 10, 11}, {1, 2, 4, 6, 8, 10, 11}},
	"jupiter": {{1, 2, 3, 4, 7, 8, 9, 10, 11}, {2, 5, 7, 9, 11}, {1, 2, 4, 7, 8, 10, 11}, {1, 2, 4, 5, 6, 9, 10, 11}, {1, 2, 3, 4, 7, 8, 10, 11}, {2, 5, 6, 9, 10, 11}, {3, 5, 6, 12}, {1, 2, 4, 5, 6, 7, 9, 10, 11}},
	"venus":   {{8, 11, 12}, {1, 2, 3, 4, 5, 8, 9, 11, 12}, {3, 5, 6, 9, 11, 12}, {3, 5, 6, 9, 11}, {5, 8, 9, 10, 11}, {1, 2, 3, 4, 5, 8, 9, 10, 11}, {3, 4, 5, 8, 9, 10, 11}, {1, 2, 3, 4, 5, 8, 9, 11}},
	"saturn":  {{1, 2, 4, 7, 8, 10, 11}, {3, 6, 11}, {3, 5, 6, 10, 11, 12}, {6, 8, 9, 10, 11, 12}, {5, 6, 11, 12}, {6, 11, 12}, {3, 5, 6, 11}, {1, 3, 4, 6, 10, 11}},
}

// AVGrahas lists the grahas with a Bhinnashtakavarga, in engine order.
var AVGrahas = []string{"sun", "moon", "mars", "mercury", "jupiter", "venus", "saturn"}

type Ashtakavarga struct {
	// Bhinna[graha][sign] bindus by zodiac sign (index 0 = Mesha).
	Bhinna map[string][12]int `json:"bhinna"`
	// Sarva[sign] is the Sarvashtakavarga total per sign.
	Sarva [12]int `json:"sarva"`
}

func ComputeAshtakavarga(c Chart) Ashtakavarga {
	gs := chartMap(c)
	pos := map[string]int{"lagna": int(c.Ascendant.Longitude/30) % 12}
	for _, id := range AVGrahas {
		pos[id] = signOf(gs[id])
	}
	av := Ashtakavarga{Bhinna: map[string][12]int{}}
	for _, g := range AVGrahas {
		var row [12]int
		for ci, contributor := range avContributors {
			for _, h := range avTable[g][ci] {
				row[(pos[contributor]+h-1)%12]++
			}
		}
		av.Bhinna[g] = row
		for s := range row {
			av.Sarva[s] += row[s]
		}
	}
	return av
}
