package corpus

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

// EditorialReference connects original modern commentary to a cleared passage
// for the SAME detected concept. It is not a claim of literal translation or
// independent human/scholarly approval.
type EditorialReference struct {
	SourceID string `json:"source_id"`
	Ref      string `json:"ref"`
}

func editorialProvenance(m Manifest, system, doc, key string, refs []EditorialReference) (string, error) {
	if len(refs) == 0 {
		return "", nil
	}
	if len(refs) > 8 {
		return "", fmt.Errorf("too many editorial references for %s", key)
	}
	maps, err := fs.Glob(files, "maps/*.json")
	if err != nil {
		return "", err
	}
	available := map[string]bool{}
	for _, name := range maps {
		b, err := files.ReadFile(name)
		if err != nil {
			return "", err
		}
		var sm sourceMap
		if err = json.Unmarshal(b, &sm); err != nil {
			return "", err
		}
		for _, e := range sm.Entries {
			if e.DocType == doc && e.Key == key {
				available[sm.SourceID+"|"+e.Ref] = true
			}
		}
	}
	seen := map[string]bool{}
	labels := []string{}
	for _, ref := range refs {
		s, ok := m.Get(ref.SourceID)
		id := ref.SourceID + "|" + ref.Ref
		if !ok || !Eligible(s.Rights) || s.System != system || !available[id] || seen[id] {
			return "", fmt.Errorf("unverified, duplicate or mismatched editorial reference %s for %s:%s", id, doc, key)
		}
		seen[id] = true
		labels = append(labels, fmt.Sprintf("%s (%d), %s", s.Title, s.Year, ref.Ref))
	}
	return "; original editorial discussion, not a translation; passage context: " + strings.Join(labels, "; "), nil
}
