import {describe,it,expect} from 'vitest';
import {houseNumber,signIndex,degreeLabel} from './houses';
import {domeDirection,wheelPoint} from './mapping';
describe('zodiac and house placement',()=>{
 it('places the May 1996 reference planets relative to Karka Lagna',()=>{
  expect(houseNumber(29.8635591,90.4806822)).toBe(10);
  expect(houseNumber(350.5216950,90.4806822)).toBe(9);
  expect(houseNumber(172.9698223,90.4806822)).toBe(3);
 });
 it('keeps whole signs intact even before the exact ascendant degree',()=>{
  expect(houseNumber(90,119.9)).toBe(1);
  expect(houseNumber(120,119.9)).toBe(2);
  expect(signIndex(359.999)).toBe(11);
  expect(signIndex(360)).toBe(0);
 });
 it('maps the returned coordinates without changing longitude',()=>{
  expect(wheelPoint(0,100).y).toBeCloseTo(-100);
  expect(wheelPoint(90,100).x).toBeCloseTo(-100);
  expect(domeDirection(90,0,100)[2]).toBeCloseTo(100);
  expect(domeDirection(0,30,100)[1]).toBeCloseTo(50);
  expect(degreeLabel(29.8635591)).toBe('29° 51′ 48″');
 });
});
