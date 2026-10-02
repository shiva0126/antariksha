package divination

import "testing"

func TestNumerologyTransparentDateAndName(t *testing.T) {
	r, err := ReadNumerology("1990-01-01", "John", 2026)
	if err != nil {
		t.Fatal(err)
	}
	if r.LifePath.Number != 3 || r.Birthday.Number != 1 || r.PersonalYear.Number != 3 {
		t.Fatalf("unexpected date numbers: %+v", r)
	}
	if r.Expression == nil || r.Expression.Number != 2 || r.SoulUrge == nil || r.SoulUrge.Number != 6 || r.Personality == nil || r.Personality.Number != 5 {
		t.Fatalf("unexpected name numbers: %+v", r)
	}
	if r.Method == "" || r.Note == "" {
		t.Fatal("calculation method and caveat must be returned")
	}
}
func TestNumerologyMasterNumbersAndInput(t *testing.T) {
	r, err := ReadNumerology("2000-02-29", "", 2026)
	if err != nil {
		t.Fatal(err)
	}
	if r.Birthday.Number != 11 {
		t.Fatalf("birthday master number reduced: %+v", r.Birthday)
	}
	for _, date := range []string{"2001-02-29", "1900-02-29", "2026-1-01"} {
		if _, err := ReadNumerology(date, "", 2026); err == nil {
			t.Errorf("accepted invalid date %s", date)
		}
	}
	if _, err = ReadNumerology("2000-02-29", "शिव", 2026); err == nil {
		t.Fatal("silently transliterated non-Latin name")
	}
}
func TestTarotDrawUniqueAndBounded(t *testing.T) {
	deck := TarotDeck()
	if len(deck) != 78 {
		t.Fatalf("deck has %d cards", len(deck))
	}
	for _, count := range []int{1, 3} {
		r, err := DrawTarot(count, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Cards) != count {
			t.Fatalf("drew %d", len(r.Cards))
		}
		seen := map[string]bool{}
		for _, c := range r.Cards {
			if seen[c.ID] {
				t.Fatalf("duplicate card %s", c.ID)
			}
			seen[c.ID] = true
			if c.Meaning == "" || c.Reflection == "" {
				t.Fatalf("empty card %+v", c)
			}
		}
	}
	if _, err := DrawTarot(2, false); err == nil {
		t.Fatal("accepted unsupported spread")
	}
}
