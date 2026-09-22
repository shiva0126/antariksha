import { useEffect, useMemo, useState } from 'react';
import { getMuhurta, getMuhurtaEvents } from '../../api/client';
import type { MuhurtaEvent, MuhurtaResponse } from '../../api/types';
import { longDate } from '../../astro/format';
import { useNames, useT } from '../../i18n';
import { PlaceSearch } from '../common/PlaceSearch';
import { defaultPlace, type Place } from '../common/places';
import { loadProfiles } from '../common/profiles';

const todayIST = () => new Intl.DateTimeFormat('en-CA', { timeZone: 'Asia/Kolkata', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date());

export function MuhurtaPage() {
  const t = useT(), n = useNames();
  const { profiles } = useMemo(loadProfiles, []);
  const [events, setEvents] = useState<MuhurtaEvent[]>([]);
  const [event, setEvent] = useState('marriage');
  const [from, setFrom] = useState(todayIST);
  const [days, setDays] = useState(30);
  const [place, setPlace] = useState<Place>(defaultPlace);
  const [who, setWho] = useState(profiles[0]?.id ?? '');
  const [showAll, setShowAll] = useState(false);
  const [data, setData] = useState<MuhurtaResponse>();
  const [error, setError] = useState(''), [busy, setBusy] = useState(false);

  useEffect(() => { getMuhurtaEvents().then(setEvents).catch(() => {}); }, []);

  async function find(e?: React.FormEvent) {
    e?.preventDefault();
    setBusy(true); setError('');
    try { setData(await getMuhurta(event, from, days, place, profiles.find(p => p.id === who))); } catch (err) { setError(err instanceof Error ? err.message : 'Could not calculate'); } finally { setBusy(false); }
  }

  const good = data?.days.filter(d => d.good) ?? [];
  const list = showAll ? data?.days ?? [] : good;
  const ev = events.find(x => x.ID === event);
  return (
    <div className="page muhurta">
      <header className="page-head">
        <div>
          <p className="kicker">Auspicious timing</p>
          <h1>{t('Muhurta')}</h1>
          <p className="muted">Finds days whose nakshatra, tithi and weekday suit the event, avoids Rikta tithis, Amavasya, Bhadra and inauspicious yogas, and, with a saved profile, checks your personal tara bala and chandra bala.</p>
        </div>
      </header>
      <form className="card muhurta-form" onSubmit={find}>
        <label className="field"><span>{t('Event')}</span>
          <select value={event} onChange={e => setEvent(e.target.value)}>{(events.length ? events : [{ ID: 'marriage', Name: 'Marriage (Vivaha)', Description: '' }]).map(x => <option key={x.ID} value={x.ID}>{x.Name}</option>)}</select></label>
        <PlaceSearch label="Where it happens" value={place} onChange={setPlace} compact />
        <label className="field"><span>From</span><input type="date" value={from} onChange={e => setFrom(e.target.value)} /></label>
        <label className="field"><span>Range</span>
          <select value={days} onChange={e => setDays(Number(e.target.value))}>{[15, 30, 60, 90].map(d => <option key={d} value={d}>{d} days</option>)}</select></label>
        <label className="field"><span>Personalise for</span>
          <select value={who} onChange={e => setWho(e.target.value)}>
            <option value="">No one (general muhurta)</option>
            {profiles.map(p => <option key={p.id} value={p.id}>{p.name || p.date}</option>)}
          </select></label>
        <button className="primary" disabled={busy}>{busy ? 'Searching…' : t('Find muhurta')}</button>
      </form>
      {ev?.Description && <p className="muted small">{ev.Description}</p>}
      {error && <p role="alert" className="form-error">{error}</p>}
      {data && (
        <section aria-label="Muhurta results">
          <div className="results-head">
            <h2>{good.length} suitable day{good.length === 1 ? '' : 's'} for {data.event.name}{data.personalised ? ' (personalised)' : ''}</h2>
            <label className="consent"><input type="checkbox" checked={showAll} onChange={e => setShowAll(e.target.checked)} /> <span>Show every day, with reasons</span></label>
          </div>
          {list.length === 0 && <div className="card"><p>No day in this range meets all the rules. Try a longer range{data.personalised ? ' or the general muhurta' : ''}.</p></div>}
          <div className="muhurta-list">
            {list.map(d => (
              <article key={d.date} className={'card muhurta-day ' + (d.good ? 'good' : 'bad')}>
                <header><h3>{longDate(d.date)} · {n(d.vaara)}</h3><span className={'badge ' + (d.good ? 'badge-strong' : 'badge-caution')}>{d.good ? 'Suitable' : 'Not suitable'}</span></header>
                <p className="muted small">{n(d.tithi)} ({n(d.paksha)}) · {n(d.nakshatra)} · {n(d.yoga)} yoga</p>
                {d.reasons.length > 0 && <ul className="reasons">{d.reasons.map(r => <li key={r}>✓ {r}</li>)}</ul>}
                {d.cautions.length > 0 && <ul className="cautions">{d.cautions.map(r => <li key={r}>✕ {r}</li>)}</ul>}
                {d.good && <p className="small"><b>Best windows:</b> {d.windows.map(w => `${w.start}–${w.end}`).join(', ')} · <b>avoid</b> {d.avoid.map(w => `${w.start}–${w.end}`).join(', ')}</p>}
              </article>
            ))}
          </div>
          <p className="disclaimer">Day-level classical rules. Month-level rules (Kharmas, Chaturmas, combust Jupiter or Venus for marriage), lagna shuddhi and local custom are not evaluated; for important events, confirm with a trusted astrologer.</p>
        </section>
      )}
    </div>
  );
}
