import { useEffect, useState } from 'react';
import { memberAPI } from './community/shared';
import { localDataKeys, setProfileOwner } from './common/profiles';
import {LegacyImport} from './common/LegacyImport';
const fields = {name:'Display name',city:'City',bio:'About me',hobbies:'Hobbies and interests',goals:'What I want to learn or experience'};
export function AccountPage(){
 const [member,setMember]=useState<any>(),[profile,setProfile]=useState<Record<string,string>>({}),[message,setMessage]=useState(''),[busy,setBusy]=useState(false);
 useEffect(()=>{memberAPI('/api/me').then(me=>{setMember(me);setProfile(me.profile||{});}).catch(e=>setMessage(e.message));},[]);
 async function action(fn:()=>Promise<void>){setBusy(true);setMessage('');try{await fn();}catch(e){setMessage(e instanceof Error?e.message:'Request failed');}finally{setBusy(false);}}
 function signedOut(){setProfileOwner('signed-out');window.dispatchEvent(new Event('antariksha-signed-out'));}
 return <section style={{maxWidth:680,margin:'3rem auto',padding:'0 1.5rem'}}><p>YOUR PRIVATE SPACE</p><h1>My Antariksha</h1>
 {member&&<><p>Signed in as <strong>{member.email||member.handle}</strong> · Private</p><p>Birth date: {member.birth_date||'Not recorded'} · Time: {member.birth_time?.slice(0,5)||'Not recorded'}</p><p>These optional details stay separate from login. Community sharing is controlled in <a href="#community">Community → Profile & interests</a>.</p>
 <form onSubmit={e=>{e.preventDefault();void action(async()=>{await memberAPI('/api/me','PUT',profile);setMessage('Private profile saved.');});}}>
 {Object.entries(fields).map(([key,label])=><label key={key} style={{display:'grid',gap:8,marginBottom:16}}>{label}<textarea rows={key==='name'||key==='city'?1:3} maxLength={key==='name'||key==='city'?100:key==='bio'?2000:1000} value={profile[key]||''} onChange={e=>setProfile({...profile,[key]:e.target.value})}/></label>)}<button disabled={busy}>Save private profile</button></form>
 <p><a href="#community">Security, recovery key and full export</a></p><button disabled={busy} onClick={()=>void action(async()=>{await memberAPI('/api/auth/logout','POST',{});signedOut();})}>Sign out</button>
 <details><summary>Delete this account</summary><p>Permanently removes this account, posts, owned family groups, messages and conversations. Copies others saved outside the app cannot be removed.</p><button disabled={busy} onClick={()=>{if(confirm('Permanently delete this account and its data?'))void action(async()=>{await memberAPI('/api/me','DELETE');for(const key of localDataKeys())localStorage.removeItem(key);signedOut();});}}>Delete my account</button></details></>}
 <LegacyImport/><p role="status">{message}</p></section>;
}
