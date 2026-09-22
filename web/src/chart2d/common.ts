import { useEffect, useState } from 'react';
import type { ChartResponse, Graha } from '../api/types';
import { grahaShort } from '../astro/format';
export interface ChartProps { chart: ChartResponse; onSelect?: (g: Graha) => void; selected?: string; title?: string }
export const deg = (g: Graha) => `${Math.floor(g.rashi_degree)}°${String(Math.floor((g.rashi_degree % 1) * 60)).padStart(2, '0')}′`;
export const keyAct = (fn: () => void) => (e: React.KeyboardEvent) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); fn(); } };

/** True on narrow screens, where the chart renders ~330px wide and needs
 * short, larger labels to stay legible. */
export function useCompact(query = '(max-width: 560px)') {
  const get = () => typeof matchMedia === 'function' && matchMedia(query).matches;
  const [compact, setCompact] = useState(get);
  useEffect(() => {
    if (typeof matchMedia !== 'function') return;
    const m = matchMedia(query), on = () => setCompact(m.matches);
    m.addEventListener('change', on);
    return () => m.removeEventListener('change', on);
  }, [query]);
  return compact;
}

/** Planet label: full name and minutes on wide screens, "Su 29°" when compact. */
export const label = (g: Graha, compact: boolean) =>
  compact ? `${grahaShort[g.id]} ${Math.floor(g.rashi_degree)}°${g.retrograde ? '℞' : ''}` : `${g.name} ${deg(g)}${g.retrograde ? ' ℞' : ''}`;
