import { useEffect, useState } from 'react';
import { Avatar, Feedback, Field, memberAPI, useAction } from './shared';
type Person = { id: string; handle: string; public_bio: string; avatar: string; accent: string; interests: string[]; links: string[]; following: string; accepted?: boolean };
export function People() {
 const action = useAction(), [people, setPeople] = useState<Person[]>([]), [requests, setRequests] = useState<Person[]>([]), [q, setQ] = useState('');
 const load = async () => { setPeople(await memberAPI(`/api/community/people?q=${encodeURIComponent(q)}`)); setRequests(await memberAPI('/api/community/follows')); };
 useEffect(() => { void action.run(load); }, []);
 const follow = (target: string, value: string) => void action.run(async () => { await memberAPI('/api/community/follows', 'POST', { target, action: value }); await load(); });
 return <><div className="card"><h2>Find your people</h2><p>Connect through shared interests. Profiles appear here only after their owners join the community.</p><form onSubmit={e => { e.preventDefault(); void action.run(load); }}><Field label="Search handles"><input value={q} onChange={e => setQ(e.target.value)} /></Field><button>Search</button></form><Feedback {...action} /></div>
 <div className="grid-3">{people.map(p => <article className="card" key={p.id}><Avatar kind={p.avatar} accent={p.accent} /><h3>@{p.handle}</h3><p>{p.public_bio}</p><div className="chips">{p.interests.map(i => <span key={i} className="tag">{i}</span>)}</div><p>{p.links.map((link, i) => <a key={link} href={link} target="_blank" rel="noopener noreferrer">Social link {i + 1} (self-declared) </a>)}</p><button disabled={action.busy} onClick={() => follow(p.id, p.following === 'none' ? 'request' : 'unfollow')}>{p.following === 'none' ? 'Request to follow' : p.following === 'pending' ? 'Cancel request' : 'Unfollow'}</button></article>)}</div>
 <section className="card"><h3>Followers and requests</h3>{requests.length === 0 && <p>No follow requests yet.</p>}{requests.map(p => <div className="community-row" key={p.id}><span>@{p.handle} · {p.accepted ? 'Approved' : 'Pending'}</span>{!p.accepted && <button disabled={action.busy} onClick={() => follow(p.id, 'accept')}>Approve</button>}<button disabled={action.busy} onClick={() => follow(p.id, 'remove')}>Remove</button></div>)}</section></>;
}
