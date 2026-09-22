package corpus

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/example/panchang/engine"
)

const rawDir = "raw"

func haveRaw(t *testing.T) {
	m, _ := LoadManifest()
	s, _ := m.Get("brihat-jataka-iyer-1885")
	if _, err := os.Stat(RawPath(rawDir, s)); err != nil {
		t.Skip("public-domain source not acquired; run: go run ./cmd/corpus acquire")
	}
}

func TestManifestDecisionsAreAuditable(t *testing.T) {
	m, err := LoadManifest()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Sources) < 10 {
		t.Fatalf("manifest should list the canonical texts, got %d", len(m.Sources))
	}
	s, ok := m.Get("bphs-santhanam")
	if !ok || s.Rights != RightsBlocked || Eligible(s.Rights) {
		t.Fatalf("Santhanam BPHS must be recorded as blocked: %+v", s)
	}
}

func TestRightsGateRejects(t *testing.T) {
	cases := map[string]Source{
		"modern translation marked PD": {ID: "x", Title: "BPHS", Translator: "R. Santhanam", Rights: RightsPublicDomain, RightsBasis: "old", System: "parashari", Year: 1900},
		"PD without age basis":         {ID: "x", Title: "T", Rights: RightsPublicDomain, RightsBasis: "trust me", System: "parashari", Year: 1984},
		"PD download without sha pin":  {ID: "x", Title: "T", Rights: RightsPublicDomain, RightsBasis: "1885", System: "parashari", Year: 1885, URL: "https://example.org/a.txt"},
		"missing basis":                {ID: "x", Title: "T", Rights: RightsSelfAuthored, System: "parashari"},
		"unknown system":               {ID: "x", Title: "T", Rights: RightsSelfAuthored, RightsBasis: "own", System: "kp"},
	}
	for name, s := range cases {
		if CheckDecision(s) == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if CheckDecision(Source{ID: "x", Title: "T", Rights: RightsPublicDomain, RightsBasis: "translator died 1930; life+60", System: "parashari", Year: 1950}) != nil {
		t.Error("explicit life+60 basis should be accepted")
	}
	m, _ := LoadManifest()
	err := gate(m, vocabulary(), Entry{System: "parashari", DocType: "yoga", Key: "gajakesari", SourceID: "bphs-santhanam", Body: "x", Source: "x"})
	if !errors.Is(err, errGate) {
		t.Fatalf("blocked source entry passed gate: %v", err)
	}
	if gate(m, vocabulary(), Entry{System: "parashari", DocType: "yoga", Key: "not_a_token", SourceID: "antariksha-authored-v1", Body: "x", Source: "x"}) == nil {
		t.Fatal("unknown engine token accepted")
	}
	if gate(m, vocabulary(), Entry{System: "jaimini", DocType: "yoga", Key: "gajakesari", SourceID: "antariksha-authored-jaimini-v1", Body: "x", Source: "x"}) == nil {
		t.Fatal("jaimini entry allowed into a parashari doc_type")
	}
}

func TestAuthoredCoverageIsGreen(t *testing.T) {
	entries, rep, err := Build(BuildOptions{RawDir: t.TempDir(), AllowMissingRaw: true})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Authored == 0 || rep.Jaimini == 0 {
		t.Fatalf("report %+v", rep)
	}
	cov := Coverage(CoveredKeys(entries))
	if !cov.Green() {
		var miss []string
		for _, v := range cov.Missing {
			miss = append(miss, v.String())
		}
		t.Fatalf("coverage %d/%d, missing %v", cov.RequiredCovered, cov.Required, miss)
	}
	for _, e := range entries {
		if e.Source == "" || !Eligible(e.Rights) {
			t.Fatalf("entry without provenance/eligible rights: %+v", e)
		}
		if strings.Contains(strings.ToLower(e.Body), "santhanam") {
			t.Fatalf("modern translation referenced in body: %s", e.ID())
		}
	}
}

func TestParashariKeysNeverResolveToJaimini(t *testing.T) {
	entries, _, err := Build(BuildOptions{RawDir: t.TempDir(), AllowMissingRaw: true})
	if err != nil {
		t.Fatal(err)
	}
	vocab := vocabulary()
	for _, e := range entries {
		_, inVocab := vocab[engine.CorpusKey{DocType: e.DocType, Key: e.Key}]
		if e.System == "jaimini" && inVocab {
			t.Fatalf("jaimini entry %s collides with an engine token", e.ID())
		}
	}
}

func TestSegmentation(t *testing.T) {
	// OCR-style input: garbled numeral heading, a running page header inside a
	// stanza, "2*" for stanza 2, "8." misread for 3, a hyphenated line break,
	// and a "5." that is out of sequence (a note number, not a stanza).
	raw := "CHAPTER  Xlir.\nOn Chandra Yogas.\n1.  If the Moon occupy\n182  BBIHAT  JATAKA.  [OH.    13.\nthe Kendra houses.\n2*  Excepting the Sun if\n8.  If the benefic occu-\npy the 6th.\n5.  Such as Kemadruma.\n"
	segs := SegmentLines(CleanLines(raw))
	if len(segs) != 3 {
		t.Fatalf("segments %+v", segs)
	}
	if segs[0].Ref != "13.1" || strings.Contains(segs[0].Text, "JATAKA") || !strings.Contains(segs[0].Text, "Kendra houses") {
		t.Fatalf("first %+v", segs[0])
	}
	if segs[1].Ref != "13.2" || segs[2].Ref != "13.3" {
		t.Fatalf("refs %+v", segs)
	}
	if !strings.Contains(segs[2].Text, "occupy the 6th") || !strings.Contains(segs[2].Text, "Such as Kemadruma") {
		t.Fatalf("third %+v", segs[2])
	}
}

func TestMatchScoreSeparatesExcerptsFromParaphrase(t *testing.T) {
	ocr := "A person born with tho Moon in sign Libra will reapocb t.ho Devas, Brahmins and holy men ; will bo latelligenfc will never covet the property of other men; will lead a religious life"
	if s := MatchScore("A person born with the Moon in sign Libra will respect the Devas, Brahmins, and holy men; will be intelligent, will never covet the property of other men", ocr); s < 0.8 {
		t.Fatalf("corrected excerpt scored %.2f", s)
	}
	if s := MatchScore("Those born with the Moon in Libra honour gods, teachers and saints and follow a spiritual path.", ocr); s > 0.4 {
		t.Fatalf("paraphrase scored %.2f", s)
	}
}

func TestPublicDomainBuild(t *testing.T) {
	haveRaw(t)
	entries, rep, err := Build(BuildOptions{RawDir: rawDir})
	if err != nil {
		t.Fatal(err)
	}
	if rep.PublicDomain < 90 {
		t.Fatalf("expected the mapped Brihat Jataka chapters, got %d", rep.PublicDomain)
	}
	multi := map[engine.CorpusKey]map[string]bool{}
	for _, e := range entries {
		k := engine.CorpusKey{DocType: e.DocType, Key: e.Key}
		if multi[k] == nil {
			multi[k] = map[string]bool{}
		}
		multi[k][e.SourceID] = true
	}
	if len(multi[engine.CorpusKey{DocType: "yoga", Key: "kemadruma"}]) != 2 {
		t.Fatal("kemadruma should merge authored and classical sources")
	}
}
