package divination

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type TarotCard struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Meaning    string `json:"meaning"`
	Reflection string `json:"reflection"`
	Reversed   bool   `json:"reversed"`
	Position   string `json:"position"`
}
type TarotReading struct {
	Cards  []TarotCard `json:"cards"`
	Deck   string      `json:"deck"`
	Method string      `json:"method"`
	Note   string      `json:"note"`
}

var majors = [][3]string{
	{"The Fool", "A fresh start invites curiosity and a sensible first step.", "What would you try if you could start small?"},
	{"The Magician", "Notice the skills and resources you already have.", "Which resource could you put to use today?"},
	{"The High Priestess", "Give yourself room to reflect before acting.", "What needs closer attention before you decide?"},
	{"The Empress", "Care, creativity and nourishment are the themes here.", "What in your life would benefit from consistent care?"},
	{"The Emperor", "Clear boundaries and a workable structure can support you.", "Which boundary or plan would help you feel organised?"},
	{"The Hierophant", "Consider what you learn from traditions, teachers and shared values.", "Which guidance fits your values, and which deserves a question?"},
	{"The Lovers", "Focus on an honest choice and the values behind it.", "What choice would be consistent with what matters to you?"},
	{"The Chariot", "Choose a direction and bring competing priorities into balance.", "What is the next step toward a goal you have chosen?"},
	{"Strength", "Patience and a calm response can be forms of courage.", "Where could gentleness work better than force?"},
	{"The Hermit", "Quiet time can help you sort your own priorities.", "What becomes clearer when you step back from outside opinions?"},
	{"Wheel of Fortune", "Circumstances change; consider what remains within your control.", "What can you adapt when a plan changes?"},
	{"Justice", "Look at evidence, fairness and responsibility.", "Have you considered the consequences for everyone involved?"},
	{"The Hanged Man", "A pause may let you see the situation from another angle.", "What assumption could you reconsider?"},
	{"Death", "This card is a symbol of endings and transition. It does not predict a death.", "What habit or chapter are you ready to leave behind?"},
	{"Temperance", "Find a sustainable balance between competing needs.", "What would a moderate, workable approach look like?"},
	{"The Devil", "Notice habits or pressures that make your choices feel smaller.", "What practical boundary could restore a sense of choice?"},
	{"The Tower", "Question a shaky assumption and prepare for change thoughtfully.", "Which plan would benefit from a backup option?"},
	{"The Star", "Hope can be supported by small acts of care and renewal.", "What helps you recover your sense of direction?"},
	{"The Moon", "Uncertainty calls for patience and checking assumptions.", "What do you know, and what still needs clarification?"},
	{"The Sun", "Make room for enjoyment, openness and recognition of progress.", "What progress can you acknowledge today?"},
	{"Judgement", "Reflect on what you have learned and what you want to do differently.", "Which lesson will you carry into your next decision?"},
	{"The World", "Recognise a completed stage and the work it took to get here.", "What deserves closure or celebration?"},
}
var minorThemes = [][]string{
	{"a new creative spark", "planning a direction", "looking ahead", "a shared milestone", "competing ideas", "recognising effort", "protecting your position", "moving plans forward", "persistence with boundaries", "carrying too much", "curious experimentation", "enthusiastic action", "confident encouragement", "responsible leadership"},
	{"openness to connection", "mutual understanding", "friendship and support", "noticing overlooked options", "acknowledging disappointment", "memories and familiar comforts", "sorting wishes from realistic options", "leaving an unfulfilling pattern", "appreciating what is enough", "shared belonging", "gentle curiosity", "expressing feelings thoughtfully", "empathetic listening", "steady emotional judgement"},
	{"a moment of clarity", "a decision needing information", "processing hurt through support", "rest and perspective", "the cost of winning an argument", "moving toward calmer circumstances", "honesty about a strategy", "questioning a limiting assumption", "addressing a worry with grounded support", "closing a difficult chapter", "asking careful questions", "slowing down a rushed response", "clear boundaries and communication", "fair, reasoned decisions"},
	{"a practical opportunity", "balancing daily commitments", "learning through teamwork", "security without excessive control", "asking for practical support", "giving and receiving fairly", "reviewing patient effort", "practising a useful skill", "enjoying earned independence", "long-term shared foundations", "learning a practical skill", "steady follow-through", "practical care", "responsible stewardship"},
}

func TarotDeck() []TarotCard {
	out := make([]TarotCard, 0, 78)
	for i, m := range majors {
		out = append(out, TarotCard{ID: fmt.Sprintf("major_%02d", i), Name: m[0], Meaning: m[1], Reflection: m[2]})
	}
	ranks := []string{"Ace", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten", "Page", "Knight", "Queen", "King"}
	for s, suit := range []string{"Wands", "Cups", "Swords", "Pentacles"} {
		for r, rank := range ranks {
			theme := minorThemes[s][r]
			out = append(out, TarotCard{ID: fmt.Sprintf("minor_%d_%02d", s, r), Name: rank + " of " + suit, Meaning: "Use this card to reflect on " + theme + ".", Reflection: "How does this theme relate to your situation, and what small action would be helpful?"})
		}
	}
	return out
}
func DrawTarot(count int, reversals bool) (TarotReading, error) {
	if count != 1 && count != 3 {
		return TarotReading{}, fmt.Errorf("choose one card or three cards")
	}
	deck := TarotDeck()
	out := TarotReading{Cards: []TarotCard{}, Deck: "78-card structure; original Astrisk wording", Method: "Random cards drawn without replacement. Reversals invite reflection on a blocked or overused theme.", Note: "Use the cards as reflection prompts. They do not establish facts about other people or predict events."}
	labels := []string{"Focus", "Challenge to consider", "A possible next step"}
	for i := 0; i < count; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(deck))))
		if err != nil {
			return TarotReading{}, err
		}
		j := int(n.Int64())
		c := deck[j]
		deck = append(deck[:j], deck[j+1:]...)
		c.Position = labels[i]
		if reversals {
			n, err = rand.Int(rand.Reader, big.NewInt(2))
			if err != nil {
				return TarotReading{}, err
			}
			c.Reversed = n.Int64() == 1
			if c.Reversed {
				c.Reflection = "Where might this theme feel blocked or be taken too far? " + c.Reflection
			}
		}
		out.Cards = append(out.Cards, c)
	}
	return out, nil
}
