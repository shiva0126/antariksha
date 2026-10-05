import {useEffect,useState} from 'react';
import {memberAPI,Feedback,useAction} from './shared';
type Notice={id:number;kind:'follow'|'family'|'interest'|'message'|'accepted';handle:string;created_at:string;read:boolean};
const labels={follow:'sent you a follow request',family:'invited you to a private family group',interest:'sent a matrimony interest',message:'sent you a message',accepted:'accepted your matrimony interest'};
const destinations={follow:'People',family:'Family',interest:'Matrimony',message:'Matrimony',accepted:'Matrimony'};
export function Notifications({navigate}:{navigate:(tab:string)=>void}){
 const [items,setItems]=useState<Notice[]>([]),[more,setMore]=useState(false);const action=useAction();
 async function load(before?:number){const rows=await memberAPI<Notice[]>('/api/me/notifications'+(before?`?before=${before}`:''));setItems(old=>before?[...old,...rows]:rows);setMore(rows.length===50);}
 useEffect(()=>{void action.run(()=>load());},[]);
 async function update(id:number,operation:string){await memberAPI(`/api/me/notifications/${id}`,'POST',{action:operation});setItems(old=>operation==='dismiss'?old.filter(n=>n.id!==id):old.map(n=>n.id===id?{...n,read:true}:n));}
 return <section className="card"><p className="kicker">Your updates</p><h2>Notifications</h2><p>Private invitations and messages. Revoked or blocked connections disappear automatically. No message contents are included here.</p><button disabled={action.busy} onClick={()=>void action.run(()=>load())}>Refresh notifications</button><Feedback {...action}/>
 {!items.length&&!action.busy&&<p>No notifications yet.</p>}
 <ul>{items.map(n=><li key={n.id} style={{padding:'1rem 0'}}><strong>@{n.handle}</strong> {labels[n.kind]}{!n.read&&<span> · New</span>}<p className="muted small">{new Date(n.created_at).toLocaleString()}</p><div className="community-actions"><button disabled={action.busy} onClick={()=>void action.run(async()=>{await update(n.id,'read');navigate(destinations[n.kind]);})}>Open {destinations[n.kind]}</button>{!n.read&&<button disabled={action.busy} onClick={()=>void action.run(()=>update(n.id,'read'))}>Mark read</button>}<button disabled={action.busy} onClick={()=>void action.run(()=>update(n.id,'dismiss'))}>Dismiss</button></div></li>)}</ul>
 {more&&<button disabled={action.busy} onClick={()=>void action.run(()=>load(items[items.length-1].id))}>Older notifications</button>}</section>;
}
