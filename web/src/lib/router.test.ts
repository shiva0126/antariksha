import { describe, expect, it } from 'vitest';
import { parse, pick } from './router';

describe('router', () => {
  it('parses sections and tabs', () => {
    expect(parse('#kundali/chart')).toEqual({ section: 'kundali', sub: 'chart', rest: [] });
    expect(parse('#matrimony/interests/abc')).toEqual({ section: 'matrimony', sub: 'interests', rest: ['abc'] });
  });
  it('redirects old addresses', () => {
    expect(parse('#day')).toMatchObject({ section: 'panchang', sub: 'today' });
    expect(parse('#month')).toMatchObject({ section: 'panchang', sub: 'calendar' });
    expect(parse('#account')).toMatchObject({ section: 'me', sub: 'profile' });
    expect(parse('#readings')).toMatchObject({ section: 'kundali', sub: 'systems' });
    expect(parse('#match')).toMatchObject({ section: 'matching', sub: '' });
  });
  it('falls back to kundali and the first tab', () => {
    expect(parse('#nope').section).toBe('kundali');
    expect(pick('x', ['a', 'b'] as const)).toBe('a');
  });
});
