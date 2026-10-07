import { useFeedback } from '../../ds';
import { useEffect,useId,useState } from 'react';
import { Feedback,Field,memberAPI,useAction } from './shared';

type Source={id:string;platform:string;url:string;text:string};
type Profile={name:string;motto:string;avatar:string;accent:string;backdrop:string;interests:string[];goals:string;values:string;communication:string;weekly_minutes:number;sources:Source[]};
type Idea={interest:string;activity:string;together:string;reason:string};
type Saved={account_id:string;revision:number;profile:Profile;ideas:Idea[];hobby_options:Idea[]};
const fresh=():Profile=>({name:'',motto:'',avatar:'star',accent:'#d6b467',backdrop:'orbits',interests:[],goals:'',values:'',communication:'',weekly_minutes:30,sources:[]});
function sourceLinkOK(value:string){if(!value)return true;try{const u=new URL(value);return u.protocol==='https:'&&!u.username&&!u.password;}catch{return false;}}

function CharacterPortrait({profile:p}:{profile:Profile}) {
 const gradient=useId();
 return <svg viewBox="0 0 260 260" className="character-portrait" role="img" aria-label={`${p.avatar} character in ${p.accent}`}>
  <defs><radialGradient id={gradient}><stop stopColor={p.accent} stopOpacity=".24"/><stop offset="1" stopColor="#080c17"/></radialGradient></defs>
  <circle cx="130" cy="130" r="125" fill={`url(#${gradient})`} stroke={p.accent} strokeOpacity=".3"/>
  {p.backdrop==='orbits'&&[0,60,120].map(angle=><ellipse key={angle} cx="130" cy="130" rx="110" ry="50" transform={`rotate(${angle} 130 130)`} fill="none" stroke={p.accent} strokeOpacity=".3"/>)}
  {p.backdrop==='rays'&&Array.from({length:12},(_,i)=><path key={i} d="M130 28v23" transform={`rotate(${i*30} 130 130)`} stroke={p.accent} strokeWidth="2"/>)}
  <circle cx="130" cy="130" r="57" fill="#111727" stroke={p.accent}/>
  <text x="130" y="153" textAnchor="middle" fontSize="68" fill={p.accent}>{({sun:'☀',moon:'☾',star:'✧',leaf:'❧',mountain:'△'} as Record<string,string>)[p.avatar]}</text>
 </svg>;
}

export function MyCharacter(){
 const action=useAction(), fb = useFeedback(),[saved,setSaved]=useState<Saved>(),[profile,setProfile]=useState<Profile>(fresh),[consent,setConsent]=useState(false),[custom,setCustom]=useState('');
 const [source,setSource]=useState<Omit<Source,'id'>>({platform:'instagram',url:'',text:''}),[sourceConsent,setSourceConsent]=useState(false),[topics,setTopics]=useState<string[]|null>(null),[selected,setSelected]=useState<string[]>([]);
 const load=async()=>{const data=await memberAPI<Saved>('/api/me/character');setSaved(data);setProfile(data.profile);setConsent(false);};
 useEffect(()=>{void action.run(load);},[]);
 const toggle=(interest:string)=>setProfile(p=>({...p,interests:p.interests.includes(interest)?p.interests.filter(x=>x!==interest):p.interests.length<20?[...p.interests,interest]:p.interests}));
 return <div className="character-layout"><aside className="card character-preview"><p className="kicker">Your own constellation</p><CharacterPortrait profile={profile}/><h2>{profile.name||'Your character'}</h2><p>{profile.motto||'A little expression of who you choose to be.'}</p><div className="chips">{profile.interests.map(x=><span className="tag" key={x}>{x}</span>)}</div><p className="small muted">Private creative profile. You choose what to share in your community and matrimony introductions.</p></aside>
 <div><section className="card"><h2>My character & interests</h2><p>Choose a look, describe your goals and find activities you might enjoy. Suggestions use the interests you select here.</p>
 <form onSubmit={e=>{e.preventDefault();void action.run(async()=>{const result=await memberAPI<{revision:number;ideas:Idea[]}>('/api/me/character','PUT',{account_id:saved!.account_id,revision:saved!.revision,profile,consent});setSaved({...saved!,revision:result.revision,profile,ideas:result.ideas});action.setMessage('Private character saved.');});}}><fieldset className="contents" disabled={!saved} aria-busy={!saved}>
 <div className="grid-2"><Field label="Character name"><input maxLength={100} value={profile.name} onChange={e=>setProfile({...profile,name:e.target.value})}/></Field><Field label="Personal motto"><input maxLength={200} value={profile.motto} onChange={e=>setProfile({...profile,motto:e.target.value})}/></Field>
 <Field label="Character symbol"><select value={profile.avatar} onChange={e=>setProfile({...profile,avatar:e.target.value})}>{['sun','moon','star','leaf','mountain'].map(v=><option key={v}>{v}</option>)}</select></Field><Field label="Character colour"><input type="color" value={profile.accent} onChange={e=>setProfile({...profile,accent:e.target.value})}/></Field>
 <Field label="Character backdrop"><select value={profile.backdrop} onChange={e=>setProfile({...profile,backdrop:e.target.value})}>{['orbits','rays','plain'].map(v=><option key={v}>{v}</option>)}</select></Field>
 <Field label="Minutes each week for a hobby"><input type="number" min={15} max={600} value={profile.weekly_minutes} onChange={e=>setProfile({...profile,weekly_minutes:+e.target.value})}/></Field></div>
 <fieldset><legend>Interests I choose</legend><div className="interest-choices">{saved?.hobby_options.map(h=><label key={h.interest}><input type="checkbox" checked={profile.interests.includes(h.interest)} onChange={()=>toggle(h.interest)}/>{h.interest}</label>)}</div></fieldset>
 <Field label="Another interest"><input maxLength={40} value={custom} onChange={e=>setCustom(e.target.value)}/></Field><button type="button" disabled={!custom.trim()||profile.interests.length>=20} onClick={()=>{const v=custom.trim().toLowerCase();if(!profile.interests.includes(v))setProfile({...profile,interests:[...profile.interests,v]});setCustom('');}}>Add interest</button>
 <div className="chips">{profile.interests.map(x=><button type="button" key={x} onClick={()=>toggle(x)} aria-label={`Remove interest ${x}`}>{x} ×</button>)}</div>
 <Field label="Goals and things I want to learn"><textarea maxLength={1000} value={profile.goals} onChange={e=>setProfile({...profile,goals:e.target.value})}/></Field><Field label="Values in my own words"><textarea maxLength={1000} value={profile.values} onChange={e=>setProfile({...profile,values:e.target.value})}/></Field><Field label="How I like to communicate"><textarea maxLength={500} value={profile.communication} onChange={e=>setProfile({...profile,communication:e.target.value})}/></Field>
 <label><input type="checkbox" required checked={consent} onChange={e=>setConsent(e.target.checked)}/> Save these details and my chosen source notes privately in my account.</label><button disabled={action.busy||!saved}>Save private character</button>
 <button type="button" disabled={action.busy} onClick={async()=>{if(await fb.confirm({title:'Reload your saved character? Unsaved edits will be replaced.' }))void action.run(load);}}>Reload saved character</button><Feedback {...action}/>
 </fieldset></form></section>
 <section className="card"><h2>Bring your own social story</h2><p>Paste a bio or post you wrote. Review the interest words it mentions and select those that describe you. A profile link alone does not import posts.</p>
 <p className="small muted">Instagram and LinkedIn connections are not enabled yet. These notes are labelled user-provided and remain private.</p>
 <Field label="Source platform"><select value={source.platform} onChange={e=>setSource({...source,platform:e.target.value})}><option value="instagram">Instagram</option><option value="linkedin">LinkedIn</option><option value="own">My own notes</option></select></Field>
 <Field label="Source link (optional)"><input type="url" maxLength={300} value={source.url} onChange={e=>setSource({...source,url:e.target.value})} placeholder="https://"/></Field>
 <Field label="My bio or post text"><textarea maxLength={2000} rows={5} value={source.text} onChange={e=>{setSource({...source,text:e.target.value});setTopics(null);setSelected([]);}}/></Field>
 <label><input type="checkbox" checked={sourceConsent} onChange={e=>setSourceConsent(e.target.checked)}/> I wrote this text and want to use it for my private profile.</label>
 <button disabled={action.busy||!sourceConsent||!source.text.trim()} onClick={()=>void action.run(async()=>{const result=await memberAPI<{mentioned_topics:string[]}>('/api/me/character/preview','POST',{text:source.text});setTopics(result.mentioned_topics);setSelected([]);})}>Preview mentioned interests</button>
 {topics&&<div><p>{topics.length?'Select the mentioned interests you actually enjoy:':'No listed interest words found. You can add your interests above.'}</p>{topics.map(t=><label key={t}><input type="checkbox" checked={selected.includes(t)} onChange={e=>setSelected(e.target.checked?[...selected,t]:selected.filter(x=>x!==t))}/>{t}</label>)}<button disabled={!sourceConsent||!sourceLinkOK(source.url)||profile.sources.length>=8||new Set([...profile.interests,...selected]).size>20} onClick={()=>{setProfile({...profile,sources:[...profile.sources,{...source,id:crypto.randomUUID()}],interests:[...new Set([...profile.interests,...selected])]});setSource({platform:source.platform,url:'',text:''});setTopics(null);setSelected([]);setSourceConsent(false);action.setMessage('Source added to your draft. Save your private character to keep it.');}}>Add reviewed note to my draft</button></div>}
 {profile.sources.map(s=><article className="character-source" key={s.id}><strong>{s.platform} · User-provided</strong><p>{s.text}</p>{s.url&&<a href={s.url} target="_blank" rel="noopener noreferrer">Source link</a>}<button onClick={()=>{setProfile({...profile,sources:profile.sources.filter(x=>x.id!==s.id)});action.setMessage('Source removed from this draft. Save to remove it from your account.');}}>Remove source note</button></article>)}
 </section>
 <section className="card"><h2>Ideas for your interests</h2><p>Based on your last saved choices. Pick what fits your circumstances.</p>{!saved?.ideas.length&&<p>Save a listed interest to see activity and conversation ideas.</p>}{saved?.ideas.map(i=><article className="character-source" key={i.interest}><h3>{i.interest}</h3><p>{i.activity}</p><p>Enjoy together: {i.together}</p><p className="small muted">{i.reason}</p></article>)}</section>
 </div></div>;
}
