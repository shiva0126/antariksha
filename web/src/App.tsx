import {lazy,Suspense,useEffect,useState} from 'react';
import {getChart} from './api/client';
import type {ChartInput,ChartResponse,Graha} from './api/types';
import {BirthForm} from './ui/BirthForm';
import {CircularWheel} from './chart2d/CircularWheel';
import {NorthIndian} from './chart2d/NorthIndian';
import {SouthIndian} from './chart2d/SouthIndian';
import {GrahaInfoPanel} from './ui/GrahaInfoPanel';
import {PlanetTable} from './ui/PlanetTable';
import {PanchangPage} from './ui/PanchangPage';
const Dome=lazy(()=>import('./scene/CelestialDome'));
type Page='kundali'|'day'|'month';
function webgl(){try{const c=document.createElement('canvas');return!!(c.getContext('webgl2')||c.getContext('webgl'))}catch{return false}}
function initialPage():Page {const p=location.hash.slice(1);return p==='day'||p==='month'?p:'kundali'}
export default function App(){
 const [page,setPage]=useState<Page>(initialPage);
 const [lastInput,setLastInput]=useState<ChartInput>();
 useEffect(()=>{const changed=()=>setPage(initialPage());window.addEventListener('hashchange',changed);return()=>window.removeEventListener('hashchange',changed)},[]);
 const [chart,setChart]=useState<ChartResponse>(),[busy,setBusy]=useState(false),[error,setError]=useState('');
 const [view,setView]=useState<'chart'|'3d'>('chart'),[layout,setLayout]=useState<'circular'|'north'|'south'>('south');
 const [selected,setSelected]=useState<Graha>(),[can3d]=useState(webgl);
 function navigate(p:Page){setPage(p);location.hash=p;setSelected(undefined)}
 async function submit(input:ChartInput){setLastInput(input);setBusy(true);setError('');setSelected(undefined);try{setChart(await getChart(input))}catch(e){setError(e instanceof Error?e.message:'Calculation failed')}finally{setBusy(false)}}
 return <main className={page==='kundali'&&!chart?'landing':'explore'}>
 <header><button className="brand" onClick={()=>navigate('kundali')}><span>अ</span><b>ANTARIKSHA</b></button><nav className="top-nav" aria-label="Main navigation"><button className={page==='kundali'?'active':''} onClick={()=>navigate('kundali')}>Kundali</button><button className={page==='day'?'active':''} onClick={()=>navigate('day')}>Daily Panchang</button><button className={page==='month'?'active':''} onClick={()=>navigate('month')}>Hindu Calendar</button></nav></header>
 {page!=='kundali'?<PanchangPage mode={page}/>:!chart?<section className="hero"><div className="orb"/><p className="kicker">A MAP OF THE MOMENT YOU ARRIVED</p><h1>The sky remembers.</h1><p className="lede">Enter your exact birth details to see your Rashi Kundali, planetary degrees, nakshatras and Lagna.</p><BirthForm onSubmit={submit} busy={busy} initial={lastInput}/>{error&&<p role="alert" className="error">{error}</p>}<footer>LAHIRI SIDEREAL · WHOLE-SIGN HOUSES</footer></section>:<>
 <section className="viewer revealed">
 <nav className="view-nav"><button className={view==='3d'?'active':''} disabled={!can3d} onClick={()=>setView('3d')}>Celestial dome</button><button className={view==='chart'?'active':''} onClick={()=>setView('chart')}>Kundli chart</button></nav>
 {view==='3d'?<Suspense fallback={<div className="loading">Assembling the heavens…</div>}><Dome chart={chart} onSelect={setSelected}/></Suspense>:<div className="chart-stage"><div className="layout-nav">{(['circular','north','south'] as const).map(x=><button key={x} className={layout===x?'active':''} onClick={()=>setLayout(x)}>{x==='north'?'North Indian':x==='south'?'South Indian':'Circular'}</button>)}</div>{layout==='circular'?<CircularWheel chart={chart} onSelect={setSelected}/>:layout==='north'?<NorthIndian chart={chart} onSelect={setSelected}/>:<SouthIndian chart={chart} onSelect={setSelected}/>}<div className="chart-legend"><span><i className="lagna-dot"/>Lagna {chart.ascendant.rashi} {chart.ascendant.degree.toFixed(3)}°</span><span>Rashi / D1 · Whole-sign houses</span></div></div>}
 <button className="reset" onClick={()=>setChart(undefined)}>← Edit birth details</button>
 </section><PlanetTable chart={chart} onSelect={setSelected}/></>}
 {selected&&<GrahaInfoPanel graha={selected} onClose={()=>setSelected(undefined)}/>}
 </main>;
}
