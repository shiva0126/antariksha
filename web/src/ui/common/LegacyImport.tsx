import {useState} from 'react';
import {loadProfiles,saveProfiles,newId,type Profile} from './profiles';

// Old device data is untrusted and has no authenticated owner. Never import
// automatically, copy session IDs, or turn chart subjects into member accounts.
export function legacyCandidates():Profile[]{
 const result:Profile[]=[];
 for(const key of ['antariksha.profiles','antariksha.birth']){
  try{
   const raw=localStorage.getItem(key);if(!raw||raw.length>500000)continue;
   const parsed:unknown=JSON.parse(raw);const rows=Array.isArray(parsed)?parsed:[parsed];
   for(const value of rows.slice(0,100)){
    if(!value||typeof value!=='object')continue;
    const p=value as Record<string,unknown>;
    if(typeof p.date!=='string'||!/^\d{4}-\d{2}-\d{2}$/.test(p.date)||new Date(p.date+'T00:00:00Z').toISOString().slice(0,10)!==p.date)continue;
    if(typeof p.time!=='string'||!/^([01]\d|2[0-3]):[0-5]\d$/.test(p.time))continue;
    if(typeof p.lat!=='number'||!Number.isFinite(p.lat)||Math.abs(p.lat)>90||typeof p.lon!=='number'||!Number.isFinite(p.lon)||Math.abs(p.lon)>180||typeof p.tz!=='string')continue;
    try{new Intl.DateTimeFormat('en',{timeZone:p.tz});}catch{continue;}
    const candidate={id:newId(),name:typeof p.name==='string'?p.name.slice(0,100):'Imported chart',place:typeof p.place==='string'?p.place.slice(0,200):'',date:p.date,time:p.time,lat:p.lat,lon:p.lon,tz:p.tz};
    if(!result.some(row=>sameChart(row,candidate)))result.push(candidate);
   }
  }catch{/* Malformed old data remains untouched. */}
 }
 return result;
}
const sameChart=(a:Profile,b:Profile)=>a.date===b.date&&a.time===b.time&&a.lat===b.lat&&a.lon===b.lon&&a.tz===b.tz;
export function LegacyImport(){
 const [rows,setRows]=useState<Profile[]|null>(null),[selected,setSelected]=useState<string[]>([]),[consent,setConsent]=useState(false),[message,setMessage]=useState('');
 return <details style={{margin:'1.5rem 0'}}><summary>Import charts saved before accounts existed</summary><p>Old charts on this device have no verified owner. Review only if this is your device and you have permission. Imports stay private to this account on this browser; no old conversations or social profiles are imported.</p>
 <button onClick={()=>{setRows(legacyCandidates());setSelected([]);setConsent(false);setMessage('');}}>Review old device charts</button>
 {rows&&<>{rows.length===0?<p>No valid legacy charts found on this device.</p>:<><ul>{rows.map(p=><li key={p.id}><label><input type="checkbox" checked={selected.includes(p.id)} onChange={e=>setSelected(old=>e.target.checked?[...old,p.id]:old.filter(id=>id!==p.id))}/>{p.name||'Unnamed chart'} · {p.date} {p.time} · {p.place||`${p.lat}, ${p.lon}`} · {p.tz}</label></li>)}</ul><label><input type="checkbox" checked={consent} onChange={e=>setConsent(e.target.checked)}/> These are my charts or I have permission to import them.</label><p><button disabled={!consent||!selected.length} onClick={()=>{const current=loadProfiles();const additions=rows.filter(p=>selected.includes(p.id)&&!current.profiles.some(old=>sameChart(old,p)));saveProfiles([...current.profiles,...additions],current.active||additions[0]?.id);setMessage(`Imported ${additions.length} chart(s). Existing duplicates were skipped. Original device data was preserved.`);setSelected([]);setConsent(false);}}>Import selected charts</button></p></> }</>}
 {message&&<p role="status">{message}</p>}</details>;
}
