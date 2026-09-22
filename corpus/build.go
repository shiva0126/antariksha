package corpus

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/example/panchang/engine"
)

// Entry is one astro_corpus row: one concept, keyed to an engine token, with
// the provenance of the passage that grounds it.
type Entry struct {
	System   string `json:"system"`
	DocType  string `json:"doc_type"`
	Key      string `json:"key"`
	Language string `json:"language"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	SourceID string `json:"source_id"`
	Ref      string `json:"ref"`
	Source   string `json:"source"`
	Rights   string `json:"rights"`
}

func (e Entry) ID() string {
	return strings.Join([]string{e.System, e.DocType, e.Key, e.Language, e.SourceID, e.Ref}, "|")
}

func (e Entry) ContentHash() string {
	h := sha256.Sum256([]byte(e.Title + "\x00" + e.Body + "\x00" + e.Source))
	return hex.EncodeToString(h[:])
}

// EmbeddingText is what gets embedded: the title carries the concept, the
// body the interpretation.
func (e Entry) EmbeddingText() string { return e.Title + "\n" + e.Body }

type authoredFile struct {
	SourceID string `json:"source_id"`
	System   string `json:"system"`
	Language string `json:"language"`
	Entries  []struct {
		DocType string `json:"doc_type"`
		Key     string `json:"key"`
		Title   string `json:"title"`
		Body    string `json:"body"`
	} `json:"entries"`
}

// MapEntry maps a hand-corrected excerpt of a public-domain verse to the
// engine token it describes.
type MapEntry struct {
	Ref     string `json:"ref"`
	DocType string `json:"doc_type"`
	Key     string `json:"key"`
	// Text is a verbatim (OCR-corrected) excerpt; "…" marks an omission and
	// each part must match the cited stanza independently.
	Text string `json:"text"`
	// Note is an optional editor's gloss, stored visibly labelled.
	Note string `json:"note,omitempty"`
}

type sourceMap struct {
	SourceID  string     `json:"source_id"`
	Language  string     `json:"language"`
	Threshold float64    `json:"threshold"`
	Entries   []MapEntry `json:"entries"`
}

type MatchResult struct {
	SourceID string  `json:"source_id"`
	Ref      string  `json:"ref"`
	Key      string  `json:"key"`
	Score    float64 `json:"score"`
}

type BuildOptions struct {
	RawDir string
	// AllowMissingRaw builds the authored corpus alone when a public-domain
	// source has not been acquired yet (offline builds, tests).
	AllowMissingRaw bool
	// AuditDir, when set, receives cleaned text and segments per source.
	AuditDir string
}

type BuildReport struct {
	Authored       int           `json:"authored"`
	PublicDomain   int           `json:"public_domain"`
	Jaimini        int           `json:"jaimini"`
	SkippedSources []string      `json:"skipped_sources,omitempty"`
	Matches        []MatchResult `json:"matches"`
}

var errGate = errors.New("rights gate")

func vocabulary() map[engine.CorpusKey]engine.VocabEntry {
	out := map[engine.CorpusKey]engine.VocabEntry{}
	for _, v := range engine.CorpusVocabulary() {
		out[v.CorpusKey] = v
	}
	return out
}

func provenance(s Source, ref string) string {
	parts := []string{s.Title}
	if s.Translator != "" {
		parts = append(parts, "tr. "+s.Translator)
	}
	if s.Year > 0 {
		parts = append(parts, fmt.Sprint(s.Year))
	}
	if ref != "" {
		parts = append(parts, "ch."+strings.Replace(ref, ".", " v.", 1))
	}
	return strings.Join(parts, ", ") + " [" + s.Rights + "]"
}

// gate is the rights-clearance check every entry passes before it can exist.
func gate(m Manifest, vocab map[engine.CorpusKey]engine.VocabEntry, e Entry) error {
	s, ok := m.Get(e.SourceID)
	if !ok {
		return fmt.Errorf("%w: entry %s cites unknown source %q", errGate, e.Key, e.SourceID)
	}
	if !Eligible(s.Rights) {
		return fmt.Errorf("%w: source %s is %s and cannot be ingested", errGate, s.ID, s.Rights)
	}
	if s.System != e.System {
		return fmt.Errorf("%w: entry %s system %s differs from source %s system %s", errGate, e.Key, e.System, s.ID, s.System)
	}
	if strings.TrimSpace(e.Body) == "" || strings.TrimSpace(e.Source) == "" {
		return fmt.Errorf("entry %s:%s: body and source provenance are required", e.DocType, e.Key)
	}
	switch e.System {
	case engine.SystemParashari:
		if _, ok := vocab[engine.CorpusKey{DocType: e.DocType, Key: e.Key}]; !ok {
			return fmt.Errorf("entry %s:%s is not an engine detection token", e.DocType, e.Key)
		}
	case "jaimini":
		if e.DocType != "karaka" {
			return fmt.Errorf("jaimini entry %s must use the karaka namespace, not %s", e.Key, e.DocType)
		}
	default:
		return fmt.Errorf("entry %s: unknown system %q", e.Key, e.System)
	}
	return nil
}

// Authored returns the embedded self-authored entries. They carry no external
// dependency, so they also back the in-memory corpus when no database exists.
func Authored() ([]Entry, error) {
	m, err := LoadManifest()
	if err != nil {
		return nil, err
	}
	return authored(m)
}

func authored(m Manifest) ([]Entry, error) {
	names, err := fs.Glob(files, "authored/*.json")
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	var out []Entry
	for _, name := range names {
		b, err := files.ReadFile(name)
		if err != nil {
			return nil, err
		}
		var f authoredFile
		if err = json.Unmarshal(b, &f); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		s, ok := m.Get(f.SourceID)
		if !ok {
			return nil, fmt.Errorf("%s: unknown source %q", name, f.SourceID)
		}
		lang := f.Language
		if lang == "" {
			lang = "en"
		}
		for _, x := range f.Entries {
			out = append(out, Entry{System: f.System, DocType: x.DocType, Key: x.Key, Language: lang, Title: x.Title, Body: strings.TrimSpace(x.Body), SourceID: s.ID, Source: provenance(s, ""), Rights: s.Rights})
		}
	}
	return out, nil
}

// Build runs segment → restructure over every rights-cleared source and
// returns gate-checked entries. Any gate failure aborts the whole build.
func Build(opts BuildOptions) ([]Entry, BuildReport, error) {
	var rep BuildReport
	m, err := LoadManifest()
	if err != nil {
		return nil, rep, err
	}
	vocab := vocabulary()
	entries, err := authored(m)
	if err != nil {
		return nil, rep, err
	}
	maps, err := fs.Glob(files, "maps/*.json")
	if err != nil {
		return nil, rep, err
	}
	sort.Strings(maps)
	for _, name := range maps {
		b, err := files.ReadFile(name)
		if err != nil {
			return nil, rep, err
		}
		var sm sourceMap
		if err = json.Unmarshal(b, &sm); err != nil {
			return nil, rep, fmt.Errorf("%s: %w", name, err)
		}
		s, ok := m.Get(sm.SourceID)
		if !ok {
			return nil, rep, fmt.Errorf("%s: unknown source %q", name, sm.SourceID)
		}
		if !Eligible(s.Rights) {
			return nil, rep, fmt.Errorf("%w: map %s targets %s source %s", errGate, name, s.Rights, s.ID)
		}
		raw, err := os.ReadFile(RawPath(opts.RawDir, s))
		if err != nil {
			if opts.AllowMissingRaw && errors.Is(err, fs.ErrNotExist) {
				rep.SkippedSources = append(rep.SkippedSources, s.ID)
				continue
			}
			return nil, rep, fmt.Errorf("source %s not acquired (run: corpus acquire): %w", s.ID, err)
		}
		if err = VerifyRaw(RawPath(opts.RawDir, s), s); err != nil {
			return nil, rep, err
		}
		lines := CleanLines(string(raw))
		segs := SegmentLines(lines)
		if opts.AuditDir != "" {
			if err = writeAudit(opts.AuditDir, s.ID, lines, segs); err != nil {
				return nil, rep, err
			}
		}
		byRef := map[string]Segment{}
		for _, sg := range segs {
			byRef[sg.Ref] = sg
		}
		threshold := sm.Threshold
		if threshold == 0 {
			threshold = 0.85
		}
		lang := sm.Language
		if lang == "" {
			lang = "en"
		}
		for _, me := range sm.Entries {
			sg, ok := byRef[me.Ref]
			if !ok {
				return nil, rep, fmt.Errorf("%s: %s cites %s but segmentation found no such stanza", name, me.Key, me.Ref)
			}
			score := 1.0
			for _, part := range strings.Split(me.Text, "…") {
				if strings.TrimSpace(part) != "" {
					score = min(score, MatchScore(part, sg.Text))
				}
			}
			rep.Matches = append(rep.Matches, MatchResult{s.ID, me.Ref, me.Key, score})
			if score < threshold {
				return nil, rep, fmt.Errorf("%s: excerpt for %s does not match OCR of %s (score %.2f < %.2f)", name, me.Key, me.Ref, score, threshold)
			}
			title := s.Title + " " + me.Ref
			if v, ok := vocab[engine.CorpusKey{DocType: me.DocType, Key: me.Key}]; ok {
				title = v.Label + " — " + s.Title + " " + me.Ref
			}
			body := strings.TrimSpace(me.Text)
			src := provenance(s, me.Ref)
			if me.Note != "" {
				body += "\n[Editor's note: " + strings.TrimSpace(me.Note) + "]"
				src += "; editor's note self-authored"
			}
			entries = append(entries, Entry{System: s.System, DocType: me.DocType, Key: me.Key, Language: lang, Title: title, Body: body, SourceID: s.ID, Ref: me.Ref, Source: src, Rights: s.Rights})
		}
	}
	seen := map[string]bool{}
	for _, e := range entries {
		if err := gate(m, vocab, e); err != nil {
			return nil, rep, err
		}
		if seen[e.ID()] {
			return nil, rep, fmt.Errorf("duplicate entry %s", e.ID())
		}
		seen[e.ID()] = true
		switch {
		case e.System == "jaimini":
			rep.Jaimini++
		case e.Rights == RightsPublicDomain:
			rep.PublicDomain++
		default:
			rep.Authored++
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID() < entries[j].ID() })
	return entries, rep, nil
}

func writeAudit(dir, id string, lines []string, segs []Segment) error {
	d := path.Join(dir, id)
	if err := os.MkdirAll(d, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path.Join(d, "clean.txt"), []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return err
	}
	f, err := os.Create(path.Join(d, "segments.jsonl"))
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	for _, s := range segs {
		if err = enc.Encode(s); err != nil {
			f.Close()
			return err
		}
	}
	return f.Close()
}

// CoverageReport lists engine tokens with no Parashari corpus entry.
type CoverageReport struct {
	Required, RequiredCovered int
	Optional, OptionalCovered int
	Missing                   []engine.VocabEntry
	ByDocType                 map[string][2]int // covered, required
}

func (c CoverageReport) Green() bool { return c.RequiredCovered == c.Required }

// Coverage answers "for each engine token, does a corpus entry exist?".
func Coverage(covered map[engine.CorpusKey]bool) CoverageReport {
	rep := CoverageReport{ByDocType: map[string][2]int{}}
	for _, v := range engine.CorpusVocabulary() {
		ok := covered[v.CorpusKey]
		if !v.Required {
			rep.Optional++
			if ok {
				rep.OptionalCovered++
			}
			continue
		}
		rep.Required++
		c := rep.ByDocType[v.DocType]
		c[1]++
		if ok {
			rep.RequiredCovered++
			c[0]++
		} else {
			rep.Missing = append(rep.Missing, v)
		}
		rep.ByDocType[v.DocType] = c
	}
	return rep
}

func CoveredKeys(entries []Entry) map[engine.CorpusKey]bool {
	out := map[engine.CorpusKey]bool{}
	for _, e := range entries {
		if e.System == engine.SystemParashari {
			out[engine.CorpusKey{DocType: e.DocType, Key: e.Key}] = true
		}
	}
	return out
}
