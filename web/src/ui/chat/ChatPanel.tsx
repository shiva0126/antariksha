import { useEffect, useRef, useState } from 'react';
import { getChatHistory, postChat } from '../../api/client';
import type { ChartInput, ChatMessage } from '../../api/types';

const suggestions = [
  'What does my chart say about my career?',
  'What about marriage and relationships?',
  'Which dasha am I running and what does it mean?',
  'Do I have Mangal dosha?',
  'Am I in Sade Sati?',
  'What is my nakshatra like?',
  'Explain my yogas',
  'What does my Saturn mean?',
];

const storageKey = (b: ChartInput) => `antariksha.chat.${b.date}.${b.time}.${b.lat}.${b.lon}.${b.tz}`;
const load = (k: string) => { try { return localStorage.getItem(k) ?? undefined; } catch { return undefined; } };
const save = (k: string, v?: string) => { try { v ? localStorage.setItem(k, v) : localStorage.removeItem(k); } catch { /* private mode */ } };

function Answer({ m }: { m: ChatMessage }) {
  return (
    <>
      {m.content.split(/\n\n+/).map((p, i) => <p key={i}>{p}</p>)}
      {m.sources.length > 0 && (
        <details className="sources">
          <summary>{m.sources.length} source{m.sources.length > 1 ? 's' : ''} · {m.model === 'grounded-corpus' ? 'grounded answer' : m.model}</summary>
          <ul>{m.sources.map(s => <li key={s.doc_type + s.key + s.source}><b>{s.title}</b><span>{s.source.replace(/ \[.*?\]$/, '')}</span></li>)}</ul>
        </details>
      )}
    </>
  );
}

export function ChatPanel({ birth, name }: { birth: ChartInput; name?: string }) {
  const key = storageKey(birth);
  const [sessionId, setSessionId] = useState(() => load(key));
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [draft, setDraft] = useState('');
  const [pending, setPending] = useState(false);
  const [loading, setLoading] = useState(Boolean(sessionId));
  const [error, setError] = useState('');
  const logRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!sessionId) { setLoading(false); return; }
    let live = true;
    getChatHistory(sessionId)
      .then(h => { if (live) setMessages(h.messages); })
      .catch(() => { if (live) { save(key); setSessionId(undefined); } })
      .finally(() => { if (live) setLoading(false); });
    return () => { live = false; };
    // Only on mount / chart change: later messages arrive through ask().
  }, [key]);

  // Keep the newest message in view by scrolling the log itself, not the page.
  useEffect(() => { const el = logRef.current; if (el) el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' }); }, [messages, pending]);

  async function ask(q: string) {
    const question = q.trim();
    if (!question || pending) return;
    setDraft(''); setError(''); setPending(true);
    const optimistic: ChatMessage = { id: -Date.now(), role: 'user', content: question, topics: [], sources: [], created_at: new Date().toISOString() };
    setMessages(m => [...m, optimistic]);
    try {
      const r = await postChat(birth, question, sessionId);
      if (r.session_id !== sessionId) { setSessionId(r.session_id); save(key, r.session_id); }
      setMessages(m => [...m.filter(x => x.id !== optimistic.id), r.question, r.answer]);
    } catch (e) {
      setMessages(m => m.filter(x => x.id !== optimistic.id));
      setDraft(question);
      setError(e instanceof Error ? e.message : 'Could not reach Antariksha');
    } finally {
      setPending(false);
    }
  }

  function reset() { save(key); setSessionId(undefined); setMessages([]); setError(''); }
  const questions = messages.filter(m => m.role === 'user' && m.id > 0);

  return (
    <div className="chat">
      <aside className="card chat-history" aria-label="Your questions">
        <div className="chat-history-head">
          <h3>Your questions <span className="muted">{questions.length}</span></h3>
          {messages.length > 0 && <button className="ghost small" onClick={reset}>New conversation</button>}
        </div>
        {questions.length === 0 ? <p className="muted small">Questions you ask are saved here with their answers, for this chart on this device.</p> : (
          <ol>{questions.map(q => (
            <li key={q.id}><button onClick={() => { const el = document.getElementById('msg-' + q.id), log = logRef.current; if (el && log) log.scrollTo({ top: el.offsetTop - log.offsetTop - 8, behavior: 'smooth' }); }}>{q.content}</button>
              <time>{new Date(q.created_at).toLocaleString(undefined, { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })}</time></li>
          ))}</ol>
        )}
      </aside>
      <section className="card chat-main" aria-label="Ask Antariksha">
        <div className="chat-log" ref={logRef} aria-live="polite">
          {loading && <p className="muted">Loading your conversation…</p>}
          {!loading && messages.length === 0 && (
            <div className="chat-welcome">
              <p className="kicker">Ask Antariksha</p>
              <h2>Questions about {name ? `${name}'s` : 'your'} chart</h2>
              <p className="muted">Answers come from your computed chart and the Antariksha corpus of classical and interpretive texts. They describe tendencies for reflection, not certainties.</p>
              <div className="chips">{suggestions.map(s => <button key={s} className="chip" onClick={() => ask(s)}>{s}</button>)}</div>
            </div>
          )}
          {messages.map(m => (
            <div key={m.id} id={'msg-' + m.id} className={'msg msg-' + m.role}>
              <div className="msg-bubble">{m.role === 'assistant' ? <Answer m={m} /> : <p>{m.content}</p>}</div>
            </div>
          ))}
          {pending && <div className="msg msg-assistant"><div className="msg-bubble typing" aria-label="Antariksha is answering"><i /><i /><i /></div></div>}
        </div>
        {messages.length > 0 && !pending && (
          <div className="chips chips-inline">{suggestions.filter(s => !messages.some(m => m.content === s)).slice(0, 3).map(s => <button key={s} className="chip" onClick={() => ask(s)}>{s}</button>)}</div>
        )}
        {error && <p role="alert" className="form-error">{error}</p>}
        <form className="composer" onSubmit={e => { e.preventDefault(); ask(draft); }}>
          <label className="sr-only" htmlFor="chat-input">Your question</label>
          <textarea id="chat-input" rows={1} value={draft} maxLength={600} placeholder="Ask about career, marriage, a planet, your dasha…"
            onChange={e => setDraft(e.target.value)}
            onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); ask(draft); } }} />
          <button className="primary" disabled={pending || !draft.trim()}>Ask</button>
        </form>
      </section>
    </div>
  );
}
