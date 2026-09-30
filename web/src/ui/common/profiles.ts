import type { ChartInput } from '../../api/types';

export interface Profile extends ChartInput { id: string; name: string; place: string }

let owner = 'signed-out';
let birth: { date: string; time: string } | undefined;
export function setProfileOwner(id: string, details?: { date: string; time: string }) { owner = id; birth = details; }
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

export const chatKey = (b: ChartInput) => `${prefix()}chat.${b.date}.${b.time}.${b.lat}.${b.lon}.${b.tz}`;

/** Every key this app stores on the device. */
export function localDataKeys(): string[] {
  const out: string[] = [];
  try { for (let i = 0; i < localStorage.length; i++) { const k = localStorage.key(i); if (k?.startsWith(prefix())) out.push(k); } } catch { /* ignore */ }
  return out;
}
