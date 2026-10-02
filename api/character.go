package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type characterSource struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	URL      string `json:"url"`
	Text     string `json:"text"`
}
type characterProfile struct {
	Name          string            `json:"name"`
	Motto         string            `json:"motto"`
	Avatar        string            `json:"avatar"`
	Accent        string            `json:"accent"`
	Backdrop      string            `json:"backdrop"`
	Interests     []string          `json:"interests"`
	Goals         string            `json:"goals"`
	Values        string            `json:"values"`
	Communication string            `json:"communication"`
	WeeklyMinutes int               `json:"weekly_minutes"`
	Sources       []characterSource `json:"sources"`
}
type hobbyIdea struct {
	Interest string `json:"interest"`
	Activity string `json:"activity"`
	Together string `json:"together"`
	Reason   string `json:"reason,omitempty"`
}

var hobbyCatalog = []hobbyIdea{
	{"reading", "Read a chapter and note one idea you want to discuss.", "Swap a favourite book recommendation.", ""},
	{"hiking", "Plan a short walk on a familiar local trail.", "Choose a public daytime trail and agree on a comfortable pace.", ""},
	{"photography", "Make a small photo series around one everyday theme.", "Try a photo walk with permission before photographing people.", ""},
	{"cooking", "Try one new recipe using ingredients you already enjoy.", "Share a recipe and talk about favourite family dishes.", ""},
	{"music", "Set aside a session to practise or listen closely to a favourite piece.", "Make a short playlist together and explain your choices.", ""},
	{"art", "Sketch or paint something from your surroundings.", "Visit a public exhibition or exchange sketches.", ""},
	{"gardening", "Care for one plant and keep a simple growth journal.", "Exchange plant-care tips or visit a community garden.", ""},
	{"astronomy", "Learn a constellation and check when it is visible locally.", "Plan an evening at a public observatory or astronomy club.", ""},
	{"technology", "Build a small project around something you use every day.", "Share a demo or attend a public maker event.", ""},
	{"writing", "Write a short journal entry or story from a prompt.", "Exchange a favourite poem or attend a writing group.", ""},
	{"languages", "Practise a few phrases in a language you chose to learn.", "Try a language exchange and take turns helping each other.", ""},
	{"sports", "Practise a sport you already enjoy at your preferred pace.", "Join a beginner-friendly public game or watch a match together.", ""},
	{"dance", "Practise a short routine or explore a style that interests you.", "Try a beginner class together.", ""},
	{"crafts", "Make a small handmade item with materials you have.", "Attend a local workshop or share a project idea.", ""},
	{"volunteering", "Find a cause you personally want to support.", "Discuss a community activity you would both enjoy helping with.", ""},
	{"travel", "Plan a local day trip within your chosen budget.", "Compare the places you would each like to explore.", ""},
}
var sourceWords = regexp.MustCompile(`[\p{L}]+`)

func characterIdeas(p characterProfile) []hobbyIdea {
	out := []hobbyIdea{}
	for _, h := range hobbyCatalog {
		for _, interest := range p.Interests {
			if strings.EqualFold(interest, h.Interest) {
				h.Reason = fmt.Sprintf("You selected %s and set aside %d minutes per week.", h.Interest, p.WeeklyMinutes)
				out = append(out, h)
				break
			}
		}
	}
	return out
}
func characterDefaults() characterProfile {
	return characterProfile{Avatar: "star", Accent: "#d6b467", Backdrop: "orbits", Interests: []string{}, Sources: []characterSource{}, WeeklyMinutes: 30}
}
func validCharacter(p characterProfile) bool {
	if len(p.Name) > 100 || len(p.Motto) > 200 || len(p.Goals) > 1000 || len(p.Values) > 1000 || len(p.Communication) > 500 || len(p.Interests) > 20 || len(p.Sources) > 8 || p.WeeklyMinutes < 15 || p.WeeklyMinutes > 600 {
		return false
	}
	if !strings.Contains("|sun|moon|star|leaf|mountain|", "|"+p.Avatar+"|") || p.Avatar == "" {
		return false
	}
	if p.Backdrop != "orbits" && p.Backdrop != "rays" && p.Backdrop != "plain" {
		return false
	}
	if len(p.Accent) != 7 || p.Accent[0] != '#' {
		return false
	}
	if _, err := strconv.ParseUint(p.Accent[1:], 16, 32); err != nil {
		return false
	}
	seen := map[string]bool{}
	for _, s := range p.Sources {
		if !chartID.MatchString(s.ID) || seen[s.ID] || len(s.Text) > 2000 || strings.TrimSpace(s.Text) == "" || (s.Platform != "instagram" && s.Platform != "linkedin" && s.Platform != "own") {
			return false
		}
		seen[s.ID] = true
		if s.URL != "" {
			u, e := url.Parse(s.URL)
			if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || len(s.URL) > 300 {
				return false
			}
		}
	}
	seen = map[string]bool{}
	for _, v := range p.Interests {
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "" || len(v) > 40 || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}
func (s *Server) characterRoutes() {
	s.memberRoute("GET /api/me/character", s.getCharacter)
	s.memberRoute("PUT /api/me/character", s.saveCharacter)
	s.memberRoute("POST /api/me/character/preview", s.previewCharacterSource)
}
func (s *Server) getCharacter(w http.ResponseWriter, r *http.Request, id string) {
	var raw []byte
	var revision int64
	err := s.membersDB().QueryRow(r.Context(), `SELECT c.profile,COALESCE(c.revision,0) FROM member_accounts a LEFT JOIN member_characters c ON c.account_id=a.id WHERE a.id=$1`, id).Scan(&raw, &revision)
	if err != nil {
		problem(w, 500, fmt.Errorf("character unavailable"))
		return
	}
	p := characterDefaults()
	if len(raw) > 0 {
		if err = json.Unmarshal(raw, &p); err != nil {
			problem(w, 500, fmt.Errorf("character unavailable"))
			return
		}
	}
	writeJSON(w, 200, map[string]any{"account_id": id, "revision": revision, "profile": p, "ideas": characterIdeas(p), "hobby_options": hobbyCatalog, "source_method": "self_provided"})
}
func (s *Server) saveCharacter(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		AccountID string           `json:"account_id"`
		Revision  int64            `json:"revision"`
		Profile   characterProfile `json:"profile"`
		Consent   bool             `json:"consent"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10))
	dec.DisallowUnknownFields()
	if dec.Decode(&in) != nil || dec.Decode(new(any)) != io.EOF || in.AccountID != id || in.Revision < 0 || !in.Consent || !validCharacter(in.Profile) {
		problem(w, 400, fmt.Errorf("check character fields and agree to private storage"))
		return
	}
	for i, v := range in.Profile.Interests {
		in.Profile.Interests[i] = strings.ToLower(strings.TrimSpace(v))
	}
	if in.Profile.Sources == nil {
		in.Profile.Sources = []characterSource{}
	}
	if in.Profile.Interests == nil {
		in.Profile.Interests = []string{}
	}
	raw, _ := json.Marshal(in.Profile)
	tx, err := s.membersDB().Begin(r.Context())
	if err != nil {
		problem(w, 500, fmt.Errorf("save unavailable"))
		return
	}
	defer tx.Rollback(r.Context())
	var owner string
	if err = tx.QueryRow(r.Context(), `SELECT id FROM member_accounts WHERE id=$1 FOR UPDATE`, id).Scan(&owner); err != nil {
		problem(w, 401, fmt.Errorf("sign in required"))
		return
	}
	var current int64
	if err = tx.QueryRow(r.Context(), `SELECT COALESCE((SELECT revision FROM member_characters WHERE account_id=$1),0)`, id).Scan(&current); err != nil {
		problem(w, 500, fmt.Errorf("save unavailable"))
		return
	}
	if current != in.Revision {
		problem(w, 409, fmt.Errorf("character changed on another device; reload before saving"))
		return
	}
	_, err = tx.Exec(r.Context(), `INSERT INTO member_characters(account_id,profile) VALUES($1,$2) ON CONFLICT(account_id) DO UPDATE SET profile=$2,revision=member_characters.revision+1,updated_at=now()`, id, raw)
	if err != nil || tx.Commit(r.Context()) != nil {
		problem(w, 500, fmt.Errorf("character could not be saved"))
		return
	}
	writeJSON(w, 200, map[string]any{"revision": current + 1, "ideas": characterIdeas(in.Profile)})
}
func (s *Server) previewCharacterSource(w http.ResponseWriter, r *http.Request, id string) {
	var in struct {
		Text string `json:"text"`
	}
	if !memberInput(w, r, &in) {
		return
	}
	if len(in.Text) > 2000 {
		problem(w, 400, fmt.Errorf("preview up to 2000 characters"))
		return
	}
	words := map[string]bool{}
	for _, v := range sourceWords.FindAllString(strings.ToLower(in.Text), -1) {
		words[v] = true
	}
	topics := []string{}
	for _, h := range hobbyCatalog {
		if words[h.Interest] {
			topics = append(topics, h.Interest)
		}
	}
	writeJSON(w, 200, map[string]any{"mentioned_topics": topics, "stored": false})
}
