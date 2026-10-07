package reading

import (
	"fmt"
	"strings"

	"github.com/example/panchang/engine"
)

// A dasha lord gives the results of the houses it rules and the house it
// occupies, coloured by its strength: the classical principle behind
// reading a period for a particular chart rather than for the planet in
// general.

// ruledHouses lists the houses from the lagna that a graha rules.
func (in *insight) ruledHouses(id string) []int {
	var hs []int
	for h := 1; h <= 12; h++ {
		if engine.HouseLord(in.f.Chart, h) == id {
			hs = append(hs, h)
		}
	}
	return hs
}

// houseTone is +1 for the fortunate houses (trines and angles), -1 for the
// harder ones (6th, 8th, 12th) and 0 otherwise.
func houseTone(h int) float64 {
	switch {
	case inInts(h, []int{1, 5, 9, 4, 7, 10}):
		return 1
	case inInts(h, []int{6, 8, 12}):
		return -1
	}
	return 0
}

func houseLabel(h int) string {
	return fmt.Sprintf("%s house (%s)", ordinal(h), houseArea[h])
}

// dashaStory describes a dasha lord for this chart: the houses it rules
// and occupies, and its strength, with an overall lean.
func (in *insight) dashaStory(id string) (string, float64) {
	if id == "" {
		return "", 0
	}
	placed := in.house(id)
	ruled := in.ruledHouses(id)
	tone := 0.0
	for _, h := range ruled {
		tone += houseTone(h)
	}
	tone += 0.5 * houseTone(placed)
	var b strings.Builder
	b.WriteString(upper(theName(id)))
	if len(ruled) > 0 {
		labels := make([]string, len(ruled))
		for i, h := range ruled {
			labels[i] = houseLabel(h)
		}
		fmt.Fprintf(&b, " rules your %s, and", joinAnd(labels))
	}
	fmt.Fprintf(&b, " sits in your %s", houseLabel(placed))
	if c := in.condition(id); c != "" {
		fmt.Fprintf(&b, ", %s", c)
	}
	switch in.f.Dignities[id].State {
	case "exalted", "own":
		tone += 1
	case "debilitated":
		if !in.f.Dignities[id].NeechaBhanga {
			tone -= 1
		}
	}
	b.WriteString(". ")
	switch {
	case tone >= 1.5:
		b.WriteString("Its period traditionally brings out the good side of these areas.")
	case tone <= -1:
		b.WriteString("Its period traditionally asks for patience in these areas.")
	default:
		b.WriteString("Its period traditionally brings a mix in these areas.")
	}
	return b.String(), tone
}

// dashaHarmony describes how the sub-period lord stands from the main
// period lord: 6/8 apart traditionally means friction, angles and trines
// harmony.
func (in *insight) dashaHarmony(maha, antara string) string {
	if maha == "" || antara == "" || maha == antara {
		return ""
	}
	d := (in.signIdx(antara)-in.signIdx(maha)+12)%12 + 1
	switch {
	case d == 6 || d == 8:
		return fmt.Sprintf("%s sits %s from %s in your chart, a 6/8 relationship that traditionally makes the two periods pull in different directions.", upper(theName(antara)), ordinal(d), theName(maha))
	case inInts(d, []int{1, 4, 5, 7, 9, 10}):
		return fmt.Sprintf("%s sits %s from %s in your chart, a supportive relationship: the sub-period works with the main period.", upper(theName(antara)), ordinal(d), theName(maha))
	}
	return ""
}
