// Package divination contains explicitly documented symbolic reading methods.
// Its meanings are original reflective writing, separate from astronomical facts.
package divination

import (
	"fmt"
	"golang.org/x/text/unicode/norm"
	"strings"
	"time"
	"unicode"
)

type NumberReading struct {
	Number     int    `json:"number"`
	Steps      []int  `json:"steps"`
	Meaning    string `json:"meaning"`
	Reflection string `json:"reflection"`
}
type Numerology struct {
	Method       string         `json:"method"`
	NameUsed     string         `json:"name_used"`
	Year         int            `json:"year"`
	LifePath     NumberReading  `json:"life_path"`
	Birthday     NumberReading  `json:"birthday"`
	PersonalYear NumberReading  `json:"personal_year"`
	Expression   *NumberReading `json:"expression,omitempty"`
	SoulUrge     *NumberReading `json:"soul_urge,omitempty"`
	Personality  *NumberReading `json:"personality,omitempty"`
	// Indian numerology (ank jyotish) reduces the birth day and the whole date
	// to 1–9 and gives each root number a ruling graha.
	Mulank        int    `json:"mulank"`
	MulankGraha   string `json:"mulank_graha"`
	Bhagyank      int    `json:"bhagyank"`
	BhagyankGraha string `json:"bhagyank_graha"`
	Note          string `json:"note"`
}

// numberGrahas is the usual Indian numerology (ank jyotish) assignment of the
// root numbers 1–9 to the nine grahas.
var numberGrahas = [10]string{"", "sun", "moon", "jupiter", "rahu", "mercury", "venus", "ketu", "saturn", "mars"}

// RootMeaning returns the theme and reflection prompt for a root number 1–9.
func RootMeaning(n int) (string, string) {
	m := numberMeanings[Root(n)]
	return m[0], m[1]
}

// Root reduces a positive number to a single digit 1–9, master numbers included.
func Root(n int) int {
	for n > 9 {
		n = digitSum(n)
	}
	return n
}

// RulingGraha returns the graha id that Indian numerology assigns to n's root.
func RulingGraha(n int) string {
	if n <= 0 {
		return ""
	}
	return numberGrahas[Root(n)]
}

var numberMeanings = map[int][2]string{
	1:  {"Initiative and independence are the themes associated with this number.", "What small project would you like to take the lead on?"},
	2:  {"Cooperation and listening are the themes associated with this number.", "Where could a clear conversation help you and someone else work together?"},
	3:  {"Expression, play and creativity are the themes associated with this number.", "What would you enjoy making or sharing without needing it to be perfect?"},
	4:  {"Structure, patience and steady work are the themes associated with this number.", "Which simple routine would make your week easier?"},
	5:  {"Curiosity, variety and freedom are the themes associated with this number.", "What new experience fits your responsibilities and budget?"},
	6:  {"Care, home and responsibility are the themes associated with this number.", "How can you support others while protecting time for yourself?"},
	7:  {"Study, reflection and independent thinking are the themes associated with this number.", "What question would you like to explore more deeply?"},
	8:  {"Organisation, ambition and practical responsibility are the themes associated with this number.", "What goal needs a clearer plan and a realistic measure of progress?"},
	9:  {"Compassion, perspective and completion are the themes associated with this number.", "What would you like to finish or contribute to a cause you care about?"},
	11: {"Some numerology traditions keep 11 as a master number and associate it with inspiration and sensitivity.", "How could you turn an idea into one grounded, manageable action?"},
	22: {"Some numerology traditions keep 22 as a master number and associate it with building something useful over time.", "What long-term goal can you break into a small first step?"},
	33: {"Some numerology traditions keep 33 as a master number and associate it with teaching and compassionate service.", "Where can you help within your skills and personal limits?"},
}

func digitSum(n int) int {
	s := 0
	for n > 0 {
		s += n % 10
		n /= 10
	}
	return s
}
func numberReading(n int, master bool) NumberReading {
	steps := []int{n}
	for n > 9 && !(master && (n == 11 || n == 22 || n == 33)) {
		n = digitSum(n)
		steps = append(steps, n)
	}
	m := numberMeanings[n]
	return NumberReading{n, steps, m[0], m[1]}
}
func ReadNumerology(date, name string, year int) (Numerology, error) {
	d, err := time.Parse("2006-01-02", date)
	if err != nil || d.Format("2006-01-02") != date || d.Year() < 1800 || d.Year() > 2399 || year < 1800 || year > 2399 {
		return Numerology{}, fmt.Errorf("enter a valid date and reading year (1800–2399)")
	}
	if len(name) > 200 {
		return Numerology{}, fmt.Errorf("name must be under 200 bytes")
	}
	out := Numerology{Method: "Pythagorean letters A–Z = 1–9 repeatedly. Life path uses the sum of all date digits. Reductions keep 11, 22 and 33, except personal year (1–9). A, E, I, O, U are vowels; Y is counted as a consonant.", Year: year, LifePath: numberReading(digitSum(d.Year())+digitSum(int(d.Month()))+digitSum(d.Day()), true), Birthday: numberReading(d.Day(), true), PersonalYear: numberReading(digitSum(int(d.Month()))+digitSum(d.Day())+digitSum(year), false), Note: "Symbolic prompts for reflection, not a test of personality or a prediction. Other numerology conventions may give different results."}
	out.Mulank, out.Bhagyank = Root(d.Day()), Root(out.LifePath.Number)
	out.MulankGraha, out.BhagyankGraha = RulingGraha(out.Mulank), RulingGraha(out.Bhagyank)
	total, vowels, consonants := 0, 0, 0
	var used strings.Builder
	for _, r := range norm.NFD.String(strings.ToUpper(strings.TrimSpace(name))) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			v := int(r-'A')%9 + 1
			total += v
			used.WriteRune(r)
			if strings.ContainsRune("AEIOU", r) {
				vowels += v
			} else {
				consonants += v
			}
			continue
		}
		if unicode.IsSpace(r) || r == '-' || r == '\'' || r == '’' || r == '.' {
			continue
		}
		return Numerology{}, fmt.Errorf("use the name's Latin-letter spelling; other scripts and digits are not silently converted")
	}
	if strings.TrimSpace(name) != "" && total == 0 {
		return Numerology{}, fmt.Errorf("enter a name containing letters")
	}
	out.NameUsed = used.String()
	if total > 0 {
		x := numberReading(total, true)
		out.Expression = &x
	}
	if vowels > 0 {
		x := numberReading(vowels, true)
		out.SoulUrge = &x
	}
	if consonants > 0 {
		x := numberReading(consonants, true)
		out.Personality = &x
	}
	return out, nil
}
