import type { ChartInput, ChartResponse, ChatMessage, MatchResponse, MuhurtaEvent, MuhurtaResponse, PlaceHit, ReadingResponse, TodayResponse, VargaResponse } from './types';

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const r = await fetch(path, init);
  const body = await r.json().catch(() => ({ error: r.statusText }));
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

export const postChat = (birth: ChartInput, question: string, sessionId?: string) =>
  request<{ session_id: string; question: ChatMessage; answer: ChatMessage }>('/api/chat', {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ birth, question, session_id: sessionId ?? '' }),
  });

export const getChatHistory = (sessionId: string) =>
  request<{ session_id: string; messages: ChatMessage[] }>(`/api/chat/history?session_id=${encodeURIComponent(sessionId)}`);

export const deleteChatSession = (sessionId: string) =>
  request<{ deleted: string }>(`/api/chat/session?session_id=${encodeURIComponent(sessionId)}`, { method: 'DELETE' });

export const searchPlaces = (q: string, signal?: AbortSignal) =>
  request<{ places: PlaceHit[]; attribution: string }>(`/api/places?q=${encodeURIComponent(q)}&limit=8`, { signal });

export const getVarga = (input: ChartInput, n: number, signal?: AbortSignal) =>
  request<VargaResponse>(`/api/chart/varga?${query(input, { n: String(n) })}`, { signal });

export const postMatch = (boy: ChartInput, girl: ChartInput) =>
  request<MatchResponse>('/api/match', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ boy, girl }) });

export const getMuhurtaEvents = () => request<MuhurtaEvent[]>('/api/muhurta/events');

export function getMuhurta(event: string, from: string, days: number, place: { lat: number; lon: number; tz: string }, birth?: ChartInput, signal?: AbortSignal) {
  const q = new URLSearchParams({ event, date: from, days: String(days), lat: String(place.lat), lon: String(place.lon), tz: place.tz });
  if (birth) { q.set('bdate', birth.date); q.set('btime', birth.time); q.set('blat', String(birth.lat)); q.set('blon', String(birth.lon)); q.set('btz', birth.tz); }
  return request<MuhurtaResponse>(`/api/muhurta?${q}`, { signal });
}

export const getToday = (input: ChartInput, signal?: AbortSignal) => request<TodayResponse>(`/api/today?${query(input)}`, { signal });

export const calendarURL = (year: number, place: { lat: number; lon: number; tz: string }, vrat = true) =>
  `/api/calendar.ics?${new URLSearchParams({ year: String(year), lat: String(place.lat), lon: String(place.lon), tz: place.tz, vrat: vrat ? '1' : '0' })}`;
