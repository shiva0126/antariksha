import type { PanchangDay } from '../../api/panchang';
import { longDate } from '../../astro/format';

const goodChog = new Set(['Amrit', 'Shubh', 'Labh']);
const badChog = new Set(['Rog', 'Kala', 'Udveg']);
const recurring = new Set(['Ekadashi', 'Pradosh Vrat', 'Sankashti Chaturthi', 'Purnima', 'Amavasya']);

export function FestivalChips({ names }: { names: string[] }) {
  if (!names.length) return null;
  return <ul className="tag-list">{names.map(f => <li key={f} className={'tag ' + (recurring.has(f) ? 'tag-vrat' : f.includes('Sankranti') ? 'tag-solar' : 'tag-festival')}>{f}</li>)}</ul>;
}

export function DayDetails({ day, placeName }: { day: PanchangDay; placeName: string }) {
  const limbs: [string, string, string][] = [
    ['Vaara', day.vaara, 'Weekday, sunrise to sunrise'],
    ['Tithi', day.tithi.name, `${day.paksha} paksha · until ${day.tithi.ends_at}`],
    ['Nakshatra', day.nakshatra.name, `Until ${day.nakshatra.ends_at}`],
    ['Yoga', day.yoga.name, `Until ${day.yoga.ends_at}`],
    ['Karana', day.karana.name, `Until ${day.karana.ends_at}`],
  ];
  const windows: [string, { start: string; end: string }, 'good' | 'bad'][] = [
    ['Abhijit muhurta', day.abhijit, 'good'], ['Rahu kaal', day.rahu_kaal, 'bad'], ['Yamaganda', day.yamaganda, 'bad'], ['Gulika kaal', day.gulika, 'bad'],
  ];
  const chog = (d: boolean) => day.choghadiya.slice(d ? 0 : 8, d ? 8 : 16);
  return (
    <section className="day-details" aria-label={`Panchang for ${day.date}`}>
      <header className="day-head">
        <div>
          <h2>{longDate(day.date)} · {day.vaara}</h2>
          <p className="muted">{day.lunar_month} month (Amanta) · {day.paksha} paksha · {placeName} · {day.location.tz}</p>
        </div>
        <FestivalChips names={day.festivals} />
      </header>
      <div className="limb-grid">
        {limbs.map(([k, v, note]) => <article className="card limb" key={k}><span className="kicker">{k}</span><h3>{v}</h3><p className="muted small">{note}</p></article>)}
      </div>
      <div className="grid-3">
        <article className="card">
          <h3>Sun and Moon</h3>
          <dl className="kv">
            <div><dt>Sunrise</dt><dd>{day.sunrise}</dd></div><div><dt>Sunset</dt><dd>{day.sunset}</dd></div>
            <div><dt>Moonrise</dt><dd>{day.moonrise || 'None this day'}</dd></div><div><dt>Moonset</dt><dd>{day.moonset || 'None this day'}</dd></div>
          </dl>
        </article>
        <article className="card">
          <h3>Muhurta windows</h3>
          <dl className="kv">{windows.map(([k, w, kind]) => <div key={k} className={'win-' + kind}><dt>{k}</dt><dd>{w.start} – {w.end}</dd></div>)}</dl>
          <p className="muted small">Abhijit is auspicious; Rahu kaal, Yamaganda and Gulika are traditionally avoided for new beginnings.</p>
        </article>
        <article className="card">
          <h3>Choghadiya</h3>
          <div className="chog">
            {[true, false].map(d => (
              <div key={String(d)}><h4>{d ? 'Day' : 'Night'}</h4>
                <ul>{chog(d).map((w, i) => <li key={i} className={goodChog.has(w.name) ? 'good' : badChog.has(w.name) ? 'bad' : ''}><span>{w.name}</span><b>{w.start}–{w.end}</b></li>)}</ul>
              </div>
            ))}
          </div>
        </article>
      </div>
    </section>
  );
}
