export interface Place { name: string; region: string; lat: number; lon: number; tz: string }

const IN = 'Asia/Kolkata';
const india: [string, string, number, number][] = [
  ['Agra', 'Uttar Pradesh', 27.1767, 78.0081], ['Ahmedabad', 'Gujarat', 23.0225, 72.5714], ['Ajmer', 'Rajasthan', 26.4499, 74.6399],
  ['Amritsar', 'Punjab', 31.634, 74.8723], ['Aurangabad', 'Maharashtra', 19.8762, 75.3433], ['Belagavi', 'Karnataka', 15.8497, 74.4977],
  ['Bengaluru', 'Karnataka', 12.9716, 77.5946], ['Bhopal', 'Madhya Pradesh', 23.2599, 77.4126], ['Bhubaneswar', 'Odisha', 20.2961, 85.8245],
  ['Chandigarh', 'Chandigarh', 30.7333, 76.7794], ['Chennai', 'Tamil Nadu', 13.0827, 80.2707], ['Coimbatore', 'Tamil Nadu', 11.0168, 76.9558],
  ['Dehradun', 'Uttarakhand', 30.3165, 78.0322], ['Delhi', 'Delhi', 28.6139, 77.209], ['Dhanbad', 'Jharkhand', 23.7957, 86.4304],
  ['Guwahati', 'Assam', 26.1445, 91.7362], ['Gwalior', 'Madhya Pradesh', 26.2183, 78.1828], ['Hubballi', 'Karnataka', 15.3647, 75.124],
  ['Hyderabad', 'Telangana', 17.385, 78.4867], ['Indore', 'Madhya Pradesh', 22.7196, 75.8577], ['Jabalpur', 'Madhya Pradesh', 23.1815, 79.9864],
  ['Jaipur', 'Rajasthan', 26.9124, 75.7873], ['Jammu', 'Jammu and Kashmir', 32.7266, 74.857], ['Jamshedpur', 'Jharkhand', 22.8046, 86.2029],
  ['Jodhpur', 'Rajasthan', 26.2389, 73.0243], ['Kanpur', 'Uttar Pradesh', 26.4499, 80.3319], ['Kochi', 'Kerala', 9.9312, 76.2673],
  ['Kolkata', 'West Bengal', 22.5726, 88.3639], ['Kozhikode', 'Kerala', 11.2588, 75.7804], ['Lucknow', 'Uttar Pradesh', 26.8467, 80.9462],
  ['Ludhiana', 'Punjab', 30.901, 75.8573], ['Madurai', 'Tamil Nadu', 9.9252, 78.1198], ['Mangaluru', 'Karnataka', 12.9141, 74.856],
  ['Meerut', 'Uttar Pradesh', 28.9845, 77.7064], ['Mumbai', 'Maharashtra', 19.076, 72.8777], ['Mysuru', 'Karnataka', 12.2958, 76.6394],
  ['Nagpur', 'Maharashtra', 21.1458, 79.0882], ['Nashik', 'Maharashtra', 19.9975, 73.7898], ['Patna', 'Bihar', 25.5941, 85.1376],
  ['Prayagraj', 'Uttar Pradesh', 25.4358, 81.8463], ['Puducherry', 'Puducherry', 11.9416, 79.8083], ['Pune', 'Maharashtra', 18.5204, 73.8567],
  ['Raipur', 'Chhattisgarh', 21.2514, 81.6296], ['Rajkot', 'Gujarat', 22.3039, 70.8022], ['Ranchi', 'Jharkhand', 23.3441, 85.3096],
  ['Shimla', 'Himachal Pradesh', 31.1048, 77.1734], ['Srinagar', 'Jammu and Kashmir', 34.0837, 74.7973], ['Surat', 'Gujarat', 21.1702, 72.8311],
  ['Thiruvananthapuram', 'Kerala', 8.5241, 76.9366], ['Tiruchirappalli', 'Tamil Nadu', 10.7905, 78.7047], ['Tirupati', 'Andhra Pradesh', 13.6288, 79.4192],
  ['Udaipur', 'Rajasthan', 24.5854, 73.7125], ['Udupi', 'Karnataka', 13.3409, 74.7421], ['Ujjain', 'Madhya Pradesh', 23.1765, 75.7885],
  ['Vadodara', 'Gujarat', 22.3072, 73.1812], ['Varanasi', 'Uttar Pradesh', 25.3176, 82.9739], ['Vijayawada', 'Andhra Pradesh', 16.5062, 80.648],
  ['Visakhapatnam', 'Andhra Pradesh', 17.6868, 83.2185],
];
const world: Place[] = [
  { name: 'Colombo', region: 'Sri Lanka', lat: 6.9271, lon: 79.8612, tz: 'Asia/Colombo' },
  { name: 'Dhaka', region: 'Bangladesh', lat: 23.8103, lon: 90.4125, tz: 'Asia/Dhaka' },
  { name: 'Kathmandu', region: 'Nepal', lat: 27.7172, lon: 85.324, tz: 'Asia/Kathmandu' },
  { name: 'Dubai', region: 'UAE', lat: 25.2048, lon: 55.2708, tz: 'Asia/Dubai' },
  { name: 'Singapore', region: 'Singapore', lat: 1.3521, lon: 103.8198, tz: 'Asia/Singapore' },
  { name: 'Kuala Lumpur', region: 'Malaysia', lat: 3.139, lon: 101.6869, tz: 'Asia/Kuala_Lumpur' },
  { name: 'London', region: 'United Kingdom', lat: 51.5072, lon: -0.1276, tz: 'Europe/London' },
  { name: 'New York', region: 'USA', lat: 40.7128, lon: -74.006, tz: 'America/New_York' },
  { name: 'Chicago', region: 'USA', lat: 41.8781, lon: -87.6298, tz: 'America/Chicago' },
  { name: 'San Francisco', region: 'USA', lat: 37.7749, lon: -122.4194, tz: 'America/Los_Angeles' },
  { name: 'Toronto', region: 'Canada', lat: 43.6532, lon: -79.3832, tz: 'America/Toronto' },
  { name: 'Sydney', region: 'Australia', lat: -33.8688, lon: 151.2093, tz: 'Australia/Sydney' },
];

export const places: Place[] = [...india.map(([name, region, lat, lon]) => ({ name, region, lat, lon, tz: IN })), ...world];
export const placeLabel = (p: Place) => `${p.name}, ${p.region}`;
export const findPlace = (label: string) => places.find(p => placeLabel(p).toLowerCase() === label.trim().toLowerCase() || p.name.toLowerCase() === label.trim().toLowerCase());
export const nearestPlace = (lat: number, lon: number) => places.find(p => Math.abs(p.lat - lat) < 1e-3 && Math.abs(p.lon - lon) < 1e-3);
export const defaultPlace = places.find(p => p.name === 'Bengaluru')!;

export function timeZones(): string[] {
  try {
    const zones = (Intl as unknown as { supportedValuesOf?: (k: string) => string[] }).supportedValuesOf?.('timeZone');
    if (zones?.length) return zones;
  } catch { /* older browsers */ }
  return Array.from(new Set(places.map(p => p.tz))).sort();
}
