import type {ChartResponse,Graha} from '../api/types';
import {degreeLabel,houseNumber} from '../astro/houses';
import {grahaColor,grahaGlyph} from '../astro/rashi';
export function PlanetTable({chart,onSelect}:{chart:ChartResponse;onSelect?:(g:Graha)=>void}) {
  return <section className="planet-table"><h2>Graha positions</h2><p>{chart.input.date} · {chart.input.time} · {chart.input.tz} · Lahiri sidereal · True lunar nodes</p><div className="table-scroll"><table><thead><tr><th>Graha</th><th>Rashi</th><th>Degree in sign</th><th>Nakshatra / Pada</th><th>House</th><th>Longitude</th><th>Motion</th></tr></thead><tbody>{chart.grahas.map(g=><tr key={g.id}><td><button onClick={()=>onSelect?.(g)} style={{color:grahaColor[g.id]}}>{grahaGlyph[g.id]} {g.name}</button></td><td>{g.rashi}</td><td>{degreeLabel(g.rashi_degree)}</td><td>{g.nakshatra} / {g.nakshatra_pada}</td><td>{houseNumber(g.longitude,chart.ascendant.longitude)}</td><td>{g.longitude.toFixed(5)}°</td><td>{g.retrograde?'Retrograde ℞':'Direct'}</td></tr>)}</tbody></table></div><p>Rahu and Ketu are lunar nodes. House 1 begins at 0° of the Lagna’s sign; the Lagna marker shows its exact degree.</p></section>;
}
