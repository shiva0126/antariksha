import { describe, expect, it } from 'vitest';
import { nameIn, transliterate } from './names';
import { KEYS, UI } from './strings';

describe('Indic names', () => {
  it('transliterates Sanskrit terms to standard spellings', () => {
    expect(transliterate('मेष', 'kn')).toBe('ಮೇಷ');
    expect(transliterate('अश्विनी', 'te')).toBe('అశ్వినీ'); // exact Sanskrit long ī
    expect(transliterate('रोहिणी', 'ml')).toBe('രോഹിണീ');
    expect(transliterate('कृत्तिका', 'gu')).toBe('કૃત્તિકા');
    expect(transliterate('विशाखा', 'bn')).toBe('বিশাখা'); // va → ba in Bengali
    expect(transliterate('पूर्णिमा', 'kn')).toBe('ಪೂರ್ಣಿಮಾ');
    expect(transliterate('सिंह', 'te')).toBe('సింహ');
  });
  it('uses regional names and the Tamil table', () => {
    expect(nameIn('ta', 'Ardra')).toBe('திருவாதிரை');
    expect(nameIn('ta', 'Purnima')).toBe('பௌர்ணமி');
    expect(nameIn('kn', 'Ravivara')).toBe('ಭಾನುವಾರ');
    expect(nameIn('ml', 'Somavara')).toBe('തിങ്കളാഴ്ച');
    expect(nameIn('mr', 'Mangalavara')).toBe('मंगळवार');
    expect(nameIn('hi', 'Karka')).toBe('कर्क');
    expect(nameIn('en', 'Karka')).toBeUndefined();
  });
  it('covers every term in every language', () => {
    const terms = ['Mesha', 'Revati', 'Amavasya', 'Vaidhriti', 'Kimstughna', 'Shanivara', 'Phalguna', 'Ketu', 'Udveg'];
    for (const lang of ['hi', 'mr', 'kn', 'ta', 'te', 'ml', 'gu', 'bn']) {
      for (const t of terms) expect(nameIn(lang, t), `${lang}:${t}`).toBeTruthy();
    }
  });
  it('translates every UI key in every language', () => {
    for (const [lang, list] of Object.entries(UI)) expect(list.length, lang).toBe(KEYS.length);
  });
});
