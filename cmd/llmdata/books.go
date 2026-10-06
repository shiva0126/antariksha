package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/panchang/corpus"
	"github.com/example/panchang/engine"
)

// bookPassage is one passage of a public-domain book, for continued
// pretraining: the model reads the classics themselves.
type bookPassage struct {
	Source     string `json:"source"`
	Title      string `json:"title"`
	Translator string `json:"translator,omitempty"`
	Year       int    `json:"year,omitempty"`
	Ref        string `json:"ref"`
	Text       string `json:"text"`
}

// chunks splits text into pieces of about n words at sentence ends.
func chunks(text string, n int) []string {
	words := strings.Fields(text)
	var out []string
	for len(words) > 0 {
		end := min(len(words), n)
		// Prefer to stop at a sentence end within the last third.
		for i := end; i > end*2/3 && end < len(words); i-- {
			if w := words[i-1]; strings.HasSuffix(w, ".") || strings.HasSuffix(w, ";") {
				end = i
				break
			}
		}
		out = append(out, strings.Join(words[:end], " "))
		words = words[end:]
	}
	return out
}

// books writes every public-domain source that is present on disk as
// passages: by chapter and verse where the OCR segments well, otherwise in
// numbered parts.
func books(rawDir, out string) error {
	m, err := corpus.LoadManifest()
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	fh, err := os.Create(out)
	if err != nil {
		return err
	}
	defer fh.Close()
	enc := json.NewEncoder(fh)
	for _, s := range m.Sources {
		if s.Rights != corpus.RightsPublicDomain || s.File == "" {
			continue
		}
		path := corpus.RawPath(rawDir, s)
		if err := corpus.VerifyRaw(path, s); err != nil {
			log.Printf("skip %s: %v", s.ID, err)
			continue
		}
		raw, _ := os.ReadFile(path)
		lines := corpus.CleanLines(string(raw))
		segs := corpus.SegmentFor(s, lines)
		total, inSegs := len(strings.Join(lines, " ")), 0
		for _, sg := range segs {
			inSegs += len(sg.Text)
		}
		count, words := 0, 0
		emit := func(ref, text string) error {
			for i, c := range chunks(text, 260) {
				r := ref
				if i > 0 {
					r = fmt.Sprintf("%s (cont. %d)", ref, i)
				}
				words += len(strings.Fields(c))
				count++
				if err := enc.Encode(bookPassage{s.ID, s.Title, s.Translator, s.Year, r, c}); err != nil {
					return err
				}
			}
			return nil
		}
		if inSegs*10 >= total*4 { // the verse split covers at least 40% of the book
			for _, sg := range segs {
				if err := emit(sg.Ref, sg.Text); err != nil {
					return err
				}
			}
		} else {
			if err := emit("part", strings.Join(lines, " ")); err != nil {
				return err
			}
		}
		log.Printf("%s: %d passages, %d words", s.ID, count, words)
	}
	return nil
}

// bookQA writes question-and-answer pairs for every checked book rule (the
// map entries with an exact clause and a plain summary), several phrasings
// each, answered in the app's plain format with the citation.
func bookQA(entries, out string) error {
	f, err := os.Open(entries)
	if err != nil {
		return fmt.Errorf("%v (run: go run ./cmd/corpus build)", err)
	}
	defer f.Close()
	if err = os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	fh, err := os.Create(out)
	if err != nil {
		return err
	}
	defer fh.Close()
	enc := json.NewEncoder(fh)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	count := 0
	for sc.Scan() {
		var e corpus.Entry
		if json.Unmarshal(sc.Bytes(), &e) != nil || e.Rights != corpus.RightsPublicDomain {
			continue
		}
		body, plain, ok := strings.Cut(e.Body, "\n"+corpus.PlainLabel)
		if !ok {
			continue
		}
		plain = strings.TrimSuffix(strings.TrimSpace(plain), "]")
		book := "Brihat Jataka"
		var qs []string
		switch e.DocType {
		case engine.DocGrahaInHouse:
			parts := strings.SplitN(e.Key, "_in_", 2)
			var h int
			fmt.Sscan(parts[1], &h)
			g, ord := engine.GrahaEnglish(parts[0]), ordinalWord(h)
			qs = []string{
				fmt.Sprintf("What does %s say about %s in the %s house?", book, g, ord),
				fmt.Sprintf("I have %s in my %s house. What do the classical texts say?", g, ord),
				fmt.Sprintf("According to Varahamihira, what happens when %s is in the %s house from the ascendant?", g, ord),
			}
		case engine.DocNakshatra:
			n := strings.ReplaceAll(e.Key, "_", " ")
			qs = []string{fmt.Sprintf("What does %s say about people born in %s nakshatra?", book, n), fmt.Sprintf("My birth star is %s. What do the classics say?", n)}
		case engine.DocGrahaInSign:
			sign := strings.TrimPrefix(e.Key, "moon_in_")
			qs = []string{fmt.Sprintf("What does %s say about the Moon in %s?", book, sign), fmt.Sprintf("My Moon sign is %s. What do the classical texts say?", sign)}
		case engine.DocYoga:
			y := strings.ReplaceAll(e.Key, "_", " ")
			qs = []string{fmt.Sprintf("What does %s say about %s yoga?", book, y), fmt.Sprintf("Explain %s yoga according to the classics.", y)}
		case engine.DocDignity:
			st := strings.TrimPrefix(e.Key, "dignity_")
			qs = []string{fmt.Sprintf("What does %s say about a planet in its %s sign?", book, st)}
		default:
			continue
		}
		quote := firstWords(body, 70)
		answer := "In short\n" + plain + "\n\nWhat this means for you\n• " + book + " " + e.Ref + " (tr. Chidambaram Iyer, 1885) says: \"" + quote + "\"\n• Old texts state results as certain; read them as traditional tendencies, not facts about you.\n\nFor reflection, not certainty."
		a, _ := json.Marshal(map[string]string{"answer": answer})
		for _, q := range qs {
			if err := enc.Encode(example{[]message{{"system", "You are Astrisk. Answer in plain language and cite the classical source. Return JSON only: {\"answer\": string}."}, {"user", q}, {"assistant", string(a)}}, map[string]any{"task": "book", "token": e.DocType + ":" + e.Key, "ref": e.Ref}}); err != nil {
				return err
			}
			count++
		}
	}
	log.Printf("wrote %d book question-answer pairs to %s", count, out)
	return nil
}

func ordinalWord(h int) string {
	return []string{"", "1st", "2nd", "3rd", "4th", "5th", "6th", "7th", "8th", "9th", "10th", "11th", "12th"}[h]
}

func firstWords(s string, n int) string {
	w := strings.Fields(s)
	if len(w) <= n {
		return strings.Join(w, " ")
	}
	return strings.Join(w[:n], " ") + " …"
}
