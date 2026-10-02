package reading

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/example/panchang/engine"
)

// MatchFactor contains server-owned evidence. Never accept these from a client
// or let generated prose replace the score, status or source.
type MatchFactor struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Evidence    string `json:"evidence"`
	Explanation string `json:"explanation"`
	Question    string `json:"question"`
	Source      string `json:"source"`
}
type MatchExplanation struct {
	Summary     string        `json:"summary"`
	Factors     []MatchFactor `json:"factors"`
	Limitations []string      `json:"limitations"`
	Disclaimer  string        `json:"disclaimer"`
	Model       string        `json:"model"`
	AIStatus    string        `json:"ai_status"`
}

var matchGuides = map[string][2]string{
	"Varna":        {"This is a historical Moon-sign classification. It does not identify anyone's caste, social standing, ability or worth. Its gender hierarchy must not be used to rank partners.", "How will you make decisions as equals and respect each other's beliefs?"},
	"Vashya":       {"This tradition groups Moon signs to discuss influence. It does not establish attraction or give either person authority over the other.", "How do you handle disagreement while preserving each other's independence?"},
	"Tara":         {"This compares the spacing of birth stars. A lower result is a traditional classification, not evidence of bad luck or a prediction of events.", "What kind of support helps each of you during stressful changes?"},
	"Yoni":         {"Birth stars are assigned symbolic animal groups. These are not observations of sexual preferences or proof of physical compatibility. Only voluntary conversations can establish comfort and boundaries.", "How would you talk about affection, boundaries and consent at a pace comfortable for both?"},
	"Graha Maitri": {"This compares the traditional relationships between the rulers of the Moon signs. It is a symbolic prompt about communication, not a personality assessment.", "When a conversation becomes difficult, do you prefer time to think or talking things through immediately?"},
	"Gana":         {"These birth-star groups are traditional symbols. Their names do not make anyone good, bad, gentle or aggressive. Actual temperament needs to be understood through interaction.", "Which daily routines help you feel settled, and where can you be flexible?"},
	"Bhakoot":      {"This compares the distance between Moon signs. A flagged pattern does not predict financial hardship, family conflict or relationship failure. Recorded exceptions do not change the displayed raw score.", "What expectations do you have about money, living arrangements and involvement from relatives?"},
	"Nadi":         {"This is a heavily weighted birth-star grouping. It cannot diagnose health, genetics or fertility, and must not be used to make claims about future children.", "If health or family planning matters to you, how would you discuss it respectfully with a qualified professional?"},
}

func ChartMatchExplanation(m engine.Match) MatchExplanation {
	r := MatchExplanation{Summary: fmt.Sprintf("The charts receive %g out of %g points under the app's selected Ashtakoota tables. This is a traditional score, not a percentage chance of a successful marriage. Read each factor separately and use conversations and lived experience to understand the relationship.", m.Total, m.Max), Factors: []MatchFactor{}, Limitations: []string{"Schools use different scoring tables and exceptions. Independent reference validation remains incomplete.", "This report does not assess consent, safety, shared values, behaviour or relationship success.", "Accurate birth details matter. This is not a full comparison of every house, divisional chart or timing period."}}
	for _, k := range m.Kootas {
		g := matchGuides[k.Name]
		r.Factors = append(r.Factors, MatchFactor{ID: engine.Slug(k.Name), Title: k.Name, Evidence: fmt.Sprintf("%g / %g points", k.Score, k.Max), Explanation: g[0], Question: g[1], Source: "Astrisk Ashtakoota engine; original plain-language guide v1 (not a classical quotation)"})
	}
	r.Factors = append(r.Factors, MatchFactor{ID: "mangal", Title: "Mars placement", Evidence: fmt.Sprintf("Groom flagged: %t; bride flagged: %t", m.BoyMangal, m.GirlMangal), Explanation: "The engine flags selected Mars houses under its stated rule. A flag is not evidence that someone is dangerous, angry or unsuitable. It does not predict harm to a partner.", Question: "How do you each manage frustration, repair disagreements and respect boundaries?", Source: "Astrisk MangalDosha engine; original plain-language guide v1"})
	return finishMatch(r)
}

// ProfileMatchExplanation takes comparisons only, not names, birth details,
// private messages, photos, social URLs, or free-text introductions.
func ProfileMatchExplanation(comparisons map[string]string, shared int) MatchExplanation {
	r := MatchExplanation{Summary: "This report compares what both people chose to publish. Similar answers can start a conversation; different or missing answers do not establish incompatibility.", Factors: []MatchFactor{}, Limitations: []string{"Profiles are self-declared, not identity-verified or a personality assessment.", "Matching text does not establish matching meaning. Different wording can express the same preference.", "No private character data, social-media research, birth charts, contact details or messages are used."}}
	items := [][3]string{
		{"city", "Location", "Where would you both want to live, and what practical constraints matter?"},
		{"timeline", "Relationship timeline", "What pace feels comfortable, and what needs to happen before a commitment?"},
		{"children", "Family plans", "What are your hopes or uncertainties about children, without assuming either person must agree?"},
		{"relocation", "Relocation", "Would either person consider moving, and how would work and support networks be affected?"},
		{"lifestyle", "Daily life", "Which routines, social habits and boundaries are important to each of you?"},
		{"values", "Values", "What does each stated value mean in everyday decisions? Can you give examples?"},
		{"hobbies", "Hobbies", "Which activities would you enjoy together, and which would you prefer to keep independent?"},
	}
	for _, item := range items {
		status := comparisons[item[0]]
		evidence, explanation := "Not shared by both people", "There is not enough information to compare this topic. Ask without assuming an answer; either person may choose to keep it private."
		if status == "same" {
			evidence = "Same wording in both published profiles"
			explanation = "Both people supplied the same wording. This may be a starting point, but ask what it means in practice before treating it as agreement."
		}
		if status == "different" {
			evidence = "Different wording in the published profiles"
			explanation = "The answers are worded differently. This may reflect different preferences or simply different descriptions. Clarify the practical expectations without labeling either person as wrong."
		}
		r.Factors = append(r.Factors, MatchFactor{ID: item[0], Title: item[1], Evidence: evidence, Explanation: explanation, Question: item[2], Source: "Consent-published profile fields; normalized text comparison, not an AI inference"})
	}
	r.Factors = append(r.Factors, MatchFactor{ID: "shared_interests", Title: "Shared selected interests", Evidence: fmt.Sprintf("%d shared selected interests", shared), Explanation: "Shared interest tags can help you choose a first conversation or activity. They do not show how often you participate, how important the activity is, or whether you will get along. No overlap is not a negative judgment.", Question: "Which interest matters most to you, and what would you enjoy introducing each other to?", Source: "Intersection of voluntarily published community interest tags"})
	return finishMatch(r)
}

func finishMatch(r MatchExplanation) MatchExplanation {
	r.Disclaimer = "For reflection, not certainty. Astrology is not a scientifically established predictor of relationship success. Choose freely, prioritise consent and safety, and consult qualified professionals for medical, legal or financial concerns."
	r.Model = "deterministic-guide"
	r.AIStatus = "not_requested"
	return r
}

// ExplainMatch only asks for interpretive prose; authoritative evidence and
// limitations remain server-rendered. Unchecked token streams are never shown.
func (s *Service) ExplainMatch(ctx context.Context, base MatchExplanation, useAI bool) MatchExplanation {
	if !useAI {
		return base
	}
	base.AIStatus = "unavailable"
	if s.LLM == nil {
		return base
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	b, _ := json.Marshal(base)
	prompt := `Explain these compatibility factors in warm, everyday English. Treat them as limited evidence, not traits of either person. Return JSON {"sections":[{"id":string,"explanation":string,"question":string}]}, exactly one section for every supplied factor, in the same order. Each explanation must be 2–4 useful sentences; each question must invite an open conversation. Do not restate numbers, scores, chart placements or classifications: the application displays those separately. Do not introduce any new yoga, planet, prediction, personality diagnosis, biological, caste, sexuality or health claim. Never recommend accepting or rejecting a partner, infer attraction, or predict marriage, divorce, fertility, harm or wealth. Do not infer preferences from a 'different' or 'missing' comparison. Never claim a source or book was consulted. Facts and plain-language guidance:
` + string(b)
	raw, err := s.LLM.Complete(ctx, prompt)
	if err != nil {
		return base
	}
	var out struct {
		Sections []struct {
			ID          string `json:"id"`
			Explanation string `json:"explanation"`
			Question    string `json:"question"`
		} `json:"sections"`
	}
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if dec.Decode(&out) != nil || len(out.Sections) != len(base.Factors) {
		base.AIStatus = "rejected"
		return base
	}
	for i, v := range out.Sections {
		if v.ID != base.Factors[i].ID || len(v.Explanation) < 80 || len(v.Explanation) > 1600 || len(v.Question) < 15 || len(v.Question) > 400 || unsafeMatchProse(v.Explanation+" "+v.Question) {
			base.AIStatus = "rejected"
			return base
		}
	}
	for i, v := range out.Sections {
		base.Factors[i].Explanation = v.Explanation
		base.Factors[i].Question = v.Question
	}
	base.Model = s.modelName()
	base.AIStatus = "generated"
	return base
}

// This is a conservative lexical safety check, not a proof of semantic truth.
func unsafeMatchProse(s string) bool {
	s = strings.ToLower(s)
	if strings.ContainsAny(s, "0123456789%<>") {
		return true
	}
	for _, term := range []string{"will die", "will divorce", "infertil", "genetic", "caste", "superior", "inferior", "guarantee", "destined", "soulmate", "perfect match", "should marry", "should not marry", "sexual compatibility", "yoga", "mahadasha", "antardasha", "cancer", "disease", "violent", "wealthy"} {
		if strings.Contains(s, term) {
			return true
		}
	}
	return false
}
