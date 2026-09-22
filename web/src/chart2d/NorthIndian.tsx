import type { ChartResponse, Graha } from '../api/types';
import { rashis, grahaColor } from '../astro/rashi';
import { signIndex, houseNumber } from '../astro/houses';
// Counterclockwise fixed houses; first house is the upper central diamond.
const positions = [[340,170],[180,66],[76,172],[174,334],[76,490],[180,590],[340,494],[500,590],[604,490],[506,334],[604,172],[500,66]];
export function NorthIndian({chart,onSelect}:{chart:ChartResponse;onSelect?:(g:Graha)=>void}) {
  const asc=signIndex(chart.ascendant.longitude);
  return <svg className="wheel traditional" viewBox="0 0 680 680" role="img" aria-label="North Indian Kundali with fixed houses">
    <rect x="20" y="20" width="640" height="640" fill="#0c1120" stroke="#a4864f"/>
    <path d="M20 20L660 660M660 20L20 660M340 20L660 340L340 660L20 340Z" fill="none" stroke="#8c744a"/>
    {positions.map(([x,y],i)=>{
      const sign=(asc+i)%12, planets=chart.grahas.filter(g=>houseNumber(g.longitude,chart.ascendant.longitude)===i+1);
      return <g key={i} data-house={i+1}>
        <text x={x} y={y} textAnchor="middle" fill={i===0?'#67e8f9':'#d0b275'} fontSize="12">H{i+1} · {rashis[sign][0]}</text>
        <text x={x} y={y+15} textAnchor="middle" fill="#929cb5" fontSize="10">Sign {sign+1}{i===0?' · Lagna':''}</text>
        {planets.map((g,j)=><text key={g.id} x={x} y={y+33+j*Math.min(16,65/Math.max(planets.length,1))} textAnchor="middle" fill={grahaColor[g.id]} fontSize="11" role="button" tabIndex={0} onClick={()=>onSelect?.(g)} onKeyDown={e=>e.key==='Enter'&&onSelect?.(g)}>{g.name} {g.rashi_degree.toFixed(2)}°{g.retrograde?' ℞':''}</text>)}
      </g>;
    })}
  </svg>;
}
