import { useEffect, useState } from 'react';

/** Every page and tab has an address `#section/sub`. Old addresses from
 *  before the redesign redirect to their new homes. */
export type Section = 'kundali' | 'panchang' | 'matching' | 'matrimony' | 'community' | 'me' | 'admin' | 'biodata';
export type Route = { section: Section; sub: string; rest: string[] };

export const SECTIONS: Section[] = ['kundali', 'panchang', 'matching', 'matrimony', 'community', 'me', 'admin', 'biodata'];

const LEGACY: Record<string, string> = {
  day: 'panchang/today', month: 'panchang/calendar', muhurta: 'panchang/muhurta', match: 'matching',
  readings: 'kundali/systems', account: 'me/profile', privacy: 'me/privacy',
};

export function parse(hash: string): Route {
  let h = hash.replace(/^#/, '');
  const [head, ...tail] = h.split('/');
  if (LEGACY[head]) h = [LEGACY[head], ...tail].join('/');
  const [section, sub = '', ...rest] = h.split('/');
  return SECTIONS.includes(section as Section) ? { section: section as Section, sub, rest } : { section: 'kundali', sub: '', rest: [] };
}

export function href(section: Section, sub?: string) { return '#' + section + (sub ? '/' + sub : ''); }

export function go(section: Section, sub?: string) {
  const next = href(section, sub);
  if (location.hash !== next) location.hash = next;
}

/** The current route; re-renders on every hash change and scrolls to the top
 *  when the section changes (not when a tab inside it changes). */
export function useRoute(): Route {
  const [route, setRoute] = useState(() => parse(location.hash));
  useEffect(() => {
    const legacy = parse(location.hash);
    const canonical = href(legacy.section, legacy.sub);
    if (location.hash && LEGACY[location.hash.slice(1).split('/')[0]]) history.replaceState(null, '', canonical + location.search);
    const changed = () => setRoute(prev => {
      const next = parse(location.hash);
      if (next.section !== prev.section) window.scrollTo({ top: 0 });
      return next;
    });
    window.addEventListener('hashchange', changed);
    return () => window.removeEventListener('hashchange', changed);
  }, []);
  return route;
}

/** A tab id from the route, falling back to the first allowed one. */
export const pick = <K extends string>(sub: string, allowed: readonly K[]): K => (allowed.includes(sub as K) ? sub as K : allowed[0]);
