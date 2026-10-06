import {useEffect,useState} from 'react';
import {api} from '../../lib/api';
import {Button,Card} from '../../ds';
type Data={metrics:{as_of:string;registrations_7d:number;registrations_30d:number;verified_emails:number;accounts_with_active_sessions:number;active_matrimony_profiles:number;incomplete_birthplaces:number;interests_created_7d:number;those_interests_accepted:number;pending_photo_reviews:number;oldest_photo_review:string|null;open_reports_over_24h:number;admins_with_authenticator:number};email_delivery_configured:boolean;phone_verification_configured:boolean;admin_authenticator_available:boolean};
export function Operations(){
 const [data,setData]=useState<Data>(),[error,setError]=useState(''),[busy,setBusy]=useState(false);
 async function load(){setBusy(true);setError('');try{setData(await api<Data>('/api/admin/operations'));}catch(e){setError(e instanceof Error?e.message:'Could not load operations');}finally{setBusy(false);}}
 useEffect(()=>{void load();},[]);
 const m=data?.metrics;
 return <Card title="Operations overview"><Button busy={busy} onClick={()=>void load()}>Refresh operations</Button>{error&&<p role="alert">{error}</p>}{data&&m&&<>
 <p>As of {new Date(m.as_of).toLocaleString()}. Aggregate counts only; private messages are not accessible here.</p>
 <dl className="kv">{[
 ['New accounts: last 7 / 30 days',`${m.registrations_7d} / ${m.registrations_30d}`],
 ['Verified email addresses',m.verified_emails],['Accounts with unexpired sessions',m.accounts_with_active_sessions],
 ['Active, visible matrimony profiles',m.active_matrimony_profiles],['Horoscope-sharing profiles missing birthplace',m.incomplete_birthplaces],
 ['Interests created in last 7 days',m.interests_created_7d],['Of those interests, currently accepted',m.those_interests_accepted],
 ['Pending photo reviews',m.pending_photo_reviews],['Oldest pending photo review',m.oldest_photo_review?new Date(m.oldest_photo_review).toLocaleString():'None'],
 ['Open reports older than 24 hours',m.open_reports_over_24h],['Superadmins with authenticator enabled',m.admins_with_authenticator],
 ['Email delivery',data.email_delivery_configured?'Configured (delivery still needs monitoring)':'Not configured'],
 ['Phone verification',data.phone_verification_configured?'Configured':'Not configured'],
 ['Authenticator encryption',data.admin_authenticator_available?'Available':'Not configured']
 ].map(([key,value])=><div key={key}><dt>{key}</dt><dd>{value}</dd></div>)}</dl>
 <p className="small">An unexpired session does not mean someone is active today. The interest counts describe a creation-date cohort, not a marriage-success rate. Provider configuration is not proof of successful delivery.</p>
 </>}</Card>;
}
