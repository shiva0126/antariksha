import { test, expect, type Page } from '@playwright/test';

let suiteCookies: Awaited<ReturnType<ReturnType<Page['context']>['cookies']>> = [];
test.beforeEach(async ({ page, baseURL }) => {
 if(suiteCookies.length){await page.context().addCookies(suiteCookies);return;}
 const response=await page.request.post('/api/auth/register',{headers:{Origin:baseURL!},data:{email:`browser_${Date.now()}_${Math.random().toString(36).slice(2)}@example.com`,password:'browser-test-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}});
 expect(response.status()).toBe(201);
 suiteCookies=await page.context().cookies();
});
test.afterAll(async ({request,baseURL})=>{await request.delete('/api/me',{headers:{Origin:baseURL!,Cookie:suiteCookies.map(c=>`${c.name}=${c.value}`).join('; ')}});});

// Language and month settings live in Me → Settings; they persist per device.
async function setting(page: Page, label: 'Language' | 'Month system', value: string) {
  const back = page.url();
  await page.goto('/#me/settings');
  await page.getByLabel(label, { exact: true }).selectOption(value);
  if (!back.includes('#me/settings')) await page.goto(back);
}

async function openChart(page: Page) {
  await page.goto('/');
  await page.evaluate(() => localStorage.clear());
  await page.goto('/');
  await page.getByRole('checkbox').check();
  await page.getByRole('button', { name: 'Reveal my chart' }).click();
  await expect(page.locator('.chart-bar')).toBeVisible();
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
  await expect(page.locator('.chart-bar')).toContainText('Karka');
  await expect(page.locator('.chart-bar')).toContainText('Revati');
  await page.getByRole('tab', { name: 'Chart', exact: true }).click();
  await page.getByRole('tab', { name: 'North Indian' }).click();
  await expect(page.locator('svg [data-house="10"]')).toContainText('Surya');
  await expect(page.locator('svg [data-house="9"]')).toContainText('Chandra');
  await page.getByRole('tab', { name: 'Circular' }).click();
  await expect(page.locator('svg [data-longitude]')).toHaveCount(9);
  await page.getByRole('tab', { name: 'Planets' }).click();
  await expect(page.locator('.planet-table').first().locator('tbody tr')).toHaveCount(9);
  await page.getByRole('tab', { name: 'Timing' }).click();
  // Ketu must follow the Mercury birth dasha; Venus–Jupiter runs in 2026.
  await expect(page.locator('.dasha-list li').nth(1)).toContainText('Ketu');
  await page.getByRole('tab', { name: 'Reading' }).click();
  await expect(page.locator('.reading')).not.toContainText('engine fact for reflective');
  await page.getByText('Sources and interpretation limits',{exact:true}).click();
  await expect(page.locator('.reading')).toContainText('not literal translations');
  await expect(page.locator('.reading')).toContainText('passage context:');
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  expect(errors).toEqual([]);
});

test('chat answers from the chart and keeps history across reloads', async ({ page }) => {
  await openChart(page);
  await page.getByRole('tab', { name: 'Ask Astrisk' }).click();
  await page.getByRole('button', { name: 'What does my chart say about my career?' }).click();
  await expect(page.locator('.msg-assistant:not(:has(.typing))').first()).toContainText('work, responsibility');
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

test('new kundali tabs: today, transits, divisional charts, timing, ashtakavarga, report', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', e => errors.push(e.message));
  await openChart(page);
  await page.getByRole('tab', { name: 'Overview' }).click();
  await expect(page.locator('.today')).toContainText('Tara bala');
  await expect(page.locator('.today .planet-table tbody tr')).toHaveCount(9);
  // The planets read from the Moon, with the book's view (Brihat Samhita 104).
  await expect(page.locator('.transit-now li')).toHaveCount(6);
  await expect(page.locator('.transit-now')).toContainText('from the Moon');
  await expect(page.locator('.transit-now')).toContainText('Brihat Samhita 104.');
  await page.getByRole('tab', { name: 'Chart', exact: true }).click();
  await page.getByLabel('Divisional chart').selectOption('9');
  await expect(page.locator('.chart-svg').first()).toContainText('Navamsha');
  await page.getByRole('tab', { name: 'Timing' }).click();
  // Forecast periods: the running one first, each with a summary.
  await expect(page.locator('.forecast-period').first()).toContainText('Now');
  await expect(page.locator('.forecast-summary').first()).not.toBeEmpty();
  await expect(page.locator('.timing-events li').first()).toBeVisible();
  await expect(page.locator('.dasha')).toContainText('Pratyantardashas');
  await expect(page.locator('.dasha')).toContainText('Yogini dasha');
  await page.getByRole('tab', { name: 'Planets' }).click();
  await expect(page.locator('.av-total')).toContainText('337');
  await page.getByRole('tab', { name: 'Reading' }).click();
  await page.getByRole('button', { name: /printable kundali/ }).click();
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
  await expect(page.locator('.chart-bar h1')).toContainText('Ravi');
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
  await setting(page, 'Language', 'hi');
  await setting(page, 'Month system', 'purnimanta');
  await page.getByLabel('Panchang date').fill('2026-09-05');
  await expect(page.locator('.ds-page-header')).toContainText('दैनिक पंचांग');
  await expect(page.locator('.shell-nav')).toContainText('विवाह');
  // Krishna paksha of Amanta Shravana is Purnimanta Bhadrapada.
  await expect(page.locator('.day-head')).toContainText('भाद्रपद');
  await setting(page, 'Language', 'en');
  await setting(page, 'Month system', 'amanta');
});

test('delete all my data removes chats from the server', async ({ page, request }) => {
  await openChart(page);
  await page.getByRole('tab', { name: 'Ask Astrisk' }).click();
  await page.getByRole('button', { name: 'Explain my yogas' }).click();
  await expect(page.locator('.msg-assistant:not(:has(.typing))')).toHaveCount(1);
  expect(await page.evaluate(() => Object.keys(localStorage).some(k => k.includes('.chat.')))).toBe(true);
  const sid = await page.evaluate(() => Object.keys(localStorage).filter(k => k.includes('.chat.')).map(k => localStorage.getItem(k))[0]);
  await page.goto('/#privacy');
  await page.getByRole('button', { name: 'Delete charts on this device' }).click();
  await expect(page.getByRole('status')).toContainText('Deleted 1 conversation');
  const r = await page.request.get(`/api/chat/history?session_id=${sid}`);
  expect(r.status()).toBe(404);
});

test('strength tab shows shadbala for seven grahas', async ({ page }) => {
  const errors: string[] = [];
  page.on('pageerror', e => errors.push(e.message));
  await openChart(page);
  await page.getByRole('tab', { name: 'Planets' }).click();
  await expect(page.locator('.sb-row[role=listitem]')).toHaveCount(7);
  await expect(page.locator('.strength tbody tr')).toHaveCount(7);
  await expect(page.locator('.sb-notes')).toContainText('Weekday lord Mars');
  expect(errors).toEqual([]);
});

test('regional languages: Tamil, Kannada, Bengali', async ({ page }) => {
  const day = async (lang: string) => { await setting(page, 'Language', lang); await page.goto('/#panchang/today'); await page.getByLabel('Panchang date').fill('2026-09-22'); };
  await day('ta');
  await expect(page.locator('.ds-page-header')).toContainText('தினசரி பஞ்சாங்கம்');
  await expect(page.locator('.limb-grid')).toContainText('செவ்வாய்'); // Tuesday
  await expect(page.locator('.limb-grid')).toContainText('ஏகாதசி');
  await day('kn');
  await expect(page.locator('.limb-grid')).toContainText('ಮಂಗಳವಾರ');
  await day('bn');
  await expect(page.locator('.limb-grid')).toContainText('একাদশী');
  await setting(page, 'Language', 'en');
});
