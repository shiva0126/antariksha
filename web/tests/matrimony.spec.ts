import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

// Runs against scripts/test-browser.sh (disposable database, non-secure cookies).
const bengaluru = { name: 'Bengaluru, Karnataka, IN', lat: 12.97, lon: 77.59, tz: 'Asia/Kolkata' };

async function member(request: APIRequestContext, base: string, name: string, dob: string, tob: string, kind: string, extra: Record<string, unknown> = {}) {
  const headers = { Origin: base };
  expect((await request.post('/api/auth/register', { headers, data: { email: `${name.toLowerCase()}_${Date.now()}@example.com`, password: 'matrimony-browser-password', birth_date: dob, birth_time: tob, consent: true, name, birth_place: bengaluru } })).status()).toBe(201);
  expect((await request.put('/api/community/settings', { headers, data: { community: true, birth_date: dob, avatar: 'sun', interests: ['music'] } })).status()).toBe(200);
  expect((await request.put('/api/matrimony/me', { headers, data: { active: true, consent: true, horoscope: true, details: { display_name: name, profile_kind: kind, city: 'Bengaluru', diet: 'vegetarian', religion: 'hindu', mother_tongue: 'kannada', introduction: `Hello, I am ${name}.`, min_age: 18, max_age: 60, ...extra } } })).status()).toBe(200);
  return (await request.get('/api/me')).json();
}

test('horoscope-aware discovery, interest with a note, acceptance and chat', async ({ page, browser, baseURL }) => {
  const other = await browser.newContext({ baseURL });
  const bridePage: Page = await other.newPage();
  await member(page.request, baseURL!, 'Rohan', '1994-11-02', '06:40', 'groom');
  await member(other.request, baseURL!, 'Ananya', '1996-05-14', '10:15', 'bride');
  try {
    await page.goto('/#matrimony');
    const card = page.getByRole('article', { name: 'Ananya' });
    await expect(card.locator('.guna-chip')).toContainText('19.5');
    await expect(card.locator('.reason-chips')).toContainText('Same city');
    await expect(card.locator('.reason-chips')).toContainText('Both vegetarian');

    // Filters: a minimum guna above the pair's score hides the card.
    await page.getByLabel('Minimum Guna score').selectOption('21');
    await page.getByRole('button', { name: 'Apply filters' }).click();
    await expect(page.getByText('No profiles match these filters yet.')).toBeVisible();
    await page.getByRole('button', { name: 'Clear' }).click();
    await expect(card).toBeVisible();

    // Full comparison with Ask Astrisk.
    await card.getByRole('button', { name: 'Compare horoscopes' }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.locator('.score-ring')).toContainText('19.5');
    await dialog.getByRole('button', { name: 'What does numerology say about us?' }).click();
    await expect(dialog.locator('.msg-assistant').last()).toContainText('root number (mulank)');
    await dialog.getByRole('button', { name: 'Close' }).click();

    // Interest with a note.
    await card.getByRole('button', { name: 'Express interest' }).click();
    await page.getByRole('dialog').getByLabel('A short note (optional)').fill('We both enjoy music. Would you like to talk?');
    await page.getByRole('dialog').getByRole('button', { name: 'Send' }).click();
    await expect(card.getByText('Interest sent')).toBeVisible();

    // The bride sees the note, accepts and starts the chat.
    await bridePage.goto('/#matrimony/interests');
    await expect(bridePage.getByText('“We both enjoy music. Would you like to talk?”')).toBeVisible();
    await bridePage.getByRole('button', { name: 'Accept' }).click();
    await bridePage.getByRole('button', { name: 'Open chat' }).click();
    await bridePage.getByPlaceholder('Message Rohan').fill('Hello Rohan!');
    await bridePage.getByRole('region', { name: 'Conversation with Rohan' }).getByRole('button', { name: 'Send' }).click();
    await expect(bridePage.locator('.mat-messages')).toContainText('Hello Rohan!');

    await page.goto('/#matrimony/interests');
    await page.getByRole('button', { name: 'Open chat' }).click();
    await expect(page.locator('.mat-messages')).toContainText('Hello Rohan!');
    await page.getByRole('button', { name: 'Share my contact' }).click();
    await page.getByLabel('Phone or email to share with Ananya only').fill('+91 98765 43210');
    await page.getByRole('region', { name: 'Conversation with Ananya' }).getByRole('button', { name: 'Share my contact' }).click();
    await expect(page.getByText('You shared: +91 98765 43210')).toBeVisible();

    // Profile editor and printable biodata.
    await page.goto('/#matrimony/profile');
    await expect(page.locator('.mat-progress')).toContainText('Profile completeness');
    await expect(page.getByLabel('Show horoscope matching')).toBeChecked();
    await page.goto('/#biodata');
    await expect(page.locator('.biodata-sheet h1')).toHaveText('Rohan');
    await expect(page.locator('.biodata-kundali')).toContainText('Janma kundali');
  } finally {
    await page.request.delete('/api/me', { headers: { Origin: baseURL! } });
    await other.request.delete('/api/me', { headers: { Origin: baseURL! } });
    await other.close();
  }
});

test('sign-up stores the birthplace; phone layout has a bottom tab bar', async ({ page, baseURL }) => {
  await page.goto('/');
  await page.getByRole('button', { name: 'New here? Create an account' }).click();
  await page.getByLabel('Your name').fill('Meera');
  await page.getByLabel('Email').fill(`meera_${Date.now()}@example.com`);
  await page.getByLabel('Password').fill('sign-up-browser-password');
  await page.getByRole('button', { name: 'Continue' }).click();
  await page.getByLabel('Date of birth').fill('1997-03-03');
  await page.getByLabel('Birth time').fill('08:00');
  await page.getByLabel('Birthplace').fill('Mysuru');
  await page.locator('.place-list').getByRole('option').first().click();
  await page.getByRole('button', { name: 'Create account' }).click();
  await expect(page.locator('.shell-header')).toBeVisible();
  try {
    const me = await (await page.request.get('/api/me')).json();
    expect(me.birth_place.name).toContain('Mysuru');
    expect(me.profile.name).toBe('Meera');
    await page.setViewportSize({ width: 390, height: 844 });
    await expect(page.locator('.shell-tabbar')).toBeVisible();
    await page.locator('.shell-tabbar').getByRole('link', { name: /Matrimony/ }).click();
    await expect(page.getByRole('heading', { name: 'Matrimony', level: 1 })).toBeVisible();
    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1);
    expect(overflow).toBe(false);
    await page.locator('.shell-me__toggle').click();
    await expect(page.getByRole('menuitem', { name: 'Saved charts' })).toBeVisible();
  } finally {
    await page.request.delete('/api/me', { headers: { Origin: baseURL! } });
  }
});
