const english: Record<string, string> = { sun: 'Sun', moon: 'Moon', mars: 'Mars', mercury: 'Mercury', jupiter: 'Jupiter', venus: 'Venus', saturn: 'Saturn', rahu: 'Rahu', ketu: 'Ketu' };
export const grahaEnglish = (id: string) => english[id] ?? id;
export const grahaShort: Record<string, string> = { sun: 'Su', moon: 'Mo', mars: 'Ma', mercury: 'Me', jupiter: 'Ju', venus: 'Ve', saturn: 'Sa', rahu: 'Ra', ketu: 'Ke' };

/** 2026-09-22 → "22 Sep 2026" */
export function longDate(iso: string) {
  const [y, m, d] = iso.split('-').map(Number);
  if (!y || !m || !d) return iso;
  return new Date(Date.UTC(y, m - 1, d)).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric', timeZone: 'UTC' });
}

export const titleCase = (s: string) => s.replace(/(^|[\s_-])(\w)/g, (_, p, c) => (p === '_' ? ' ' : p) + c.toUpperCase());
