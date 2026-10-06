package corpus

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// Every planet-in-house passage must speak about its own house: by number,
// as the ascendant for the 1st, or through a rule that names the houses it
// covers. A verse that lists several houses must not be cited whole for
// one of them (sun_in_8 once carried only the Sun-in-7th clause).
func TestHouseExcerptsNameTheirHouse(t *testing.T) {
	b, err := files.ReadFile("maps/brihat_jataka_iyer_1885.json")
	if err != nil {
		t.Fatal(err)
	}
	var sm sourceMap
	if err = json.Unmarshal(b, &sm); err != nil {
		t.Fatal(err)
	}
	key := regexp.MustCompile(`^(\w+)_in_(\d+)$`)
	ord := []string{"", "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th", "10th", "11th", "12th"}
	jupiter := map[int]string{1: "learned", 2: "of good speech", 3: "a niggard", 4: "will live in comfort", 5: "will be intelligent", 6: "will have no enemies", 7: "superior to those of his father", 8: "unsuited to his rank", 9: "a devotee", 10: "possessed of wealth", 11: "full of gain", 12: "fearful deeds"}
	seen := map[string]bool{}
	for _, me := range sm.Entries {
		if me.DocType != "graha_in_house" {
			continue
		}
		m := key.FindStringSubmatch(me.Key)
		if m == nil {
			t.Fatalf("bad key %s", me.Key)
		}
		var h int
		fmt.Sscan(m[2], &h)
		seen[me.Key] = true
		texts := []string{me.Text}
		for _, v := range me.Via {
			texts = append(texts, v.Text)
		}
		all := strings.ToLower(strings.Join(texts, " "))
		ok := strings.Contains(all, ord[h]+" house") || h == 1 && strings.Contains(all, "ascendant") || h == 3 && strings.Contains(all, "third house")
		if strings.Contains(all, "12 signs from the ascendant") {
			ok = strings.HasSuffix(strings.TrimSpace(strings.ToLower(texts[len(texts)-1])), jupiter[h]) || strings.Contains(texts[len(texts)-1], jupiter[h])
		}
		if !ok {
			t.Errorf("%s (%s) does not speak about the %s house: %q", me.Key, me.Ref, ord[h], all)
		}
		if strings.TrimSpace(me.Plain) == "" {
			t.Errorf("%s has no plain-language summary", me.Key)
		}
	}
	for _, g := range []string{"sun", "moon", "mars", "mercury", "jupiter", "venus", "saturn"} {
		for h := 1; h <= 12; h++ {
			if !seen[fmt.Sprintf("%s_in_%d", g, h)] {
				t.Errorf("no Brihat Jataka passage for %s in the %s house", g, ord[h])
			}
		}
	}
}

// Every transit passage (Brihat Samhita 104) must speak about its own house
// from the Moon and its own planet, for the Sun to Saturn in all 12 houses.
func TestTransitExcerptsNameTheirHouseAndPlanet(t *testing.T) {
	b, err := files.ReadFile("maps/brihat_samhita_iyer_1884.json")
	if err != nil {
		t.Fatal(err)
	}
	var sm sourceMap
	if err = json.Unmarshal(b, &sm); err != nil {
		t.Fatal(err)
	}
	key := regexp.MustCompile(`^(\w+)_transit_(\d+)$`)
	ord := []string{"", "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th", "10th", "11th", "12th"}
	name := map[string]string{"sun": "sun", "moon": "moon", "mars": "mars", "mercury": "mercury", "jupiter": "jupiter", "venus": "venus", "saturn": "saturn"}
	seen := map[string]bool{}
	for _, me := range sm.Entries {
		m := key.FindStringSubmatch(me.Key)
		if me.DocType != "transit" || m == nil {
			t.Fatalf("bad transit entry %s:%s", me.DocType, me.Key)
		}
		var h int
		fmt.Sscan(m[2], &h)
		seen[me.Key] = true
		text := strings.ToLower(me.Text)
		house := strings.Contains(text, ord[h]+" house") || strings.Contains(text, ord[h]+" or ") ||
			h == 1 && (strings.Contains(text, "sign occupied by the moon") || strings.Contains(text, "sign occupied by herself")) ||
			h == 3 && strings.Contains(text, "third house")
		if !house {
			t.Errorf("%s (%s) does not speak about the %s house: %q", me.Key, me.Ref, ord[h], me.Text)
		}
		if !strings.Contains(text, name[m[1]]) && !strings.HasPrefix(text, "when he ") && !strings.HasPrefix(text, "when she ") {
			t.Errorf("%s (%s) does not name its planet: %q", me.Key, me.Ref, me.Text)
		}
		if strings.TrimSpace(me.Plain) == "" {
			t.Errorf("%s has no plain summary", me.Key)
		}
	}
	for _, g := range []string{"sun", "moon", "mars", "mercury", "jupiter", "venus", "saturn"} {
		for h := 1; h <= 12; h++ {
			if !seen[fmt.Sprintf("%s_transit_%d", g, h)] {
				t.Errorf("no Brihat Samhita passage for %s transiting the %s house", g, ord[h])
			}
		}
	}
}
