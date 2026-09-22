export interface ChartInput { date: string; time: string; lat: number; lon: number; tz: string }

export interface Graha {
  id: string; name: string; longitude: number; latitude: number; distance_au: number;
  rashi: string; rashi_degree: number; nakshatra: string; nakshatra_pada: number;
  retrograde: boolean; speed: number;
}

export interface ChartResponse {
  schema_version: number; input: ChartInput; ayanamsa: 'lahiri';
  ascendant: { longitude: number; rashi: string; degree: number };
  grahas: Graha[]; houses: { system: 'whole_sign'; cusps: number[] };
}

export interface DashaPeriod { lord: string; maha?: string; antara?: string; pratyantara?: string; from: string; to: string; level: string }

export interface ChartFacts {
  chart: ChartResponse;
  dignities: Record<string, { state: string; sign: string; neecha_bhanga?: boolean }>;
  combustion: Record<string, boolean>;
  retrograde: string[];
  vimshottari: {
    birth_balance: { lord: string; years_remaining: number };
    current: DashaPeriod; upcoming: DashaPeriod; sequence: DashaPeriod[];
    antaras: DashaPeriod[]; pratyantaras: DashaPeriod[];
  };
  yogini: { current: YoginiPeriod; sequence: YoginiPeriod[] };
  ashtakavarga: { bhinna: Record<string, number[]>; sarva: number[] };
  yogas: { name: string; type: string; planets: string[] | null; houses: string[]; strength: string; geometry: Record<string, unknown> }[];
  as_of: string;
}

export interface Reading {
  summary: string; lagna_and_moon: string;
  grahas: { graha: string; placement: string; meaning: string }[];
  yogas: { name: string; meaning: string; effect: string; strength: string }[];
  dashas: { current: string; upcoming: string };
  themes: { career: string; relationships: string; strengths: string; growth_areas: string };
  disclaimer: string;
}

export interface Source { doc_type: string; key: string; title: string; source: string; ref?: string }

export interface ReadingResponse {
  chart_hash: string; cached: boolean; model: string; facts: ChartFacts; reading: Reading; grounding: Source[] | null;
}

export interface ChatMessage {
  id: number; role: 'user' | 'assistant'; content: string; topics: string[];
  sources: Source[]; model?: string; created_at: string;
}

export interface YoginiPeriod { yogini: string; lord: string; from: string; to: string }

export interface Koota { name: string; max: number; score: number; boy: string; girl: string; description: string }
export interface MatchResponse {
  match: { kootas: Koota[]; total: number; max: number; verdict: string; doshas: string[]; exceptions: string[]; boy_mangal_dosha: boolean; girl_mangal_dosha: boolean; mangal_note: string; boy_moon: string; girl_moon: string };
  boy: { lagna: string; moon_sign: string; nakshatra: string; pada: number; mars_house: number };
  girl: { lagna: string; moon_sign: string; nakshatra: string; pada: number; mars_house: number };
}

export interface MuhurtaDay { date: string; vaara: string; tithi: string; paksha: string; nakshatra: string; yoga: string; good: boolean; score: number; reasons: string[]; cautions: string[]; windows: { start: string; end: string }[]; avoid: { start: string; end: string }[] }
export interface MuhurtaResponse { event: { id: string; name: string; description: string }; personalised: boolean; days: MuhurtaDay[] }
export interface MuhurtaEvent { ID: string; Name: string; Description: string }

export interface Transit { id: string; name: string; rashi: string; degree: number; nakshatra: string; retrograde: boolean; house_from_lagna: number; house_from_moon: number }
export interface TodayResponse {
  now: string; panchang: import('./panchang').PanchangDay; transits: Transit[];
  tara_bala: { number: number; name: string; favourable: boolean; moon_nakshatra: string };
  chandra_bala: { house: number; favourable: boolean; moon_sign: string };
  sade_sati: { active: boolean; phase: number };
  dasha: DashaPeriod; upcoming: { date: string; festivals: string[] }[];
}

export interface PlaceHit { name: string; region: string; country: string; lat: number; lon: number; tz: string; population: number }
export interface VargaResponse { varga: number; name: string; theme: string; chart: ChartResponse }
