package corpus

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEditorialReferencesMustMatchClearedConcept(t *testing.T) {
	m, err := LoadManifest()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, system, doc, key, source, ref string
		valid                               bool
	}{
		{"matching passage", "parashari", "yoga", "sunapha", "brihat-jataka-iyer-1885", "13.5", true},
		{"wrong concept", "parashari", "yoga", "gajakesari", "brihat-jataka-iyer-1885", "13.5", false},
		{"made up verse", "parashari", "yoga", "sunapha", "brihat-jataka-iyer-1885", "99.99", false},
		{"pending book", "parashari", "yoga", "sunapha", "bhrigu-sutras", "13.5", false},
		{"blocked book", "parashari", "yoga", "sunapha", "bphs-santhanam", "13.5", false},
		{"different school", "jaimini", "yoga", "sunapha", "brihat-jataka-iyer-1885", "13.5", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := editorialProvenance(m, tc.system, tc.doc, tc.key, []EditorialReference{{tc.source, tc.ref}})
			if (err == nil) != tc.valid {
				t.Fatalf("%s %v", out, err)
			}
			if tc.valid && (!strings.Contains(out, "not a translation") || strings.Contains(out, "[public_domain]")) {
				t.Fatal("editorial was misrepresented as classical")
			}
		})
	}
	if _, err = editorialProvenance(m, "parashari", "yoga", "sunapha", []EditorialReference{{"brihat-jataka-iyer-1885", "13.5"}, {"brihat-jataka-iyer-1885", "13.5"}}); err == nil {
		t.Fatal("duplicate accepted")
	}
}

func TestPlainLanguageCorpusAndSourceContext(t *testing.T) {
	entries, err := Authored()
	if err != nil {
		t.Fatal(err)
	}
	linked, dignities, nakshatras := 0, 0, 0
	for _, e := range entries {
		if strings.Contains(e.Source, "passage context:") {
			linked++
		}
		if e.DocType == "dignity" || e.DocType == "nakshatra" {
			if e.DocType == "dignity" {
				dignities++
			} else {
				nakshatras++
			}
			if !strings.Contains(e.Body, "Reflection:") || len(e.Body) < 180 {
				t.Fatalf("thin or unhelpful entry %s", e.Key)
			}
			for _, bad := range []string{"It gives ", "will be ", "guarantees ", "its success is slow but permanent"} {
				if strings.Contains(e.Body, bad) {
					t.Fatalf("deterministic claim in %s: %s", e.Key, bad)
				}
			}
		}
	}
	if linked != 34 || dignities != 38 || nakshatras != 27 {
		t.Fatalf("linked=%d dignity=%d nakshatra=%d", linked, dignities, nakshatras)
	}
	// Entries may gain content without pretending additional books were loaded.
	b, _ := files.ReadFile("authored/yogas.json")
	var f authoredFile
	if json.Unmarshal(b, &f) != nil {
		t.Fatal("invalid yogas")
	}
	for _, e := range f.Entries {
		if len(e.References) > 0 && len(strings.Fields(e.Body)) < 75 {
			t.Fatalf("thin enriched yoga %s", e.Key)
		}
	}
}
