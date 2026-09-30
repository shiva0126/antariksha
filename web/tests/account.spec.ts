import { test, expect } from '@playwright/test';

test('private account registration, profile editing and deletion', async ({ page }) => {
  const email='browser_' + Date.now()+'@example.com';
  await page.goto('/#account');
  await expect(page).toHaveTitle('Astrisk · Kundali & Panchang');
  await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href','https://astrisk.space/');
  await expect(page.getByRole('heading',{name:'Welcome to Astrisk'})).toBeVisible();
  await expect(page.locator('.main-nav')).toHaveCount(0);
  const anonymous=await page.request.get('/api/chart');expect(anonymous.status()).toBe(401);
  await page.getByRole('button', { name: 'New here? Create an account', exact: true }).click();
  await page.getByLabel('Email', { exact: true }).fill(email);
  await page.getByLabel('Password', { exact: true }).fill('browser-test-password');
  await page.getByLabel('Date of birth',{exact:true}).fill('1996-05-14');
  await page.getByLabel('Birth time',{exact:true}).fill('10:15');
  await expect(page.getByLabel('Hobbies and interests')).toHaveCount(0);
  await page.getByRole('button', { name: 'Create account', exact: true }).click();
  await expect(page.getByText('Signed in as')).toBeVisible();
  try {
    await page.getByLabel('Hobbies and interests').fill('Reading and hiking');
    await page.getByRole('button', { name: 'Save private profile' }).click();
    await expect(page.getByRole('status')).toHaveText('Private profile saved.');
    await page.reload();
    await expect(page.getByLabel('Hobbies and interests')).toHaveValue('Reading and hiking');
    await page.evaluate(()=>localStorage.setItem('antariksha.birth',JSON.stringify({name:'Legacy chart',date:'1990-01-01',time:'06:30',lat:12.97,lon:77.59,tz:'Asia/Kolkata',place:'Bengaluru'})));
    await page.getByText('Import charts saved before accounts existed',{exact:true}).click();
    await page.getByRole('button',{name:'Review old device charts'}).click();
    await expect(page.getByRole('button',{name:'Import selected charts'})).toBeDisabled();
    await page.getByLabel(/Legacy chart ·/).check();
    await page.getByLabel('These are my charts or I have permission to import them.').check();
    await page.getByRole('button',{name:'Import selected charts'}).click();
    await expect(page.getByText(/Imported 1 chart/)).toBeVisible();
    expect(await page.evaluate(()=>localStorage.getItem('antariksha.birth'))).not.toBeNull();
    await page.getByRole('button',{name:'Sign out',exact:true}).click();
    await expect(page.getByRole('heading',{name:'Welcome to Astrisk'})).toBeVisible();
    await page.getByLabel('Email',{exact:true}).fill(email);
    await page.getByLabel('Password',{exact:true}).fill('browser-test-password');
    await page.getByRole('button',{name:'Sign in',exact:true}).click();
    await expect(page.getByLabel('Hobbies and interests')).toHaveValue('Reading and hiking');
  } finally {
    await page.getByText('Delete this account', { exact: true }).click();
    page.once('dialog', dialog => dialog.accept());
    await page.getByRole('button', { name: 'Delete my account', exact: true }).click();
    await expect(page.getByRole('heading',{name:'Welcome to Astrisk'})).toBeVisible();
  }
});

test('switching accounts never displays the previous saved chart',async({page})=>{
 const headers={Origin:'http://127.0.0.1:3000'};
 const register=async(suffix:string)=>{const response=await page.request.post('/api/auth/register',{headers,data:{email:`isolation_${Date.now()}_${suffix}@example.com`,password:'isolation-test-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}});expect(response.status()).toBe(201);};
 await register('a');
 const first=await page.context().cookies();
 try{
  await page.goto('/');
  await page.getByRole('checkbox').check();
  await page.getByRole('button',{name:'Reveal my chart'}).click();
  await expect(page.locator('.profile')).toBeVisible();
  await register('b');
  await page.evaluate(()=>window.dispatchEvent(new Event('focus')));
  await expect(page.getByRole('button',{name:'Reveal my chart'})).toBeVisible();
  await expect(page.locator('.profile')).toHaveCount(0);
 }finally{
  await page.request.delete('/api/me',{headers});
  await page.context().addCookies(first);
  await page.request.delete('/api/me',{headers});
 }
});
