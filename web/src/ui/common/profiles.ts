import type { ChartInput } from '../../api/types';

export interface Profile extends ChartInput { id: string; name: string; place: string }

let owner = 'signed-out';
type AccountBirth = { date: string; time: string; place?: { name: string; lat: number; lon: number; tz: string } };
let birth: AccountBirth | undefined;
export function setProfileOwner(id: string, details?: AccountBirth) { owner = id; birth = details; }
export function accountBirth() { return birth; }
const prefix = () => `antariksha.member.${owner}.`;

export function loadProfiles(): { profiles: Profile[]; active?: string } {
  try {
    const profiles: Profile[] = JSON.parse(localStorage.getItem(prefix() + 'profiles') || '[]');
    const active = localStorage.getItem(prefix() + 'active') || profiles[0]?.id;
    return { profiles, active: profiles.some(p => p.id === active) ? active : profiles[0]?.id };
  } catch { return { profiles: [] }; }
}

export function saveProfiles(profiles: Profile[], active?: string) {
  try {
    localStorage.setItem(prefix() + 'profiles', JSON.stringify(profiles));
    active ? localStorage.setItem(prefix() + 'active', active) : localStorage.removeItem(prefix() + 'active');
  } catch { /* private mode: profiles live for this visit only */ }
}

export const newId = () => Math.random().toString(36).slice(2, 10);

/** Existing device charts are uploaded only by the explicit account-save action. */
export async function initializeAccountCharts() {
  const account = owner, key = prefix();
  if (account === 'signed-out') return;
  const response = await fetch('/api/me/charts', {credentials:'same-origin'});
  if (!response.ok) throw new Error('Saved account charts are unavailable.');
  const data = await response.json();
  if (data.account_id !== account) throw new Error('Your session changed. Sign in again.');
  if (owner !== account) return;
  if (localStorage.getItem(key+'profiles') === null) {
    saveProfiles(data.profiles);
    localStorage.setItem(key+'charts-revision', String(data.revision));
  }
}
export async function saveAccountCharts() {
  const account = owner, key = prefix();
  if (account === 'signed-out') throw new Error('Sign in first.');
  const {profiles} = loadProfiles();
  const revision = Number(localStorage.getItem(key+'charts-revision') || 0);
  const response = await fetch('/api/me/charts', {method:'PUT',credentials:'same-origin',headers:{'Content-Type':'application/json'},body:JSON.stringify({profiles,revision,account_id:account})});
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || 'Unable to save account charts.');
  if (owner !== account) return;
  localStorage.setItem(key+'charts-revision',String(data.revision));
}
export async function loadAccountCharts() {
  const account = owner, key = prefix();
  const response = await fetch('/api/me/charts',{credentials:'same-origin'});
  if (!response.ok) throw new Error('Unable to load account charts.');
  const data = await response.json();
  if (data.account_id !== account) throw new Error('Your session changed. Sign in again.');
  if (owner !== account) return;
  localStorage.setItem(key+'charts-before-load',JSON.stringify(loadProfiles()));
  saveProfiles(data.profiles);
  localStorage.setItem(key+'charts-revision',String(data.revision));
  window.dispatchEvent(new Event('astrisk-charts-loaded'));
}

export function restoreDeviceCharts() {
  const raw = localStorage.getItem(prefix()+'charts-before-load');
  if (!raw) throw new Error('No recovery copy is available for this account on this device.');
  const saved: {profiles: Profile[]; active?: string} = JSON.parse(raw);
  if (!Array.isArray(saved.profiles)) throw new Error('The recovery copy is invalid.');
  saveProfiles(saved.profiles, saved.active);
  window.dispatchEvent(new Event('astrisk-charts-loaded'));
}

export const chatKey = (b: ChartInput) => `${prefix()}chat.${b.date}.${b.time}.${b.lat}.${b.lon}.${b.tz}`;

/** Every key this app stores on the device. */
export function localDataKeys(): string[] {
  const out: string[] = [];
  try { for (let i = 0; i < localStorage.length; i++) { const k = localStorage.key(i); if (k?.startsWith(prefix())) out.push(k); } } catch { /* ignore */ }
  return out;
}
