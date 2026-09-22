import {useState} from 'react';
import type {ChartInput} from '../api/types';
import {locations} from './locations';
export function BirthForm({onSubmit,busy,initial}:{onSubmit:(x:ChartInput)=>void;busy:boolean;initial?:ChartInput}) {
 const [town,setTown]=useState(()=>initial?locations.find(x=>x.lat===initial.lat&&x.lon===initial.lon)?.name??'Custom':'Bengaluru');
 const city=locations.find(x=>x.name===town)??locations[0];
 function submit(e:React.FormEvent<HTMLFormElement>){e.preventDefault();const f=new FormData(e.currentTarget);onSubmit({date:String(f.get('date')),time:String(f.get('time')),lat:town==='Custom'?Number(f.get('lat')):city.lat,lon:town==='Custom'?Number(f.get('lon')):city.lon,tz:town==='Custom'?String(f.get('tz')):city.tz})}
 return <form className="birth-form" onSubmit={submit}>
 <label><span>Birth date</span><input required name="date" type="date" defaultValue={initial?.date??'1996-05-14'}/></label>
 <label><span>Exact time</span><input required name="time" type="time" defaultValue={initial?.time??'10:15'}/><small>Enter local clock time, including any daylight saving.</small></label>
 <label><span>Birthplace</span><select value={town} onChange={e=>setTown(e.target.value)}>{locations.map(c=><option key={c.name}>{c.name}</option>)}<option value="Custom">Custom coordinates</option></select><small>{town==='Custom'?'Enter your precise birthplace below':city.lat.toFixed(4)+'°, '+city.lon.toFixed(4)+'° · '+city.tz}</small></label>
 {town==='Custom'&&<><label><span>Latitude</span><input name="lat" type="number" step="any" min="-90" max="90" required defaultValue={initial?.lat??12.9716}/></label><label><span>Longitude</span><input name="lon" type="number" step="any" min="-180" max="180" required defaultValue={initial?.lon??77.5946}/></label><label><span>IANA timezone</span><input name="tz" required defaultValue={initial?.tz??'Asia/Kolkata'}/></label></>}
 <button disabled={busy}><i/>{busy?'Reading the sky…':'Reveal my sky'}</button></form>;
}
