import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';

export type Lang = 'en' | 'hi';
export type MonthSystem = 'amanta' | 'purnimanta';

const hi: Record<string, string> = {
  'Kundali': 'कुंडली', 'Matching': 'कुंडली मिलान', 'Muhurta': 'मुहूर्त', 'Daily Panchang': 'दैनिक पंचांग', 'Hindu Calendar': 'हिंदू कैलेंडर',
  'Today': 'आज', 'Chart': 'चार्ट', 'Planets': 'ग्रह', 'Dasha': 'दशा', 'Ashtakavarga': 'अष्टकवर्ग', 'Reading': 'फलादेश', 'Ask Antariksha': 'प्रश्न पूछें', 'Report': 'रिपोर्ट',
  'Lagna': 'लग्न', 'Moon sign': 'चंद्र राशि', 'Nakshatra': 'नक्षत्र', 'Sun sign': 'सूर्य राशि', 'Current dasha': 'वर्तमान दशा',
  'Edit birth details': 'जन्म विवरण बदलें', 'Your chart': 'आपकी कुंडली', 'Janma kundali': 'जन्म कुंडली',
  'Reveal my chart': 'कुंडली देखें', 'Birth date': 'जन्म तिथि', 'Birth time': 'जन्म समय', 'Birthplace': 'जन्म स्थान', 'Name': 'नाम',
  'Vaara': 'वार', 'Tithi': 'तिथि', 'Yoga': 'योग', 'Karana': 'करण', 'Sunrise': 'सूर्योदय', 'Sunset': 'सूर्यास्त', 'Moonrise': 'चंद्रोदय', 'Moonset': 'चंद्रास्त',
  'Sun and Moon': 'सूर्य और चंद्र', 'Muhurta windows': 'मुहूर्त काल', 'Choghadiya': 'चौघड़िया', 'Day': 'दिन', 'Night': 'रात',
  'Abhijit muhurta': 'अभिजित मुहूर्त', 'Rahu kaal': 'राहु काल', 'Yamaganda': 'यमगंड', 'Gulika kaal': 'गुलिक काल',
  'Location': 'स्थान', 'Date': 'तारीख', 'Month': 'माह', 'Panchang calendar': 'पंचांग कैलेंडर', 'Festivals and observances this month': 'इस माह के पर्व और व्रत',
  'Shukla': 'शुक्ल', 'Krishna': 'कृष्ण', 'paksha': 'पक्ष', 'month': 'मास', 'Profiles': 'प्रोफ़ाइल', 'Add profile': 'नई प्रोफ़ाइल', 'Language': 'भाषा', 'Months': 'मास पद्धति',
  'Groom': 'वर', 'Bride': 'वधू', 'Match charts': 'मिलान करें', 'Find muhurta': 'मुहूर्त खोजें', 'Event': 'कार्य',
};

const names: Record<string, string> = {
  Mesha: 'मेष', Vrishabha: 'वृषभ', Mithuna: 'मिथुन', Karka: 'कर्क', Simha: 'सिंह', Kanya: 'कन्या', Tula: 'तुला', Vrishchika: 'वृश्चिक', Dhanu: 'धनु', Makara: 'मकर', Kumbha: 'कुंभ', Meena: 'मीन',
  Ashwini: 'अश्विनी', Bharani: 'भरणी', Krittika: 'कृत्तिका', Rohini: 'रोहिणी', Mrigashira: 'मृगशिरा', Ardra: 'आर्द्रा', Punarvasu: 'पुनर्वसु', Pushya: 'पुष्य', Ashlesha: 'आश्लेषा', Magha: 'मघा',
  'Purva Phalguni': 'पूर्व फाल्गुनी', 'Uttara Phalguni': 'उत्तर फाल्गुनी', Hasta: 'हस्त', Chitra: 'चित्रा', Swati: 'स्वाति', Vishakha: 'विशाखा', Anuradha: 'अनुराधा', Jyeshtha: 'ज्येष्ठा', Mula: 'मूल',
  'Purva Ashadha': 'पूर्वाषाढ़ा', 'Uttara Ashadha': 'उत्तराषाढ़ा', Shravana: 'श्रवण', Dhanishtha: 'धनिष्ठा', Shatabhisha: 'शतभिषा', 'Purva Bhadrapada': 'पूर्व भाद्रपद', 'Uttara Bhadrapada': 'उत्तर भाद्रपद', Revati: 'रेवती',
  Pratipada: 'प्रतिपदा', Dwitiya: 'द्वितीया', Tritiya: 'तृतीया', Chaturthi: 'चतुर्थी', Panchami: 'पंचमी', Shashthi: 'षष्ठी', Saptami: 'सप्तमी', Ashtami: 'अष्टमी', Navami: 'नवमी', Dashami: 'दशमी',
  Ekadashi: 'एकादशी', Dwadashi: 'द्वादशी', Trayodashi: 'त्रयोदशी', Chaturdashi: 'चतुर्दशी', Purnima: 'पूर्णिमा', Amavasya: 'अमावस्या',
  Ravivara: 'रविवार', Somavara: 'सोमवार', Mangalavara: 'मंगलवार', Budhavara: 'बुधवार', Guruvara: 'गुरुवार', Shukravara: 'शुक्रवार', Shanivara: 'शनिवार',
  Chaitra: 'चैत्र', Vaishakha: 'वैशाख', Jyeshtha_m: 'ज्येष्ठ', Ashadha: 'आषाढ़', Bhadrapada: 'भाद्रपद', Ashwina: 'आश्विन', Kartika: 'कार्तिक', Margashirsha: 'मार्गशीर्ष', Pausha: 'पौष', Phalguna: 'फाल्गुन', Adhika: 'अधिक',
  Surya: 'सूर्य', Chandra: 'चंद्र', Mangala: 'मंगल', Budha: 'बुध', Guru: 'गुरु', Shukra: 'शुक्र', Shani: 'शनि', Rahu: 'राहु', Ketu: 'केतु',
  Shukla: 'शुक्ल', Krishna: 'कृष्ण', Amrit: 'अमृत', Shubh: 'शुभ', Labh: 'लाभ', Chara: 'चर', Rog: 'रोग', Kala: 'काल', Udveg: 'उद्वेग',
  Vishkambha: 'विष्कंभ', Priti: 'प्रीति', Ayushman: 'आयुष्मान', Saubhagya: 'सौभाग्य', Shobhana: 'शोभन', Atiganda: 'अतिगंड', Sukarma: 'सुकर्मा', Dhriti: 'धृति', Shula: 'शूल', Ganda: 'गंड',
  Vriddhi: 'वृद्धि', Dhruva: 'ध्रुव', Vyaghata: 'व्याघात', Harshana: 'हर्षण', Vajra: 'वज्र', Siddhi: 'सिद्धि', Vyatipata: 'व्यतीपात', Variyana: 'वरीयान', Parigha: 'परिघ', Shiva: 'शिव',
  Siddha: 'सिद्ध', Sadhya: 'साध्य', Shubha: 'शुभ', Brahma: 'ब्रह्म', Indra: 'इंद्र', Vaidhriti: 'वैधृति',
  Bava: 'बव', Balava: 'बालव', Kaulava: 'कौलव', Taitila: 'तैतिल', Garaja: 'गर', Vanija: 'वणिज', Vishti: 'विष्टि', Shakuni: 'शकुनि', Chatushpada: 'चतुष्पद', Naga: 'नाग', Kimstughna: 'किंस्तुघ्न',
};

export const MONTHS = ['Chaitra', 'Vaishakha', 'Jyeshtha', 'Ashadha', 'Shravana', 'Bhadrapada', 'Ashwina', 'Kartika', 'Margashirsha', 'Pausha', 'Magha', 'Phalguna'];

interface Settings { lang: Lang; months: MonthSystem; setLang: (l: Lang) => void; setMonths: (m: MonthSystem) => void }
const Ctx = createContext<Settings>({ lang: 'en', months: 'amanta', setLang: () => {}, setMonths: () => {} });

const read = <T extends string>(k: string, d: T): T => { try { return (localStorage.getItem(k) as T) || d; } catch { return d; } };

export function SettingsProvider({ children }: { children: ReactNode }) {
  const [lang, setLang] = useState<Lang>(() => read('antariksha.lang', 'en'));
  const [months, setMonths] = useState<MonthSystem>(() => read('antariksha.months', 'amanta'));
  useEffect(() => { try { localStorage.setItem('antariksha.lang', lang); localStorage.setItem('antariksha.months', months); } catch { /* ignore */ } document.documentElement.lang = lang; }, [lang, months]);
  return <Ctx.Provider value={{ lang, months, setLang, setMonths }}>{children}</Ctx.Provider>;
}

export function useSettings() { return useContext(Ctx); }

/** t(): UI string in the chosen language (English is the key). */
export function useT() {
  const { lang } = useContext(Ctx);
  return (s: string) => (lang === 'hi' ? hi[s] ?? s : s);
}

/** n(): a Sanskrit term (rashi, nakshatra, tithi, graha…) in Devanagari when Hindi is chosen. */
export function useNames() {
  const { lang } = useContext(Ctx);
  return (s: string) => {
    if (lang !== 'hi' || !s) return s;
    if (names[s]) return names[s];
    return s.split(' ').map(w => names[w] ?? (w === 'Jyeshtha' ? names.Jyeshtha : w)).join(' ');
  };
}

/** Month name in the chosen system. Amanta months end at new moon; in the
 * Purnimanta system the dark fortnight belongs to the following month. */
export function useMonthName() {
  const { months } = useContext(Ctx);
  return (amanta: string, paksha: string) => {
    if (months === 'amanta' || paksha !== 'Krishna') return amanta;
    const adhika = amanta.startsWith('Adhika ');
    const base = adhika ? amanta.slice(7) : amanta;
    const i = MONTHS.indexOf(base);
    if (i < 0) return amanta;
    return (adhika ? 'Adhika ' : '') + MONTHS[(i + 1) % 12];
  };
}
