import { useEffect, useState } from 'react';
import { api } from '../../lib/api';

export function AdminAuthenticator() {
 const [status,setStatus]=useState<{enabled:boolean;available:boolean}>(),[password,setPassword]=useState(''),[code,setCode]=useState(''),[secret,setSecret]=useState(''),[busy,setBusy]=useState(false),[error,setError]=useState('');
 useEffect(()=>{void api<{enabled:boolean;available:boolean}>('/api/me/admin-totp').then(setStatus).catch(e=>setError(e.message));},[]);
 async function act(action:string){setBusy(true);setError('');try{const out=await api<{secret?:string;signed_out?:boolean}>('/api/me/admin-totp','POST',{action,password,code});setCode('');if(out.secret)setSecret(out.secret);if(out.signed_out){setPassword('');setSecret('');window.dispatchEvent(new Event('antariksha-signed-out'));}}catch(e){setError(e instanceof Error?e.message:'Authenticator change failed');}finally{setBusy(false);}}
 return <section className="card"><h2>Administrator two-factor authentication</h2>
 <p>{status?.enabled?'Enabled: your password and an authenticator code are required for sign-in.':'Not enabled. Protect your administrator account with an authenticator app.'}</p>
 <p>This protects sign-in across the app. Password recovery does not remove this factor. If you lose your authenticator, recovery requires the server operator. Enabling or disabling it signs out all devices.</p>
 {status&&!status.available&&<p role="alert">Server encryption key is unavailable. Authenticator changes are disabled; contact the server operator.</p>}
 <form onSubmit={e=>{e.preventDefault();void act(status?.enabled?'disable':secret?'confirm':'begin');}}>
 <label className="field">Current password for authenticator setup<input required type="password" autoComplete="current-password" maxLength={72} value={password} onChange={e=>setPassword(e.target.value)}/></label>
 {secret&&<div><p>Add a time-based account named Astrisk to your authenticator app using this setup key. Keep it private. Setup expires in 10 minutes.</p><code style={{overflowWrap:'anywhere'}}>{secret}</code></div>}
 {(secret||status?.enabled)&&<label className="field">Six-digit authenticator code<input required inputMode="numeric" autoComplete="one-time-code" pattern="[0-9]{6}" maxLength={6} value={code} onChange={e=>setCode(e.target.value)}/></label>}
 {error&&<p role="alert">{error}</p>}
 <button disabled={busy||!status?.available}>{busy?'Please wait…':status?.enabled?'Disable authenticator and sign out':secret?'Confirm authenticator and sign out':'Set up authenticator'}</button>
 </form></section>;
}
