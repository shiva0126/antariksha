import { useEffect, useState } from 'react';
import { Feedback, memberAPI, useAction } from './shared';

export function Moderation() {
 const action = useAction(), [reports, setReports] = useState<{ id: number; post_id: number; reason: string; caption: string }[]>([]);
 const load = async () => setReports(await memberAPI('/api/community/moderation'));
 useEffect(() => { void action.run(load); }, []);
 return <section className="card"><h2>Moderation queue</h2><p>Access is restricted to configured moderators. Actions are recorded.</p>{reports.map(report => <article className="card" key={report.id}><h3>Post {report.post_id}</h3><p>{report.caption}</p><p>Report: {report.reason}</p><button onClick={() => void action.run(async () => { await memberAPI('/api/community/moderation', 'POST', { report: report.id, hide: true }); await load(); })}>Hide post</button><button onClick={() => void action.run(async () => { await memberAPI('/api/community/moderation', 'POST', { report: report.id, hide: false }); await load(); })}>Dismiss report</button></article>)}<Feedback {...action} /></section>;
}
