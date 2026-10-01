import { useState } from 'react';
import { loadAccountCharts, restoreDeviceCharts, saveAccountCharts } from './profiles';
export function ChartStorage() {
 const [busy,setBusy]=useState(false),[message,setMessage]=useState('');
 async function run(action:()=>Promise<void>|void,success:string){setBusy(true);setMessage('');try{await action();setMessage(success);}catch(e){setMessage(e instanceof Error?e.message:'Unable to sync charts.');}finally{setBusy(false);}}
 return <aside className="card" aria-label="Private chart storage"><p>Save your charts privately to use them on another device. Only save charts you own or have permission to store.</p><div className="chips">
 <button disabled={busy} onClick={()=>void run(saveAccountCharts,'Charts saved to your private account.')}>Save charts to my account</button>
 <button disabled={busy} onClick={()=>{if(confirm('Load the charts saved in your account? A recovery copy of current device charts will be kept on this device.'))void run(loadAccountCharts,'Account charts loaded.');}}>Load account charts</button>
 <button disabled={busy} onClick={()=>{if(confirm('Restore device charts from before your last account load? This does not change the account copy.'))void run(restoreDeviceCharts,'Previous device charts restored. Account copy unchanged.');}}>Restore previous device charts</button>
 </div>{message && <p role="status">{message}</p>}</aside>;
}
