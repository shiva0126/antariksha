/** The one HTTP client. Sends JSON (or FormData), keeps the session cookie,
 *  rejects on HTTP errors with the server's message, signals a signed-out
 *  session, and lets a cancelled request reject instead of resolving. */
// eslint-disable-next-line @typescript-eslint/no-explicit-any -- callers name T; untyped legacy calls get any
export async function api<T = any>(path: string, method = 'GET', body?: unknown, signal?: AbortSignal): Promise<T> {
  const form = body instanceof FormData;
  const r = await fetch(path, { method, signal, credentials: 'same-origin', headers: form || body === undefined ? {} : { 'Content-Type': 'application/json' }, body: body === undefined ? undefined : form ? body : JSON.stringify(body) });
  let data: any;
  try { data = await r.json(); } catch (e) {
    if (e instanceof DOMException && e.name === 'AbortError') throw e;
    if (r.ok) throw new Error('The server sent an unreadable response. Please try again.');
    data = { error: r.statusText };
  }
  if (r.status === 401) window.dispatchEvent(new Event('antariksha-signed-out'));
  if (!r.ok) throw new Error(data?.error || `Request failed (${r.status})`);
  return data as T;
}

/** Saves JSON as a file the user downloads. */
export function downloadJSON(name: string, data: unknown) {
  const url = URL.createObjectURL(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' }));
  const a = document.createElement('a'); a.href = url; a.download = name; a.click(); URL.revokeObjectURL(url);
}
