import {test,expect} from '@playwright/test';

test('private numerology, tropical Western chart and tarot explain their methods',async({page,baseURL})=>{
 const headers={Origin:baseURL!};
 expect((await page.request.post('/api/auth/register',{headers,data:{email:`readings_browser_${Date.now()}@example.com`,password:'readings-browser-password',birth_date:'1996-05-14',birth_time:'10:15',consent:true}})).status()).toBe(201);
 try{
  await page.goto('/#readings');await expect(page.getByRole('heading',{name:'More readings'})).toBeVisible();
  await page.getByLabel('Name for numerology (optional)').fill('John');
  await page.getByRole('button',{name:'Read my numbers'}).click();
  await page.getByText('Our numerology convention',{exact:true}).click();
  await expect(page.getByText(/Pythagorean letters A–Z/)).toBeVisible();
  await expect(page.getByText('Expression',{exact:true})).toBeVisible();
  await page.getByRole('tab',{name:'Western astrology'}).click();
  await page.getByRole('button',{name:'Read my Western chart'}).click();
  await expect(page.getByRole('heading',{name:/Your Sun, Moon and rising sign/})).toBeVisible();
  await expect(page.getByText(/tropical zodiac/)).toBeVisible();
  await page.getByRole('tab',{name:'Tarot'}).click();
  await page.getByRole('button',{name:'Draw cards'}).click();
  await expect(page.getByText('About this draw',{exact:true})).toBeVisible();
  await expect(page.locator('.tarot-card')).toHaveCount(3);
 }finally{await page.request.delete('/api/me',{headers});}
});
