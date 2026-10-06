import { useState, type ReactNode } from 'react';
import { api } from '../../lib/api';
export const memberAPI = api;
export function useAction() {
  const [busy, setBusy] = useState(false), [message, setMessage] = useState('');
  async function run(fn: () => Promise<void>) { setBusy(true); setMessage(''); try { await fn(); } catch (e) { setMessage(e instanceof Error ? e.message : 'Request failed'); } finally { setBusy(false); } }
  return { busy, message, setMessage, run };
}
export function Feedback({ busy, message }: { busy: boolean; message: string }) { return <p role="status" className="community-feedback">{busy ? 'Working…' : message}</p>; }
export function Field({ label, children }: { label: string; children: ReactNode }) { return <label className="field"><span>{label}</span>{children}</label>; }
export function Avatar({ kind = 'sun', accent = '#d6b467' }: { kind?: string; accent?: string }) {
  return <span className="member-avatar" style={{ color: accent }} aria-label={`${kind} avatar`}>{({ sun: '☀', moon: '☾', star: '✧', leaf: '❧', mountain: '△' } as Record<string, string>)[kind] || '☀'}</span>;
}
export type Family = { id: string; name: string; mine: boolean; accepted: boolean; owner_handle: string };
