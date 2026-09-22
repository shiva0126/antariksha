export const signIndex = (longitude: number) => Math.floor(((longitude % 360) + 360) % 360 / 30);
export const houseNumber = (longitude: number, ascendant: number) => (signIndex(longitude) - signIndex(ascendant) + 12) % 12 + 1;
export const degreeLabel = (degree: number) => {
  const seconds = Math.floor(degree * 3600 + 1e-6);
  return `${Math.floor(seconds / 3600)}° ${String(Math.floor(seconds / 60) % 60).padStart(2, '0')}′ ${String(seconds % 60).padStart(2, '0')}″`;
};
