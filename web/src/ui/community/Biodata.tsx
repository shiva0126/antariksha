import { useFeedback } from '../../ds';
import { useEffect, useState } from 'react';
import { Feedback, Field, memberAPI, useAction } from './shared';
import { OPTIONS } from '../matrimony/options';

export type Biodata = {
 display_name:string; profile_kind:string; introduction:string; city:string;
 occupation:string; education:string; languages:string; timeline:string; children:string;
 relocation:string; lifestyle:string; values:string; hobbies:string; family_about:string;
 social_links:string[]; min_age:number; max_age:number;
};
export const emptyBiodata:Biodata={display_name:'',profile_kind:'person',introduction:'',city:'',occupation:'',education:'',languages:'',timeline:'',children:'',relocation:'',lifestyle:'',values:'',hobbies:'',family_about:'',social_links:[],min_age:18,max_age:70};
const fields = [
 ['display_name','Display name',100],['introduction','Introduction',1000],['city','City',100],
 ['occupation','Occupation',200],['education','Education',200],['languages','Languages',200],
 ['hobbies','Hobbies and interests',500],['family_about','About the family',1000],
 ['lifestyle','Lifestyle',200],['values','Values',500],
] as const;
// Structured fields share the option lists of the member's own profile editor.
const selects = [['religion','Religion'],['mother_tongue','Mother tongue'],['diet','Diet'],['marital_status','Marital status'],['timeline','Marriage timeline'],['children','Children'],['relocation','Relocation']] as const;

export function BiodataFields({value,onChange}:{value:Biodata;onChange:(value:Biodata)=>void}) {
 return <div className="grid-2">
  <Field label="Profile type"><select value={value.profile_kind} onChange={e=>onChange({...value,profile_kind:e.target.value})}><option value="person">Person seeking a partner</option><option value="bride">Bride</option><option value="groom">Groom</option></select></Field>
  {fields.map(([key,label,max])=><Field key={key} label={label}><textarea maxLength={max} rows={max>200?3:1} value={value[key]} onChange={e=>onChange({...value,[key]:e.target.value})}/></Field>)}
  {selects.map(([key,label])=><Field key={key} label={label}><select value={(value as unknown as Record<string,string>)[key]??''} onChange={e=>onChange({...value,[key]:e.target.value})}><option value="">Not shared</option>{OPTIONS[key].map(([v,l])=><option key={v} value={v}>{l}</option>)}</select></Field>)}
  <Field label="Preferred minimum age"><input type="number" min={18} max={100} value={value.min_age} onChange={e=>onChange({...value,min_age:+e.target.value})}/></Field>
  <Field label="Preferred maximum age"><input type="number" min={18} max={100} value={value.max_age} onChange={e=>onChange({...value,max_age:+e.target.value})}/></Field>
  <Field label="Social profile links"><textarea rows={3} value={(value.social_links||[]).join('\n')} onChange={e=>onChange({...value,social_links:e.target.value.split('\n')})} placeholder="One HTTPS link per line"/></Field>
  <p className="small muted">Share only links the person has chosen to include. Links are self-declared.</p>
 </div>;
}
export const cleanBiodata=(value:Biodata):Biodata=>({...value,social_links:value.social_links.map(x=>x.trim()).filter(Boolean)});
export type MatrimonyPhoto={id:string;alt:string;published?:boolean};
export function PhotoGallery({photos,forOwner}:{photos:MatrimonyPhoto[];forOwner?:string}) {
 return <div className="biodata-gallery">{photos.map(p=><a key={p.id} href={`/api/matrimony/photos/${p.id}${forOwner?'?for='+encodeURIComponent(forOwner):''}`} target="_blank" rel="noopener noreferrer"><img loading="lazy" src={`/api/matrimony/photos/${p.id}${forOwner?'?for='+encodeURIComponent(forOwner):''}`} alt={p.alt||'Matrimony profile photo'}/></a>)}</div>;
}
export function PhotoEditor({photos,onChange,draftId}:{photos:MatrimonyPhoto[];onChange:(photos:MatrimonyPhoto[])=>void;draftId?:string}) {
 const action=useAction(),[file,setFile]=useState<File|null>(null),[alt,setAlt]=useState(''),[consent,setConsent]=useState(false),[reset,setReset]=useState(0);
 return <section aria-label="Matrimony photos"><h3>Photos</h3><p>Up to six JPEG or PNG photos, each under 8 MB. New uploads start private.</p>
  <div className="biodata-gallery">{photos.map(p=><figure key={p.id}><img loading="lazy" src={`/api/matrimony/photos/${p.id}`} alt={p.alt||'Your uploaded photo'}/><figcaption>{p.alt}</figcaption>
   {!draftId&&<label><input type="checkbox" checked={!!p.published} onChange={e=>onChange(photos.map(x=>x.id===p.id?{...x,published:e.target.checked}:x))}/> Include when I publish</label>}
   <button type="button" disabled={action.busy} onClick={()=>void action.run(async()=>{await memberAPI('/api/matrimony/photos/'+p.id,'DELETE');onChange(photos.filter(x=>x.id!==p.id));})}>Remove photo</button>
  </figure>)}</div>
  {photos.length<6&&<><Field label="Choose profile photo"><input key={reset} type="file" accept="image/jpeg,image/png" onChange={e=>setFile(e.target.files?.[0]||null)}/></Field>
  <Field label="Photo description"><input maxLength={300} value={alt} onChange={e=>setAlt(e.target.value)}/></Field>
  <label><input type="checkbox" checked={consent} onChange={e=>setConsent(e.target.checked)}/> I have permission from the adult pictured to store this photo.</label>
  <button type="button" disabled={action.busy||!file||!consent} onClick={()=>void action.run(async()=>{const body=new FormData();body.set('photo',file!);body.set('alt',alt);body.set('consent','true');if(draftId)body.set('draft_id',draftId);const result=await memberAPI('/api/matrimony/photos','POST',body);onChange([...photos,{id:result.id,alt,published:false}]);setFile(null);setAlt('');setConsent(false);setReset(x=>x+1);action.setMessage('Photo saved privately.');})}>Upload photo</button></>}
  <Feedback {...action}/>
 </section>;
}

type Draft={id:string;relationship:string;details:Biodata;status:'draft'|'pending';mine:boolean;creator_handle:string;recipient_handle?:string;photos:MatrimonyPhoto[]};
function DraftCard({draft,onChange,onImported}:{draft:Draft;onChange:()=>Promise<void>;onImported:()=>Promise<void>}) {
 const action=useAction(), fb = useFeedback(),[handle,setHandle]=useState('');
 return <article className="card"><h3>{draft.details.display_name}</h3><p>{draft.mine?`Prepared by you · ${draft.status==='draft'?'Private draft':'Awaiting review by @'+draft.recipient_handle}`:`Prepared by @${draft.creator_handle} (${draft.relationship})`}</p>
  <details><summary>Review biodata</summary><dl className="biodata-review">{fields.filter(([key])=>draft.details[key]).map(([key,label])=><div key={key}><dt>{label}</dt><dd>{draft.details[key]}</dd></div>)}<div><dt>Preferred age range</dt><dd>{draft.details.min_age}–{draft.details.max_age}</dd></div></dl>{draft.details.social_links?.map(link=><p key={link}><a href={link} target="_blank" rel="noopener noreferrer">{link}</a></p>)}</details>
  {draft.mine&&draft.status==='draft'?<><PhotoEditor photos={draft.photos} draftId={draft.id} onChange={()=>void onChange()}/>
   <Field label="Member's Astrisk handle"><input value={handle} onChange={e=>setHandle(e.target.value)} placeholder="Their handle after they register"/></Field>
   <p>The member will receive this in their biodata inbox and review the details and photos before publishing.</p>
   <button disabled={action.busy||!handle.trim()} onClick={()=>void action.run(async()=>{await memberAPI('/api/matrimony/drafts/'+draft.id,'POST',{action:'send',handle:handle.replace(/^@/,'')});await onChange();})}>Send for their review</button>
  </>:<PhotoGallery photos={draft.photos}/>}
  {!draft.mine&&<><p>Accepting replaces your current biodata and adds these photos privately. Your profile will be paused for you to review and publish.</p><button disabled={action.busy} onClick={async()=>{if(await fb.confirm({title:'Is this your biodata? Replace your current biodata with this draft and add its photos? Your profile will be paused.' }))void action.run(async()=>{await memberAPI('/api/matrimony/drafts/'+draft.id,'POST',{action:'accept',consent:true});await onChange();await onImported();});}}>Use this biodata privately</button></>}
  <button disabled={action.busy} onClick={async()=>{if(await fb.confirm({title:'Delete this draft and its photos?' }))void action.run(async()=>{await memberAPI('/api/matrimony/drafts/'+draft.id,'DELETE');await onChange();});}}>{draft.mine?'Delete draft':'Decline and delete'}</button><Feedback {...action}/>
 </article>;
}

export function FamilyBiodata({onImported}:{onImported:()=>Promise<void>}) {
 const action=useAction(),[drafts,setDrafts]=useState<Draft[]>([]),[details,setDetails]=useState<Biodata>({...emptyBiodata}),[relationship,setRelationship]=useState('parent'),[consent,setConsent]=useState(false);
 const load=async()=>setDrafts(await memberAPI('/api/matrimony/drafts'));
 useEffect(()=>{void action.run(load);},[]);
 return <section className="card"><p className="kicker">Prepared together · chosen personally</p><h2>Family biodata & review inbox</h2><p>Prepare a bride or groom's profile with their permission. Save a private draft, add their photos, then send it to their Astrisk handle. They decide whether to publish.</p>
  <details><summary>Prepare biodata for a family member</summary><form onSubmit={e=>{e.preventDefault();void action.run(async()=>{await memberAPI('/api/matrimony/drafts','POST',{details:cleanBiodata(details),relationship,consent});setDetails({...emptyBiodata});setConsent(false);await load();action.setMessage('Private draft created. Add photos below, then send it for review.');});}}>
   <Field label="Your relationship to them"><select value={relationship} onChange={e=>setRelationship(e.target.value)}>{['parent','sibling','relative','friend'].map(x=><option key={x}>{x}</option>)}</select></Field>
   <BiodataFields value={details} onChange={setDetails}/><label><input type="checkbox" required checked={consent} onChange={e=>setConsent(e.target.checked)}/> This person is an adult and has given me permission to prepare and store this biodata.</label><button disabled={action.busy||!details.display_name.trim()}>Create private biodata draft</button>
  </form></details><button disabled={action.busy} onClick={()=>void action.run(load)}>Refresh biodata inbox</button><Feedback {...action}/>
  {drafts.map(d=><DraftCard key={d.id} draft={d} onChange={load} onImported={onImported}/>)}
  {!drafts.length&&<p>No drafts or review requests yet.</p>}
 </section>;
}
