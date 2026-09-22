import type { ChartResponse, Graha } from '../api/types';
import { rashis, grahaColor } from '../astro/rashi';
import { signIndex, houseNumber } from '../astro/houses';
const cells = [[1,0],[2,0],[3,0],[3,1],[3,2],[3,3],[2,3],[1,3],[0,3],[0,2],[0,1],[0,0]];
export function SouthIndian({chart,onSelect}:{chart:ChartResponse;onSelect?:(g:Graha)=>void}) {
  const asc = signIndex(chart.ascendant.longitude);
  return <svg className="wheel traditional" viewBox="0 0 680 680" role="img" aria-label="South Indian Kundali with fixed zodiac signs">
    {cells.map(([col,row], sign) => {
      const x=20+col*160,y=20+row*160, planets=chart.grahas.filter(g=>signIndex(g.longitude)===sign);
      return <g key={sign} data-sign={sign+1}>
        <rect x={x} y={y} width="160" height="160" fill={asc===sign?'#102d36':'#0c1120'} stroke="#76613d"/>
        {asc===sign&&<path d={`M${x} ${y+32} L${x+32} ${y}`} stroke="#67e8f9" strokeWidth="3"/>}
        <text x={x+80} y={y+23} textAnchor="middle" fill="#d0b275" fontSize="13">{rashis[sign][0]} · H{houseNumber(sign*30,chart.ascendant.longitude)}</text>
        {asc===sign&&<text x={x+80} y={y+40} textAnchor="middle" fill="#67e8f9" fontSize="11">Lagna {chart.ascendant.degree.toFixed(2)}°</text>}
        {planets.map((g,i)=><text key={g.id} x={x+80} y={y+59+i*Math.min(18,90/Math.max(planets.length,1))} textAnchor="middle" fill={grahaColor[g.id]} fontSize="12" role="button" tabIndex={0} onClick={()=>onSelect?.(g)} onKeyDown={e=>e.key==='Enter'&&onSelect?.(g)}>{g.name} {g.rashi_degree.toFixed(2)}°{g.retrograde?' ℞':''}</text>)}
      </g>;
    })}
    <text x="340" y="320" textAnchor="middle" fill="#d8c293" fontSize="25">Rashi • D1</text>
    <text x="340" y="353" textAnchor="middle" fill="#929cb5" fontSize="13">South Indian · Fixed signs</text>
    <text x="340" y="378" textAnchor="middle" fill="#67e8f9" fontSize="13">Lagna: {chart.ascendant.rashi}</text>
  </svg>;
}
