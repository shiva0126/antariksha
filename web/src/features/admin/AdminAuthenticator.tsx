import { useEffect, useState } from 'react';
import { api } from '../../lib/api';
import {Button,Card,Field,Input,Notice} from '../../ds';

export function AdminAuthenticator() {
 const [status,setStatus]=useState<{enabled:boolean;available:boolean}>(),[password,setPassword]=useState(''),[code,setCode]=useState(''),[secret,setSecret]=useState(''),[busy,setBusy]=useState(false),[error,setError]=useState('');
 useEffect(()=>{void api<{enabled:boolean;available:boolean}>('/api/me/admin-totp').then(setStatus).catch(e=>setError(e.message));},[]);
 async function act(action:string){setBusy(true);setError('');try{const out=await api<{secret?:string;signed_out?:boolean}>('/api/me/admin-totp','POST',{action,password,code});setCode('');if(out.secret)setSecret(out.secret);if(out.signed_out){setPassword('');setSecret('');window.dispatchEvent(new Event('antariksha-signed-out'));}}catch(e){setError(e instanceof Error?e.message:'Authenticator change failed');}finally{setBusy(false);}}
 return <Card title="Administrator two-factor authentication">
 <p>{status?.enabled?'Enabled: your password and an authenticator code are required for sign-in.':'Not enabled. Protect your administrator account with an authenticator app.'}</p>
 <p>This protects sign-in across the app. Password recovery does not remove this factor. If you lose your authenticator, recovery requires the server operator. Enabling or disabling it signs out all devices.</p>
 {status&&!status.available&&<p role="alert">Server encryption key is unavailable. Authenticator changes are disabled; contact the server operator.</p>}
 <form onSubmit={e=>{e.preventDefault();void act(status?.enabled?'disable':secret?'confirm':'begin');}}>
 <Field label="Current password for authenticator setup"><Input required type="password" autoComplete="current-password" maxLength={72} value={password} onChange={e=>setPassword(e.target.value)}/></Field>
 {secret&&<div><p>Add a time-based account named Astrisk to your authenticator app using this setup key. Keep it private. Setup expires in 10 minutes.</p><code className="admin-authenticator-secret">{secret}</code></div>}
 {(secret||status?.enabled)&&<Field label="Six-digit authenticator code"><Input required inputMode="numeric" autoComplete="one-time-code" pattern="[0-9]{6}" maxLength={6} value={code} onChange={e=>setCode(e.target.value)}/></Field>}
 {error&&<Notice tone="danger">{error}</Notice>}
 <Button type="submit" busy={busy} disabled={!status?.available}>{status?.enabled?'Disable authenticator and sign out':secret?'Confirm authenticator and sign out':'Set up authenticator'}</Button>
 {secret&&<Button type="button" variant="ghost" disabled={busy} onClick={()=>{setSecret('');setCode('');setError('');}}>Restart setup</Button>}
 </form></Card>;
}
