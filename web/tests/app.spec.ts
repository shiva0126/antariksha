import { test, expect, type Page } from '@playwright/test';

async function openChart(page: Page) {
  await page.goto('/');
  await page.evaluate(() => localStorage.clear());
  await page.goto('/');
  await page.getByRole('button', { name: 'Reveal my chart' }).click();
  await expect(page.locator('.profile')).toBeVisible();
}

test('kundali: placement, dasha, tabs and no horizontal overflow', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', e => errors.push(e.message));
  await openChart(page);
  await expect(page.getByRole('img', { name: 'South Indian Kundali with fixed zodiac signs' })).toBeVisible();
  // Reference chart (14 May 1996 10:15 Bengaluru): Karka lagna, Sun in Mesha, Moon in Meena.
  await expect(page.locator('svg [data-sign="1"]')).toContainText('Surya');
  await expect(page.locator('svg [data-sign="12"]')).toContainText('Chandra');
  await expect(page.locator('.profile')).toContainText('Karka');
  await expect(page.locator('.profile')).toContainText('Revati');
  await page.getByRole('tab', { name: 'North Indian' }).click();
  await expect(page.locator('svg [data-house="10"]')).toContainText('Surya');
  await expect(page.locator('svg [data-house="9"]')).toContainText('Chandra');
  await page.getByRole('tab', { name: 'Circular' }).click();
  await expect(page.locator('svg [data-longitude]')).toHaveCount(9);
  await page.getByRole('tab', { name: 'Planets' }).click();
  await expect(page.locator('.planet-table tbody tr')).toHaveCount(9);
  await page.getByRole('tab', { name: 'Dasha' }).click();
  // Ketu must follow the Mercury birth dasha; Venus–Jupiter runs in 2026.
  await expect(page.locator('.dasha-list li').nth(1)).toContainText('Ketu');
  await page.getByRole('tab', { name: 'Reading' }).click();
  await expect(page.locator('.reading')).not.toContainText('engine fact for reflective');
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  expect(errors).toEqual([]);
});

test('chat answers from the chart and keeps history across reloads', async ({ page }) => {
  await openChart(page);
  await page.getByRole('tab', { name: 'Ask Antariksha' }).click();
  await page.getByRole('button', { name: 'What does my chart say about my career?' }).click();
  await expect(page.locator('.msg-assistant').first()).toContainText('10th house is Mesha');
  await page.getByRole('textbox', { name: 'Your question' }).fill('Do I have Mangal dosha?');
  await page.getByRole('textbox', { name: 'Your question' }).press('Enter');
  await expect(page.locator('.msg-assistant')).toHaveCount(2);
  await expect(page.locator('.msg-assistant').nth(1)).toContainText('Mangal dosha');
  await page.reload();
  await expect(page.locator('.msg')).toHaveCount(4);
  await expect(page.locator('.chat-history li')).toHaveCount(2);
});

test('panchang: day limbs, calendar festivals and mobile layout', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', e => errors.push(e.message));
  await page.goto('/#day');
  await page.getByLabel('Panchang date').fill('2026-09-22');
  await expect(page.locator('.limb-grid')).toContainText('Ekadashi');
  await expect(page.locator('.day-head')).toContainText('Ekadashi');
  await page.goto('/#month');
  await page.getByLabel('Calendar month').fill('2026-09');
  await expect(page.locator('.calendar-day')).toHaveCount(30);
  await expect(page.locator('.festival-list')).toContainText('Ganesh Chaturthi');
  await page.getByRole('button', { name: /2026-09-14 / }).click();
  await expect(page.locator('.day-details h2')).toContainText('14 Sept 2026');
  await expect(page.locator('.day-head')).toContainText('Ganesh Chaturthi');
  await page.setViewportSize({ width: 390, height: 844 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  expect(errors).toEqual([]);
});

test('chart remains usable without WebGL', async ({ page }) => {
  await page.addInitScript(() => {
    const old = HTMLCanvasElement.prototype.getContext;
    HTMLCanvasElement.prototype.getContext = function (kind: string, ...args: unknown[]) {
      if (String(kind).includes('webgl')) return null;
      return (old as (...a: unknown[]) => unknown).call(this, kind, ...args);
    } as typeof old;
  });
  await openChart(page);
  await expect(page.getByRole('tab', { name: 'Celestial dome' })).toBeDisabled();
  await expect(page.getByRole('img', { name: 'South Indian Kundali with fixed zodiac signs' })).toBeVisible();
});
