export interface Limb {name:string;number:number;ends_at:string}
export interface Window {start:string;end:string}
export interface PanchangDay {
 date:string; location:{lat:number;lon:number;tz:string};
 sunrise:string;sunset:string;moonrise?:string;moonset?:string;
 vaara:string;paksha:string;lunar_month:string;
 tithi:Limb;nakshatra:Limb;yoga:Limb;karana:Limb;
 rahu_kaal:Window;yamaganda:Window;gulika:Window;abhijit:Window;
 choghadiya:(Window&{name:string})[];festivals:string[];
}
export interface MonthDay {date:string;tithi:string;paksha:string;festivals:string[]}
export async function fetchJSON<T>(path:string, signal?:AbortSignal):Promise<T>{
 const r=await fetch(path,{signal});const value=await r.json();
 if(!r.ok)throw new Error(value.error??'Could not load this date');
 return value;
}
