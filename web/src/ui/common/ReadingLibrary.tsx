import {useEffect,useState} from 'react';
import {memberAPI} from '../community/shared';

type Library={llm_configured:boolean;database_available:boolean;corpus_entries:number;embedded_entries:number;compatible_embeddings:number;semantic_enabled:boolean;semantic_ready:boolean;embedding_model?:string;sources:{id:string;title:string;rights:string;system:string;entries:number;note:string}[]};

export function ReadingLibrary(){
 const [data,setData]=useState<Library>(),[error,setError]=useState('');
 useEffect(()=>{let active=true;void memberAPI<Library>('/api/reading/status').then(d=>{if(active)setData(d)}).catch(()=>{if(active)setError('Library status is temporarily unavailable. No source or AI availability is assumed.');});return()=>{active=false};},[]);
 return <details className="card" style={{overflowWrap:'anywhere'}} aria-label="Reading library status"><summary>Reading library & AI availability</summary>
  {error?<p role="status">{error}</p>:!data?<p role="status">Checking the library…</p>:<>
   <p>{data.llm_configured?'AI provider configured. Availability still depends on the provider; a failed request uses a labeled guide.':'AI provider not configured. Readings and matchmaking explanations currently use the fact-based guides.'}</p>
   <p>{data.database_available?`${data.corpus_entries} database entries · ${data.embedded_entries} stored vectors`:'Database library unavailable. Embedded original guides remain available.'}</p>
   <p>{data.semantic_ready?`Semantic search is ready: ${data.compatible_embeddings} passages are indexed${data.embedding_model?.startsWith('bge-small')?' with a free embedding model that runs on this server':''}. Ask Astrisk uses it to add the passages closest to your question; exact fact lookup remains primary.`:'Semantic enrichment is not ready. Exact fact lookup works without embeddings.'}</p>
   <p>This is a searchable library, not model training. A source listed below is not necessarily ingested. Original guides are modern editorial reflections, not literal translations or independently certified readings.</p>
   <ul>{data.sources.map(s=><li key={s.id}><strong>{s.title}</strong> — {s.entries>0?`${s.entries} database entries`:s.rights==='copyrighted_blocked'?'Blocked from ingestion':s.rights==='not_acquired'?'Not acquired':'No database entries'} · {s.system}<details><summary>Source status</summary><p>{s.rights.replaceAll('_',' ')}</p>{s.note&&<p>{s.note}</p>}</details></li>)}</ul>
   <p className="disclaimer">Source coverage is not evidence that astrology predicts outcomes. External chart comparisons and expert content review remain separate validation tasks.</p>
  </>}
 </details>;
}
