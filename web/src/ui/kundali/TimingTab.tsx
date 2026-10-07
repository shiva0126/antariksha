import { useEffect, useState } from 'react';
import { getTransits, getVarshaphal } from '../../api/client';
import type { AreaForecast, ChartInput, TransitEvent, TransitsResponse, VarshaphalResponse } from '../../api/types';
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
    case 'exact': {
      const point = e.target === 'lagna' ? 'rising degree (lagna)' : `natal ${e.target === 'sun' ? 'Sun' : 'Moon'}`;
      const meaning = e.graha === 'saturn' ? (e.target === 'moon' ? ' This is the heart of Sade Sati: a time for patience and steady effort.' : ' A time to slow down and build carefully.')
        : e.graha === 'jupiter' ? ' Traditionally a supportive, expanding moment.' : ' A time of change and new directions.';
      return `${who[0].toUpperCase() + who.slice(1)} passes exactly over your ${point} at ${e.degree.toFixed(1)}° ${e.rashi}.${meaning}`;
    }
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
const localDay = (iso: string, zone: string) => longDate(new Date(iso).toLocaleDateString('en-CA', { timeZone: zone }));

const munthaChip = { good: 'success', mixed: 'neutral', hard: 'warning' } as const;

/** The Tajika year chart from one birthday to the next (Varshaphal). */
function YearCard({ birth }: { birth: ChartInput }) {
  const [year, setYear] = useState<number>();
  const [data, setData] = useState<VarshaphalResponse>();
  const [error, setError] = useState('');
  useEffect(() => {
    const ctrl = new AbortController();
    setError('');
    getVarshaphal(birth, year, ctrl.signal).then(setData).catch(e => { if (e.name !== 'AbortError') setError(e.message); });
    return () => ctrl.abort();
  }, [birth, year]);
  if (error) return <Notice tone="danger">Year chart unavailable: {error}</Notice>;
  if (!data) return <Card title="Your year"><Skeleton lines={4} /></Card>;
  const v = data.varsha;
  const zone = birth.tz;
  const when = (iso: string) => new Date(iso).toLocaleString('en-IN', { timeZone: zone, day: 'numeric', month: 'long', year: 'numeric', hour: 'numeric', minute: '2-digit' });
  return (
    <Card title={`Your year: ${localDay(v.return_at, zone)} to ${localDay(data.until, zone)}`}
      sub={`Varshaphal: the chart for the moment the Sun returns to its birth position (${when(v.return_at)}), read for the year that follows. Age ${v.age}.`}>
      <div className="varsha">
        <p><b>The year is ruled by {the(v.year_lord)}.</b> {v.year_lord_reason}</p>
        <p><Chip tone={munthaChip[v.muntha.tone]}>Muntha in {v.muntha.rashi}</Chip> {v.muntha.detail}</p>
        {v.strengths.length > 0 && <><b className="small">Working for you</b><ul className="forecast-reasons">{v.strengths.map(x => <li key={x}>{x}</li>)}</ul></>}
        {v.cautions.length > 0 && <><b className="small">Go carefully</b><ul className="forecast-reasons">{v.cautions.map(x => <li key={x}>{x}</li>)}</ul></>}
        <b className="small">The year, stretch by stretch (Mudda dasha)</b>
        <ol className="forecast mudda">
          {v.mudda.map(m => {
            const current = Date.now() >= Date.parse(m.from) && Date.now() < Date.parse(m.to);
            return (
              <li key={m.from} className={'forecast-period' + (current ? ' forecast-period--now' : '')}>
                <div className="forecast-when">
                  <Chip tone={toneChip[m.tone]}>{toneWord[m.tone]}</Chip>
                  <b>{grahaEnglish(m.lord)}</b>
                  <span className="muted small">{current ? 'Now, until ' + localDay(m.to, zone) : localDay(m.from, zone) + ' to ' + localDay(m.to, zone)}</span>
                </div>
                <span className="small">{m.detail}</span>
              </li>
            );
          })}
        </ol>
        <details>
          <summary>Chart details</summary>
          <p className="small">Year lagna {v.chart.ascendant.rashi} {v.chart.ascendant.degree.toFixed(1)}°, a {v.day_chart ? 'day' : 'night'} chart. The five office-holders:</p>
          <ul className="forecast-reasons">
            {v.offices.map(o => <li key={o.role}>{o.role}: {grahaEnglish(o.graha)}, {ord(o.house)} house, strength {o.strength.toFixed(2)}{o.aspect === 'conjunct' ? ', in the lagna' : o.aspect ? `, ${o.aspect} aspect to the lagna` : ', no aspect to the lagna'}</li>)}
          </ul>
          <p className="muted small">{v.method} Mudda dasha: the Vimshottari order compressed into the year, starting from the birth star's lord advanced one place per year of age.</p>
        </details>
        <div className="ds-row">
          <Button size="sm" variant="ghost" disabled={v.age <= 1} onClick={() => setYear(v.year - 1)}>Previous year</Button>
          <Button size="sm" variant="ghost" onClick={() => setYear(v.year + 1)}>Next year</Button>
        </div>
      </div>
    </Card>
  );
}
const cap = (s: string) => s[0].toUpperCase() + s.slice(1);
const day = (iso: string) => longDate(iso.slice(0, 10));

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
  const events = data.events.filter(e => (slow.has(e.graha) && e.graha !== 'ketu') || e.kind.endsWith('eclipse') || e.kind === 'exact');
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
      <YearCard birth={birth} />
      {own && data.past_periods && <LookingBack periods={data.past_periods} />}
      <Card title="Sign changes, exact passes and eclipses" sub="The slow planets and the eclipses of the next two years, including when they cross your natal Moon, Sun and rising degree">
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
