import { useEffect,useState } from 'react';
export function EmailRecovery({action,token,onBack}:{action:'request'|'reset'|'verify';token?:string;onBack:()=>void}){
 const [email,setEmail]=useState(''),[password,setPassword]=useState(''),[message,setMessage]=useState(''),[busy,setBusy]=useState(false),[done,setDone]=useState(false);
 return <main className="login-screen"><section className="login-card"><p className="kicker">ASTRISK.SPACE</p><h1>{action==='verify'?'Verify email':action==='reset'?'Choose a new password':'Reset password by email'}</h1>
 {!done&&<form onSubmit={async e=>{e.preventDefault();setBusy(true);setMessage('');try{const res=await fetch('/api/auth/email/'+(action==='request'?'reset-request':action),{method:'POST',credentials:'same-origin',headers:{'Content-Type':'application/json'},body:JSON.stringify(action==='request'?{email}:action==='reset'?{token,password}:{token})});const out=await res.json();if(!res.ok)throw new Error(out.error||'Request failed');setDone(true);setPassword('');setMessage(out.message||(action==='reset'?'Password updated. Sign in with your new password.':'Email verified.'));}catch(err){setMessage(err instanceof Error?err.message:'Request failed');}finally{setBusy(false);}}}>
 {action==='request'&&<label className="field">Email<input type="email" autoComplete="email" required maxLength={254} value={email} onChange={e=>setEmail(e.target.value)}/></label>}
 {action==='reset'&&<label className="field">New password<input type="password" autoComplete="new-password" required minLength={12} maxLength={72} value={password} onChange={e=>setPassword(e.target.value)}/></label>}
 <button disabled={busy}>{action==='verify'?'Verify my email':action==='reset'?'Update password':'Send reset link'}</button></form>}
 <p role="status">{message}</p><button onClick={onBack}>Back to sign in</button></section></main>;
}
export function EmailLinkGate({children}:{children:React.ReactNode}){
 const parse=()=>{const m=location.hash.match(/^#(reset|verify)\/([a-f0-9]{64})$/);return m?{action:m[1] as 'reset'|'verify',token:m[2]}:null;};
 const [link,setLink]=useState(parse);
 useEffect(()=>{const update=()=>setLink(parse());window.addEventListener('hashchange',update);return()=>window.removeEventListener('hashchange',update);},[]);
 if(link)return <EmailRecovery {...link} onBack={()=>{history.replaceState(null,'',location.pathname);setLink(null);window.location.reload();}}/>;
 return <>{children}</>;
}
