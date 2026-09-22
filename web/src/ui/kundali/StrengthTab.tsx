import { useEffect, useState } from 'react';
import { getShadbala } from '../../api/client';
import type { ChartInput, ShadbalaResponse } from '../../api/types';
import { grahaEnglish } from '../../astro/format';

const cols: [keyof ShadbalaResponse['rows'][number], string][] = [['sthana', 'Sthana'], ['dig', 'Dig'], ['kala', 'Kala'], ['chesta', 'Chesta'], ['naisargika', 'Naisargika'], ['drik', 'Drik']];

export function StrengthTab({ birth }: { birth: ChartInput }) {
  const [data, setData] = useState<ShadbalaResponse>();
  const [error, setError] = useState('');
  const [hover, setHover] = useState<string>();
  useEffect(() => {
    const ctrl = new AbortController();
    getShadbala(birth, ctrl.signal).then(setData).catch(e => { if (e.name !== 'AbortError') setError(e.message); });
    return () => ctrl.abort();
  }, [birth]);
  if (error) return <p role="alert" className="form-error">{error}</p>;
  if (!data) return <div className="card loading-card">Calculating Shadbala…</div>;
  const rows = [...data.rows].sort((a, b) => a.rank - b.rank);
  const max = Math.max(1.6, ...rows.map(r => r.ratio));
  return (
    <div className="strength">
      <div className="card">
        <h3>Shadbala: strength against the classical requirement</h3>
        <p className="muted small">Each bar is a graha's total six-fold strength as a percentage of its BPHS minimum (Sun 5, Moon 6, Mars 5, Mercury 7, Jupiter 6.5, Venus 5.5, Saturn 5 rupas). Past the line, the graha meets its requirement.</p>
        <div className="sb-chart" role="list" aria-label="Shadbala as a percentage of required strength">
          <div className="sb-row sb-axis" aria-hidden><span /><span className="sb-track"><span className="sb-ref" style={{ left: `${(1 / max) * 100}%` }}><span>100% of required</span></span></span><span /></div>
          {rows.map(r => (
            <div key={r.graha} role="listitem" className={'sb-row' + (hover === r.graha ? ' is-hover' : '')}
              onMouseEnter={() => setHover(r.graha)} onMouseLeave={() => setHover(undefined)} onFocus={() => setHover(r.graha)} onBlur={() => setHover(undefined)} tabIndex={0}
              aria-label={`${grahaEnglish(r.graha)}: ${r.rupas.toFixed(2)} rupas, ${(r.ratio * 100).toFixed(0)}% of required, rank ${r.rank}`}>
              <span className="sb-name">{r.rank}. {grahaEnglish(r.graha)}</span>
              <span className="sb-track"><span className="sb-bar" style={{ width: `${(r.ratio / max) * 100}%` }} /><span className="sb-ref" style={{ left: `${(1 / max) * 100}%` }} /></span>
              <span className="sb-value">{(r.ratio * 100).toFixed(0)}% <em className={r.ratio >= 1 ? 'meets' : 'below'}>{r.ratio >= 1 ? '✓ meets' : '▽ below'}</em></span>
              {hover === r.graha && (
                <span className="sb-tip" role="tooltip">
                  <b>{grahaEnglish(r.graha)}</b> {r.rupas.toFixed(2)} of {r.required} rupas · Ishta {r.ishta_phala.toFixed(1)} · Kashta {r.kashta_phala.toFixed(1)}{r.chesta_motion ? ` · motion ${r.chesta_motion}` : ''}
                </span>
              )}
            </div>
          ))}
        </div>
      </div>
      <div className="card table-card">
        <p className="muted table-caption">The six balas in virupas (60 virupas = 1 rupa)</p>
        <div className="table-scroll">
          <table className="planet-table">
            <thead><tr><th>Graha</th>{cols.map(([, l]) => <th key={l}>{l}</th>)}<th>Total</th><th>Rupas</th><th>Required</th><th>Ishta / Kashta</th></tr></thead>
            <tbody>{rows.map(r => (
              <tr key={r.graha}><td>{grahaEnglish(r.graha)}</td>{cols.map(([k]) => <td key={k} className="num">{(r[k] as number).toFixed(1)}</td>)}
                <td className="num"><b>{r.total.toFixed(1)}</b></td><td className="num">{r.rupas.toFixed(2)}</td><td className="num">{r.required}</td>
                <td className="num">{r.ishta_phala.toFixed(1)} / {r.kashta_phala.toFixed(1)}</td></tr>
            ))}</tbody>
          </table>
        </div>
        <ul className="muted small sb-notes">{data.notes.map(n => <li key={n}>{n}</li>)}</ul>
      </div>
      <details className="card">
        <summary><b>What the six balas mean</b></summary>
        <dl className="kv">
          <div><dt>Sthana</dt><dd>positional: exaltation, dignity across seven vargas, odd/even signs, kendra position, decanate</dd></div>
          <div><dt>Dig</dt><dd>directional: Sun and Mars strongest in the 10th, Jupiter and Mercury in the 1st, Saturn in the 7th, Moon and Venus in the 4th</dd></div>
          <div><dt>Kala</dt><dd>temporal: day or night birth, lunar phase, the day's thirds, year/month/weekday/hora lords, declination, planetary war</dd></div>
          <div><dt>Chesta</dt><dd>motional: retrograde and slow grahas gain strength</dd></div>
          <div><dt>Naisargika</dt><dd>natural brightness order, Sun strongest to Saturn</dd></div>
          <div><dt>Drik</dt><dd>aspects received: benefic aspects add, malefic aspects subtract</dd></div>
        </dl>
      </details>
    </div>
  );
}
