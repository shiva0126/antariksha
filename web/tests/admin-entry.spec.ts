import { test, expect, type Page } from '@playwright/test';

// UI fixtures never create accounts or change permissions in the live database.
async function mockAPI(page: Page, role?: string) {
  let signedIn = Boolean(role);
  await page.route(url => url.pathname.startsWith('/api/'), async route => {
    const path = new URL(route.request().url()).pathname;
    let body: unknown = [], status = 200;
    if (path === '/api/auth/options') body = { email_delivery_available: false };
    else if (path === '/api/auth/login') { signedIn = true; body = { ok: true }; }
    else if (path === '/api/me') {
      status = signedIn ? 200 : 401;
      body = signedIn ? { id: 'fixture', handle: 'owner', email: 'owner@example.com', role: role || 'superadmin', profile: { name: 'Owner' } } : { error: 'Sign in required' };
    } else if (path === '/api/me/charts') body = { account_id: 'fixture', profiles: [], revision: 0 };
    else if (path === '/api/admin/summary') body = { users: 8, suspended: 0, moderators: 0, superadmins: 1, active_sessions: 1, open_post_reports: 0, open_profile_reports: 0 };
    await route.fulfill({ status, json: body });
  });
}

test('administrator entry uses normal login and keeps the admin destination', async ({ page }) => {
  await mockAPI(page);
  await page.goto('/');
  await page.getByRole('link', { name: 'Administrator sign in', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Administrator sign in' })).toBeVisible();
  await page.getByLabel('Email', { exact: true }).fill('owner@example.com');
  await page.getByLabel('Password', { exact: true }).fill('fixture-password');
  const login = page.waitForRequest(r => new URL(r.url()).pathname === '/api/auth/login');
  await page.getByRole('button', { name: 'Sign in', exact: true }).click();
  expect((await login).postDataJSON()).toEqual({ email: 'owner@example.com', password: 'fixture-password' });
  await expect(page.getByRole('heading', { name: 'Superadmin', exact: true })).toBeVisible();
  await expect(page.getByRole('heading', { name: '8 users', exact: true })).toBeVisible();
  const entry = page.getByRole('link', { name: 'Superadmin console', exact: true });
  await expect(entry).toBeVisible();
  for (const width of [390, 320, 768, 1024]) {
    await page.setViewportSize({ width, height: 844 });
    await expect(entry).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), `no overflow at ${width}px`).toBe(true);
  }
});

test('ordinary members see an explanation, not administration controls', async ({ page }) => {
  await mockAPI(page, 'member');
  const adminRequests: string[] = [];
  page.on('request', r => { if (new URL(r.url()).pathname.startsWith('/api/admin/')) adminRequests.push(r.url()); });
  await page.goto('/#admin');
  await expect(page.getByRole('heading', { name: 'Admin access required' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Superadmin console' })).toHaveCount(0);
  expect(adminRequests).toEqual([]);
});
