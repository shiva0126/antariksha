import type { ChartInput } from '../../api/types';

export interface Profile extends ChartInput { id: string; name: string; place: string }

const KEY = 'antariksha.profiles', ACTIVE = 'antariksha.active';

export function loadProfiles(): { profiles: Profile[]; active?: string } {
  try {
    let profiles: Profile[] = JSON.parse(localStorage.getItem(KEY) || '[]');
    // Migrate the single saved chart from earlier versions.
    const legacy = localStorage.getItem('antariksha.birth');
    if (!profiles.length && legacy) {
      profiles = [{ ...JSON.parse(legacy), id: newId() }];
      localStorage.removeItem('antariksha.birth');
      saveProfiles(profiles, profiles[0].id);
    }
    const active = localStorage.getItem(ACTIVE) || profiles[0]?.id;
    return { profiles, active: profiles.some(p => p.id === active) ? active : profiles[0]?.id };
  } catch { return { profiles: [] }; }
}

export function saveProfiles(profiles: Profile[], active?: string) {
  try {
    localStorage.setItem(KEY, JSON.stringify(profiles));
    active ? localStorage.setItem(ACTIVE, active) : localStorage.removeItem(ACTIVE);
  } catch { /* private mode: profiles live for this visit only */ }
}

export const newId = () => Math.random().toString(36).slice(2, 10);

export const chatKey = (b: ChartInput) => `antariksha.chat.${b.date}.${b.time}.${b.lat}.${b.lon}.${b.tz}`;

/** Every key this app stores on the device. */
export function localDataKeys(): string[] {
  const out: string[] = [];
  try { for (let i = 0; i < localStorage.length; i++) { const k = localStorage.key(i); if (k?.startsWith('antariksha.')) out.push(k); } } catch { /* ignore */ }
  return out;
}
