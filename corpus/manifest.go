// Package corpus builds the classical-text knowledge base behind readings:
// rights-cleared sources → cleaned, verse-segmented text → entries keyed to the
// engine's detection vocabulary → PostgreSQL + pgvector.
//
// Nothing enters the corpus without passing the rights gate: every entry names
// a manifest source whose rights decision is public_domain, self_authored or
// licensed, and every public-domain excerpt must match the segmented OCR of the
// pinned (sha256) source file it cites.
package corpus

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed sources.json authored/*.json maps/*.json
var files embed.FS

const (
	RightsPublicDomain = "public_domain"
	RightsSelfAuthored = "self_authored"
	RightsLicensed     = "licensed"
	RightsBlocked      = "copyrighted_blocked"
	RightsNotAcquired  = "not_acquired"
)

// Source is one row of the auditable sources manifest.
type Source struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Edition     string `json:"edition,omitempty"`
	Translator  string `json:"translator,omitempty"`
	Year        int    `json:"year,omitempty"`
	URL         string `json:"url,omitempty"`
	File        string `json:"file,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
	Rights      string `json:"rights"`
	RightsBasis string `json:"rights_basis"`
	System      string `json:"system"`
	Notes       string `json:"notes,omitempty"`
}

type Manifest struct {
	Sources []Source `json:"sources"`
}

func LoadManifest() (Manifest, error) {
	var m Manifest
	b, err := files.ReadFile("sources.json")
	if err != nil {
		return m, err
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return m, fmt.Errorf("sources.json: %w", err)
	}
	seen := map[string]bool{}
	for _, s := range m.Sources {
		if seen[s.ID] {
			return m, fmt.Errorf("sources.json: duplicate source id %q", s.ID)
		}
		seen[s.ID] = true
		if err := CheckDecision(s); err != nil {
			return m, err
		}
	}
	return m, nil
}

func (m Manifest) Get(id string) (Source, bool) {
	for _, s := range m.Sources {
		if s.ID == id {
			return s, true
		}
	}
	return Source{}, false
}

// Modern translations known to be in copyright. A manifest row naming one of
// these may only carry the copyrighted_blocked decision.
var blockedTranslators = []string{"santhanam", "girish chand sharma", "g. c. sharma", "usha & sashi", "usha and sashi"}

// Eligible reports whether content under this rights decision may be ingested.
func Eligible(rights string) bool {
	return rights == RightsPublicDomain || rights == RightsSelfAuthored || rights == RightsLicensed
}

// CheckDecision validates that a manifest row records a coherent rights
// decision. It does not decide eligibility; see Eligible.
func CheckDecision(s Source) error {
	if s.ID == "" || s.Title == "" {
		return fmt.Errorf("source %q: id and title are required", s.ID)
	}
	switch s.Rights {
	case RightsPublicDomain, RightsSelfAuthored, RightsLicensed, RightsBlocked, RightsNotAcquired:
	default:
		return fmt.Errorf("source %s: unknown rights decision %q", s.ID, s.Rights)
	}
	if strings.TrimSpace(s.RightsBasis) == "" {
		return fmt.Errorf("source %s: rights_basis is required for an auditable decision", s.ID)
	}
	if s.System != "parashari" && s.System != "jaimini" {
		return fmt.Errorf("source %s: system must be parashari or jaimini", s.ID)
	}
	who := strings.ToLower(s.Translator + " " + s.Edition)
	for _, b := range blockedTranslators {
		if strings.Contains(who, b) && s.Rights != RightsBlocked {
			return fmt.Errorf("source %s: %q is an in-copyright modern translation and must be copyrighted_blocked", s.ID, s.Translator)
		}
	}
	if s.Rights == RightsPublicDomain {
		// Age heuristic from the ingestion codex: pre-1929 publication, or an
		// explicit author-life+60 basis recorded by a human.
		if (s.Year == 0 || s.Year > 1929) && !strings.Contains(strings.ToLower(s.RightsBasis), "life+60") {
			return fmt.Errorf("source %s: public_domain needs year <= 1929 or a recorded life+60 basis", s.ID)
		}
		if s.URL != "" && (s.SHA256 == "" || s.File == "") {
			return fmt.Errorf("source %s: downloadable public-domain source must pin file and sha256", s.ID)
		}
	}
	return nil
}
