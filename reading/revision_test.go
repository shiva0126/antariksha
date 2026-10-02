package reading

import (
	"strings"
	"testing"
)

func TestGroundingRevisionTracksContentAndProvenance(t *testing.T) {
	a := Rule{DocType: "yoga", Key: "sunapha", Body: "original", Source: "source"}
	b := Rule{DocType: "yoga", Key: "anapha", Body: "other", Source: "source"}
	first := GroundingHash([]Rule{a, b})
	if first != GroundingHash([]Rule{b, a}) {
		t.Fatal("order changed revision")
	}
	for _, field := range []string{"body", "source", "ref", "title"} {
		c := a
		switch field {
		case "body":
			c.Body += " changed"
		case "source":
			c.Source += " changed"
		case "ref":
			c.Ref = "13.5"
		case "title":
			c.Title = "new"
		}
		if first == GroundingHash([]Rule{c, b}) {
			t.Fatalf("%s not covered", field)
		}
	}
	if NewService(nil, nil).CacheModel() != FallbackModel {
		t.Fatal("disabled model reused hosted cache")
	}
}

func TestUnicodeExcerptNeverUsesByteOffsetAsRuneIndex(t *testing.T) {
	s := strings.Repeat("आ", 45) + ". " + strings.Repeat("अ", 80)
	got := firstSentence(s, 60)
	if got != strings.Repeat("आ", 45)+"." {
		t.Fatalf("broken excerpt %q", got)
	}
}
