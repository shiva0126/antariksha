import type { ChartFacts } from '../../api/types';
import { rashis } from '../../astro/rashi';
import { grahaEnglish } from '../../astro/format';
import { useNames } from '../../i18n';

const order = ['sun', 'moon', 'mars', 'mercury', 'jupiter', 'venus', 'saturn'];

export function AshtakavargaTab({ facts }: { facts: ChartFacts }) {
  const n = useNames();
  const av = facts.ashtakavarga;
  const lagna = Math.floor(facts.chart.ascendant.longitude / 30);
  const house = (s: number) => ((s - lagna + 12) % 12) + 1;
  const total = av.sarva.reduce((a, b) => a + b, 0);
  const strongest = av.sarva.map((v, i) => [v, i]).sort((a, b) => b[0] - a[0]).slice(0, 3).map(([, i]) => i);
  return (
    <div className="ashtaka">
      <div className="card">
        <h3>Sarvashtakavarga</h3>
        <p className="muted small">Bindus per sign from all seven grahas and the Lagna (total {total}; 337 in every chart). Signs above 28 bindus traditionally give good results for the houses they hold and for transits through them; below 25 they need more care.</p>
        <div className="sav-grid">
          {av.sarva.map((v, s) => (
            <div key={s} className={'sav ' + (v >= 28 ? 'strong' : v < 25 ? 'weak' : '')}>
              <b>{v}</b><span>{n(rashis[s][0])}</span><em>House {house(s)}</em>
            </div>
          ))}
        </div>
        <p className="muted small">Strongest: {strongest.map(s => `${rashis[s][0]} (house ${house(s)})`).join(', ')}.</p>
      </div>
      <div className="card table-card">
        <p className="muted table-caption">Bhinnashtakavarga: each graha's bindus by sign</p>
        <div className="table-scroll">
          <table className="planet-table av-table">
            <thead><tr><th>Graha</th>{rashis.map(r => <th key={r[0]} title={r[0]}>{r[0].slice(0, 3)}</th>)}<th>Total</th></tr></thead>
            <tbody>
              {order.map(g => {
                const row = av.bhinna[g] ?? [];
                return <tr key={g}><td>{grahaEnglish(g)}</td>{row.map((v, i) => <td key={i} className={'num av-' + v}>{v}</td>)}<td className="num"><b>{row.reduce((a, b) => a + b, 0)}</b></td></tr>;
              })}
              <tr className="av-total"><td>Sarva</td>{av.sarva.map((v, i) => <td key={i} className="num"><b>{v}</b></td>)}<td className="num"><b>{total}</b></td></tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
