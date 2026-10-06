import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';

// Guard rails from the frontend audit: they fail the test run when a change
// brings back the patterns that made the old UI inconsistent.
const root = __dirname;
const files = (dir: string): string[] => readdirSync(dir).flatMap(f => {
  const p = join(dir, f);
  return statSync(p).isDirectory() ? files(p) : [p];
});
const all = files(root).filter(f => /\.(tsx?|css)$/.test(f) && !f.endsWith('.test.ts'));
const rel = (f: string) => f.slice(root.length + 1);

describe('UI rules', () => {
  it('never uses browser pop-ups (use useFeedback().confirm / prompt / toast)', () => {
    const bad = all.filter(f => f.endsWith('.tsx')).filter(f => /(^|[^.\w])(prompt|confirm|alert)\((?!\s*\{)/m.test(readFileSync(f, 'utf8').replace(/\/\/.*$/gm, '')));
    expect(bad.map(rel)).toEqual([]);
  });
  it('features use design-system controls, not raw buttons or inline styles', () => {
    const feature = all.filter(f => rel(f).startsWith('features/') && !rel(f).startsWith('features/shell/') && f.endsWith('.tsx'));
    const bad = feature.filter(f => { const s = readFileSync(f, 'utf8'); return /<button[\s>]/.test(s) || /style=\{\{/.test(s); });
    expect(bad.map(rel)).toEqual([]);
  });
  it('colours in new code come from tokens.css', () => {
    const bad = all.filter(f => (rel(f).startsWith('features/') || rel(f).startsWith('ds/')) && f.endsWith('.css') && !f.endsWith('tokens.css'))
      .filter(f => /#[0-9a-fA-F]{3,8}\b/.test(readFileSync(f, 'utf8').replace(/#[0-9a-fA-F]{6}[0-9a-fA-F]{2}\b/g, '')));
    expect(bad.map(rel)).toEqual([]);
  });
});
