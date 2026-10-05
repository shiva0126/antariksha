import { useEffect,useState } from 'react';
import { Feedback,memberAPI,useAction } from './shared';
export function MatrimonyModeration(){
 const action=useAction(),[reports,setReports]=useState<{id:number;handle:string;reason:string;details:{introduction:string}}[]>([]);
 const load=async()=>setReports(await memberAPI('/api/matrimony/moderation'));
 useEffect(()=>{void action.run(load);},[]);
 return <section className="card"><h2>Matrimony reports</h2>{reports.map(r=><article key={r.id}><h3>@{r.handle}</h3><p>{r.details.introduction}</p><p>{r.reason}</p><button disabled={action.busy} onClick={()=>void action.run(async()=>{await memberAPI('/api/matrimony/moderation','POST',{report:r.id,hide:true});await load();})}>Hide profile and photos</button><button disabled={action.busy} onClick={()=>void action.run(async()=>{await memberAPI('/api/matrimony/moderation','POST',{report:r.id,hide:false});await load();})}>Dismiss profile report</button></article>)}{!reports.length&&<p>No open profile reports.</p>}<Feedback {...action}/></section>;
}

type Pending={account_id:string;handle:string;display_name:string|null;created_at:string;photos:string[]};
/** Moderators compare a live selfie with published photos; the selfie is deleted after review. */
export function VerificationQueue(){
 const action=useAction(),[items,setItems]=useState<Pending[]>([]);
 const load=async()=>setItems(await memberAPI('/api/matrimony/verifications'));
 useEffect(()=>{void action.run(load);},[]);
 const review=(id:string,approve:boolean)=>void action.run(async()=>{await memberAPI('/api/matrimony/verifications/'+id,'POST',{approve,note:approve?'':'Selfie did not match the profile photos'});await load();});
 return <section className="card"><h2>Photo verification</h2><p className="muted">Approve only when the selfie clearly shows the same person as the published photos. The selfie is deleted after your decision.</p>
  {items.map(v=><article key={v.account_id} className="verify-item"><h3>{v.display_name||'@'+v.handle}</h3><div className="verify-photos"><figure><img src={`/api/matrimony/verifications/${v.account_id}/photo`} alt="Selfie for verification"/><figcaption>Selfie</figcaption></figure>{v.photos.map(p=><figure key={p}><img src={`/api/matrimony/photos/${p}`} alt="Published profile photo"/><figcaption>Profile photo</figcaption></figure>)}</div>
  <button disabled={action.busy} onClick={()=>review(v.account_id,true)}>Approve: same person</button><button disabled={action.busy} onClick={()=>review(v.account_id,false)}>Reject</button></article>)}
  {!items.length&&<p>No selfies waiting.</p>}<Feedback {...action}/></section>;
}
