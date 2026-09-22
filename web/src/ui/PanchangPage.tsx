import {useEffect,useState} from 'react';
import {locations} from './locations';
import {fetchJSON,type MonthDay,type PanchangDay} from '../api/panchang';
import {getChart} from '../api/client';
import type {ChartResponse} from '../api/types';
import {SouthIndian} from '../chart2d/SouthIndian';
import {PlanetTable} from './PlanetTable';
function today(){return new Intl.DateTimeFormat('en-CA',{timeZone:'Asia/Kolkata',year:'numeric',month:'2-digit',day:'2-digit'}).format(new Date())}
export function PanchangPage({mode}:{mode:'day'|'month'}) {
 const [date,setDate]=useState(today),[city,setCity]=useState(0),[month,setMonth]=useState(()=>today().slice(0,7));
 const [time,setTime]=useState('12:00'),[day,setDay]=useState<PanchangDay>(),[days,setDays]=useState<MonthDay[]>([]),[chart,setChart]=useState<ChartResponse>();
 const [error,setError]=useState(''),[busy,setBusy]=useState(false),[chartError,setChartError]=useState('');
 const location=locations[city];
 useEffect(()=>{const ctrl=new AbortController();setBusy(true);setDay(undefined);setError('');
 const q=new URLSearchParams({date,lat:String(location.lat),lon:String(location.lon),tz:location.tz});
 fetchJSON<PanchangDay>('/api/panchang?'+q,ctrl.signal).then(setDay).catch(e=>{if(e.name!=='AbortError')setError(e.message)}).finally(()=>{if(!ctrl.signal.aborted)setBusy(false)});
 return()=>ctrl.abort();},[date,location]);
 useEffect(()=>{if(mode!=='month')return;const ctrl=new AbortController();setDays([]);
 const [year,m]=month.split('-');const q=new URLSearchParams({year,month:m,lat:String(location.lat),lon:String(location.lon),tz:location.tz});
 fetchJSON<MonthDay[]>('/api/month?'+q,ctrl.signal).then(setDays).catch(e=>{if(e.name!=='AbortError')setError(e.message)});
 return()=>ctrl.abort();},[month,location,mode]);
 useEffect(()=>{const ctrl=new AbortController();setChart(undefined);setChartError('');
 getChart({date,time,...location},ctrl.signal).then(setChart).catch(e=>{if(e.name!=='AbortError')setChartError(e.message)});
 return()=>ctrl.abort();},[date,time,location]);
 const start=new Date(month+'-01T12:00:00Z').getUTCDay();
 const choose=(value:string)=>{setDate(value);setMonth(value.slice(0,7))};
 return <div className="panchang-page">
 <div className="page-heading"><div><p className="kicker">THE HINDU CALENDAR</p><h1>{mode==='month'?'Panchang calendar':'Daily Panchang'}</h1><p>Five limbs at local sunrise. Explore the grahas at your chosen time.</p></div><div className="panchang-filters"><label>Location<select value={city} onChange={e=>setCity(Number(e.target.value))}>{locations.map((x,i)=><option value={i} key={x.name}>{x.name}</option>)}</select></label><label>{mode==='month'?'Month':'Date'}<input aria-label={mode==='month'?'Calendar month':'Panchang date'} type={mode==='month'?'month':'date'} value={mode==='month'?month:date} onChange={e=>{if(e.target.value)choose(mode==='month'?e.target.value+'-01':e.target.value)}}/></label></div></div>
 {error&&<p role="alert" className="error">{error}</p>}
 {mode==='month'&&<div className="calendar-grid" aria-label="Hindu month calendar">{['Sun','Mon','Tue','Wed','Thu','Fri','Sat'].map(d=><div className="weekday" key={d}>{d}</div>)}{Array.from({length:start},(_,i)=><div key={'blank'+i} className="calendar-empty"/>)}{days.map(d=><button className={'calendar-day '+(d.date===date?'selected':'')} key={d.date} onClick={()=>setDate(d.date)} aria-label={d.date+' '+d.tithi}><b>{Number(d.date.slice(-2))}</b><span>{d.tithi}</span><small>{d.paksha} paksha</small>{d.festivals.map(f=><em key={f}>{f}</em>)}</button>)}{!days.length&&!error&&<p>Loading the month…</p>}</div>}
 {busy&&<p role="status">Calculating Panchang…</p>}
 {day&&<section className="day-details"><div className="day-title"><h2>{day.date} · {day.vaara}</h2><span>{day.lunar_month} (Amanta) · {day.paksha} paksha · {location.name} · {location.tz}</span></div>
 <div className="limb-grid">{(['tithi','nakshatra','yoga','karana'] as const).map(key=><article key={key}><span>{key}</span><h3>{day[key].name}</h3><p>Until {day[key].ends_at}</p></article>)}</div>
 <div className="timing-grid"><article><h3>Sun & Moon</h3>{[['Sunrise',day.sunrise],['Sunset',day.sunset],['Moonrise',day.moonrise||'No rise this Panchang day'],['Moonset',day.moonset||'No set this Panchang day']].map(([k,v])=><p key={k}><span>{k}</span><b>{v}</b></p>)}</article><article><h3>Daily windows</h3>{(['rahu_kaal','yamaganda','gulika','abhijit'] as const).map(k=><p key={k}><span>{k.replace('_',' ')}</span><b>{day[k].start} – {day[k].end}</b></p>)}</article><article><h3>Choghadiya</h3><div className="chog-list">{day.choghadiya.map((w,i)=><p key={i}><span>{i<8?'Day':'Night'} · {w.name}</span><b>{w.start} – {w.end}</b></p>)}</div></article></div>
 </section>}
 <section className="transits"><div className="day-title"><div><h2>Graha positions on {date}</h2><p>Geocentric sidereal positions at the selected local time; these are transits, not a birth chart.</p></div><label>Time · {location.tz}<input type="time" value={time} aria-label="Planet positions time" onChange={e=>e.target.value&&setTime(e.target.value)}/></label></div>{chartError&&<p className="error">{chartError}</p>}{chart?<><div className="transit-chart"><SouthIndian chart={chart}/></div><PlanetTable chart={chart}/></>:!chartError&&<p>Calculating planet positions…</p>}</section>
 </div>;
}
