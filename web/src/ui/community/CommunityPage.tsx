import { useEffect, useState } from 'react';
import { Feed } from './Feed';
import { Notifications } from './Notifications';
import { People } from './People';
import { FamilySpace } from './FamilySpace';
import { Matrimony } from './Matrimony';
import { CommunitySettings, Moderation, Security } from './Settings';
import { memberAPI } from './shared';
import './community.css';
const tabs = ['Feed', 'Notifications', 'People', 'Family', 'Matrimony', 'Profile & interests', 'Security', 'Moderation'];
export function CommunityPage() {
 const [moderator,setModerator]=useState(false);
 useEffect(()=>{memberAPI('/api/me/security').then(s=>setModerator(s.moderator)).catch(()=>{});},[]);
 const [tab, setTab] = useState('Feed'), [handle, setHandle] = useState<string | null>(null), [ready, setReady] = useState(false);
 useEffect(() => { memberAPI('/api/me').then(m => setHandle(m.handle)).catch(() => {}).finally(() => setReady(true)); }, []);
 return <div className="page community-page"><header className="community-heading"><div><p className="kicker">People · stories · belonging</p><h1>Your Antariksha community</h1><p>Share your world. Find common ground. Stay in control.</p></div>{handle && <a className="ghost" href="#account">@{handle}</a>}</header>
 {!ready ? <p>Loading…</p> : !handle ? <div className="card"><h2>A place for your story</h2><p>Sign in to share photos, follow people, build your private family tree and explore matrimony.</p><a className="primary" href="#account">Sign in or create account</a></div> : <><nav aria-label="Community sections" className="community-tabs">{tabs.filter(t=>t!=='Moderation'||moderator).map(t => <button key={t} className={tab === t ? 'active' : ''} onClick={() => setTab(t)}>{t}</button>)}</nav>
 {tab === 'Feed' && <Feed />}{tab === 'People' && <People />}{tab === 'Family' && <FamilySpace />}{tab === 'Matrimony' && <Matrimony />}{tab === 'Profile & interests' && <CommunitySettings />}{tab === 'Security' && <Security />}{tab === 'Moderation' && <Moderation />}
 {tab === 'Notifications' && <Notifications navigate={setTab}/>}
 </>}
 </div>;
}
