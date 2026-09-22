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

export interface DashaPeriod { lord: string; maha?: string; antara?: string; from: string; to: string; level: string }

export interface ChartFacts {
  chart: ChartResponse;
  dignities: Record<string, { state: string; sign: string; neecha_bhanga?: boolean }>;
  combustion: Record<string, boolean>;
  retrograde: string[];
  vimshottari: {
    birth_balance: { lord: string; years_remaining: number };
    current: DashaPeriod; upcoming: DashaPeriod; sequence: DashaPeriod[];
  };
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
