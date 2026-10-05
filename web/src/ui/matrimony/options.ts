// Structured biodata options. Values match api/matrimony.go matrimonyEnums.
export type Option = [value: string, label: string];

export const OPTIONS: Record<string, Option[]> = {
  religion: [['hindu', 'Hindu'], ['muslim', 'Muslim'], ['christian', 'Christian'], ['sikh', 'Sikh'], ['jain', 'Jain'], ['buddhist', 'Buddhist'], ['parsi', 'Parsi'], ['jewish', 'Jewish'], ['spiritual', 'Spiritual, not religious'], ['none', 'No religion'], ['other', 'Other'], ['prefer_not', 'Prefer not to say']],
  mother_tongue: [['hindi', 'Hindi'], ['marathi', 'Marathi'], ['kannada', 'Kannada'], ['tamil', 'Tamil'], ['telugu', 'Telugu'], ['malayalam', 'Malayalam'], ['gujarati', 'Gujarati'], ['bengali', 'Bengali'], ['punjabi', 'Punjabi'], ['odia', 'Odia'], ['urdu', 'Urdu'], ['konkani', 'Konkani'], ['tulu', 'Tulu'], ['assamese', 'Assamese'], ['english', 'English'], ['other', 'Other']],
  diet: [['vegetarian', 'Vegetarian'], ['eggetarian', 'Eggetarian'], ['non_vegetarian', 'Non-vegetarian'], ['vegan', 'Vegan'], ['jain', 'Jain'], ['other', 'Other']],
  marital_status: [['never_married', 'Never married'], ['divorced', 'Divorced'], ['widowed', 'Widowed'], ['separated', 'Separated'], ['awaiting_divorce', 'Awaiting divorce']],
  education_level: [['high_school', 'High school'], ['diploma', 'Diploma'], ['bachelors', "Bachelor's degree"], ['masters', "Master's degree"], ['doctorate', 'Doctorate'], ['other', 'Other']],
  occupation_category: [['it_software', 'IT / software'], ['engineering', 'Engineering'], ['medicine', 'Medicine / healthcare'], ['business', 'Business / self-employed'], ['government', 'Government / public sector'], ['education', 'Education / research'], ['finance', 'Finance / accounts'], ['law', 'Law'], ['arts_media', 'Arts / media'], ['defence', 'Defence'], ['other', 'Other'], ['not_working', 'Not working']],
  income_band: [['under_3l', 'Under ₹3 lakh'], ['3_6l', '₹3–6 lakh'], ['6_10l', '₹6–10 lakh'], ['10_20l', '₹10–20 lakh'], ['20_35l', '₹20–35 lakh'], ['35_50l', '₹35–50 lakh'], ['over_50l', 'Over ₹50 lakh'], ['prefer_not', 'Prefer not to say']],
  family_type: [['joint', 'Joint family'], ['nuclear', 'Nuclear family'], ['other', 'Other']],
  timeline: [['within_6_months', 'Within 6 months'], ['within_year', 'Within a year'], ['one_to_two_years', 'In 1–2 years'], ['not_sure', 'Not sure yet']],
  relocation: [['open', 'Open to relocating'], ['within_country', 'Within the country'], ['no', 'Prefer not to relocate'], ['discuss', 'Happy to discuss']],
  children: [['want', 'Want children'], ['dont_want', "Don't want children"], ['open', 'Open / undecided'], ['have_children', 'Have children']],
};

export const label = (field: string, value?: string) => OPTIONS[field]?.find(o => o[0] === value)?.[1] ?? '';

export const heightLabel = (cm?: number) => {
  if (!cm) return '';
  const inches = Math.round(cm / 2.54);
  return `${Math.floor(inches / 12)}′${inches % 12}″ (${cm} cm)`;
};

export const GRAHA: Record<string, string> = { sun: 'Sun', moon: 'Moon', mars: 'Mars', mercury: 'Mercury', jupiter: 'Jupiter', venus: 'Venus', saturn: 'Saturn', rahu: 'Rahu', ketu: 'Ketu' };
