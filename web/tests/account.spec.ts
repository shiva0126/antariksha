import { test, expect } from '@playwright/test';

test('private account registration, profile editing and deletion', async ({ page }) => {
  await page.goto('/#account');
  await page.getByRole('button', { name: 'Create account', exact: true }).click();
  await page.getByLabel('Handle', { exact: true }).fill('browser_' + Date.now());
  await page.getByLabel('Password', { exact: true }).fill('browser-test-password');
  await page.getByRole('checkbox').check();
  await page.getByRole('button', { name: 'Create account', exact: true }).click();
  await expect(page.getByText('Signed in as')).toBeVisible();
  try {
    await page.getByLabel('Hobbies and interests').fill('Reading and hiking');
    await page.getByRole('button', { name: 'Save private profile' }).click();
    await expect(page.getByRole('status')).toHaveText('Private profile saved.');
    await page.reload();
    await expect(page.getByLabel('Hobbies and interests')).toHaveValue('Reading and hiking');
  } finally {
    await page.getByText('Delete this account', { exact: true }).click();
    page.once('dialog', dialog => dialog.accept());
    await page.getByRole('button', { name: 'Delete my account', exact: true }).click();
    await expect(page.getByRole('status')).toHaveText('Account deleted.');
  }
});
