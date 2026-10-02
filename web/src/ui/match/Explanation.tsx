import {useState} from 'react';
import type {MatchExplanation} from '../../api/types';
import {Feedback, memberAPI, useAction} from '../community/shared';

export function Explanation({report}:{report:MatchExplanation}) {
 return <section className="card" aria-label="Detailed compatibility explanation">
  <p className="kicker">Understand the comparison</p><h2>Detailed compatibility explanation</h2>
  <p>{report.summary}</p>
  <p role="status" className="small muted">{report.ai_status==='generated'?'AI-assisted explanation · '+report.model:report.ai_status==='not_requested'?'Fact-based guide · no AI call':report.ai_status==='rejected'?'AI output did not pass checks. Showing the fact-based guide.':'AI is unavailable. Showing the complete fact-based guide.'}</p>
  {report.factors.map(f=><details key={f.id}><summary>{f.title} — {f.evidence}</summary><p>{f.explanation}</p><p><strong>Talk about it:</strong> {f.question}</p><p className="small muted">Basis: {f.source}</p></details>)}
  <h3>What this cannot tell you</h3><ul>{report.limitations.map(l=><li key={l}>{l}</li>)}</ul><p className="disclaimer">{report.disclaimer}</p>
 </section>;
}

export function ProfileExplanation({peer}:{peer:string}) {
 const action=useAction(),[report,setReport]=useState<MatchExplanation>(),[useAI,setUseAI]=useState(false);
 return <div>
  <label><input type="checkbox" checked={useAI} onChange={e=>setUseAI(e.target.checked)}/> Use AI to expand the explanation</label>
  <p className="small muted">Optional AI receives only comparison labels and the shared-interest count. Names, profile text, photos, contacts and social links are not sent. No explanation is saved.</p>
  <button disabled={action.busy} onClick={()=>void action.run(async()=>{setReport(undefined);setReport(await memberAPI<MatchExplanation>('/api/matrimony/explanation/'+peer,'POST',{use_ai:useAI}));})}>Explain this match</button>
  <Feedback {...action}/>{report&&<Explanation report={report}/>}
 </div>;
}
