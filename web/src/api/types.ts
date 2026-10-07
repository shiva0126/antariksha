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

export interface ChatAnswer { answer: string; topics: string[]; sources: Source[]; model: string }

export interface YoginiPeriod { yogini: string; lord: string; from: string; to: string }

export interface Koota { name: string; max: number; score: number; boy: string; girl: string; description: string }
export interface MatchResponse {
	explanation?: MatchExplanation;
  match: { kootas: Koota[]; total: number; max: number; verdict: string; doshas: string[]; exceptions: string[]; boy_mangal_dosha: boolean; girl_mangal_dosha: boolean; mangal_note: string; boy_moon: string; girl_moon: string;
    poruthams: Porutham[]; porutham_good: number; boy_kuja: KujaReport; girl_kuja: KujaReport };
  boy: { lagna: string; moon_sign: string; nakshatra: string; pada: number; mars_house: number };
  girl: { lagna: string; moon_sign: string; nakshatra: string; pada: number; mars_house: number };
}

export interface MatchExplanation {
  summary: string;
  factors: {id:string;title:string;evidence:string;result:string;explanation:string;question:string;source:string}[];
  limitations: string[];
  disclaimer: string;
  model: string;
  ai_status: 'not_requested'|'unavailable'|'rejected'|'generated';
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

export interface ShadbalaRow { graha: string; sthana: number; dig: number; kala: number; chesta: number; naisargika: number; drik: number; total: number; rupas: number; required_rupas: number; required?: number; ratio: number; rank: number; ishta_phala: number; kashta_phala: number; details: Record<string, number>; chesta_motion?: string }
export interface ShadbalaResponse { rows: (ShadbalaRow & { required: number })[]; notes: string[] }

/** What a classical book says for one condition, in plain words. */
export interface BookView { ref: string; source: string; plain: string }

export interface TransitPlanet {
  graha: string; rashi: string; degree: number; nakshatra: string; retrograde: boolean;
  from_moon: number; from_lagna: number;
  /** Brihat Samhita 104.4 verdict; absent for Rahu and Ketu. */
  favourable?: boolean;
  /** False while the planet is in the half of the sign where it does not yet act (104.49-50). */
  active: boolean;
  /** 104.53 conditions that weaken good results: debilitated, enemy, combust. */
  weakened: string[];
  bindus: number; sarva: number; aspects: string[];
  book?: BookView;
}

export interface TransitEvent {
  graha: string; kind: 'ingress' | 'retrograde' | 'direct' | 'solar_eclipse' | 'lunar_eclipse';
  at: string; from?: string; rashi: string; degree: number;
  from_moon?: number; from_lagna?: number; favourable?: boolean;
  on_moon_sign?: boolean; on_lagna_sign?: boolean; degrees_to_moon?: number;
  book?: BookView;
}

export interface ForecastReason { text: string; source: string; sign: number }
export interface AreaForecast {
  area: { id: string; name: string; houses: number[] };
  tone: 'supportive' | 'mixed' | 'challenging';
  confidence: 'strong' | 'moderate' | 'light';
  score: number; reasons: ForecastReason[];
}
export interface ForecastPeriod { from: string; to: string; maha: string; antara: string; summary: string; areas: AreaForecast[]; book?: BookView }

export interface TransitsResponse {
  now: string; months: number;
  planets: TransitPlanet[]; events: TransitEvent[]; periods: ForecastPeriod[];
  sade_sati: boolean; sade_sati_phase: number; kantaka_shani: boolean; ashtama_shani: boolean;
  double_transit: number[];
  notes: Record<string, string>;
}

/** One of the ten South Indian poruthams. Rajju and Vedha are essential. */
export interface Porutham { name: string; status: 'good' | 'medium' | 'bad'; essential: boolean; detail: string }
/** Mangal (Kuja) dosha from the lagna, Moon and Venus, after cancellations. */
export interface KujaReport { from_lagna: number; from_moon: number; from_venus: number; present: string[]; cancellations: string[]; effective: boolean }
