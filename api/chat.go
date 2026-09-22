package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/example/panchang/engine"
	"github.com/example/panchang/reading"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ChatMessage is one stored turn.
type ChatMessage struct {
	ID        int64          `json:"id"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	Topics    []string       `json:"topics"`
	Sources   []reading.Rule `json:"sources"`
	Model     string         `json:"model,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// ChatStore persists conversations. Sessions are bound to one birth chart.
type ChatStore interface {
	Session(ctx context.Context, id string) (chartHash string, ok bool, err error)
	CreateSession(ctx context.Context, id, chartHash string, birth engine.ChartInput) error
	Append(ctx context.Context, sessionID string, m ChatMessage) (ChatMessage, error)
	History(ctx context.Context, sessionID string, limit int) ([]ChatMessage, error)
	// Delete removes a session and every message in it (user data deletion).
	Delete(ctx context.Context, sessionID string) error
}

var sessionIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// MemoryChatStore keeps chats in process memory (no database configured).
type MemoryChatStore struct {
	mu       sync.Mutex
	sessions map[string]string
	messages map[string][]ChatMessage
	next     int64
}

func NewMemoryChatStore() *MemoryChatStore {
	return &MemoryChatStore{sessions: map[string]string{}, messages: map[string][]ChatMessage{}}
}

func (m *MemoryChatStore) Session(_ context.Context, id string) (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h, ok := m.sessions[id]
	return h, ok, nil
}

func (m *MemoryChatStore) CreateSession(_ context.Context, id, hash string, _ engine.ChartInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.sessions) > 10000 {
		return errors.New("chat store is full")
	}
	m.sessions[id] = hash
	return nil
}

func (m *MemoryChatStore) Append(_ context.Context, sid string, msg ChatMessage) (ChatMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	msg.ID, msg.CreatedAt = m.next, time.Now().UTC()
	m.messages[sid] = append(m.messages[sid], msg)
	return msg, nil
}

func (m *MemoryChatStore) History(_ context.Context, sid string, limit int) ([]ChatMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h := m.messages[sid]
	if len(h) > limit {
		h = h[len(h)-limit:]
	}
	return append([]ChatMessage{}, h...), nil
}

func (m *MemoryChatStore) Delete(_ context.Context, sid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sid)
	delete(m.messages, sid)
	return nil
}

// PostgresChatStore persists chats in chat_sessions / chat_messages.
type PostgresChatStore struct{ Pool *pgxpool.Pool }

func (p PostgresChatStore) Session(ctx context.Context, id string) (string, bool, error) {
	var h string
	err := p.Pool.QueryRow(ctx, `SELECT chart_hash FROM chat_sessions WHERE id=$1`, id).Scan(&h)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}
	return h, err == nil, err
}

func (p PostgresChatStore) CreateSession(ctx context.Context, id, hash string, birth engine.ChartInput) error {
	b, _ := json.Marshal(birth)
	_, err := p.Pool.Exec(ctx, `INSERT INTO chat_sessions(id,chart_hash,birth) VALUES($1,$2,$3)`, id, hash, b)
	return err
}

func (p PostgresChatStore) Append(ctx context.Context, sid string, m ChatMessage) (ChatMessage, error) {
	src, _ := json.Marshal(m.Sources)
	if m.Topics == nil {
		m.Topics = []string{}
	}
	err := p.Pool.QueryRow(ctx, `INSERT INTO chat_messages(session_id,role,content,topics,sources,model) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,created_at`,
		sid, m.Role, m.Content, m.Topics, src, m.Model).Scan(&m.ID, &m.CreatedAt)
	if err == nil {
		_, _ = p.Pool.Exec(ctx, `UPDATE chat_sessions SET updated_at=now() WHERE id=$1`, sid)
	}
	return m, err
}

func (p PostgresChatStore) Delete(ctx context.Context, sid string) error {
	_, err := p.Pool.Exec(ctx, `DELETE FROM chat_sessions WHERE id=$1`, sid) // messages cascade
	return err
}

func (p PostgresChatStore) History(ctx context.Context, sid string, limit int) ([]ChatMessage, error) {
	rows, err := p.Pool.Query(ctx, `SELECT id,role,content,topics,sources,model,created_at FROM (SELECT * FROM chat_messages WHERE session_id=$1 ORDER BY id DESC LIMIT $2) t ORDER BY id`, sid, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ChatMessage{}
	for rows.Next() {
		var m ChatMessage
		var src []byte
		if err = rows.Scan(&m.ID, &m.Role, &m.Content, &m.Topics, &src, &m.Model, &m.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(src, &m.Sources)
		if m.Sources == nil {
			m.Sources = []reading.Rule{}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type chatRequest struct {
	SessionID string            `json:"session_id"`
	Birth     engine.ChartInput `json:"birth"`
	Question  string            `json:"question"`
}

func (s *Server) chat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&req); err != nil {
		problem(w, 400, fmt.Errorf("invalid JSON body"))
		return
	}
	c, err := s.engine.BirthChart(req.Birth)
	if err != nil {
		problem(w, 400, err)
		return
	}
	now := time.Now()
	facts, rules, err := s.reading.BuildFacts(r.Context(), c, now)
	if err != nil {
		problem(w, 500, err)
		return
	}
	hash := reading.ChartHash(req.Birth, "en")
	sid := req.SessionID
	if sid != "" {
		h, ok, err := s.chats.Session(r.Context(), sid)
		if err != nil {
			problem(w, 500, err)
			return
		}
		if !ok || h != hash {
			sid = "" // unknown session, or one that belongs to another chart
		}
	}
	if sid == "" {
		sid = newSessionID()
		if err = s.chats.CreateSession(r.Context(), sid, hash, req.Birth); err != nil {
			problem(w, 500, err)
			return
		}
	}
	prior, err := s.chats.History(r.Context(), sid, 12)
	if err != nil {
		problem(w, 500, err)
		return
	}
	turns := make([]reading.ChatTurn, len(prior))
	for i, m := range prior {
		turns[i] = reading.ChatTurn{Role: m.Role, Content: m.Content}
	}
	cc := reading.ChatContext{}
	loc, _ := time.LoadLocation(req.Birth.TZ)
	nowLocal := now.In(loc)
	if t, err := s.engine.BirthChart(engine.ChartInput{Date: nowLocal.Format("2006-01-02"), Time: nowLocal.Format("15:04"), Lat: req.Birth.Lat, Lon: req.Birth.Lon, TZ: req.Birth.TZ}); err == nil {
		cc.Transit = &t
	}
	if sc, ok := s.engine.(ShadbalaCalculator); ok {
		if sb, err := sc.Shadbala(c); err == nil {
			cc.Shadbala = &sb
		}
	}
	ans, err := s.reading.Answer(r.Context(), facts, rules, req.Question, turns, cc)
	if err != nil {
		problem(w, 400, err)
		return
	}
	q, err := s.chats.Append(r.Context(), sid, ChatMessage{Role: "user", Content: req.Question, Topics: []string{}, Sources: []reading.Rule{}})
	if err != nil {
		problem(w, 500, err)
		return
	}
	a, err := s.chats.Append(r.Context(), sid, ChatMessage{Role: "assistant", Content: ans.Answer, Topics: ans.Topics, Sources: ans.Sources, Model: ans.Model})
	if err != nil {
		problem(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"session_id": sid, "question": q, "answer": a})
}

func (s *Server) chatHistory(w http.ResponseWriter, r *http.Request) {
	sid := r.URL.Query().Get("session_id")
	if !sessionIDPattern.MatchString(sid) {
		problem(w, 400, fmt.Errorf("invalid session_id"))
		return
	}
	if _, ok, err := s.chats.Session(r.Context(), sid); err != nil {
		problem(w, 500, err)
		return
	} else if !ok {
		problem(w, 404, fmt.Errorf("chat session not found"))
		return
	}
	msgs, err := s.chats.History(r.Context(), sid, 500)
	if err != nil {
		problem(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"session_id": sid, "messages": msgs})
}
