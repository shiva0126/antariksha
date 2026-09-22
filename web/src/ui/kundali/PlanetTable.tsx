import type { ChartFacts, ChartResponse, Graha } from '../../api/types';
import { degreeLabel, houseNumber } from '../../astro/houses';
import { grahaColor, grahaGlyph } from '../../astro/rashi';
import { grahaEnglish, titleCase } from '../../astro/format';

export function PlanetTable({ chart, facts, onSelect, selected, caption }: { chart: ChartResponse; facts?: ChartFacts; onSelect?: (g: Graha) => void; selected?: string; caption?: string }) {
  return (
    <div className="card table-card">
      {caption && <p className="muted table-caption">{caption}</p>}
      <div className="table-scroll">
        <table className="planet-table">
          <thead><tr><th>Graha</th><th>Rashi</th><th>Degree</th><th>Nakshatra · pada</th><th>House</th>{facts && <th>Dignity</th>}<th>Motion</th></tr></thead>
          <tbody>
            {chart.grahas.map(g => {
              const d = facts?.dignities[g.id];
              const notes = [facts?.combustion[g.id] ? 'combust' : '', d?.neecha_bhanga ? 'neecha bhanga' : ''].filter(Boolean).join(', ');
              return (
                <tr key={g.id} className={selected === g.id ? 'is-selected' : ''}>
                  <td><button className="graha-link" onClick={() => onSelect?.(g)} style={{ color: grahaColor[g.id] }} disabled={!onSelect}>
                    <span aria-hidden>{grahaGlyph[g.id]}</span> {g.name} <small>{grahaEnglish(g.id)}</small></button></td>
                  <td>{g.rashi}</td>
                  <td className="num">{degreeLabel(g.rashi_degree)}</td>
                  <td>{g.nakshatra} · {g.nakshatra_pada}</td>
                  <td className="num">{houseNumber(g.longitude, chart.ascendant.longitude)}</td>
                  {facts && <td>{d ? titleCase(d.state) : '—'}{notes && <small className="note"> {notes}</small>}</td>}
                  <td>{g.id === 'rahu' || g.id === 'ketu' ? 'Node' : g.retrograde ? 'Retrograde ℞' : 'Direct'}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      <p className="muted small">Lahiri sidereal · whole-sign houses from the Lagna sign · true lunar nodes (always retrograde by nature).</p>
    </div>
  );
}
