import { useEffect,useState } from 'react';
import { Feedback,memberAPI,useAction } from './shared';
export function MatrimonyModeration(){
 const action=useAction(),[reports,setReports]=useState<{id:number;handle:string;reason:string;details:{introduction:string}}[]>([]);
 const load=async()=>setReports(await memberAPI('/api/matrimony/moderation'));
 useEffect(()=>{void action.run(load);},[]);
 return <section className="card"><h2>Matrimony reports</h2>{reports.map(r=><article key={r.id}><h3>@{r.handle}</h3><p>{r.details.introduction}</p><p>{r.reason}</p><button disabled={action.busy} onClick={()=>void action.run(async()=>{await memberAPI('/api/matrimony/moderation','POST',{report:r.id,hide:true});await load();})}>Hide profile and photos</button><button disabled={action.busy} onClick={()=>void action.run(async()=>{await memberAPI('/api/matrimony/moderation','POST',{report:r.id,hide:false});await load();})}>Dismiss profile report</button></article>)}{!reports.length&&<p>No open profile reports.</p>}<Feedback {...action}/></section>;
}
