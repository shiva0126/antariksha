import { test, expect, type Page } from '@playwright/test';

async function openChart(page: Page) {
  await page.goto('/');
  await page.evaluate(() => localStorage.clear());
  await page.goto('/');
  await page.getByRole('checkbox').check();
  await page.getByRole('button', { name: 'Reveal my chart' }).click();
  await expect(page.locator('.profile')).toBeVisible();
}

test('kundali: placement, dasha, tabs and no horizontal overflow', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', e => errors.push(e.message));
  await openChart(page);
  await page.getByRole('tab', { name: 'Chart', exact: true }).click();
  await expect(page.getByRole('img', { name: 'South Indian Kundali with fixed zodiac signs' })).toBeVisible();
  // Reference chart (14 May 1996 10:15 Bengaluru): Karka lagna, Sun in Mesha, Moon in Meena.
  await expect(page.locator('svg [data-sign="1"]')).toContainText('Surya');
  await expect(page.locator('svg [data-sign="12"]')).toContainText('Chandra');
  await expect(page.locator('.profile')).toContainText('Karka');
  await expect(page.locator('.profile')).toContainText('Revati');
  await page.getByRole('tab', { name: 'Chart', exact: true }).click();
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
  await expect(page.locator('.msg-assistant:not(:has(.typing))').first()).toContainText('10th house is Mesha');
  await page.getByRole('textbox', { name: 'Your question' }).fill('Do I have Mangal dosha?');
  await page.getByRole('textbox', { name: 'Your question' }).press('Enter');
  await expect(page.locator('.msg-assistant:not(:has(.typing))')).toHaveCount(2);
  await expect(page.locator('.msg-assistant:not(:has(.typing))').nth(1)).toContainText('Mangal dosha');
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

test('consent is required before a chart is created', async ({ page }) => {
  await page.goto('/');
  await page.evaluate(() => localStorage.clear());
  await page.goto('/');
  await page.getByRole('button', { name: 'Reveal my chart' }).click();
  await expect(page.getByRole('alert')).toContainText('consent');
});

test('new kundali tabs: today, divisional charts, dasha levels, ashtakavarga, report', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', e => errors.push(e.message));
  await openChart(page);
  await page.getByRole('tab', { name: 'Today' }).click();
  await expect(page.locator('.today')).toContainText('Tara bala');
  await expect(page.locator('.today .planet-table tbody tr')).toHaveCount(9);
  await page.getByRole('tab', { name: 'Chart', exact: true }).click();
  await page.getByLabel('Divisional chart').selectOption('9');
  await expect(page.locator('svg')).toContainText('Navamsha');
  await page.getByRole('tab', { name: 'Dasha' }).click();
  await expect(page.locator('.dasha')).toContainText('Pratyantardashas');
  await expect(page.locator('.dasha')).toContainText('Yogini dasha');
  await page.getByRole('tab', { name: 'Ashtakavarga' }).click();
  await expect(page.locator('.av-total')).toContainText('337');
  await page.getByRole('tab', { name: 'Report' }).click();
  await expect(page.locator('.report')).toContainText('Planetary positions');
  expect(errors).toEqual([]);
});

test('profiles: add a second chart and switch', async ({ page }) => {
  await openChart(page);
  await page.getByLabel('Profiles').selectOption('__new');
  await page.getByPlaceholder('Whose chart is this?').fill('Ravi');
  await page.getByRole('textbox', { name: 'Birth date' }).fill('1990-01-01');
  await page.getByRole('checkbox').check();
  await page.getByRole('button', { name: 'Reveal my chart' }).click();
  await expect(page.locator('.profile h1')).toContainText('Ravi');
  await expect(page.getByLabel('Profiles').locator('option')).toHaveCount(3);
});

test('matching and muhurta pages', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', e => errors.push(e.message));
  await page.goto('/#match');
  await page.getByRole('button', { name: 'Match charts' }).click();
  await expect(page.locator('.score-ring')).toBeVisible();
  await expect(page.locator('.koota-table tbody tr')).toHaveCount(8);
  await page.goto('/#muhurta');
  await page.getByRole('button', { name: 'Find muhurta' }).click();
  await expect(page.locator('.results-head')).toContainText('suitable');
  expect(errors).toEqual([]);
});

test('place search finds towns by old names', async ({ page }) => {
  await page.goto('/');
  await page.evaluate(() => localStorage.clear());
  await page.goto('/');
  const input = page.getByRole('combobox', { name: 'Birthplace' });
  await input.fill('bombay');
  await expect(page.locator('.place-list').getByRole('option').first()).toContainText('Mumbai');
  await page.locator('.place-list').getByRole('option').first().click();
  await expect(page.locator('.place-search small')).toContainText('Asia/Kolkata');
});

test('hindi interface and purnimanta months', async ({ page }) => {
  await page.goto('/#day');
  await page.getByLabel('Language').selectOption('hi');
  await page.getByLabel('Month system').selectOption('purnimanta');
  await page.getByLabel('Panchang date').fill('2026-09-05');
  await expect(page.locator('.main-nav')).toContainText('दैनिक पंचांग');
  // Krishna paksha of Amanta Shravana is Purnimanta Bhadrapada.
  await expect(page.locator('.day-head')).toContainText('भाद्रपद');
  await page.getByLabel('Language').selectOption('en');
  await page.getByLabel('Month system').selectOption('amanta');
});

test('delete all my data removes chats from the server', async ({ page, request }) => {
  await openChart(page);
  await page.getByRole('tab', { name: 'Ask Antariksha' }).click();
  await page.getByRole('button', { name: 'Explain my yogas' }).click();
  await expect(page.locator('.msg-assistant:not(:has(.typing))')).toHaveCount(1);
  expect(await page.evaluate(() => Object.keys(localStorage).some(k => k.startsWith('antariksha.chat.')))).toBe(true);
  const sid = await page.evaluate(() => Object.keys(localStorage).filter(k => k.startsWith('antariksha.chat.')).map(k => localStorage.getItem(k))[0]);
  await page.goto('/#privacy');
  await page.getByRole('button', { name: 'Delete all my data' }).click();
  await expect(page.getByRole('status')).toContainText('Deleted 1 conversation');
  const r = await request.get(`/api/chat/history?session_id=${sid}`);
  expect(r.status()).toBe(404);
});

test('strength tab shows shadbala for seven grahas', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', e => errors.push(e.message));
  await openChart(page);
  await page.getByRole('tab', { name: 'Strength' }).click();
  await expect(page.locator('.sb-row[role=listitem]')).toHaveCount(7);
  await expect(page.locator('.strength tbody tr')).toHaveCount(7);
  await expect(page.locator('.sb-notes')).toContainText('Weekday lord Mars');
  expect(errors).toEqual([]);
});

test('regional languages: Tamil, Kannada, Bengali', async ({ page }) => {
  await page.goto('/#day');
  await page.getByLabel('Panchang date').fill('2026-09-22');
  await page.getByLabel('Language').selectOption('ta');
  await expect(page.locator('.main-nav')).toContainText('தினசரி பஞ்சாங்கம்');
  await expect(page.locator('.limb-grid')).toContainText('செவ்வாய்'); // Tuesday
  await expect(page.locator('.limb-grid')).toContainText('ஏகாதசி');
  await page.getByLabel('Language').selectOption('kn');
  await expect(page.locator('.limb-grid')).toContainText('ಮಂಗಳವಾರ');
  await page.getByLabel('Language').selectOption('bn');
  await expect(page.locator('.limb-grid')).toContainText('একাদশী');
  await page.getByLabel('Language').selectOption('en');
});
