import { useState, type ReactNode } from 'react';
export async function memberAPI<T = any>(path: string, method = 'GET', body?: unknown): Promise<T> {
  const form = body instanceof FormData;
  const response = await fetch(path, { method, credentials: 'same-origin', headers: form ? {} : { 'Content-Type': 'application/json' }, body: body === undefined ? undefined : form ? body : JSON.stringify(body) });
  let result: any;
  try { result = await response.json(); } catch {
    if (response.ok) throw new Error('The server sent an unreadable response. Please try again.');
    result = { error: 'The server could not complete this request.' };
  }
  if(response.status===401)window.dispatchEvent(new Event('antariksha-signed-out'));
  if (!response.ok) throw new Error(result.error || 'Request failed');
  return result;
}
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
