import type { ChartInput, ChartResponse, ChatMessage, ReadingResponse } from './types';

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
