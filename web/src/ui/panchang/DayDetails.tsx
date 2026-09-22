import type { PanchangDay } from '../../api/panchang';
import { longDate } from '../../astro/format';
import { useMonthName, useNames, useSettings, useT } from '../../i18n';

const goodChog = new Set(['Amrit', 'Shubh', 'Labh']);
const badChog = new Set(['Rog', 'Kala', 'Udveg']);
const recurring = new Set(['Pradosh Vrat', 'Sankashti Chaturthi', 'Purnima', 'Amavasya']);
const isVrat = (f: string) => recurring.has(f) || f.endsWith('Ekadashi');

export function FestivalChips({ names }: { names: string[] }) {
  if (!names.length) return null;
  return <ul className="tag-list">{names.map(f => <li key={f} className={'tag ' + (isVrat(f) ? 'tag-vrat' : f.includes('Sankranti') ? 'tag-solar' : 'tag-festival')}>{f}</li>)}</ul>;
}

export function DayDetails({ day, placeName }: { day: PanchangDay; placeName: string }) {
  const t = useT(), n = useNames(), monthName = useMonthName();
  const months = useSettings().months === 'amanta' ? 'Amanta' : 'Purnimanta';
  const limbs: [string, string, string][] = [
    [t('Vaara'), n(day.vaara), 'Weekday, sunrise to sunrise'],
    [t('Tithi'), n(day.tithi.name), `${n(day.paksha)} ${t('paksha')} · until ${day.tithi.ends_at}`],
    [t('Nakshatra'), n(day.nakshatra.name), `Until ${day.nakshatra.ends_at}`],
    [t('Yoga'), n(day.yoga.name), `Until ${day.yoga.ends_at}`],
    [t('Karana'), n(day.karana.name), `Until ${day.karana.ends_at}${day.karana.name === 'Vishti' ? ' · Bhadra' : ''}`],
  ];
  const windows: [string, { start: string; end: string }, 'good' | 'bad'][] = [
    [t('Abhijit muhurta'), day.abhijit, 'good'], [t('Rahu kaal'), day.rahu_kaal, 'bad'], [t('Yamaganda'), day.yamaganda, 'bad'], [t('Gulika kaal'), day.gulika, 'bad'],
  ];
  const chog = (d: boolean) => day.choghadiya.slice(d ? 0 : 8, d ? 8 : 16);
  return (
    <section className="day-details" aria-label={`Panchang for ${day.date}`}>
      <header className="day-head">
        <div>
          <h2>{longDate(day.date)} · {n(day.vaara)}</h2>
          <p className="muted">{n(monthName(day.lunar_month, day.paksha))} {t('month')} ({months}) · {n(day.paksha)} {t('paksha')} · {placeName} · {day.location.tz}</p>
        </div>
        <FestivalChips names={day.festivals} />
      </header>
      <div className="limb-grid">
        {limbs.map(([k, v, note]) => <article className="card limb" key={k}><span className="kicker">{k}</span><h3>{v}</h3><p className="muted small">{note}</p></article>)}
      </div>
      <div className="grid-3">
        <article className="card">
          <h3>{t('Sun and Moon')}</h3>
          <dl className="kv">
            <div><dt>{t('Sunrise')}</dt><dd>{day.sunrise}</dd></div><div><dt>{t('Sunset')}</dt><dd>{day.sunset}</dd></div>
            <div><dt>{t('Moonrise')}</dt><dd>{day.moonrise || 'None this day'}</dd></div><div><dt>{t('Moonset')}</dt><dd>{day.moonset || 'None this day'}</dd></div>
          </dl>
        </article>
        <article className="card">
          <h3>{t('Muhurta windows')}</h3>
          <dl className="kv">{windows.map(([k, w, kind]) => <div key={k} className={'win-' + kind}><dt>{k}</dt><dd>{w.start} – {w.end}</dd></div>)}</dl>
          <p className="muted small">Abhijit is auspicious; Rahu kaal, Yamaganda and Gulika are traditionally avoided for new beginnings.</p>
        </article>
        <article className="card">
          <h3>{t('Choghadiya')}</h3>
          <div className="chog">
            {[true, false].map(d => (
              <div key={String(d)}><h4>{d ? t('Day') : t('Night')}</h4>
                <ul>{chog(d).map((w, i) => <li key={i} className={goodChog.has(w.name) ? 'good' : badChog.has(w.name) ? 'bad' : ''}><span>{n(w.name)}</span><b>{w.start}–{w.end}</b></li>)}</ul>
              </div>
            ))}
          </div>
        </article>
      </div>
    </section>
  );
}
