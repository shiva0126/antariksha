package corpus

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Acquire downloads a public-domain source into rawDir/<id>/<file> and refuses
// to keep it unless its sha256 matches the pinned manifest value. Sources
// without an eligible rights decision are never downloaded.
func Acquire(ctx context.Context, s Source, rawDir string, hc *http.Client) (string, error) {
	if s.Rights != RightsPublicDomain {
		return "", fmt.Errorf("source %s: rights %s; only public_domain sources are acquired", s.ID, s.Rights)
	}
	if s.URL == "" {
		return "", fmt.Errorf("source %s: no url", s.ID)
	}
	path := RawPath(rawDir, s)
	if err := VerifyRaw(path, s); err == nil {
		return path, nil
	}
	if hc == nil {
		hc = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("source %s: HTTP %d", s.ID, resp.StatusCode)
	}
	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	tmp := path + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(f, resp.Body)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp)
		return "", err
	}
	if err = VerifyRaw(tmp, s); err != nil {
		os.Remove(tmp)
		return "", err
	}
	return path, os.Rename(tmp, path)
}

func RawPath(rawDir string, s Source) string { return filepath.Join(rawDir, s.ID, s.File) }

func VerifyRaw(path string, s Source) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != s.SHA256 {
		return fmt.Errorf("source %s: sha256 %s does not match manifest %s", s.ID, got, s.SHA256)
	}
	return nil
}

// Segment is one natural unit of a classical text: a numbered stanza together
// with the translator's notes on it.
type Segment struct {
	Chapter int    `json:"chapter"`
	Verse   int    `json:"verse"`
	Ref     string `json:"ref"`
	Text    string `json:"text"`
}

var (
	// Running page headers: "182  BBIHAT  JATAKA.  [OH.    20." and variants
	// the OCR produced ("ImtHAT jAtAKA", "UTAKir", "JATAW").
	headerRe  = regexp.MustCompile(`(?i)(j\s?a\s?t\s?a\s?[kw]|utak|jatak|jauk)`)
	chapterRe = regexp.MustCompile(`^.{0,3}CHA\S{0,5}\s+([IVXLZHilr!1]+)\W*`)
	verseRe   = regexp.MustCompile(`^[\s.(]*([0-9]{1,2})\s*[.*,]\)?\s+([A-Za-z(].*)$`)
	junkRe    = regexp.MustCompile(`[\^\\|~_{}<>»«•€]`)
	spaceRe   = regexp.MustCompile(`\s+`)
)

// CleanLines removes page furniture (running headers, folio numbers, margin
// noise), joins words hyphenated across line breaks and strips OCR junk. Line
// structure is preserved so segmentation can find stanza starts.
func CleanLines(raw string) []string {
	in := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(in))
	for _, l := range in {
		l = junkRe.ReplaceAllString(l, "")
		l = strings.TrimRight(l, " \t")
		t := strings.TrimSpace(l)
		letters := 0
		for _, r := range t {
			if unicode.IsLetter(r) {
				letters++
			}
		}
		if letters < 3 {
			continue // folio numbers, signatures, stray marks
		}
		if len(t) < 70 && headerRe.MatchString(t) && strings.ContainsAny(t, "0123456789[]") {
			continue
		}
		out = append(out, l)
	}
	// Join "occu-" + "py" across line breaks.
	for i := 0; i+1 < len(out); i++ {
		cur := strings.TrimRight(out[i], " ")
		if strings.HasSuffix(cur, "-") && len(cur) > 1 && unicode.IsLetter(rune(cur[len(cur)-2])) {
			next := strings.TrimLeft(out[i+1], " ")
			if next != "" && unicode.IsLower(rune(next[0])) {
				word, rest, _ := strings.Cut(next, " ")
				out[i] = strings.TrimSuffix(cur, "-") + word
				out[i+1] = rest
			}
		}
	}
	return out
}

// romanOCR maps characters the OCR substitutes inside chapter numerals.
var romanOCR = strings.NewReplacer("i", "I", "l", "I", "L", "I", "r", "I", "!", "I", "1", "I", "Z", "X", "H", "II")

func parseRoman(s string) int {
	s = romanOCR.Replace(s)
	vals := map[byte]int{'I': 1, 'V': 5, 'X': 10}
	total := 0
	for i := 0; i < len(s); i++ {
		v, ok := vals[s[i]]
		if !ok {
			return 0
		}
		if i+1 < len(s) && vals[s[i+1]] > v {
			total -= v
		} else {
			total += v
		}
	}
	return total
}

// Digits the OCR commonly confuses in stanza numbers ("8." for 3, "0." for 9).
var digitConfusion = map[[2]byte]bool{{'3', '8'}: true, {'8', '3'}: true, {'9', '0'}: true, {'0', '9'}: true, {'5', '6'}: true, {'6', '5'}: true, {'1', '7'}: true, {'7', '1'}: true}

func numberMatches(got string, want int) bool {
	w := strconv.Itoa(want)
	if got == w {
		return true
	}
	if len(got) != len(w) {
		return false
	}
	for i := range got {
		if got[i] != w[i] && !digitConfusion[[2]byte{got[i], w[i]}] {
			return false
		}
	}
	return true
}

// Segment splits cleaned lines into chapter.verse units. Chapter headings must
// increase; stanza numbers must run consecutively within a chapter (allowing
// known OCR digit confusions), so page numbers and note markers are not
// mistaken for stanzas.
func SegmentLines(lines []string) []Segment {
	var out []Segment
	chapter, verse := 0, 0
	var cur *Segment
	var buf []string
	flush := func() {
		if cur != nil {
			cur.Text = strings.TrimSpace(spaceRe.ReplaceAllString(strings.Join(buf, " "), " "))
			out = append(out, *cur)
		}
		cur, buf = nil, nil
	}
	for _, l := range lines {
		if m := chapterRe.FindStringSubmatch(l); m != nil {
			if n := parseRoman(m[1]); n > chapter && (chapter == 0 || n <= chapter+3) {
				flush()
				chapter, verse = n, 0
				continue
			}
		}
		if chapter > 0 {
			// A lowercase stanza start ("3. person born…", the OCR dropped "A")
			// is accepted only on an exact number, never a confusable one.
			if m := verseRe.FindStringSubmatch(l); m != nil && numberMatches(m[1], verse+1) && (m[1] == strconv.Itoa(verse+1) || !unicode.IsLower(rune(m[2][0]))) {
				flush()
				verse++
				cur = &Segment{Chapter: chapter, Verse: verse, Ref: fmt.Sprintf("%d.%d", chapter, verse)}
				buf = []string{m[2]}
				continue
			}
		}
		if cur != nil {
			buf = append(buf, l)
		}
	}
	flush()
	return out
}

func normTokens(s string) []string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return strings.Fields(b.String())
}

// near reports whether an OCR token plausibly spells the corrected token:
// equal, or within one edit for words of four or more letters (two for eight+).
func near(a, b string) bool {
	if a == b {
		return true
	}
	limit := 0
	if len(b) >= 8 {
		limit = 2
	} else if len(b) >= 4 {
		limit = 1
	}
	if limit == 0 || abs(len(a)-len(b)) > limit {
		return false
	}
	return levenshtein(a, b) <= limit
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func levenshtein(a, b string) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

// MatchScore measures how much of a hand-corrected excerpt is attested, in
// order, in the OCR text of the cited segment: the best fraction of excerpt
// tokens greedily aligned within a window of the segment. A corrected excerpt
// of a real passage scores ~0.9+; invented or copied-from-elsewhere text fails.
func MatchScore(excerpt, segment string) float64 {
	ex, sg := normTokens(excerpt), normTokens(segment)
	if len(ex) == 0 || len(sg) == 0 {
		return 0
	}
	best := 0
	for start := range sg {
		matched, j := 0, start
		for i := 0; i < len(ex); i++ {
			// Allow the OCR a little slack: look ahead up to three tokens. A
			// compound split in the excerpt ("one quarter") may be a single OCR
			// token after line-break hyphen joining ("onequarter").
			for k := j; k < len(sg) && k < j+4; k++ {
				if near(sg[k], ex[i]) {
					matched++
					j = k + 1
					break
				}
				if i+1 < len(ex) && near(sg[k], ex[i]+ex[i+1]) {
					matched += 2
					i++
					j = k + 1
					break
				}
			}
		}
		if matched > best {
			best = matched
		}
	}
	return float64(best) / float64(len(ex))
}
