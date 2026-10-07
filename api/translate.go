package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"unicode"

	"github.com/example/panchang/reading"
)

// Machine translation of chat answers into the member's language, through
// the local translation service (scripts/translate-server.py, IndicTrans2,
// loopback only, free). Only the everyday parts of an answer are translated:
// "In short" and "What this means for you". The fixed headings and the
// closing line stay in English as markers, which the app shows in the
// member's language, and the technical chart details stay in English. When
// the service is missing or slow the English answer is served unchanged.

type Translator struct {
	URL  string
	HTTP *http.Client

	mu    sync.Mutex
	cache map[string]string
}

const translateCacheSize = 20000

func NewTranslator(url string, client *http.Client) *Translator {
	return &Translator{URL: strings.TrimRight(url, "/"), HTTP: client, cache: map[string]string{}}
}

// Translate turns English sentences into lang (an app language code),
// answering repeated sentences from memory.
func (t *Translator) Translate(ctx context.Context, lang string, texts []string) ([]string, error) {
	out := make([]string, len(texts))
	var todo []string
	var at []int
	t.mu.Lock()
	for i, x := range texts {
		if v, ok := t.cache[lang+"\x00"+x]; ok {
			out[i] = v
		} else {
			todo, at = append(todo, x), append(at, i)
		}
	}
	t.mu.Unlock()
	if len(todo) == 0 {
		return out, nil
	}
	body, _ := json.Marshal(map[string]any{"lang": lang, "texts": todo})
	req, err := http.NewRequestWithContext(ctx, "POST", t.URL+"/translate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := t.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("translation service: %s", res.Status)
	}
	var got struct {
		Texts []string `json:"texts"`
	}
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		return nil, err
	}
	if len(got.Texts) != len(todo) {
		return nil, fmt.Errorf("translation service returned %d of %d texts", len(got.Texts), len(todo))
	}
	t.mu.Lock()
	if len(t.cache)+len(todo) > translateCacheSize {
		t.cache = map[string]string{}
	}
	for j, i := range at {
		out[i] = got.Texts[j]
		t.cache[lang+"\x00"+todo[j]] = got.Texts[j]
	}
	t.mu.Unlock()
	return out, nil
}

const (
	ansShort   = "In short"
	ansPoints  = "What this means for you"
	ansDetails = "Chart details"
	ansClosing = "For reflection, not certainty."
)

// localize translates an answer's everyday parts into lang. It reports
// false, leaving the English answer to be served, when lang is English or
// unsupported, no translator is configured, or translation fails.
func (s *Server) localize(ctx context.Context, answer, lang string) (string, bool) {
	if s.translator == nil || answer == "" {
		return "", false
	}
	if _, ok := reading.LanguageNames[lang]; !ok {
		return "", false
	}
	blocks := strings.Split(answer, "\n\n")
	plain := strings.HasPrefix(answer, ansShort+"\n")
	// Each translatable line is split into sentences; slots remember where
	// every sentence goes back.
	type slot struct{ block, line int }
	var sentences []string
	var where []slot
	lines := make([][]string, len(blocks))
	parts := make([][][]string, len(blocks))
	inDetails := false
	for b, blk := range blocks {
		lines[b] = strings.Split(blk, "\n")
		parts[b] = make([][]string, len(lines[b]))
		if blk == ansClosing {
			continue
		}
		if plain && strings.HasPrefix(blk, ansDetails+"\n") {
			inDetails = true
		}
		if inDetails {
			continue
		}
		for l, line := range lines[b] {
			if plain && l == 0 && (line == ansShort || line == ansPoints) {
				continue
			}
			text := strings.TrimPrefix(line, "• ")
			for _, sen := range splitSentences(text) {
				parts[b][l] = append(parts[b][l], sen)
				sentences = append(sentences, sen)
				where = append(where, slot{b, l})
			}
		}
	}
	if len(sentences) == 0 {
		return "", false
	}
	got, err := s.translator.Translate(ctx, lang, sentences)
	if err != nil {
		s.logger.Warn("translation unavailable", "lang", lang, "err", err)
		return "", false
	}
	done := make([][][]string, len(blocks))
	for b := range blocks {
		done[b] = make([][]string, len(lines[b]))
	}
	for i, w := range where {
		done[w.block][w.line] = append(done[w.block][w.line], got[i])
	}
	for b := range blocks {
		for l, line := range lines[b] {
			if parts[b][l] == nil {
				continue
			}
			prefix := ""
			if strings.HasPrefix(line, "• ") {
				prefix = "• "
			}
			lines[b][l] = prefix + strings.Join(done[b][l], " ")
		}
		blocks[b] = strings.Join(lines[b], "\n")
	}
	return strings.Join(blocks, "\n\n"), true
}

// splitSentences splits English text at ". ", "? " and "! " followed by a
// capital letter, which keeps references such as "Brihat Samhita 104.4" and
// dates whole.
func splitSentences(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	var out []string
	r := []rune(text)
	start := 0
	for i := 0; i+2 < len(r); i++ {
		if (r[i] == '.' || r[i] == '?' || r[i] == '!') && r[i+1] == ' ' && unicode.IsUpper(r[i+2]) {
			out = append(out, strings.TrimSpace(string(r[start:i+1])))
			start = i + 2
		}
	}
	return append(out, strings.TrimSpace(string(r[start:])))
}

// localizedAnswer is a chat answer as JSON, translated when possible, with
// the English original kept alongside.
func (s *Server) localizedAnswer(ctx context.Context, ans reading.ChatAnswer, lang string) any {
	t, ok := s.localize(ctx, ans.Answer, lang)
	if !ok {
		return ans
	}
	return struct {
		reading.ChatAnswer
		Original string `json:"original"`
		Lang     string `json:"lang"`
	}{reading.ChatAnswer{Answer: t, Topics: ans.Topics, Sources: ans.Sources, Model: ans.Model}, ans.Answer, lang}
}
