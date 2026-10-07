import { useEffect, useState } from 'react';
import { getTransits } from '../../api/client';
import type { AreaForecast, ChartInput, TransitEvent, TransitsResponse } from '../../api/types';
import { grahaEnglish, longDate } from '../../astro/format';
import { Button, Card, Chip, Notice, Skeleton } from '../../ds';
import { api } from '../../lib/api';

const ord = (n: number) => n + (n % 100 >= 11 && n % 100 <= 13 ? 'th' : n % 10 === 1 ? 'st' : n % 10 === 2 ? 'nd' : n % 10 === 3 ? 'rd' : 'th');
const the = (id: string) => (id === 'sun' || id === 'moon' ? 'the ' : '') + grahaEnglish(id);
const toneChip = { supportive: 'success', challenging: 'warning', mixed: 'neutral' } as const;
const toneWord = { supportive: 'Supportive', challenging: 'Asks for patience', mixed: 'Mixed' } as const;

function Area({ a }: { a: AreaForecast }) {
  return (
    <li className="forecast-area">
      <div className="forecast-area-head">
        <Chip tone={toneChip[a.tone]}>{toneWord[a.tone]}</Chip>
        <b>{cap(a.area.name)}</b>
        <span className="muted small">{a.confidence} agreement</span>
      </div>
      <details>
        <summary>Why</summary>
        <ul className="forecast-reasons">
          {a.reasons.map((r, i) => <li key={i}>{r.text} <span className="muted small">({r.source})</span></li>)}
        </ul>
      </details>
    </li>
  );
}

/** One line for a sign change, station or eclipse, in plain words. */
function eventText(e: TransitEvent): string {
  const who = the(e.graha);
  switch (e.kind) {
    case 'ingress':
      return `${who[0].toUpperCase() + who.slice(1)} moves into ${e.rashi}, your ${ord(e.from_moon ?? 0)} house from the Moon (${ord(e.from_lagna ?? 0)} from the lagna).`;
    case 'retrograde':
      return `${who[0].toUpperCase() + who.slice(1)} turns retrograde in ${e.rashi}: its themes slow down and turn inward.`;
    case 'direct':
      return `${who[0].toUpperCase() + who.slice(1)} turns direct in ${e.rashi}: its themes move forward again.`;
    default: {
      const kind = e.kind === 'solar_eclipse' ? 'Solar eclipse' : 'Lunar eclipse';
      const where = e.on_moon_sign ? ' in your Moon sign' : e.on_lagna_sign ? ' in your rising sign' : '';
      return `${kind} in ${e.rashi}${where}.`;
    }
  }
}

const slow = new Set(['jupiter', 'saturn', 'rahu', 'ketu']);

type Answer = 'yes' | 'partly' | 'no';

/** Past periods of the member's own chart: "did this feel true?" (opt-in). */
function LookingBack({ periods }: { periods: TransitsResponse['periods'] }) {
  const [answers, setAnswers] = useState<Record<string, Answer>>({});
  useEffect(() => {
    api<{ period_from: string; area: string; response: Answer }[]>('/api/me/forecast-feedback')
      .then(rows => setAnswers(Object.fromEntries(rows.map(r => [r.period_from + '|' + r.area, r.response])))).catch(() => undefined);
  }, []);
  const asked = periods.filter(p => p.areas.length > 0).slice(-4).reverse();
  if (asked.length === 0) return null;
  async function answer(p: TransitsResponse['periods'][number], response: Answer) {
    const a = p.areas[0], key = p.from.slice(0, 10) + '|' + a.area.id;
    await api('/api/me/forecast-feedback', 'PUT', { period_from: p.from.slice(0, 10), period_to: p.to.slice(0, 10), area: a.area.id, tone: a.tone, response });
    setAnswers(old => ({ ...old, [key]: response }));
  }
  return (
    <Card title="Looking back" sub="Optional: tell us whether recent periods felt as described. Only your answer, the period and the area are kept, to measure and improve these readings.">
      <ul className="forecast">
        {asked.map(p => {
          const a = p.areas[0], key = p.from.slice(0, 10) + '|' + a.area.id;
          return (
            <li key={key} className="forecast-period">
              <span className="muted small">{day(p.from)} to {day(p.to)}</span>
              <span>{cap(a.area.name)} was read as <b>{toneWord[a.tone].toLowerCase()}</b>. Did it feel that way?</span>
              <div className="ds-row">
                {(['yes', 'partly', 'no'] as const).map(r => (
                  <Button key={r} size="sm" variant={answers[key] === r ? 'primary' : 'ghost'} aria-pressed={answers[key] === r} onClick={() => void answer(p, r)}>{r === 'yes' ? 'Yes' : r === 'partly' ? 'Partly' : 'No'}</Button>
                ))}
              </div>
            </li>
          );
        })}
      </ul>
    </Card>
  );
}
const day = (iso: string) => longDate(iso.slice(0, 10));
const cap = (s: string) => s[0].toUpperCase() + s.slice(1);

export function TimingTab({ birth }: { birth: ChartInput }) {
  const [own, setOwn] = useState(false);
  useEffect(() => {
    api<{ birth_date?: string; birth_time?: string }>('/api/me').then(me => setOwn(me.birth_date === birth.date && (me.birth_time ?? '').slice(0, 5) === birth.time)).catch(() => setOwn(false));
  }, [birth]);
  const [data, setData] = useState<TransitsResponse>();
  const [error, setError] = useState('');
  const [allPeriods, setAllPeriods] = useState(false);
  const [allEvents, setAllEvents] = useState(false);
  useEffect(() => {
    const ctrl = new AbortController();
    setData(undefined); setError('');
    getTransits(birth, 24, ctrl.signal, 6).then(setData).catch(e => { if (e.name !== 'AbortError') setError(e.message); });
    return () => ctrl.abort();
  }, [birth]);
  if (error) return <Notice tone="danger">Timing unavailable: {error}</Notice>;
  if (!data) return <Card title="The coming months"><Skeleton lines={6} /></Card>;
  const events = data.events.filter(e => (slow.has(e.graha) && e.graha !== 'ketu') || e.kind.endsWith('eclipse'));
  return (
    <div className="timing">
      <Card title="The coming months" sub="Your running dasha read with the slow planets: Jupiter, Saturn, Rahu and Ketu">
        <p className="muted">A transit gives its results according to the dasha you are running (Brihat Samhita 104.46), so each period below reads the two together. It shows the areas of life that stand out and how strongly the factors agree. These are tendencies for reflection, not certain events.</p>
        <ol className="forecast">
          {(allPeriods ? data.periods : data.periods.slice(0, 4)).map((p, i) => (
            <li key={p.from} className={'forecast-period' + (i === 0 ? ' forecast-period--now' : '')}>
              <div className="forecast-when">
                <b>{i === 0 ? 'Now' : day(p.from)}</b>
                <span className="muted small">until {day(p.to)} · {grahaEnglish(p.maha)} period, {grahaEnglish(p.antara)} sub-period</span>
              </div>
              <p className="forecast-summary">{p.summary}</p>
              {(i === 0 || data.periods[i - 1]?.antara !== p.antara) && p.themes.length > 0 && (
                <details className="forecast-themes">
                  <summary>What this period is about in your chart</summary>
                  <ul className="forecast-reasons">{p.themes.map(x => <li key={x}>{x}</li>)}</ul>
                </details>
              )}
              {p.book && (i === 0 || data.periods[i - 1]?.maha !== p.maha) && <p className="muted small">The book on a {grahaEnglish(p.maha)} period ({p.book.source.replace(/, tr\..*$/, '')} {p.book.ref}): {p.book.plain}</p>}
              {p.areas.length > 0 && <ul className="forecast-areas">{p.areas.slice(0, 2).map(a => <Area key={a.area.id} a={a} />)}</ul>}
            </li>
          ))}
        </ol>
        {data.periods.length > 4 && <Button variant="ghost" size="sm" onClick={() => setAllPeriods(v => !v)}>{allPeriods ? 'Show fewer periods' : `Show all ${data.periods.length} periods`}</Button>}
      </Card>
      {own && data.past_periods && <LookingBack periods={data.past_periods} />}
      <Card title="Sign changes and eclipses" sub="The slow planets and the eclipses of the next two years">
        <ul className="timing-events">
          {(allEvents ? events : events.slice(0, 10)).map(e => (
            <li key={e.graha + e.kind + e.at}>
              <span className="timing-date">{day(e.at)}</span>
              <span>
                {eventText(e)}
                {e.favourable !== undefined && <> <Chip tone={e.favourable ? 'success' : 'warning'}>{e.favourable ? 'A good house for it' : 'A harder house for it'}</Chip></>}
                {e.book && <span className="muted small block">The book ({e.book.source.replace(/, tr\..*$/, '')} {e.book.ref}): {e.book.plain}</span>}
              </span>
            </li>
          ))}
        </ul>
        {events.length > 10 && <Button variant="ghost" size="sm" onClick={() => setAllEvents(v => !v)}>{allEvents ? 'Show fewer' : `Show all ${events.length}`}</Button>}
        <p className="muted small">{data.notes.nodes}</p>
      </Card>
    </div>
  );
}
