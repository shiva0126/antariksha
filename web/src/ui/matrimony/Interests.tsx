import { useEffect, useRef, useState } from 'react';
import { useT } from '../../i18n';
import { useFeedback } from '../../ds';
import { Avatar, memberAPI } from '../community/shared';
import { displayName, type Interest } from './types';

type Message = { id: number; body: string; mine: boolean; created_at: string };
const POLL_MS = 10_000;

/** Chat with an accepted match. New messages arrive by polling every 10 s
 *  while the page is visible; contact details are shared only on request. */
function Conversation({ peer, onClose, onChanged }: { peer: Interest; onClose: () => void; onChanged: () => void }) {
  const t = useT();
  const fb = useFeedback();
  const [messages, setMessages] = useState<Message[]>([]), [body, setBody] = useState(''), [error, setError] = useState(''), [busy, setBusy] = useState(false);
  const [contacts, setContacts] = useState<{ mine: string; theirs: string }>({ mine: '', theirs: '' }), [contact, setContact] = useState(''), [sharing, setSharing] = useState(false);
  const log = useRef<HTMLDivElement>(null);
  const name = displayName(peer);
  async function load() {
    const [m, c] = await Promise.all([memberAPI<Message[]>(`/api/matrimony/messages/${peer.id}`), memberAPI<{ mine: string; theirs: string }>(`/api/matrimony/contact/${peer.id}`)]);
    setMessages(m); setContacts(c);
  }
  useEffect(() => {
    void load().catch(e => setError(e.message));
    const timer = setInterval(() => { if (document.visibilityState === 'visible') void load().catch(() => {}); }, POLL_MS);
    return () => clearInterval(timer);
  }, [peer.id]);
  useEffect(() => { log.current?.scrollTo({ top: log.current.scrollHeight }); }, [messages.length]);
  async function act(fn: () => Promise<void>) { setBusy(true); setError(''); try { await fn(); } catch (e) { setError(e instanceof Error ? e.message : 'Request failed'); } finally { setBusy(false); } }
  return (
    <section className="card mat-conversation" aria-label={`Conversation with ${name}`}>
      <header className="mat-conv-head">
        <div><h3>{name}</h3><p className="muted small">Messages refresh automatically. Never send money or share bank details.</p></div>
        <button className="ghost small" onClick={onClose}>✕</button>
      </header>
      <div className="member-messages mat-messages" ref={log} aria-live="polite">
        {messages.length === 0 && <p className="muted">Say hello. A good start: something from their biodata you would like to know more about.</p>}
        {messages.map(m => <p key={m.id} className={m.mine ? 'mine' : ''}><span>{m.body}</span><time>{new Date(m.created_at).toLocaleString(undefined, { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })}</time></p>)}
      </div>
      <form className="composer" onSubmit={e => { e.preventDefault(); if (!body.trim()) return; void act(async () => { await memberAPI(`/api/matrimony/messages/${peer.id}`, 'POST', { body }); setBody(''); await load(); }); }}>
        <label className="sr-only" htmlFor="mat-msg">Message</label>
        <textarea id="mat-msg" rows={1} required maxLength={2000} value={body} onChange={e => setBody(e.target.value)} placeholder={`Message ${name}`} />
        <button className="primary" disabled={busy || !body.trim()}>{t('Send')}</button>
      </form>
      <div className="mat-contact">
        {contacts.theirs ? <p><b>{name} shared:</b> {contacts.theirs}</p> : <p className="muted small">{name} has not shared contact details yet.</p>}
        {contacts.mine ? (
          <p className="small">You shared: {contacts.mine} <button className="link-button" onClick={() => void act(async () => { await memberAPI(`/api/matrimony/contact/${peer.id}`, 'DELETE'); await load(); })}>Stop sharing</button></p>
        ) : sharing ? (
          <form className="field-row" onSubmit={e => { e.preventDefault(); void act(async () => { await memberAPI(`/api/matrimony/contact/${peer.id}`, 'PUT', { contact }); setSharing(false); await load(); }); }}>
            <label className="field"><span>Phone or email to share with {name} only</span><input required minLength={3} maxLength={200} value={contact} onChange={e => setContact(e.target.value)} /></label>
            <button className="primary" disabled={busy}>{t('Share my contact')}</button>
          </form>
        ) : <button className="ghost" onClick={() => setSharing(true)}>{t('Share my contact')}</button>}
      </div>
      <div className="mat-actions">
        <button className="ghost" disabled={busy} onClick={() => void act(async () => { await memberAPI('/api/matrimony/interests', 'POST', { target: peer.id, action: 'unmatch' }); onChanged(); onClose(); })}>Unmatch</button>
        <button className="ghost danger-text" disabled={busy} onClick={async () => { if (await fb.confirm({ title: `Block ${name}? You will no longer see each other or be able to message.` })) void act(async () => { await memberAPI('/api/community/blocks', 'POST', { target: peer.id, block: true }); onChanged(); onClose(); }); }}>Block</button>
      </div>
      {error && <p role="alert" className="form-error">{error}</p>}
    </section>
  );
}

export function Interests({ open }: { open?: string }) {
  const t = useT();
  const [items, setItems] = useState<Interest[]>([]), [peer, setPeer] = useState<Interest>(), [error, setError] = useState(''), [loading, setLoading] = useState(true);
  async function load() { try { const rows = await memberAPI<Interest[]>('/api/matrimony/interests'); setItems(rows); if (open) setPeer(rows.find(r => r.id === open && r.status === 'accepted')); } catch (e) { setError(e instanceof Error ? e.message : 'Request failed'); } finally { setLoading(false); } }
  useEffect(() => { void load(); const timer = setInterval(() => { if (document.visibilityState === 'visible') void load(); }, 30_000); return () => clearInterval(timer); }, []);
  const act = (target: string, action: string) => memberAPI('/api/matrimony/interests', 'POST', { target, action }).then(load).catch(e => setError(e.message));
  const received = items.filter(i => !i.outgoing && i.status === 'pending'), sent = items.filter(i => i.outgoing && i.status === 'pending'), connected = items.filter(i => i.status === 'accepted');
  const row = (i: Interest, actions: React.ReactNode) => (
    <li key={i.id + i.status} className="mat-interest">
      <Avatar kind={i.avatar} accent={i.accent} />
      <div><b>{displayName(i)}</b>{i.note && <p className="mat-note">“{i.note}”</p>}<p className="muted small">{new Date(i.last_message || i.created_at).toLocaleDateString()}</p></div>
      <div className="mat-actions">{actions}</div>
    </li>
  );
  return (
    <div className="mat-interests">
      {error && <p role="alert" className="form-error">{error}</p>}
      {loading && <p className="muted">Loading…</p>}
      <section className="card"><h3>{t('Received')} <span className="count">{received.length}</span></h3>
        {received.length === 0 ? <p className="muted">No new interests. Complete your profile and turn on horoscope matching to appear with reasons on more cards.</p> :
          <ul className="mat-list">{received.map(i => row(i, <><button className="primary" onClick={() => void act(i.id, 'accept')}>{t('Accept')}</button><button className="ghost" onClick={() => void act(i.id, 'decline')}>{t('Decline')}</button></>))}</ul>}
      </section>
      <section className="card"><h3>{t('Connected')} <span className="count">{connected.length}</span></h3>
        {connected.length === 0 ? <p className="muted">When someone accepts your interest, or you accept theirs, you can chat here.</p> :
          <ul className="mat-list">{connected.map(i => row(i, <button className="primary" onClick={() => setPeer(i)}>{t('Open chat')}</button>))}</ul>}
      </section>
      {peer && <Conversation key={peer.id} peer={peer} onClose={() => setPeer(undefined)} onChanged={() => void load()} />}
      <section className="card"><h3>{t('Sent')} <span className="count">{sent.length}</span></h3>
        {sent.length === 0 ? <p className="muted">Interests you send appear here until they are answered.</p> : <ul className="mat-list">{sent.map(i => row(i, <span className="tag">Waiting</span>))}</ul>}
      </section>
    </div>
  );
}
