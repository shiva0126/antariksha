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
	Result      string `json:"result"`
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
		r.Factors = append(r.Factors, MatchFactor{ID: engine.Slug(k.Name), Title: k.Name, Evidence: fmt.Sprintf("%g / %g points", k.Score, k.Max), Result: kootaResult(k, m.Exceptions), Explanation: g[0], Question: g[1], Source: "Astrisk Ashtakoota engine; original plain-language guide v2 (not a classical quotation)"})
	}
	marsResult := "Neither chart triggers the app's Lagna-based Mars rule. This is not a complete assessment of Mars from other reference points or a statement about either person's temperament."
	if m.BoyMangal && m.GirlMangal {
		marsResult = "Both charts trigger the app's Lagna-based Mars rule. The selected tradition treats this as a mutual cancellation; that does not establish relationship safety or success."
	} else if m.BoyMangal || m.GirlMangal {
		marsResult = "Only one chart triggers the app's Lagna-based Mars rule. The comparison does not evaluate every other reference point or cancellation, and it is not a reason to reject either person."
	}
	r.Factors = append(r.Factors, MatchFactor{ID: "mangal", Title: "Mars placement", Evidence: fmt.Sprintf("Groom flagged: %t; bride flagged: %t", m.BoyMangal, m.GirlMangal), Result: marsResult, Explanation: "The engine flags selected Mars houses under its stated rule. A flag is not evidence that someone is dangerous, angry or unsuitable. It does not predict harm to a partner.", Question: "How do you each manage frustration, repair disagreements and respect boundaries?", Source: "Astrisk MangalDosha engine; original plain-language guide v2"})
	return finishMatch(r)
}

// kootaResult explains only the returned score and recorded exceptions. It
// never recomputes astronomy, adjusts points, or infers a personal trait.
func kootaResult(k engine.Koota, exceptions []string) string {
	result := "This pair receives part of the available points in the selected table. This describes the table result, not partial agreement between the people."
	if k.Score == 0 {
		result = "This pair receives no points for this factor in the selected table. That is not evidence that the relationship will fail or that either person has a fault."
	} else if k.Score == k.Max {
		result = "This pair receives all available points for this factor in the selected table. That is not proof of real-world compatibility or a reason to skip important conversations."
	}
	if k.Name == "Bhakoot" || k.Name == "Nadi" {
		if k.Score == 0 {
			result += " The engine records a " + k.Name + " flag."
			found := false
			for _, note := range exceptions {
				if strings.HasPrefix(note, k.Name+" dosha ") || (k.Name == "Nadi" && strings.HasPrefix(note, "Some traditions cancel Nadi dosha ")) {
					result += " " + note
					found = true
				}
			}
			if found {
				result += " The exception is shown separately: the raw points and total are unchanged."
			} else {
				result += " No exception is recorded by the app's implemented rules; this does not mean every tradition would reach the same conclusion."
			}
		}
	}
	return result
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
		r.Factors = append(r.Factors, MatchFactor{ID: item[0], Title: item[1], Evidence: evidence, Result: profileResult(item[0], status), Explanation: explanation, Question: item[2], Source: "Consent-published profile fields; normalized text comparison, not an AI inference"})
	}
	interestResult := "There are no shared selected tags in the published profiles. Private, unlisted or differently named interests are not compared."
	if shared > 0 {
		interestResult = "Both profiles selected at least one of the same interest tags. You can use that as an optional conversation opener, without assuming equal enthusiasm or experience."
	}
	r.Factors = append(r.Factors, MatchFactor{ID: "shared_interests", Title: "Shared selected interests", Evidence: fmt.Sprintf("%d shared selected interests", shared), Result: interestResult, Explanation: "Shared interest tags can help you choose a first conversation or activity. They do not show how often you participate, how important the activity is, or whether you will get along. No overlap is not a negative judgment.", Question: "Which interest matters most to you, and what would you enjoy introducing each other to?", Source: "Intersection of voluntarily published community interest tags"})
	return finishMatch(r)
}

func profileResult(topic, status string) string {
	if status != "same" && status != "different" {
		return "This topic remains unknown because it was not shared by both people. It is not counted as disagreement, and neither person needs to disclose more than they want."
	}
	context := map[string]string{
		"city":       "A city label does not tell you where someone wants to settle, their commute, or their willingness to move.",
		"timeline":   "A timeline can depend on work, study, family responsibilities and personal readiness. It is not a commitment or deadline.",
		"children":   "Family plans deserve a voluntary conversation, including uncertainty and the possibility of changing preferences. This field says nothing about fertility.",
		"relocation": "A willingness to move may depend on destination, work, caregiving and support. Clarify those conditions rather than assuming a promise.",
		"lifestyle":  "A short lifestyle description cannot capture daily routines. Discuss schedules, personal space and which habits are flexible.",
		"values":     "The same value can lead to different practical choices. Discuss a real example of how each person applies it, without judging their character.",
		"hobbies":    "A hobby label does not measure time, skill or enthusiasm. Separate activities can be as important as shared ones.",
	}
	return context[topic]
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
	// The caller may reuse the deterministic report after an AI attempt.
	base.Factors = append([]MatchFactor(nil), base.Factors...)
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
