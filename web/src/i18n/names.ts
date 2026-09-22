// Sanskrit terms (the engine's English keys) in each Indic script.
// A canonical Sanskrit table in Devanagari is transliterated to the other
// Brahmic scripts by code point: their Unicode blocks share one layout.
// Tamil lacks aspirated and voiced stops, so it has its own table of the
// forms Tamil almanacs use. Regional names for weekdays and Mars override
// the transliteration where they differ.

const sa: Record<string, string> = {
  Mesha: 'मेष', Vrishabha: 'वृषभ', Mithuna: 'मिथुन', Karka: 'कर्क', Simha: 'सिंह', Kanya: 'कन्या', Tula: 'तुला', Vrishchika: 'वृश्चिक', Dhanu: 'धनु', Makara: 'मकर', Kumbha: 'कुंभ', Meena: 'मीन',
  Ashwini: 'अश्विनी', Bharani: 'भरणी', Krittika: 'कृत्तिका', Rohini: 'रोहिणी', Mrigashira: 'मृगशिरा', Ardra: 'आर्द्रा', Punarvasu: 'पुनर्वसु', Pushya: 'पुष्य', Ashlesha: 'आश्लेषा', Magha: 'मघा',
  'Purva Phalguni': 'पूर्वफाल्गुनी', 'Uttara Phalguni': 'उत्तरफाल्गुनी', Hasta: 'हस्त', Chitra: 'चित्रा', Swati: 'स्वाती', Vishakha: 'विशाखा', Anuradha: 'अनुराधा', Jyeshtha: 'ज्येष्ठा', Mula: 'मूल',
  'Purva Ashadha': 'पूर्वाषाढा', 'Uttara Ashadha': 'उत्तराषाढा', Shravana: 'श्रवण', Dhanishtha: 'धनिष्ठा', Shatabhisha: 'शतभिषा', 'Purva Bhadrapada': 'पूर्वभाद्रपदा', 'Uttara Bhadrapada': 'उत्तरभाद्रपदा', Revati: 'रेवती',
  Pratipada: 'प्रतिपदा', Dwitiya: 'द्वितीया', Tritiya: 'तृतीया', Chaturthi: 'चतुर्थी', Panchami: 'पंचमी', Shashthi: 'षष्ठी', Saptami: 'सप्तमी', Ashtami: 'अष्टमी', Navami: 'नवमी', Dashami: 'दशमी',
  Ekadashi: 'एकादशी', Dwadashi: 'द्वादशी', Trayodashi: 'त्रयोदशी', Chaturdashi: 'चतुर्दशी', Purnima: 'पूर्णिमा', Amavasya: 'अमावास्या',
  Ravivara: 'रविवार', Somavara: 'सोमवार', Mangalavara: 'मंगलवार', Budhavara: 'बुधवार', Guruvara: 'गुरुवार', Shukravara: 'शुक्रवार', Shanivara: 'शनिवार',
  Chaitra: 'चैत्र', Vaishakha: 'वैशाख', Ashadha: 'आषाढ', Bhadrapada: 'भाद्रपद', Ashwina: 'आश्विन', Kartika: 'कार्तिक', Margashirsha: 'मार्गशीर्ष', Pausha: 'पौष', Phalguna: 'फाल्गुन', Adhika: 'अधिक',
  Surya: 'सूर्य', Chandra: 'चंद्र', Mangala: 'मंगल', Budha: 'बुध', Guru: 'गुरु', Shukra: 'शुक्र', Shani: 'शनि', Rahu: 'राहु', Ketu: 'केतु',
  Shukla: 'शुक्ल', Krishna: 'कृष्ण', Amrit: 'अमृत', Shubh: 'शुभ', Labh: 'लाभ', Chara: 'चर', Rog: 'रोग', Kala: 'काल', Udveg: 'उद्वेग',
  Vishkambha: 'विष्कंभ', Priti: 'प्रीति', Ayushman: 'आयुष्मान', Saubhagya: 'सौभाग्य', Shobhana: 'शोभन', Atiganda: 'अतिगंड', Sukarma: 'सुकर्मा', Dhriti: 'धृति', Shula: 'शूल', Ganda: 'गंड',
  Vriddhi: 'वृद्धि', Dhruva: 'ध्रुव', Vyaghata: 'व्याघात', Harshana: 'हर्षण', Vajra: 'वज्र', Siddhi: 'सिद्धि', Vyatipata: 'व्यतीपात', Variyana: 'वरीयान', Parigha: 'परिघ', Shiva: 'शिव',
  Siddha: 'सिद्ध', Sadhya: 'साध्य', Shubha: 'शुभ', Brahma: 'ब्रह्म', Indra: 'इंद्र', Vaidhriti: 'वैधृति',
  Bava: 'बव', Balava: 'बालव', Kaulava: 'कौलव', Taitila: 'तैतिल', Garaja: 'गर', Vanija: 'वणिज', Vishti: 'विष्टि', Shakuni: 'शकुनि', Chatushpada: 'चतुष्पद', Naga: 'नाग', Kimstughna: 'किंस्तुघ्न',
};

// Hindi uses the Sanskrit forms except where Hindi spelling differs.
const hiOverrides: Record<string, string> = { 'Purva Phalguni': 'पूर्व फाल्गुनी', 'Uttara Phalguni': 'उत्तर फाल्गुनी', 'Purva Ashadha': 'पूर्वाषाढ़ा', 'Uttara Ashadha': 'उत्तराषाढ़ा', 'Purva Bhadrapada': 'पूर्व भाद्रपद', 'Uttara Bhadrapada': 'उत्तर भाद्रपद', Amavasya: 'अमावस्या', Ashadha: 'आषाढ़', Karka: 'कर्क' };

const regional: Record<string, Record<string, string>> = {
  mr: { Mangalavara: 'मंगळवार', Mangala: 'मंगळ' },
  kn: { Ravivara: 'ಭಾನುವಾರ', Mangalavara: 'ಮಂಗಳವಾರ', Mangala: 'ಮಂಗಳ', Karka: 'ಕರ್ಕಾಟಕ' },
  te: { Ravivara: 'ఆదివారం', Somavara: 'సోమవారం', Mangalavara: 'మంగళవారం', Budhavara: 'బుధవారం', Guruvara: 'గురువారం', Shukravara: 'శుక్రవారం', Shanivara: 'శనివారం', Mangala: 'కుజ', Karka: 'కర్కాటకం' },
  ml: { Ravivara: 'ഞായറാഴ്ച', Somavara: 'തിങ്കളാഴ്ച', Mangalavara: 'ചൊവ്വാഴ്ച', Budhavara: 'ബുധനാഴ്ച', Guruvara: 'വ്യാഴാഴ്ച', Shukravara: 'വെള്ളിയാഴ്ച', Shanivara: 'ശനിയാഴ്ച', Mangala: 'ചൊവ്വ', Guru: 'വ്യാഴം', Karka: 'കർക്കടകം' },
  gu: { Mangalavara: 'મંગળવાર', Mangala: 'મંગળ' },
  bn: { Guruvara: 'বৃহস্পতিবার', Guru: 'বৃহস্পতি' },
};

const ta: Record<string, string> = {
  Mesha: 'மேஷம்', Vrishabha: 'ரிஷபம்', Mithuna: 'மிதுனம்', Karka: 'கடகம்', Simha: 'சிம்மம்', Kanya: 'கன்னி', Tula: 'துலாம்', Vrishchika: 'விருச்சிகம்', Dhanu: 'தனுசு', Makara: 'மகரம்', Kumbha: 'கும்பம்', Meena: 'மீனம்',
  Ashwini: 'அஸ்வினி', Bharani: 'பரணி', Krittika: 'கார்த்திகை', Rohini: 'ரோகிணி', Mrigashira: 'மிருகசீரிடம்', Ardra: 'திருவாதிரை', Punarvasu: 'புனர்பூசம்', Pushya: 'பூசம்', Ashlesha: 'ஆயில்யம்', Magha: 'மகம்',
  'Purva Phalguni': 'பூரம்', 'Uttara Phalguni': 'உத்திரம்', Hasta: 'அஸ்தம்', Chitra: 'சித்திரை', Swati: 'சுவாதி', Vishakha: 'விசாகம்', Anuradha: 'அனுஷம்', Jyeshtha: 'கேட்டை', Mula: 'மூலம்',
  'Purva Ashadha': 'பூராடம்', 'Uttara Ashadha': 'உத்திராடம்', Shravana: 'திருவோணம்', Dhanishtha: 'அவிட்டம்', Shatabhisha: 'சதயம்', 'Purva Bhadrapada': 'பூரட்டாதி', 'Uttara Bhadrapada': 'உத்திரட்டாதி', Revati: 'ரேவதி',
  Pratipada: 'பிரதமை', Dwitiya: 'துவிதியை', Tritiya: 'திருதியை', Chaturthi: 'சதுர்த்தி', Panchami: 'பஞ்சமி', Shashthi: 'சஷ்டி', Saptami: 'சப்தமி', Ashtami: 'அஷ்டமி', Navami: 'நவமி', Dashami: 'தசமி',
  Ekadashi: 'ஏகாதசி', Dwadashi: 'துவாதசி', Trayodashi: 'திரயோதசி', Chaturdashi: 'சதுர்த்தசி', Purnima: 'பௌர்ணமி', Amavasya: 'அமாவாசை',
  Ravivara: 'ஞாயிறு', Somavara: 'திங்கள்', Mangalavara: 'செவ்வாய்', Budhavara: 'புதன்', Guruvara: 'வியாழன்', Shukravara: 'வெள்ளி', Shanivara: 'சனி',
  Chaitra: 'சைத்ரம்', Vaishakha: 'வைசாகம்', Ashadha: 'ஆஷாடம்', Bhadrapada: 'பாத்ரபதம்', Ashwina: 'ஆஸ்வயுஜம்', Kartika: 'கார்த்திகம்', Margashirsha: 'மார்கசீர்ஷம்', Pausha: 'பௌஷ்யம்', Phalguna: 'பால்குனம்', Adhika: 'அதிக',
  Surya: 'சூரியன்', Chandra: 'சந்திரன்', Mangala: 'செவ்வாய்', Budha: 'புதன்', Guru: 'குரு', Shukra: 'சுக்கிரன்', Shani: 'சனி', Rahu: 'ராகு', Ketu: 'கேது',
  Shukla: 'சுக்ல', Krishna: 'கிருஷ்ண', Amrit: 'அமிர்தம்', Shubh: 'சுபம்', Labh: 'லாபம்', Chara: 'சரம்', Rog: 'ரோகம்', Kala: 'காலம்', Udveg: 'உத்வேகம்',
  Vishkambha: 'விஷ்கம்பம்', Priti: 'ப்ரீதி', Ayushman: 'ஆயுஷ்மான்', Saubhagya: 'சௌபாக்யம்', Shobhana: 'சோபனம்', Atiganda: 'அதிகண்டம்', Sukarma: 'சுகர்மம்', Dhriti: 'திருதி', Shula: 'சூலம்', Ganda: 'கண்டம்',
  Vriddhi: 'விருத்தி', Dhruva: 'துருவம்', Vyaghata: 'வியாகாதம்', Harshana: 'ஹர்ஷணம்', Vajra: 'வஜ்ரம்', Siddhi: 'சித்தி', Vyatipata: 'வியதீபாதம்', Variyana: 'வரீயான்', Parigha: 'பரிகம்', Shiva: 'சிவம்',
  Siddha: 'சித்தம்', Sadhya: 'சாத்தியம்', Shubha: 'சுபம்', Brahma: 'பிரம்மம்', Indra: 'ஐந்திரம்', Vaidhriti: 'வைதிருதி',
  Bava: 'பவம்', Balava: 'பாலவம்', Kaulava: 'கௌலவம்', Taitila: 'தைதுலம்', Garaja: 'கரசை', Vanija: 'வணிசை', Vishti: 'பத்திரை', Shakuni: 'சகுனி', Chatushpada: 'சதுஷ்பாதம்', Naga: 'நாகவம்', Kimstughna: 'கிம்ஸ்துக்னம்',
};

const scriptOffset: Record<string, number> = { bn: 0x0980 - 0x0900, gu: 0x0a80 - 0x0900, te: 0x0c00 - 0x0900, kn: 0x0c80 - 0x0900, ml: 0x0d00 - 0x0900 };

/** Transliterates Devanagari into another Brahmic script by code point. */
export function transliterate(s: string, lang: string): string {
  const off = scriptOffset[lang];
  if (off === undefined) return s;
  let out = '';
  for (const ch of s) {
    const c = ch.codePointAt(0)!;
    if (c === 0x093c) continue; // nukta: not needed for Sanskrit terms
    if (c < 0x0900 || c > 0x097f) { out += ch; continue; }
    if (lang === 'bn' && c === 0x0935) { out += 'ব'; continue; } // Bengali writes va as ba
    out += String.fromCodePoint(c + off);
  }
  return out;
}

export function nameIn(lang: string, key: string): string | undefined {
  if (lang === 'en') return undefined;
  if (lang === 'ta') return ta[key];
  if (lang === 'hi') return hiOverrides[key] ?? sa[key];
  const r = regional[lang]?.[key];
  if (r) return r;
  const base = sa[key];
  if (!base) return undefined;
  return lang === 'mr' ? base : transliterate(base, lang);
}
