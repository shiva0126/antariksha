package reading

import (
	"context"
	"github.com/example/panchang/engine"
	"strings"
)

type MemoryCorpus struct{ Entries []Rule }

func (m MemoryCorpus) Rules(_ context.Context, f engine.ChartFacts) ([]Rule, error) {
	keys := map[string]bool{}
	for _, y := range f.Yogas {
		keys[strings.ToLower(strings.ReplaceAll(y.Name, " ", "-"))] = true
	}
	for id, d := range f.Dignities {
		keys[id+"_"+d.State] = true
	}
	out := []Rule{}
	for _, r := range m.Entries {
		if keys[r.Key] {
			out = append(out, r)
		}
	}
	return out, nil
}

var DefaultCorpus = MemoryCorpus{Entries: []Rule{{"gajakesari", "Gajakesari", "Jupiter in a kendra from the Moon is traditionally read as a supportive combination for judgment and learning; its result depends on dignity and affliction.", "BPHS-inspired rule"}, {"budha-aditya", "Budha-Aditya", "Sun and Mercury in one sign is traditionally associated with intellect and communication. Combustion changes emphasis but does not erase the geometric fact.", "BPHS-inspired rule"}, {"ruchaka", "Ruchaka", "Mars in own or exaltation sign in a kendra is a Mahapurusha combination associated with initiative and courage.", "BPHS-inspired rule"}}}
