package engine

import (
	"strings"
	"testing"
)

// moonChart places the Moon in Mesha with the lagna in the given sign and
// the other planets at the given signs (0 = Mesha).
func moonChart(lagna int, at map[string]int) Chart {
	c := Chart{Ascendant: Point{Longitude: float64(lagna*30 + 5)}}
	for _, id := range GrahaIDs {
		s, ok := at[id]
		if !ok {
			s = 2 // Mithuna: the 3rd from the Moon, outside every rule here
		}
		if id == "moon" {
			s = 0
		}
		c.Grahas = append(c.Grahas, Graha{ID: id, Longitude: float64(s*30 + 10)})
	}
	return c
}

func hasYoga(c Chart, name string) bool {
	for _, y := range DetectYogas(c, Dignities(c)) {
		if y.Name == name {
			return true
		}
	}
	return false
}

// Brihat Jataka 13.2: benefics in the 6th, 7th and 8th from the Moon, all three.
func TestAdhiNeedsAllThreeHouses(t *testing.T) {
	full := moonChart(1, map[string]int{"mercury": 5, "jupiter": 6, "venus": 7})
	if !hasYoga(full, "Adhi Yoga") {
		t.Fatal("benefics in the 6th, 7th and 8th: no Adhi")
	}
	for name, at := range map[string]map[string]int{
		"one benefic":         {"jupiter": 6},
		"two houses":          {"jupiter": 6, "venus": 7},
		"two in the same one": {"jupiter": 6, "venus": 6, "mercury": 7},
	} {
		if hasYoga(moonChart(1, at), "Adhi Yoga") {
			t.Errorf("%s: Adhi claimed", name)
		}
	}
}

// Brihat Jataka 13.3: no planet but the Sun beside the Moon is Kemadruma;
// it stands only when no cancellation applies.
func TestKemadrumaFollowsTheBookWithCancellations(t *testing.T) {
	// Everything in the 3rd from the Moon; lagna Simha puts the Moon in the 9th.
	bare := moonChart(4, map[string]int{"sun": 1})
	if k := Kemadruma(bare); !k.Present || len(k.Cancellations) != 0 || !hasYoga(bare, "Kemadruma") {
		t.Fatalf("bare Moon: %+v", k)
	}
	if k := Kemadruma(moonChart(4, map[string]int{"venus": 1})); k.Present {
		t.Fatal("Venus in the 2nd from the Moon is Sunapha, not Kemadruma")
	}
	cases := map[string]struct {
		c    Chart
		want string
	}{
		"Moon in a kendra from the lagna": {moonChart(0, nil), "kendra from the lagna"},
		"a planet with the Moon":          {moonChart(4, map[string]int{"saturn": 0}), "Saturn is with the Moon"},
		"a planet in a kendra from Moon":  {moonChart(4, map[string]int{"jupiter": 3}), "Jupiter is in a kendra from the Moon"},
	}
	for name, tc := range cases {
		k := Kemadruma(tc.c)
		if !k.Present || !strings.Contains(strings.Join(k.Cancellations, "; "), tc.want) || hasYoga(tc.c, "Kemadruma") {
			t.Errorf("%s: %+v", name, k)
		}
	}
}
