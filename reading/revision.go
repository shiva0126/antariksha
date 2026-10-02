package reading

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// GroundingHash invalidates cached prose when its actual content or provenance
// changes, rather than relying on somebody remembering to bump a version.
func GroundingHash(rules []Rule) string {
	rows := make([]string, 0, len(rules))
	for _, r := range rules {
		b, _ := json.Marshal([]string{r.DocType, r.Key, r.Title, r.Body, r.Source, r.Ref})
		rows = append(rows, string(b))
	}
	sort.Strings(rows)
	b, _ := json.Marshal(rows)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func (s *Service) CacheModel() string {
	if s.LLM == nil {
		return FallbackModel
	}
	return s.modelName()
}
