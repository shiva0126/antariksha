import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';

import { nameIn } from './names';
import { KEYS, UI } from './strings';

export type Lang = 'en' | 'hi' | 'mr' | 'kn' | 'ta' | 'te' | 'ml' | 'gu' | 'bn';
export const LANGUAGES: [Lang, string][] = [['en', 'English'], ['hi', 'हिन्दी'], ['mr', 'मराठी'], ['kn', 'ಕನ್ನಡ'], ['ta', 'தமிழ்'], ['te', 'తెలుగు'], ['ml', 'മലയാളം'], ['gu', 'ગુજરાતી'], ['bn', 'বাংলা']];
const keyIndex = new Map<string, number>(KEYS.map((k, i) => [k, i]));

// Script fonts, loaded only for the selected language so English users pay nothing.
const scriptFont: Partial<Record<Lang, string>> = {
  hi: 'Noto Sans Devanagari', mr: 'Noto Sans Devanagari', kn: 'Noto Sans Kannada', ta: 'Noto Sans Tamil',
  te: 'Noto Sans Telugu', ml: 'Noto Sans Malayalam', gu: 'Noto Sans Gujarati', bn: 'Noto Sans Bengali',
};

function loadScriptFont(lang: Lang) {
  const family = scriptFont[lang];
  document.documentElement.style.setProperty('--script-font', family ? `'${family}'` : 'system-ui');
  if (!family || document.querySelector(`link[data-script-font="${lang}"]`)) return;
  const link = document.createElement('link');
  link.rel = 'stylesheet';
  link.dataset.scriptFont = lang;
  link.href = `https://fonts.googleapis.com/css2?family=${family.replace(/ /g, '+')}:wght@400;500;600&display=swap`;
  document.head.appendChild(link);
}
export type MonthSystem = 'amanta' | 'purnimanta';

export const MONTHS = ['Chaitra', 'Vaishakha', 'Jyeshtha', 'Ashadha', 'Shravana', 'Bhadrapada', 'Ashwina', 'Kartika', 'Margashirsha', 'Pausha', 'Magha', 'Phalguna'];

interface Settings { lang: Lang; months: MonthSystem; setLang: (l: Lang) => void; setMonths: (m: MonthSystem) => void }
const Ctx = createContext<Settings>({ lang: 'en', months: 'amanta', setLang: () => {}, setMonths: () => {} });

const read = <T extends string>(k: string, d: T): T => { try { return (localStorage.getItem(k) as T) || d; } catch { return d; } };

export function SettingsProvider({ children }: { children: ReactNode }) {
  const [lang, setLang] = useState<Lang>(() => { const l = read<Lang>('antariksha.lang', 'en'); return LANGUAGES.some(([k]) => k === l) ? l : 'en'; });
  const [months, setMonths] = useState<MonthSystem>(() => read('antariksha.months', 'amanta'));
  useEffect(() => { try { localStorage.setItem('antariksha.lang', lang); localStorage.setItem('antariksha.months', months); } catch { /* ignore */ } document.documentElement.lang = lang; loadScriptFont(lang); }, [lang, months]);
  return <Ctx.Provider value={{ lang, months, setLang, setMonths }}>{children}</Ctx.Provider>;
}

export function useSettings() { return useContext(Ctx); }

/** t(): UI string in the chosen language (English is the key). */
export function useT() {
  const { lang } = useContext(Ctx);
  return (s: string) => {
    const i = keyIndex.get(s);
    return i === undefined ? s : UI[lang]?.[i] ?? s;
  };
}

/** n(): a Sanskrit term (rashi, nakshatra, tithi, graha, …) in the chosen
 * language's script; compound names ("Adhika Bhadrapada") word by word. */
export function useNames() {
  const { lang } = useContext(Ctx);
  return (s: string) => {
    if (lang === 'en' || !s) return s;
    const whole = nameIn(lang, s);
    if (whole) return whole;
    return s.split(' ').map(w => nameIn(lang, w) ?? w).join(' ');
  };
}

/** Month name in the chosen system. Amanta months end at new moon; in the
 * Purnimanta system the dark fortnight belongs to the following month. */
export function useMonthName() {
  const { months } = useContext(Ctx);
  return (amanta: string, paksha: string) => {
    if (months === 'amanta' || paksha !== 'Krishna') return amanta;
    const adhika = amanta.startsWith('Adhika ');
    const base = adhika ? amanta.slice(7) : amanta;
    const i = MONTHS.indexOf(base);
    if (i < 0) return amanta;
    return (adhika ? 'Adhika ' : '') + MONTHS[(i + 1) % 12];
  };
}
