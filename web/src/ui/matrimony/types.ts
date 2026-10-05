import type { Biodata, MatrimonyPhoto } from '../community/Biodata';

export type Details = Biodata & {
  region?: string; religion?: string; community?: string; mother_tongue?: string; diet?: string; height_cm?: number;
  marital_status?: string; education_level?: string; occupation_category?: string; income_band?: string; family_type?: string;
};

export type Horoscope = {
  guna: number; max: number; kootas: { name: string; score: number; max: number }[]; doshas: string[]; exceptions: string[];
  their_moon: string; your_moon: string; their_mangal: 'yes' | 'no' | 'unknown'; your_mangal: 'yes' | 'no' | 'unknown';
  number_relation: string; their_root_number: number; their_root_graha: string; your_root_number: number; your_root_graha: string;
  viewer_is_groom_side: boolean;
};

export type Candidate = {
  id: string; handle: string; details: Details; avatar: string; accent: string; age: number; photos: MatrimonyPhoto[];
  shared_interests: string[]; verified: boolean; saved: '' | 'saved' | 'skipped'; interest: '' | 'sent' | 'received' | 'accepted' | 'declined';
  horoscope: Horoscope | null; horoscope_note?: string; reasons: string[];
};

export type DiscoverResponse = { items: Candidate[]; total: number; page: number; pages: number; horoscope_enabled?: boolean; needs_profile?: boolean };

export type Interest = {
  id: string; handle: string; display_name: string; avatar: string; accent: string; outgoing: boolean;
  status: 'pending' | 'accepted' | 'declined'; note: string; created_at: string; last_message?: string;
};

export type MyProfile = {
  active: boolean; details: Details; hidden: boolean; horoscope_visible: boolean; verified: boolean;
  saved_search: Record<string, string>; verification: null | 'pending' | 'approved' | 'rejected'; has_birth_place: boolean; email_alerts: boolean;
};

export const displayName = (c: { handle: string; details?: Details; display_name?: string }) => c.details?.display_name || c.display_name || '@' + c.handle;
