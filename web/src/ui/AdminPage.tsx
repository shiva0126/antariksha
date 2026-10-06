import {useEffect,useState} from 'react';
import {memberAPI} from './community/shared';
import {Moderation} from './community/Settings';
import {MatrimonyModeration,VerificationQueue} from './community/MatrimonyModeration';
import {PageHeader,Tabs} from '../ds';
import {ReadingLibrary} from './common/ReadingLibrary';
import './admin.css';
import {AdminAuthenticator} from '../features/admin/AdminAuthenticator';
import {Operations} from '../features/admin/Operations';

type User={id:string;handle:string;email:string;role:string;suspended:boolean;created_at:string;email_verified:boolean;sessions:number;community:boolean;matrimony:boolean};
type Summary={users:number;suspended:number;moderators:number;superadmins:number;active_sessions:number;open_post_reports:number;open_profile_reports:number};
type Audit={id:number;actor:string;action:string;subject:string;created_at:string;details:{handle?:string;reason?:string}};

export function AdminPage(){
 const [tab,setTab]=useState('Users'),[summary,setSummary]=useState<Summary>(),[error,setError]=useState('');
 const refresh=async()=>{setSummary(await memberAPI<Summary>('/api/admin/summary'));};
 useEffect(()=>{void refresh().catch(e=>setError(e.message));},[]);
 return <div className="page admin-page"><PageHeader kicker="Astrisk administration" title="Superadmin" description="Manage members, review reports and check the reading library."/>
 {error&&<p role="alert">{error}</p>}{summary&&<>
 <div className="grid-3"><article className="card"><h2>{summary.users} users</h2><p>{summary.suspended} suspended · {summary.active_sessions} active sessions</p></article><article className="card"><h2>{summary.moderators} moderators</h2><p>{summary.superadmins} superadmin accounts</p></article><article className="card"><h2>{summary.open_post_reports+summary.open_profile_reports} open reports</h2><p>{summary.open_post_reports} post reports · {summary.open_profile_reports} profile reports</p></article></div>
 <Tabs label="Administration sections" active={tab} onSelect={setTab} items={['Users','Operations','Security','Moderation','Photo verification','Reading library','Action log'].map(t=>({id:t,label:t}))}/>
 {tab==='Operations'&&<Operations/>}
 {tab==='Security'&&<AdminAuthenticator/>}
 {tab==='Users'&&<Users changed={refresh}/>}{tab==='Moderation'&&<><Moderation/><MatrimonyModeration/></>}{tab==='Photo verification'&&<VerificationQueue/>}{tab==='Reading library'&&<ReadingLibrary/>}{tab==='Action log'&&<ActionLog/>}
 </>}</div>;
}

function Users({changed}:{changed:()=>Promise<void>}){
 const [users,setUsers]=useState<User[]>([]),[q,setQ]=useState(''),[status,setStatus]=useState(''),[page,setPage]=useState(1),[busy,setBusy]=useState(false),[message,setMessage]=useState(''),[selected,setSelected]=useState<User>();
 const [action,setAction]=useState('revoke_sessions'),[role,setRole]=useState('member'),[reason,setReason]=useState(''),[password,setPassword]=useState(''),[confirm,setConfirm]=useState('');
 async function load(p:number){setBusy(true);setMessage('');setSelected(undefined);setPassword('');try{setUsers(await memberAPI<User[]>('/api/admin/users?'+new URLSearchParams({q,status,page:String(p)})));setPage(p);}catch(e){setUsers([]);setMessage(String(e));}finally{setBusy(false);}}
 useEffect(()=>{void load(1);},[]);
 function select(user:User){setSelected(user);setAction('revoke_sessions');setRole(user.role);setPassword('');setReason('');setConfirm('');setMessage('');}
 return <><form className="card" onSubmit={e=>{e.preventDefault();void load(1);}}><div className="grid-2"><label className="field">Search users<input value={q} maxLength={254} placeholder="Email or handle" onChange={e=>setQ(e.target.value)} disabled={busy}/></label><label className="field">Account status<select value={status} onChange={e=>setStatus(e.target.value)} disabled={busy}><option value="">All accounts</option><option value="active">Active</option><option value="suspended">Suspended</option></select></label></div><button disabled={busy}>Search users</button></form>
 <p role="status">{busy?'Loading…':message}</p>
 <div className="card table-card"><div className="table-scroll"><table className="planet-table"><caption>Registered users · page {page}</caption><thead><tr><th>Account</th><th>Role / status</th><th>Joined</th><th>Sessions</th><th>Action</th></tr></thead><tbody>{users.slice(0,25).map(u=><tr key={u.id}><td style={{overflowWrap:'anywhere'}}>@{u.handle}<br/>{u.email||'No email'}<br/><small>{u.email_verified?'Email verified':'Email unverified'}</small></td><td>{u.role} · {u.suspended?'Suspended':'Active'}</td><td>{new Date(u.created_at).toLocaleDateString()}</td><td>{u.sessions}</td><td><button disabled={busy||u.role==='superadmin'} onClick={()=>select(u)}>{u.role==='superadmin'?'Protected':'Manage'}</button></td></tr>)}</tbody></table></div>{!users.length&&!busy&&<p>No matching users.</p>}<div className="community-row"><button disabled={busy||page===1} onClick={()=>void load(page-1)}>Previous users</button><button disabled={busy||users.length<=25} onClick={()=>void load(page+1)}>Next users</button></div></div>
 {selected&&<form className="card" aria-label="Manage selected user" onSubmit={async e=>{e.preventDefault();setBusy(true);setMessage('');try{await memberAPI('/api/admin/users/'+selected.id+'/action','POST',{action,role,reason,password,confirm_handle:confirm});await load(page);await changed();setMessage('Account action saved. Sessions were revoked and the action was recorded.');}catch(e){setMessage(String(e));}finally{setPassword('');setBusy(false);}}}>
 <h2>Manage @{selected.handle}</h2><p>{selected.email}</p><fieldset disabled={busy} style={{border:0,padding:0}}>
 <label className="field">Account action<select value={action} onChange={e=>setAction(e.target.value)}><option value="revoke_sessions">Sign out all devices</option><option value={selected.suspended?'reactivate':'suspend'}>{selected.suspended?'Reactivate account':'Suspend account'}</option><option value="role">Change moderator role</option><option value="delete">Permanently delete account</option></select></label>
 {action==='role'&&<label className="field">New role<select value={role} onChange={e=>setRole(e.target.value)}><option value="member">Member</option><option value="moderator">Moderator</option></select></label>}
 {action==='suspend'&&<p>Suspension stops sign-in, signs out devices and pauses community and matrimony visibility.</p>}
 {action==='reactivate'&&<p>The member can sign in again. Community and matrimony remain paused until they choose to enable them.</p>}
 {action==='delete'&&<><p>Permanently removes this account, its posts, photos, messages, saved charts, conversations and owned family groups. The admin action remains in the log. This cannot be undone here.</p><label className="field">Confirm account handle<input required value={confirm} onChange={e=>setConfirm(e.target.value)} autoComplete="off"/></label></>}
 <label className="field">Reason<textarea required minLength={3} maxLength={500} value={reason} onChange={e=>setReason(e.target.value)}/></label><label className="field">Your current password<input type="password" required maxLength={72} autoComplete="current-password" value={password} onChange={e=>setPassword(e.target.value)}/></label>
 <button className="primary">Apply account action</button><button type="button" onClick={()=>{setSelected(undefined);setPassword('');}}>Cancel</button></fieldset></form>}
 </>;
}

function ActionLog(){
 const [rows,setRows]=useState<Audit[]>([]),[page,setPage]=useState(1),[busy,setBusy]=useState(false),[error,setError]=useState('');
 async function load(p:number){setBusy(true);setError('');try{setRows(await memberAPI('/api/admin/audit?page='+p));setPage(p);}catch(e){setError(String(e));}finally{setBusy(false);}}
 useEffect(()=>{void load(1);},[]);
 return <section className="card"><h2>Action log</h2>{error&&<p role="alert">{error}</p>}{rows.slice(0,25).map(a=><article key={a.id}><h3>{a.action.replaceAll('_',' ')}</h3><p>{new Date(a.created_at).toLocaleString()} · {a.actor}</p><p>Account: {a.details.handle||a.subject}</p>{a.details.reason&&<p>Reason: {a.details.reason}</p>}</article>)}{!rows.length&&<p>No recorded actions.</p>}<button disabled={busy||page===1} onClick={()=>void load(page-1)}>Previous actions</button><button disabled={busy||rows.length<=25} onClick={()=>void load(page+1)}>Next actions</button></section>;
}
