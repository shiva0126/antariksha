import { useEffect, useState } from 'react';
import { calendarURL, fetchJSON, getChart } from '../../api/client';
import type { MonthDay, PanchangDay } from '../../api/panchang';
import type { ChartResponse } from '../../api/types';
import { SouthIndian } from '../../chart2d/SouthIndian';
import { PlanetTable } from '../kundali/PlanetTable';
import { PlaceSearch } from '../common/PlaceSearch';
import { useT } from '../../i18n';
import { defaultPlace, type Place } from '../common/places';
import { DayDetails, FestivalChips } from './DayDetails';

const todayIn = (tz: string) => new Intl.DateTimeFormat('en-CA', { timeZone: tz, year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date());
const shift = (iso: string, days: number) => { const d = new Date(iso + 'T12:00:00Z'); d.setUTCDate(d.getUTCDate() + days); return d.toISOString().slice(0, 10); };
const shiftMonth = (ym: string, n: number) => { const [y, m] = ym.split('-').map(Number); const d = new Date(Date.UTC(y, m - 1 + n, 1)); return d.toISOString().slice(0, 7); };
const monthTitle = (ym: string) => new Date(ym + '-01T12:00:00Z').toLocaleDateString('en-GB', { month: 'long', year: 'numeric', timeZone: 'UTC' });

export function PanchangPage({ mode }: { mode: 'day' | 'month' }) {
  const t = useT();
  const [place, setPlace] = useState<Place>(defaultPlace);
  const [date, setDate] = useState(() => todayIn(defaultPlace.tz));
  const [month, setMonth] = useState(() => todayIn(defaultPlace.tz).slice(0, 7));
  const [day, setDay] = useState<PanchangDay>(), [days, setDays] = useState<MonthDay[]>([]);
  const [time, setTime] = useState('12:00'), [chart, setChart] = useState<ChartResponse>();
  const [error, setError] = useState(''), [monthError, setMonthError] = useState(''), [chartError, setChartError] = useState('');
  const today = todayIn(place.tz);

  useEffect(() => {
    const ctrl = new AbortController();
    setDay(undefined); setError('');
    fetchJSON<PanchangDay>(`/api/panchang?${new URLSearchParams({ date, lat: String(place.lat), lon: String(place.lon), tz: place.tz })}`, ctrl.signal)
      .then(d => setDay({ ...d, festivals: d.festivals ?? [] })).catch(e => { if (e.name !== 'AbortError') setError(e.message); });
    return () => ctrl.abort();
  }, [date, place]);

  useEffect(() => {
    if (mode !== 'month') return;
    const ctrl = new AbortController();
    setDays([]); setMonthError('');
    const [year, m] = month.split('-');
    fetchJSON<MonthDay[]>(`/api/month?${new URLSearchParams({ year, month: m, lat: String(place.lat), lon: String(place.lon), tz: place.tz })}`, ctrl.signal)
      .then(ds => setDays(ds.map(d => ({ ...d, festivals: d.festivals ?? [] })))).catch(e => { if (e.name !== 'AbortError') setMonthError(e.message); });
    return () => ctrl.abort();
  }, [month, place, mode]);

  useEffect(() => {
    const ctrl = new AbortController();
    setChart(undefined); setChartError('');
    getChart({ date, time, lat: place.lat, lon: place.lon, tz: place.tz }, ctrl.signal).then(setChart).catch(e => { if (e.name !== 'AbortError') setChartError(e.message); });
    return () => ctrl.abort();
  }, [date, time, place]);

  const pick = (d: string) => { setDate(d); setMonth(d.slice(0, 7)); };
  const leading = new Date(month + '-01T12:00:00Z').getUTCDay();

  return (
    <div className="page panchang">
      <header className="page-head">
        <div>
          <p className="kicker">Hindu calendar</p>
          <h1>{mode === 'month' ? t('Panchang calendar') : t('Daily Panchang')}</h1>
          <p className="muted">The five limbs at local sunrise, festivals, muhurta windows and choghadiya for your location.</p>
        </div>
        <div className="filters">
          <PlaceSearch label={t('Location')} value={place} onChange={p => setPlace(p)} compact />
          {mode === 'day' ? (
            <div className="stepper">
              <button className="ghost" aria-label="Previous day" onClick={() => pick(shift(date, -1))}>‹</button>
              <label className="field compact"><span>{t('Date')}</span><input type="date" aria-label="Panchang date" value={date} onChange={e => e.target.value && pick(e.target.value)} /></label>
              <button className="ghost" aria-label="Next day" onClick={() => pick(shift(date, 1))}>›</button>
            </div>
          ) : (
            <div className="stepper">
              <button className="ghost" aria-label="Previous month" onClick={() => setMonth(shiftMonth(month, -1))}>‹</button>
              <label className="field compact"><span>{t('Month')}</span><input type="month" aria-label="Calendar month" value={month} onChange={e => e.target.value && setMonth(e.target.value)} /></label>
              <button className="ghost" aria-label="Next month" onClick={() => setMonth(shiftMonth(month, 1))}>›</button>
            </div>
          )}
          <button className="ghost" onClick={() => pick(today)}>{t('Today')}</button>
        </div>
      </header>

      {mode === 'month' && (
        <section className="card calendar" aria-label={`Panchang calendar for ${monthTitle(month)}`}>
          <div className="calendar-head">
            <h2>{monthTitle(month)}</h2>
            <a className="ghost" href={calendarURL(Number(month.slice(0, 4)), place)} download>Add {month.slice(0, 4)} festivals to my calendar (.ics)</a>
          </div>
          {monthError && <p role="alert" className="form-error">{monthError}</p>}
          <div className="calendar-grid">
            {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map(d => <div className="weekday" key={d}>{d}</div>)}
            {Array.from({ length: leading }, (_, i) => <div key={'b' + i} className="calendar-empty" />)}
            {days.map(d => {
              const named = d.festivals.filter(f => !f.endsWith('Ekadashi') && !['Pradosh Vrat', 'Sankashti Chaturthi', 'Purnima', 'Amavasya'].includes(f));
              return (
                <button key={d.date} className={'calendar-day' + (d.date === date ? ' selected' : '') + (d.date === today ? ' today' : '') + (named.length ? ' has-festival' : '')}
                  onClick={() => setDate(d.date)} aria-label={`${d.date} ${d.tithi} ${d.paksha} ${d.festivals.join(', ')}`} aria-pressed={d.date === date}>
                  <b>{Number(d.date.slice(-2))}</b>
                  <span className="cal-tithi">{d.tithi}</span>
                  <span className={'cal-paksha ' + d.paksha.toLowerCase()}>{d.paksha}</span>
                  {d.festivals.length > 0 && <em>{d.festivals[0]}{d.festivals.length > 1 ? ` +${d.festivals.length - 1}` : ''}</em>}
                </button>
              );
            })}
          </div>
          {!days.length && !monthError && <p className="muted">Loading the month…</p>}
          {days.some(d => d.festivals.length) && (
            <details className="festival-list" open>
              <summary>{t('Festivals and observances this month')}</summary>
              <ul>{days.filter(d => d.festivals.length).map(d => <li key={d.date}><button className="link-button" onClick={() => setDate(d.date)}>{Number(d.date.slice(-2))} {monthTitle(month).split(' ')[0].slice(0, 3)}</button><FestivalChips names={d.festivals} /></li>)}</ul>
            </details>
          )}
        </section>
      )}

      {error && <p role="alert" className="form-error">{error}</p>}
      {day ? <DayDetails day={day} placeName={place.name} /> : !error && <div className="card loading-card">Calculating Panchang…</div>}

      <details className="card transits">
        <summary><h2>Planet positions on {date}</h2><span className="muted small">Transits at a chosen local time — not a birth chart</span></summary>
        <label className="field compact transit-time"><span>Time · {place.tz}</span><input type="time" value={time} aria-label="Planet positions time" onChange={e => e.target.value && setTime(e.target.value)} /></label>
        {chartError && <p className="form-error">{chartError}</p>}
        {chart ? <div className="transit-body"><div className="transit-chart"><SouthIndian chart={chart} /></div><PlanetTable chart={chart} /></div> : !chartError && <p className="muted">Calculating planet positions…</p>}
      </details>
    </div>
  );
}
