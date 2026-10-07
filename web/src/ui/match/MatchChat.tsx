import { useEffect, useRef, useState } from 'react';
import type { ChatAnswer, Source } from '../../api/types';
import { useSettings } from '../../i18n';
import { AnswerText } from '../chat/AnswerText';

const suggestions = [
  'Explain our guna score',
  'Do we have any doshas, and are they cancelled?',
  'What does numerology say about us?',
  'Explain Nadi and Bhakoot',
  'Do either of us have Mangal dosha?',
  'Compare our Moon signs and nakshatras',
  'How are our 7th houses and Venus?',
  'Which dashas are we each running?',
];

type Turn = { id: number; role: 'user' | 'assistant'; content: string; original?: string; sources?: Source[]; model?: string };
export type History = { role: 'user' | 'assistant'; content: string }[];
export type AskFn = (question: string, history: History, lang: string) => Promise<ChatAnswer>;

/** Ask Astrisk about a compared pair. Like the match itself, the conversation
 *  lives only in this page: nothing is saved on the server. */
export function MatchChat({ ask: send, intro }: { ask: AskFn; intro?: string }) {
  const { lang } = useSettings();
  const [turns, setTurns] = useState<Turn[]>([]);
  const [draft, setDraft] = useState('');
  const [pending, setPending] = useState(false);
  const [error, setError] = useState('');
  const logRef = useRef<HTMLDivElement>(null);

  useEffect(() => { const el = logRef.current; if (el) el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' }); }, [turns, pending]);

  async function ask(q: string) {
    const question = q.trim();
    if (!question || pending) return;
    setDraft(''); setError(''); setPending(true);
    const history = turns.slice(-6).map(({ role, content, original }) => ({ role, content: original ?? content }));
    setTurns(t => [...t, { id: Date.now(), role: 'user', content: question }]);
    try {
      const a = await send(question, history, lang);
      setTurns(t => [...t, { id: Date.now() + 1, role: 'assistant', content: a.answer, original: a.original, sources: a.sources, model: a.model }]);
    } catch (e) {
      setTurns(t => t.slice(0, -1));
      setDraft(question);
      setError(e instanceof Error ? e.message : 'Could not reach Astrisk');
    } finally {
      setPending(false);
    }
  }

  return (
    <section className="card chat-main match-chat" aria-label="Ask Astrisk about this match">
      <div className="chat-log" ref={logRef} aria-live="polite">
        {turns.length === 0 && (
          <div className="chat-welcome">
            <p className="kicker">Ask Astrisk</p>
            <h2>Questions about this match</h2>
            <p className="muted">{intro ?? "Answers use both computed charts, the eight koota tables and both people's numerology. This conversation is not saved; it disappears when you leave or change the details."}{lang !== 'en' ? ' Answers are machine-translated from English when the translation service is running; technical chart details stay in English.' : ''}</p>
            <div className="chips">{suggestions.map(s => <button key={s} type="button" className="chip" onClick={() => ask(s)}>{s}</button>)}</div>
          </div>
        )}
        {turns.map(m => (
          <div key={m.id} className={'msg msg-' + m.role}>
            <div className="msg-bubble">
              {m.role === 'assistant' ? <AnswerText text={m.content} original={m.original} /> : <p>{m.content}</p>}
              {m.sources && m.sources.length > 0 && (
                <details className="sources">
                  <summary>{m.sources.length} source{m.sources.length > 1 ? 's' : ''} · {m.model === 'grounded-corpus' ? 'grounded answer' : m.model}</summary>
                  <ul>{m.sources.map(s => <li key={s.doc_type + s.key + s.source}><b>{s.title}</b><span>{s.source.replace(/ \[.*?\]$/, '')}</span></li>)}</ul>
                </details>
              )}
            </div>
          </div>
        ))}
        {pending && <div className="msg msg-assistant"><div className="msg-bubble typing" aria-label="Astrisk is answering"><i /><i /><i /></div></div>}
      </div>
      {turns.length > 0 && !pending && (
        <div className="chips chips-inline">{suggestions.filter(s => !turns.some(m => m.content === s)).slice(0, 3).map(s => <button key={s} type="button" className="chip" onClick={() => ask(s)}>{s}</button>)}</div>
      )}
      {error && <p role="alert" className="form-error">{error}</p>}
      <form className="composer" onSubmit={e => { e.preventDefault(); ask(draft); }}>
        <label className="sr-only" htmlFor="match-chat-input">Your question about this match</label>
        <textarea id="match-chat-input" rows={1} value={draft} maxLength={600} placeholder="Ask about the score, a koota, doshas, numerology…"
          onChange={e => setDraft(e.target.value)}
          onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); ask(draft); } }} />
        <button className="primary" disabled={pending || !draft.trim()}>Ask</button>
      </form>
    </section>
  );
}
