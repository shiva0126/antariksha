import { useEffect, useState } from 'react';
import { getToday } from '../../api/client';
import type { ChartInput, TodayResponse } from '../../api/types';
import { grahaColor, grahaGlyph } from '../../astro/rashi';
import { grahaEnglish, longDate } from '../../astro/format';
import { useMonthName, useNames, useT } from '../../i18n';
import { FestivalChips } from '../panchang/DayDetails';

const ord = (n: number) => n + (n === 1 ? 'st' : n === 2 ? 'nd' : n === 3 ? 'rd' : 'th');

export function TodayTab({ birth }: { birth: ChartInput }) {
  const [data, setData] = useState<TodayResponse>();
  const [error, setError] = useState('');
  const t = useT(), n = useNames(), month = useMonthName();
  useEffect(() => {
    const ctrl = new AbortController();
    getToday(birth, ctrl.signal).then(setData).catch(e => { if (e.name !== 'AbortError') setError(e.message); });
    return () => ctrl.abort();
  }, [birth]);
  if (error) return <p role="alert" className="form-error">{error}</p>;
  if (!data) return <div className="card loading-card">Reading today's sky…</div>;
  const p = data.panchang, d = data.dasha;
  return (
    <div className="today">
      <div className="grid-3">
        <article className="card">
          <p className="kicker">{t('Today')} · {longDate(p.date)}</p>
          <h3>{n(p.vaara)} · {n(p.tithi.name)}</h3>
          <p>{n(p.paksha)} {t('paksha')} · {n(month(p.lunar_month, p.paksha))} {t('month')}</p>
          <p className="muted small">{t('Nakshatra')} {n(p.nakshatra.name)} until {p.nakshatra.ends_at} · {t('Sunrise')} {p.sunrise} · {t('Sunset')} {p.sunset}</p>
          <FestivalChips names={p.festivals} />
        </article>
        <article className={'card bala ' + (data.tara_bala.favourable ? 'good' : 'bad')}>
          <p className="kicker">Tara bala</p>
          <h3>{data.tara_bala.name} <small>({data.tara_bala.number})</small></h3>
          <p>The Moon is in {n(data.tara_bala.moon_nakshatra)}, the {ord(data.tara_bala.number)} tara from your birth star. It is {data.tara_bala.favourable ? 'favourable' : 'unfavourable'} for new beginnings today.</p>
        </article>
        <article className={'card bala ' + (data.chandra_bala.favourable ? 'good' : 'bad')}>
          <p className="kicker">Chandra bala</p>
          <h3>{ord(data.chandra_bala.house)} from your Moon</h3>
          <p>The Moon transits {n(data.chandra_bala.moon_sign)}, which is {data.chandra_bala.favourable ? 'supportive' : 'less supportive'} for your mind and plans today.</p>
        </article>
      </div>
      <div className="grid-2">
        <article className="card">
          <h3>Running periods</h3>
          <dl className="kv">
            <div><dt>Mahadasha</dt><dd>{grahaEnglish(d.maha ?? '')}</dd></div>
            <div><dt>Antardasha</dt><dd>{grahaEnglish(d.antara ?? '')} · until {longDate(d.to)}</dd></div>
            {d.pratyantara && <div><dt>Pratyantardasha</dt><dd>{grahaEnglish(d.pratyantara)}</dd></div>}
            <div><dt>Sade Sati</dt><dd>{data.sade_sati.active ? ['', 'Rising phase', 'Peak phase', 'Setting phase'][data.sade_sati.phase] : 'Not running'}</dd></div>
          </dl>
        </article>
        <article className="card">
          <h3>Coming up</h3>
          {data.upcoming.length === 0 ? <p className="muted">No festivals in the next two weeks.</p> :
            <ul className="upcoming">{data.upcoming.map(u => <li key={u.date}><b>{longDate(u.date)}</b><FestivalChips names={u.festivals} /></li>)}</ul>}
        </article>
      </div>
      <div className="card table-card">
        <p className="muted table-caption">Planets now (gochar), counted from your natal Lagna and Moon</p>
        <div className="table-scroll">
          <table className="planet-table">
            <thead><tr><th>Graha</th><th>Transit sign</th><th>Nakshatra</th><th>From Lagna</th><th>From Moon</th></tr></thead>
            <tbody>{data.transits.map(g => (
              <tr key={g.id}><td style={{ color: grahaColor[g.id] }}>{grahaGlyph[g.id]} {n(g.name)} <small>{grahaEnglish(g.id)}</small></td>
                <td>{n(g.rashi)} {g.degree.toFixed(1)}°{g.retrograde && g.id !== 'rahu' && g.id !== 'ketu' ? ' ℞' : ''}</td><td>{n(g.nakshatra)}</td>
                <td className="num">{ord(g.house_from_lagna)}</td><td className="num">{ord(g.house_from_moon)}</td></tr>
            ))}</tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
