package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A fake translation service that upper-cases, so tests can see exactly
// which parts were sent.
func fakeTranslator(t *testing.T, calls *int) *Translator {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*calls++
		var in struct {
			Lang  string
			Texts []string
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		out := make([]string, len(in.Texts))
		for i, x := range in.Texts {
			out[i] = in.Lang + ":" + strings.ToUpper(x)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"texts": out})
	}))
	t.Cleanup(srv.Close)
	return NewTranslator(srv.URL, srv.Client())
}

func TestLocalizeTranslatesOnlyTheEverydayParts(t *testing.T) {
	calls := 0
	s := NewServer(nil, NoCache{}, nil)
	s.SetTranslator(fakeTranslator(t, &calls))
	answer := "In short\nSaturn tests you. Patience helps (Brihat Samhita 104.4).\n\nWhat this means for you\n• Work steadily. Rest well.\n• Supportive: 1 May 2027\n\nChart details\nSaturn in the 1st house.\n\nMore details.\n\nFor reflection, not certainty."
	got, ok := s.localize(context.Background(), answer, "hi")
	if !ok {
		t.Fatal("not translated")
	}
	want := "In short\nhi:SATURN TESTS YOU. hi:PATIENCE HELPS (BRIHAT SAMHITA 104.4).\n\nWhat this means for you\n• hi:WORK STEADILY. hi:REST WELL.\n• hi:SUPPORTIVE: 1 MAY 2027\n\nChart details\nSaturn in the 1st house.\n\nMore details.\n\nFor reflection, not certainty."
	if got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	// Repeated sentences come from the cache.
	if _, ok := s.localize(context.Background(), answer, "hi"); !ok || calls != 1 {
		t.Fatalf("calls %d", calls)
	}
	for _, lang := range []string{"", "en", "xx"} {
		if _, ok := s.localize(context.Background(), answer, lang); ok {
			t.Fatalf("%q translated", lang)
		}
	}
}

func TestLocalizeFallsBackToEnglish(t *testing.T) {
	s := NewServer(nil, NoCache{}, nil)
	gone := httptest.NewServer(http.NotFoundHandler())
	gone.Close()
	s.SetTranslator(NewTranslator(gone.URL, http.DefaultClient))
	if _, ok := s.localize(context.Background(), "In short\nHello there.", "ta"); ok {
		t.Fatal("translated without a service")
	}
}

func TestChatStoresTheEnglishOriginal(t *testing.T) {
	calls := 0
	s := NewServer(realEngine(t), NoCache{}, nil)
	s.SetTranslator(fakeTranslator(t, &calls))
	code, _, body := do(t, s, "POST", "/api/chat", `{"birth":{"date":"1996-05-14","time":"10:15","lat":12.97,"lon":77.59,"tz":"Asia/Kolkata"},"question":"What does my Moon sign say?","lang":"kn"}`)
	if code != 200 {
		t.Fatal(code, body)
	}
	var out struct{ Answer ChatMessage }
	_ = json.Unmarshal([]byte(body), &out)
	if out.Answer.Lang != "kn" || !strings.HasPrefix(out.Answer.Original, "In short\n") || !strings.Contains(out.Answer.Content, "kn:") {
		t.Fatalf("%+v", out.Answer)
	}
}
