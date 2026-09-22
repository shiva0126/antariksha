import type {ChartResponse,Graha} from '../api/types';
import {grahaColor,grahaGlyph,rashis} from '../astro/rashi';
import {wheelPoint} from '../astro/mapping';
import {houseNumber} from '../astro/houses';
export function CircularWheel({chart,onSelect}:{chart:ChartResponse;onSelect:(g:Graha)=>void}) {
 const c=340, sorted=[...chart.grahas].sort((a,b)=>a.longitude-b.longitude);
 const lanes=new Map<string,number>();
 for(const g of sorted){let lane=0;while(sorted.some(p=>lanes.get(p.id)===lane&&Math.min(Math.abs(g.longitude-p.longitude),360-Math.abs(g.longitude-p.longitude))<13))lane++;lanes.set(g.id,lane);}
 const lagna=wheelPoint(chart.ascendant.longitude,310,c,c);
 return <svg className="wheel" viewBox="0 0 680 680" role="img" aria-label="Kundali wheel: exact sidereal longitudes">
 <circle cx={c} cy={c} r="320" fill="#090e1c" stroke="#987d4c"/>
 <circle cx={c} cy={c} r="268" fill="none" stroke="#65583c"/>
 {rashis.map((r,i)=>{const a=wheelPoint(i*30,320,c,c),b=wheelPoint(i*30,80,c,c),p=wheelPoint(i*30+15,294,c,c),h=wheelPoint(i*30+15,95,c,c);return <g key={r[0]}><line x1={a.x} y1={a.y} x2={b.x} y2={b.y} stroke="#433e3a"/><text x={p.x} y={p.y-4} textAnchor="middle" fill="#cfb477" fontSize="18">{r[2]}</text><text x={p.x} y={p.y+13} textAnchor="middle" fill="#b1a792" fontSize="10">{r[0]}</text><text x={h.x} y={h.y} textAnchor="middle" fill="#677b99" fontSize="11">H{houseNumber(i*30,chart.ascendant.longitude)}</text></g>})}
 {Array.from({length:108},(_,i)=>{const a=wheelPoint(i*360/108,268,c,c),b=wheelPoint(i*360/108,i%4?264:257,c,c);return <line key={i} x1={a.x} y1={a.y} x2={b.x} y2={b.y} stroke="#5b6682"/>})}
 <line x1={c} y1={c} x2={lagna.x} y2={lagna.y} stroke="#67e8f9" strokeDasharray="4 5"/>
 <circle cx={lagna.x} cy={lagna.y} r="5" fill="#67e8f9"/>
 {sorted.map(g=>{const lane=lanes.get(g.id)??0,p=wheelPoint(g.longitude,239-lane*29,c,c),a=wheelPoint(g.longitude,267,c,c);return <g key={g.id} className="planet" data-longitude={g.longitude} role="button" tabIndex={0} aria-label={g.name+' '+g.rashi+' '+g.rashi_degree.toFixed(2)+' degrees'} onClick={()=>onSelect(g)} onKeyDown={e=>e.key==='Enter'&&onSelect(g)}><title>{g.name} · {g.rashi} {g.rashi_degree.toFixed(3)}° · House {houseNumber(g.longitude,chart.ascendant.longitude)}</title><line x1={a.x} y1={a.y} x2={p.x} y2={p.y} stroke={grahaColor[g.id]} opacity=".55"/><circle cx={a.x} cy={a.y} r="2.5" fill={grahaColor[g.id]}/><circle cx={p.x} cy={p.y} r="13" fill="#101726" stroke={grahaColor[g.id]}/><text x={p.x} y={p.y+1} fill={grahaColor[g.id]} textAnchor="middle" dominantBaseline="middle" fontSize="20">{grahaGlyph[g.id]}</text><text x={p.x} y={p.y+24} fill={grahaColor[g.id]} textAnchor="middle" fontSize="9">{g.rashi_degree.toFixed(1)}°{g.retrograde?' ℞':''}</text></g>})}
 <circle cx={c} cy={c} r="67" fill="#0b1323" stroke="#28394f"/><text x={c} y={c-17} textAnchor="middle" fill="#67e8f9" fontSize="10">LAGNA</text><text x={c} y={c+5} textAnchor="middle" fill="#e1dacb" fontSize="15">{chart.ascendant.rashi}</text><text x={c} y={c+24} textAnchor="middle" fill="#9babc3" fontSize="12">{chart.ascendant.degree.toFixed(3)}°</text></svg>;
}
