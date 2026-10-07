import type { ChartInput, ChartResponse, ChatAnswer, ChatMessage, MatchResponse, MuhurtaEvent, MuhurtaResponse, PlaceHit, ReadingResponse, ShadbalaResponse, TodayResponse, TransitsResponse, VargaResponse } from './types';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const r = await fetch(path, init);
  // A cancelled request (the user changed the date mid-load) must reject, not
  // resolve with a placeholder: rendering that placeholder crashed the page.
  let body: any;
  try { body = await r.json(); } catch (e) {
    if (e instanceof DOMException && e.name === 'AbortError') throw e;
    if (r.ok) throw new Error('The server sent an unreadable response. Please try again.');
    body = { error: r.statusText };
  }
  if(r.status===401)window.dispatchEvent(new Event('antariksha-signed-out'));
  if (!r.ok) throw new Error(body.error || `Request failed (${r.status})`);
  return body as T;
}

const query = (input: ChartInput, extra: Record<string, string> = {}) =>
  new URLSearchParams({ date: input.date, time: input.time, lat: String(input.lat), lon: String(input.lon), tz: input.tz, ...extra });

export const getChart = (input: ChartInput, signal?: AbortSignal) =>
  request<ChartResponse>(`/api/chart?${query(input, { ayanamsa: 'lahiri' })}`, { signal });

export const getReading = (input: ChartInput, signal?: AbortSignal) =>
  request<ReadingResponse>(`/api/reading?${query(input)}`, { signal });

export const fetchJSON = <T,>(path: string, signal?: AbortSignal) => request<T>(path, { signal });

export const postChat = (birth: ChartInput, question: string, sessionId?: string, lang = 'en') =>
  request<{ session_id: string; question: ChatMessage; answer: ChatMessage }>('/api/chat', {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ birth, question, session_id: sessionId ?? '', lang }),
  });

export const getChatHistory = (sessionId: string) =>
  request<{ session_id: string; messages: ChatMessage[] }>(`/api/chat/history?session_id=${encodeURIComponent(sessionId)}`);

export const deleteChatSession = (sessionId: string) =>
  request<{ deleted: string }>(`/api/chat/session?session_id=${encodeURIComponent(sessionId)}`, { method: 'DELETE' });

export const searchPlaces = (q: string, signal?: AbortSignal) =>
  request<{ places: PlaceHit[]; attribution: string }>(`/api/places?q=${encodeURIComponent(q)}&limit=8`, { signal });

export const getVarga = (input: ChartInput, n: number, signal?: AbortSignal) =>
  request<VargaResponse>(`/api/chart/varga?${query(input, { n: String(n) })}`, { signal });

export const postMatch = (boy: ChartInput, girl: ChartInput, useAI=false) =>
  request<MatchResponse>('/api/match', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ boy, girl, use_ai:useAI }) });

export const postMatchChat = (boy: ChartInput, girl: ChartInput, question: string, history: { role: 'user' | 'assistant'; content: string }[], lang = 'en') =>
  request<ChatAnswer>('/api/match/chat', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ boy, girl, question, history, lang }) });

export const postProfileChat = (peer: string, question: string, history: { role: 'user' | 'assistant'; content: string }[], lang = 'en') =>
  request<ChatAnswer>(`/api/matrimony/compare/${encodeURIComponent(peer)}/chat`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ question, history, lang }) });

export const getMuhurtaEvents = () => request<MuhurtaEvent[]>('/api/muhurta/events');

export function getMuhurta(event: string, from: string, days: number, place: { lat: number; lon: number; tz: string }, birth?: ChartInput, signal?: AbortSignal) {
  const q = new URLSearchParams({ event, date: from, days: String(days), lat: String(place.lat), lon: String(place.lon), tz: place.tz });
  if (birth) { q.set('bdate', birth.date); q.set('btime', birth.time); q.set('blat', String(birth.lat)); q.set('blon', String(birth.lon)); q.set('btz', birth.tz); }
  return request<MuhurtaResponse>(`/api/muhurta?${q}`, { signal });
}

export const getToday = (input: ChartInput, signal?: AbortSignal) => request<TodayResponse>(`/api/today?${query(input)}`, { signal });

/** Today's transits read against the birth chart, the coming sign changes and eclipses, and forecast periods. */
export const getTransits = (input: ChartInput, months = 24, signal?: AbortSignal, past = 0) =>
  request<TransitsResponse>(`/api/transits?${query(input, { months: String(months), ...(past ? { past: String(past) } : {}) })}`, { signal });

export const calendarURL = (year: number, place: { lat: number; lon: number; tz: string }, vrat = true) =>
  `/api/calendar.ics?${new URLSearchParams({ year: String(year), lat: String(place.lat), lon: String(place.lon), tz: place.tz, vrat: vrat ? '1' : '0' })}`;

export const getShadbala = (input: ChartInput, signal?: AbortSignal) =>
  request<ShadbalaResponse>(`/api/chart/shadbala?${query(input)}`, { signal }).then(r => ({ ...r, rows: r.rows.map(x => ({ ...x, required: x.required_rupas })) }));
